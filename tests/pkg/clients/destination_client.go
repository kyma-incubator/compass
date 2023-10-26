package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

type DestinationServiceAPIConfig struct {
	EndpointTenantSubaccountLevelDestinations            string        `envconfig:"APP_ENDPOINT_TENANT_DESTINATIONS,default=/destination-configuration/v1/subaccountDestinations"`
	EndpointTenantInstanceLevelDestinations              string        `envconfig:"APP_ENDPOINT_TENANT_INSTANCE_LEVEL_DESTINATIONS,default=/destination-configuration/v1/instanceDestinations"`
	EndpointTenantSubaccountLevelDestinationCertificates string        `envconfig:"APP_ENDPOINT_TENANT_DESTINATION_CERTIFICATES,default=/destination-configuration/v1/subaccountCertificates"`
	EndpointTenantInstanceLevelDestinationCertificates   string        `envconfig:"APP_ENDPOINT_TENANT_INSTANCE_LEVEL_DESTINATION_CERTIFICATES,default=/destination-configuration/v1/instanceCertificates"`
	Timeout                                              time.Duration `envconfig:"APP_DESTINATIONS_TIMEOUT,default=30s"`
	SkipSSLVerify                                        bool          `envconfig:"APP_DESTINATIONS_SKIP_SSL_VERIFY,default=false"`
	OAuthTokenPath                                       string        `envconfig:"APP_DESTINATION_OAUTH_TOKEN_PATH,default=/oauth/token"`
}

type Destination struct {
	Name              string `json:"Name"`
	Type              string `json:"Type"`
	URL               string `json:"URL"`
	Authentication    string `json:"Authentication"`
	XCorrelationID    string `json:"x-correlation-id"` // from bundle
	XSystemTenantID   string `json:"x-system-id"`      // local tenant id
	XSystemTenantName string `json:"x-system-name"`    // random or application name
	XSystemType       string `json:"x-system-type"`    // application type
}

type DestinationClient struct {
	httpClient *http.Client
	apiConfig  DestinationServiceAPIConfig
	apiURL     string
}

func NewDestinationClient(instanceConfig config.InstanceConfig, apiConfig DestinationServiceAPIConfig,
	subdomain string) (*DestinationClient, error) {
	ctx := context.Background()

	baseTokenURL, err := url.Parse(instanceConfig.TokenURL)
	if err != nil {
		return nil, errors.Errorf("failed to parse auth url '%s': %v", instanceConfig.TokenURL, err)
	}
	parts := strings.Split(baseTokenURL.Hostname(), ".")
	if len(parts) < 2 {
		return nil, errors.Errorf("auth url '%s' should have a subdomain", instanceConfig.TokenURL)
	}
	originalSubdomain := parts[0]

	tokenURL := strings.Replace(instanceConfig.TokenURL, originalSubdomain, subdomain, 1) + apiConfig.OAuthTokenPath
	cfg := clientcredentials.Config{
		ClientID:  instanceConfig.ClientID,
		TokenURL:  tokenURL,
		AuthStyle: oauth2.AuthStyleInParams,
	}
	cert, err := tls.X509KeyPair([]byte(instanceConfig.Cert), []byte(instanceConfig.Key))
	if err != nil {
		return nil, errors.Errorf("failed to create destinations client x509 pair: %v", err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: apiConfig.SkipSSLVerify,
			Certificates:       []tls.Certificate{cert},
		},
	}

	mtlsClient := &http.Client{
		Transport: transport,
		Timeout:   apiConfig.Timeout,
	}

	ctx = context.WithValue(ctx, oauth2.HTTPClient, mtlsClient)

	httpClient := cfg.Client(ctx)
	httpClient.Timeout = apiConfig.Timeout

	return &DestinationClient{
		httpClient: httpClient,
		apiConfig:  apiConfig,
		apiURL:     instanceConfig.URL,
	}, nil
}

func (c *DestinationClient) CreateDestination(t *testing.T, destination Destination) {
	destinationBytes, err := json.Marshal(destination)
	require.NoError(t, err)

	url := c.apiURL + c.apiConfig.EndpointTenantSubaccountLevelDestinations
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(destinationBytes))
	require.NoError(t, err)

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func (c *DestinationClient) DeleteDestination(t *testing.T, destinationName string) {
	url := c.apiURL + c.apiConfig.EndpointTenantSubaccountLevelDestinations + "/" + url.QueryEscape(destinationName)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func (c *DestinationClient) GetDestinationByName(t *testing.T, serviceURL, destinationName, instanceID, token string, expectedStatusCode int) json.RawMessage {
	subpath := ""
	if instanceID != "" {
		subpath = c.apiConfig.EndpointTenantInstanceLevelDestinations
	} else {
		subpath = c.apiConfig.EndpointTenantSubaccountLevelDestinations
	}
	url := serviceURL + subpath + "/" + url.QueryEscape(destinationName)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	request.Header.Set(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))

	httpClient := &http.Client{}
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.apiConfig.SkipSSLVerify,
		},
	}
	httpClient.Timeout = c.apiConfig.Timeout

	resp, err := httpClient.Do(request)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, expectedStatusCode, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %s", resp.StatusCode, expectedStatusCode, string(body)))
	return body
}

func (c *DestinationClient) GetDestinationCertificateByName(t *testing.T, serviceURL, certificateName, instanceID, token string, expectedStatusCode int) json.RawMessage {
	subpath := ""
	if instanceID != "" {
		subpath = c.apiConfig.EndpointTenantInstanceLevelDestinationCertificates
	} else {
		subpath = c.apiConfig.EndpointTenantSubaccountLevelDestinationCertificates
	}
	url := serviceURL + subpath + "/" + url.QueryEscape(certificateName)
	request, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	request.Header.Set(util.AuthorizationHeader, fmt.Sprintf("Bearer %s", token))

	httpClient := &http.Client{}
	httpClient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: c.apiConfig.SkipSSLVerify,
		},
	}
	httpClient.Timeout = c.apiConfig.Timeout

	resp, err := httpClient.Do(request)
	require.NoError(t, err)
	defer func() {
		if err := resp.Body.Close(); err != nil {
			t.Logf("Could not close response body %s", err)
		}
	}()

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	require.Equal(t, expectedStatusCode, resp.StatusCode, fmt.Sprintf("actual status code %d is different from the expected one: %d. Reason: %v", resp.StatusCode, expectedStatusCode, string(body)))
	return body
}
