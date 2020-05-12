package ias

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

const (
	PathServiceProviders  = "/service/sps"
	PathCompanyGlobal     = "/service/company/global"
	PathAccess            = "/service/sps/%s/rba"
	PathIdentityProviders = "/service/idp"
	PathDelete            = "/service/sps/delete"
	PathDeleteSecret      = "/service/sps/clientSecret"
)

type (
	ClientConfig struct {
		URL    string
		ID     string
		Secret string
	}

	Client struct {
		config         ClientConfig
		httpClient     *http.Client
		closeBodyError error
	}

	Request struct {
		Method  string
		Path    string
		Body    io.Reader
		Headers map[string]string
		Delete  bool
	}
)

func NewClient(cli *http.Client, cfg ClientConfig) *Client {
	return &Client{
		config:     cfg,
		httpClient: cli,
	}
}

func (c *Client) SetOIDCConfiguration(spID string, payload OIDCType) error {
	return c.call(c.serviceProviderPath(spID), payload)
}

func (c *Client) SetSAMLConfiguration(spID string, payload SAMLType) error {
	return c.call(c.serviceProviderPath(spID), payload)
}

func (c *Client) SetAssertionAttribute(spID string, payload PostAssertionAttributes) error {
	return c.call(c.serviceProviderPath(spID), payload)
}

func (c *Client) SetSubjectNameIdentifier(spID string, payload SubjectNameIdentifier) error {
	return c.call(c.serviceProviderPath(spID), payload)
}

func (c *Client) SetAuthenticationAndAccess(spID string, payload AuthenticationAndAccess) error {
	pathAccess := fmt.Sprintf(PathAccess, spID)

	return c.call(pathAccess, payload)
}

func (c *Client) SetDefaultAuthenticatingIDP(payload DefaultAuthIDPConfig) error {
	return c.call(PathServiceProviders, payload)
}

func (c *Client) GetCompany() (_ *Company, err error) {
	company := &Company{}
	request := &Request{Method: http.MethodGet, Path: PathCompanyGlobal}

	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return company, errors.Wrap(err, "while making request to ias platform about company")
	}

	err = json.NewDecoder(response.Body).Decode(company)
	if err != nil {
		return company, errors.Wrap(err, "while decoding response body with company data")
	}

	return company, nil
}

func (c *Client) CreateServiceProvider(serviceName, companyID string) (err error) {
	payload := fmt.Sprintf("sp_name=%s&company_id=%s", serviceName, companyID)
	request := &Request{
		Method:  http.MethodPost,
		Path:    PathServiceProviders,
		Body:    strings.NewReader(payload),
		Headers: map[string]string{"content-type": "application/x-www-form-urlencoded"},
	}

	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return errors.Wrap(err, "while making request with ServiceProvider creation")
	}

	return nil
}

func (c *Client) DeleteServiceProvider(spID string) (err error) {
	request := &Request{
		Method: http.MethodPut,
		Path:   fmt.Sprintf("%s?sp_id=%s", PathDelete, spID),
		Delete: true,
	}
	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return errors.Wrap(err, "while making request to delete ServiceProvider")
	}

	return nil
}

func (c *Client) DeleteSecret(payload DeleteSecrets) (err error) {
	request, err := c.jsonRequest(PathDeleteSecret, http.MethodDelete, payload)
	if err != nil {
		return errors.Wrapf(err, "while creating json request for path %s", PathDeleteSecret)
	}
	request.Delete = true

	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return errors.Wrap(err, "while making request to delete ServiceProvider secrets")
	}

	return nil
}

func (c *Client) GenerateServiceProviderSecret(secretCfg SecretConfiguration) (_ *ServiceProviderSecret, err error) {
	secretResponse := &ServiceProviderSecret{}
	request, err := c.jsonRequest(PathServiceProviders, http.MethodPut, secretCfg)
	if err != nil {
		return secretResponse, errors.Wrap(err, "while creating request for secret provider")
	}

	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return secretResponse, errors.Wrap(err, "while creating ServiceProvider secret")
	}

	err = json.NewDecoder(response.Body).Decode(secretResponse)
	if err != nil {
		return secretResponse, errors.Wrap(err, "while decoding response with secret provider")
	}

	return secretResponse, nil
}

func (c Client) AuthenticationURL(id ProviderID) string {
	return fmt.Sprintf("%s%s/%s", c.config.URL, PathIdentityProviders, id)
}

func (c *Client) serviceProviderPath(spID string) string {
	return fmt.Sprintf("%s/%s", PathServiceProviders, spID)
}

func (c *Client) call(path string, payload interface{}) (err error) {
	request, err := c.jsonRequest(path, http.MethodPut, payload)
	if err != nil {
		return errors.Wrapf(err, "while creating json request for path %s", path)
	}

	response, err := c.do(request)
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()
	if err != nil {
		return errors.Wrapf(err, "while making request for path %s", path)
	}

	return nil
}

func (c *Client) jsonRequest(path string, method string, payload interface{}) (*Request, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(payload)
	if err != nil {
		return &Request{}, err
	}

	return &Request{
		Method:  method,
		Path:    path,
		Body:    buffer,
		Headers: map[string]string{"content-type": "application/json"},
	}, nil
}

func (c *Client) do(sciReq *Request) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", c.config.URL, sciReq.Path)
	req, err := http.NewRequest(sciReq.Method, url, sciReq.Body)
	if err != nil {
		return nil, err
	}

	req.Close = true
	req.SetBasicAuth(c.config.ID, c.config.Secret)
	for h, v := range sciReq.Headers {
		req.Header.Set(h, v)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return &http.Response{}, errors.Wrap(err, "while making request")
	}

	switch {
	case response.StatusCode == http.StatusOK ||
		response.StatusCode == http.StatusCreated ||
		response.StatusCode == http.StatusNoContent:
		return response, nil
	case sciReq.Delete && response.StatusCode == http.StatusNotFound:
		return response, nil
	case response.StatusCode == http.StatusRequestTimeout:
		return response, kebError.NewTemporaryError(c.responseErrorMessage(response))
	case response.StatusCode >= http.StatusInternalServerError:
		return response, kebError.NewTemporaryError(c.responseErrorMessage(response))
	default:
		return response, errors.Errorf("while sending request to IAS: %s", c.responseErrorMessage(response))
	}
}

func (c *Client) closeResponseBody(response *http.Response) error {
	if response.Body == nil {
		return nil
	}
	return response.Body.Close()
}

func (c *Client) responseErrorMessage(response *http.Response) string {
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Sprintf("unexpected status code %d cannot read body response: %s", response.StatusCode, err)
	}
	return fmt.Sprintf("unexpected status code %d with body: %s", response.StatusCode, string(body))
}
