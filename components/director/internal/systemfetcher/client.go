package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	urlpkg "net/url"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
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
	Endpoint                    string `envconfig:"APP_SYSTEM_INFORMATION_ENDPOINT"`
	FilterCriteria              string `envconfig:"APP_SYSTEM_INFORMATION_FILTER_CRITERIA"`
	FilterTenantCriteriaPattern string `envconfig:"APP_SYSTEM_INFORMATION_FILTER_TENANT_CRITERIA_PATTERN"`
}

type Client struct {
	apiConfig    APIConfig
	oAuth2Config OAuth2Config
}

func NewClient(apiConfig APIConfig, oAuth2Config OAuth2Config) *Client {
	return &Client{
		apiConfig:    apiConfig,
		oAuth2Config: oAuth2Config,
	}
}

func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error) {
	//TODO: See if a custom HTTP fetch with client_creds isn't a better option because this now makes new http clients on every call
	cfg := clientcredentials.Config{
		ClientID:     c.oAuth2Config.ClientID,
		ClientSecret: c.oAuth2Config.ClientSecret,
		TokenURL:     fmt.Sprintf(c.oAuth2Config.OAuthTokenEndpointPattern, tenant),
		Scopes:       scopes,
	}

	// TODO: Check token, err := cfg.Token(ctx) optimization
	httpClient := cfg.Client(ctx)

	//TODO: The double fetch is a better approach if it doesn't affect that much the performance, this is an alternative approach to loading custom fields from the env and branching from them to execute one of the fetches only
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
