package edp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	MaasConsumerEnvironmentKey = "maasConsumerEnvironment"
	MaasConsumerRegionKey      = "maasConsumerRegion"
	MaasConsumerSubAccountKey  = "maasConsumerSubAccount"

	dataTenantTmpl     = "%s/namespaces/%s/dataTenants"
	metadataTenantTmpl = "%s/namespaces/%s/dataTenants/%s/%s/metadata"

	namespaceToken = "%s/oauth2/token"
)

type Config struct {
	AuthURL     string
	AdminURL    string
	Namespace   string
	Secret      string
	Environment string `envconfig:"default=prod"`
	Required    bool   `envconfig:"default=false"`
	Disabled    bool
}

type Client struct {
	config     Config
	httpClient *http.Client
	log        logrus.FieldLogger
}

func NewClient(config Config, log logrus.FieldLogger) *Client {
	cfg := clientcredentials.Config{
		ClientID:     fmt.Sprintf("edp-namespace;%s", config.Namespace),
		ClientSecret: config.Secret,
		TokenURL:     fmt.Sprintf(namespaceToken, config.AuthURL),
		Scopes:       []string{"edp-namespace.read edp-namespace.update"},
	}
	httpClientOAuth := cfg.Client(context.Background())
	httpClientOAuth.Timeout = 30 * time.Second

	return &Client{
		config:     config,
		httpClient: httpClientOAuth,
		log:        log,
	}
}

func (c *Client) dataTenantURL() string {
	return fmt.Sprintf(dataTenantTmpl, c.config.AdminURL, c.config.Namespace)
}

func (c *Client) metadataTenantURL(name, env string) string {
	return fmt.Sprintf(metadataTenantTmpl, c.config.AdminURL, c.config.Namespace, name, env)
}

func (c *Client) CreateDataTenant(data DataTenantPayload) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "while marshaling dataTenant payload")
	}

	return c.post(c.dataTenantURL(), rawData)
}

func (c *Client) DeleteDataTenant(name, env string) error {
	URL := fmt.Sprintf("%s/%s/%s", c.dataTenantURL(), name, env)
	request, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return errors.Wrap(err, "while creating delete dataTenant request")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "while requesting about delete dataTenant")
	}

	return c.processResponse(response, true)
}

func (c *Client) CreateMetadataTenant(name, env string, data MetadataTenantPayload) error {
	rawData, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "while marshaling tenant metadata payload")
	}

	return c.post(c.metadataTenantURL(name, env), rawData)
}

func (c *Client) DeleteMetadataTenant(name, env, key string) error {
	URL := fmt.Sprintf("%s/%s", c.metadataTenantURL(name, env), key)
	request, err := http.NewRequest(http.MethodDelete, URL, nil)
	if err != nil {
		return errors.Wrap(err, "while creating delete metadata request")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "while requesting about delete metadata")
	}

	return c.processResponse(response, true)
}

func (c *Client) GetMetadataTenant(name, env string) (_ []MetadataItem, err error) {
	response, err := c.httpClient.Get(c.metadataTenantURL(name, env))
	if err != nil {
		return []MetadataItem{}, errors.Wrap(err, "while requesting about dataTenant metadata")
	}
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()

	var metadata []MetadataItem
	err = json.NewDecoder(response.Body).Decode(&metadata)
	if err != nil {
		return nil, errors.Wrap(err, "while decoding dataTenant metadata response")
	}

	return metadata, nil
}

func (c *Client) post(URL string, data []byte) (err error) {
	response, err := c.httpClient.Post(URL, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return errors.Wrapf(err, "while sending POST request on %s", URL)
	}
	defer func() {
		err = multierror.Append(err, errors.Wrap(c.closeResponseBody(response), "while trying to close body reader")).ErrorOrNil()
	}()

	return c.processResponse(response, false)
}

func (c *Client) processResponse(response *http.Response, allowNotFound bool) error {
	byteBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.Wrapf(err, "while reading response body (status code %d)", response.StatusCode)
	}
	body := string(byteBody)

	switch response.StatusCode {
	case http.StatusCreated:
		c.log.Infof("Resource created: %s", responseLog(response))
		return nil
	case http.StatusConflict:
		c.log.Infof("Resource already exist: %s", responseLog(response))
		return nil
	case http.StatusNoContent:
		c.log.Infof("Action executed correctly: %s", responseLog(response))
		return nil
	case http.StatusNotFound:
		c.log.Infof("Resource not found: %s", responseLog(response))
		if allowNotFound {
			return nil
		}
		c.log.Errorf("Body content: %s", body)
		return errors.Errorf("Not Found: %s", responseLog(response))
	case http.StatusRequestTimeout:
		c.log.Errorf("Request timeout %s: %s", responseLog(response), body)
		return kebError.NewTemporaryError("Request timeout: %s", responseLog(response))
	case http.StatusBadRequest:
		c.log.Errorf("Bad request %s: %s", responseLog(response), body)
		return errors.Errorf("Bad request: %s", responseLog(response))
	}

	if response.StatusCode >= 500 {
		c.log.Errorf("EDP server returns failed status %s: %s", responseLog(response), body)
		return kebError.NewTemporaryError("EDP server returns failed status %s", responseLog(response))
	}

	c.log.Errorf("EDP server not supported response %s: %s", responseLog(response), body)
	return errors.Errorf("Undefined/empty/notsupported status code response %s", responseLog(response))
}

func responseLog(r *http.Response) string {
	return fmt.Sprintf("Response status code: %d for request %s %s", r.StatusCode, r.Request.Method, r.Request.URL)
}

func (c *Client) closeResponseBody(response *http.Response) error {
	if response.Body == nil {
		return nil
	}

	return response.Body.Close()
}
