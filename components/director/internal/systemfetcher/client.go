package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/paging"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

// OAuth2Config missing godoc
type OAuth2Config struct {
	ClientID             string   `envconfig:"APP_OAUTH_CLIENT_ID"`
	ClientSecret         string   `envconfig:"APP_OAUTH_CLIENT_SECRET"`
	TokenEndpointPattern string   `envconfig:"APP_OAUTH_TOKEN_ENDPOINT_PATTERN"`
	TenantHeaderName     string   `envconfig:"APP_OAUTH_TENANT_HEADER_NAME"`
	ScopesClaim          []string `envconfig:"APP_OAUTH_SCOPES_CLAIM"`
}

// APIConfig missing godoc
type APIConfig struct {
	Endpoint                    string        `envconfig:"APP_SYSTEM_INFORMATION_ENDPOINT"`
	FilterCriteria              string        `envconfig:"APP_SYSTEM_INFORMATION_FILTER_CRITERIA"`
	FilterTenantCriteriaPattern string        `envconfig:"APP_SYSTEM_INFORMATION_FILTER_TENANT_CRITERIA_PATTERN"`
	Timeout                     time.Duration `envconfig:"APP_SYSTEM_INFORMATION_FETCH_TIMEOUT"`
	PageSize                    uint64        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SIZE"`
	PagingSkipParam             string        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SKIP_PARAM"`
	PagingSizeParam             string        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SIZE_PARAM"`
}

// Client missing godoc
type Client struct {
	apiConfig     APIConfig
	oAuth2Config  OAuth2Config
	clientCreator clientCreatorFunc
}

type clientCreatorFunc func(ctx context.Context, oauth2Config OAuth2Config) *http.Client

// DefaultClientCreator missing godoc
func DefaultClientCreator(ctx context.Context, oAuth2Config OAuth2Config) *http.Client {
	cfg := clientcredentials.Config{
		ClientID:     oAuth2Config.ClientID,
		ClientSecret: oAuth2Config.ClientSecret,
		TokenURL:     oAuth2Config.TokenEndpointPattern,
		Scopes:       oAuth2Config.ScopesClaim,
	}

	httpClient := cfg.Client(ctx)
	return httpClient
}

// NewClient missing godoc
func NewClient(apiConfig APIConfig, oAuth2Config OAuth2Config, clientCreator clientCreatorFunc) *Client {
	return &Client{
		apiConfig:     apiConfig,
		oAuth2Config:  oAuth2Config,
		clientCreator: clientCreator,
	}
}

// FetchSystemsForTenant fetches systems from the service by making 2 HTTP calls with different filter criteria
func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error) {
	ctx, httpClient := c.clientForTenant(ctx, tenant)

	qp1 := map[string]string{"$filter": c.apiConfig.FilterCriteria}
	var systems []System

	systemsFunc := getSystemsPagingFunc(ctx, httpClient, &systems)
	pi1 := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp1, c.apiConfig.PageSize, systemsFunc)
	if err := pi1.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems for tenant %s", tenant)
	}

	tenantFilter := fmt.Sprintf(c.apiConfig.FilterTenantCriteriaPattern, tenant)
	qp2 := map[string]string{"$filter": tenantFilter}
	var systemsByTenantFilter []System

	systemsByTenantFilterFunc := getSystemsPagingFunc(ctx, httpClient, &systemsByTenantFilter)
	pi2 := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp2, c.apiConfig.PageSize, systemsByTenantFilterFunc)
	if err := pi2.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems with tenant filter for tenant %s", tenant)
	}

	log.C(ctx).Infof("Fetched systems for tenant %s", tenant)
	return append(systems, systemsByTenantFilter...), nil
}

func (c *Client) clientForTenant(ctx context.Context, tenant string) (context.Context, *http.Client) {
	httpTokenClient := &http.Client{
		Timeout: c.apiConfig.Timeout,
		Transport: &HeaderTransport{
			tenantHeaderName: c.oAuth2Config.TenantHeaderName,
			tenant:           tenant,
			base:             http.DefaultTransport,
		},
	}
	// The client provided here is used only to get token, not for
	// the actual request to the system service
	ctx = context.WithValue(ctx, oauth2.HTTPClient, httpTokenClient)

	httpClient := c.clientCreator(ctx, c.oAuth2Config)
	httpClient.Timeout = c.apiConfig.Timeout

	return ctx, httpClient
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

func getSystemsPagingFunc(ctx context.Context, httpClient *http.Client, systems *[]System) func(string) (uint64, error) {
	return func(url string) (uint64, error) {
		currentSystems, err := fetchSystemsForTenant(ctx, httpClient, url)
		if err != nil {
			return 0, err
		}
		log.C(ctx).Infof("Fetched page of systems for URL %s", url)
		*systems = append(*systems, currentSystems...)
		return uint64(len(currentSystems)), nil
	}
}
