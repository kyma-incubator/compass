package destinationfetchersvc

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	correlationIDPrefix = "sap.s4:communicationScenario:"
	s4HANAType          = "SAP S/4HANA Cloud"
	s4HANABaseURLSuffix = "-api"
	authorizationHeader = "Authorization"
)

// DestinationServiceAPIConfig destination service api configuration
type DestinationServiceAPIConfig struct {
	GoroutineLimit                int64         `envconfig:"APP_DESTINATIONS_SENSITIVE_GOROUTINE_LIMIT,default=10"`
	RetryInterval                 time.Duration `envconfig:"APP_DESTINATIONS_RETRY_INTERVAL,default=100ms"`
	RetryAttempts                 uint          `envconfig:"APP_DESTINATIONS_RETRY_ATTEMPTS,default=3"`
	EndpointGetTenantDestinations string        `envconfig:"APP_ENDPOINT_GET_TENANT_DESTINATIONS,default=/destination-configuration/v1/subaccountDestinations"`
	EndpointFindDestination       string        `envconfig:"APP_ENDPOINT_FIND_DESTINATION,default=/destination-configuration/v1/destinations"`
	Timeout                       time.Duration `envconfig:"APP_DESTINATIONS_TIMEOUT,default=5s"`
	PageSize                      int           `envconfig:"APP_DESTINATIONS_PAGE_SIZE,default=100"`
	PagingPageParam               string        `envconfig:"APP_DESTINATIONS_PAGE_PARAM,default=$page"`
	PagingSizeParam               string        `envconfig:"APP_DESTINATIONS_PAGE_SIZE_PARAM,default=$pageSize"`
	PagingCountParam              string        `envconfig:"APP_DESTINATIONS_PAGE_COUNT_PARAM,default=$pageCount"`
	PagingCountHeader             string        `envconfig:"APP_DESTINATIONS_PAGE_COUNT_HEADER,default=Page-Count"`
	SkipSSLVerify                 bool          `envconfig:"APP_DESTINATIONS_SKIP_SSL_VERIFY,default=false"`
	OAuthTokenPath                string        `envconfig:"APP_DESTINATION_OAUTH_TOKEN_PATH,default=/oauth/token"`
	ResponseCorrelationIDHeader   string        `envconfig:"APP_DESTINATIONS_RESPONSE_CORRELATION_ID_HEADER,default=x-vcap-request-id"`
}

// Client destination client
type Client struct {
	httpClient        *http.Client
	apiConfig         DestinationServiceAPIConfig
	authConfig        config.InstanceConfig
	authToken         string
	authTokenValidity time.Time
}

func (c *Client) Close() {
	c.httpClient.CloseIdleConnections()
}

// destinationFromService destination received from destination service
type destinationFromService struct {
	Name                    string `json:"Name"`
	Type                    string `json:"Type"`
	URL                     string `json:"URL"`
	Authentication          string `json:"Authentication"`
	XFSystemName            string `json:"XFSystemName"`
	CommunicationScenarioID string `json:"communicationScenarioId"`
	ProductName             string `json:"product.name"`
	XCorrelationID          string `json:"x-correlation-id"`
	XSystemTenantID         string `json:"x-system-id"`
	XSystemTenantName       string `json:"x-system-name"`
	XSystemType             string `json:"x-system-type"`
	XSystemBaseURL          string `json:"x-system-base-url"`
}

func (d *destinationFromService) setDefaults(result *model.DestinationInput) error {
	// Set values from custom properties
	if result.XSystemType == "" {
		result.XSystemType = d.ProductName
	}
	if result.XSystemType != s4HANAType {
		return nil
	}
	if result.XCorrelationID == "" {
		if d.CommunicationScenarioID != "" {
			result.XCorrelationID = correlationIDPrefix + d.CommunicationScenarioID
		}
	}
	if result.XSystemTenantName == "" {
		result.XSystemTenantName = d.XFSystemName
	}
	if result.XSystemBaseURL != "" || result.URL == "" {
		return nil
	}

	baseURL, err := url.Parse(result.URL)
	if err != nil {
		return errors.Wrapf(err, "%s destination has invalid URL '%s'", s4HANAType, result.URL)
	}
	subdomains := strings.Split(baseURL.Hostname(), ".")
	if len(subdomains) < 2 {
		return fmt.Errorf(
			"%s destination has invalid URL '%s'. Expected at least 2 subdomains", s4HANAType, result.URL)
	}
	subdomains[0] = strings.TrimSuffix(subdomains[0], s4HANABaseURLSuffix)

	result.XSystemBaseURL = fmt.Sprintf("%s://%s", baseURL.Scheme, strings.Join(subdomains, "."))
	return nil
}

// ToModel missing godoc
func (d *destinationFromService) ToModel() (model.DestinationInput, error) {
	result := model.DestinationInput{
		Name:              d.Name,
		Type:              d.Type,
		URL:               d.URL,
		Authentication:    d.Authentication,
		XCorrelationID:    d.XCorrelationID,
		XSystemTenantID:   d.XSystemTenantID,
		XSystemTenantName: d.XSystemTenantName,
		XSystemType:       d.XSystemType,
		XSystemBaseURL:    d.XSystemBaseURL,
	}

	if err := d.setDefaults(&result); err != nil {
		return model.DestinationInput{}, err
	}

	return result, result.Validate()
}

// DestinationResponse paged response from destination service
type DestinationResponse struct {
	destinations []destinationFromService
	pageCount    int
}

func getHttpClient(instanceConfig config.InstanceConfig, apiConfig DestinationServiceAPIConfig) (*http.Client, error) {
	cert, err := tls.X509KeyPair([]byte(instanceConfig.Cert), []byte(instanceConfig.Key))
	if err != nil {
		return nil, errors.Errorf("failed to create destinations client x509 pair: %v", err)
	}
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: apiConfig.SkipSSLVerify,
				Certificates:       []tls.Certificate{cert},
			},
		},
		Timeout: apiConfig.Timeout,
	}, nil
}

func setInstanceConfigTokenURLForSubdomain(
	instanceConfig *config.InstanceConfig, apiConfig DestinationServiceAPIConfig, subdomain string) error {

	baseTokenURL, err := url.Parse(instanceConfig.TokenURL)
	if err != nil {
		return errors.Errorf("failed to parse auth url '%s': %v", instanceConfig.TokenURL, err)
	}
	parts := strings.Split(baseTokenURL.Hostname(), ".")
	if len(parts) < 2 {
		return errors.Errorf("auth url '%s' should have a subdomain", instanceConfig.TokenURL)
	}
	originalSubdomain := parts[0]

	instanceConfig.TokenURL = strings.Replace(instanceConfig.TokenURL, originalSubdomain, subdomain, 1) +
		apiConfig.OAuthTokenPath

	return nil
}

// NewClient returns new destination client
func NewClient(instanceConfig config.InstanceConfig, apiConfig DestinationServiceAPIConfig,
	subdomain string) (*Client, error) {

	if err := setInstanceConfigTokenURLForSubdomain(&instanceConfig, apiConfig, subdomain); err != nil {
		return nil, err
	}
	httpClient, err := getHttpClient(instanceConfig, apiConfig)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: httpClient,
		apiConfig:  apiConfig,
		authConfig: instanceConfig,
	}, nil
}

// FetchTenantDestinationsPage returns a page of destinations
func (c *Client) FetchTenantDestinationsPage(ctx context.Context, tenantID, page string) (*DestinationResponse, error) {
	fetchURL := c.authConfig.URL + c.apiConfig.EndpointGetTenantDestinations
	req, err := c.buildFetchRequest(ctx, fetchURL, page)
	if err != nil {
		return nil, err
	}

	destinationsPageCallStart := time.Now()
	res, err := c.sendRequestWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.C(ctx).WithError(err).Error("Unable to close response body")
		}
	}()

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received status code %d when trying to fetch destinations", res.StatusCode)
	}

	destinationsPageCallHeadersDuration := time.Since(destinationsPageCallStart)

	var destinations []destinationFromService
	if err := json.NewDecoder(res.Body).Decode(&destinations); err != nil {
		return nil, errors.Wrap(err, "failed to decode response body")
	}

	destinationsPageCallFullDuration := time.Since(destinationsPageCallStart)

	pageCountHeader := res.Header.Get(c.apiConfig.PagingCountHeader)
	if pageCountHeader == "" {
		return nil, errors.Errorf("missing '%s' header from destinations response", c.apiConfig.PagingCountHeader)
	}
	pageCount, err := strconv.Atoi(pageCountHeader)
	if err != nil {
		return nil, errors.Errorf("invalid header '%s' '%s'", c.apiConfig.PagingCountHeader, pageCountHeader)
	}

	logDuration := log.C(ctx).Debugf
	if destinationsPageCallFullDuration > c.apiConfig.Timeout/2 {
		logDuration = log.C(ctx).Warnf
	}
	destinationCorrelationHeader := c.apiConfig.ResponseCorrelationIDHeader
	logDuration("Getting tenant '%s' destinations page %s/%s took %s, %s of which for headers, %s: '%s'",
		tenantID, page, pageCountHeader, destinationsPageCallFullDuration.String(), destinationsPageCallHeadersDuration.String(),
		destinationCorrelationHeader, res.Header.Get(destinationCorrelationHeader))

	return &DestinationResponse{
		destinations: destinations,
		pageCount:    pageCount,
	}, nil
}

func (c *Client) setToken(req *http.Request) error {
	token, err := c.getToken(req.Context())
	if err != nil {
		return err
	}
	req.Header.Set(authorizationHeader, token)
	return nil
}

func (c *Client) buildFetchRequest(ctx context.Context, url string, page string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build request")
	}
	headers := correlation.HeadersForRequest(req)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	query := req.URL.Query()
	query.Add(c.apiConfig.PagingCountParam, "true")
	query.Add(c.apiConfig.PagingPageParam, page)
	query.Add(c.apiConfig.PagingSizeParam, strconv.Itoa(c.apiConfig.PageSize))
	req.URL.RawQuery = query.Encode()
	if err := c.setToken(req); err != nil {
		return nil, err
	}
	return req, nil
}

type token struct {
	AccessToken      string `json:"access_token,omitempty"`
	ExpiresInSeconds int64  `json:"expires_in,omitempty"`
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	if c.authToken != "" && time.Now().Before(c.authTokenValidity.Add(-c.apiConfig.Timeout)) {
		return c.authToken, nil
	}
	tokenRequestBody := strings.NewReader(url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.authConfig.ClientID},
		"client_secret": {c.authConfig.ClientSecret},
	}.Encode())
	tokenRequest, err := http.NewRequestWithContext(ctx, http.MethodPost, c.authConfig.TokenURL, tokenRequestBody)
	if err != nil {
		return "", errors.Wrap(err, "failed to create token request")
	}
	tokenRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	authToken, err := c.doTokenRequest(tokenRequest)
	if err != nil {
		return "", errors.Wrap(err, "failed to do token request")
	}
	c.authToken = authToken.AccessToken
	c.authTokenValidity = time.Now().Add(time.Second * time.Duration(authToken.ExpiresInSeconds))
	return "", nil
}

func (c *Client) doTokenRequest(tokenRequest *http.Request) (token, error) {
	res, err := c.httpClient.Do(tokenRequest)
	if err != nil {
		return token{}, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.C(tokenRequest.Context()).WithError(err).Error("Unable to close token response body")
		}
	}()
	if res.StatusCode != http.StatusOK {
		return token{}, fmt.Errorf("token request failed with %s", res.Status)
	}
	authToken := token{}
	if err := json.NewDecoder(res.Body).Decode(&authToken); err != nil {
		return token{}, errors.Wrap(err, "failed to decode token")
	}
	return authToken, nil
}

// FetchDestinationSensitiveData returns sensitive data of a destination
func (c *Client) FetchDestinationSensitiveData(ctx context.Context, destinationName string) ([]byte, error) {
	fetchURL := fmt.Sprintf("%s%s/%s", c.authConfig.URL, c.apiConfig.EndpointFindDestination, destinationName)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to build request")
	}
	req.Header.Set(correlation.RequestIDHeaderKey, correlation.CorrelationIDForRequest(req))
	if err := c.setToken(req); err != nil {
		return nil, err
	}
	res, err := c.sendRequestWithRetry(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := res.Body.Close(); err != nil {
			log.C(req.Context()).WithError(err).Error("Unable to close response body")
		}
	}()

	if res.StatusCode == http.StatusNotFound {
		return nil, apperrors.NewNotFoundError(resource.Destination, destinationName)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Errorf("received status code %d when trying to get destination info for %s",
			res.StatusCode, destinationName)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, errors.Wrap(err, "failed to read body of response")
	}

	return body, nil
}

func (c *Client) sendRequestWithRetry(req *http.Request) (*http.Response, error) {
	var response *http.Response
	err := retry.Do(func() error {
		res, err := c.httpClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "failed to execute HTTP request")
		}

		if err == nil && res.StatusCode < http.StatusInternalServerError {
			response = res
			return nil
		}

		defer func() {
			if err := res.Body.Close(); err != nil {
				log.C(req.Context()).WithError(err).Error("Unable to close response body")
			}
		}()
		body, err := io.ReadAll(res.Body)

		if err != nil {
			return errors.Wrap(err, "failed to read response body")
		}
		return errors.Errorf("request failed with status code %d, error message: %v", res.StatusCode, string(body))
	}, retry.Attempts(c.apiConfig.RetryAttempts), retry.Delay(c.apiConfig.RetryInterval))

	return response, err
}
