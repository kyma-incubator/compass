package clients

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// TODO Do we need all properties?
type Destination struct {
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
}

type DestinationClient struct {
	httpClient *http.Client
	apiConfig  destinationfetchersvc.DestinationServiceAPIConfig
	apiURL     string
}

func NewDestinationClient(instanceConfig config.InstanceConfig, apiConfig destinationfetchersvc.DestinationServiceAPIConfig,
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

	url := c.apiURL + c.apiConfig.EndpointGetTenantDestinations
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(destinationBytes))
	require.NoError(t, err)

	resp, err := c.httpClient.Do(request)
	require.NoError(t, err)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}
