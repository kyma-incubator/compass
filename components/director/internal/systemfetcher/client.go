package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	scopes = []string{"uaa.resource"}
)

type OAuth2Config struct {
	ClientID                  string `envconfig:"APP_OAUTH_CLIENT_ID"`
	ClientSecret              string `envconfig:"APP_OAUTH_CLIENT_SECRET"`
	OAuthTokenEndpointPattern string `envconfig:"APP_OAUTH_TOKEN_ENDPOINT_PATTERN"`
}

type APIConfig struct {
	Endpoint                    string        `envconfig:"APP_SYSTEM_INFORMATION_ENDPOINT"`
	FilterCriteria              string        `envconfig:"APP_SYSTEM_INFORMATION_FILTER_CRITERIA"`
	FilterTenantCriteriaPattern string        `envconfig:"APP_SYSTEM_INFORMATION_FILTER_TENANT_CRITERIA_PATTERN"`
	Timeout                     time.Duration `envconfig:"APP_SYSTEM_INFORMATION_FETCH_TIMEOUT"`
}

type Client struct {
	apiConfig         APIConfig
	oAuth2Config      OAuth2Config
	clientCreatorFunc ClientCreatorFunc
}

type ClientCreatorFunc func(ctx context.Context, oauth2Config OAuth2Config, scopes []string, tenant string) *http.Client

func DefaultClientCreator(ctx context.Context, oAuth2Config OAuth2Config, scopes []string, tenant string) *http.Client {
	cfg := clientcredentials.Config{
		ClientID:     oAuth2Config.ClientID,
		ClientSecret: oAuth2Config.ClientSecret,
		TokenURL:     oAuth2Config.OAuthTokenEndpointPattern,
		Scopes:       scopes,
	}

	httpClient := cfg.Client(ctx)
	return httpClient
}

func NewClient(apiConfig APIConfig, oAuth2Config OAuth2Config, clientCreatorFunc ClientCreatorFunc) *Client {
	return &Client{
		apiConfig:         apiConfig,
		oAuth2Config:      oAuth2Config,
		clientCreatorFunc: clientCreatorFunc,
	}
}

func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error) {
	httpTokenClient := &http.Client{
		Timeout: c.apiConfig.Timeout,
		Transport: &HeaderTransport{
			tenant: tenant,
			base:   http.DefaultTransport,
		},
	}
	// The client provided here is used only to get token, not for
	// the actual request to the system service
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpTokenClient)

	httpClient := c.clientCreatorFunc(ctx, c.oAuth2Config, scopes, tenant)
	httpClient.Timeout = c.apiConfig.Timeout

	url := c.apiConfig.Endpoint + "?$filter=" + urlpkg.QueryEscape(c.apiConfig.FilterCriteria)
	systems, err := fetchSystemsForTenant(ctx, httpClient, url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems from %s", url)
	}

	tenantFilter := fmt.Sprintf(c.apiConfig.FilterTenantCriteriaPattern, tenant)
	url = c.apiConfig.Endpoint + "?$filter=" + urlpkg.QueryEscape(tenantFilter)
	systemsByTenantFilter, err := fetchSystemsForTenant(ctx, httpClient, url)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems from %s", url)
	}

	return append(systems, systemsByTenantFilter...), nil
}

func fetchSystemsForTenant(ctx context.Context, httpClient *http.Client, url string) ([]System, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HTTP request")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute HTTP request")
	}
	defer func() {
		err := resp.Body.Close()
		if err != nil {
			log.C(ctx).Println("Failed to close HTTP response body")
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: expected: %d, but got: %d", http.StatusOK, resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse HTTP response body")
	}

	var systems []System
	if err = json.Unmarshal(respBody, &systems); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal systems response")
	}

	return systems, nil
}
