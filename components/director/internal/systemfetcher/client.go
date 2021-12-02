package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/auth"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/paging"
	"github.com/pkg/errors"
)

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
	apiConfig  APIConfig
	httpClient http.Client
}

// NewClient missing godoc
func NewClient(apiConfig APIConfig, client http.Client) *Client {
	return &Client{
		apiConfig:  apiConfig,
		httpClient: client,
	}
}

// FetchSystemsForTenant fetches systems from the service by making 2 HTTP calls with different filter criteria
func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error) {
	qp1 := map[string]string{"$filter": c.apiConfig.FilterCriteria}
	var systems []System

	systemsFunc := c.getSystemsPagingFunc(ctx, &systems, tenant)
	pi1 := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp1, c.apiConfig.PageSize, systemsFunc)
	if err := pi1.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems for tenant %s", tenant)
	}

	tenantFilter := fmt.Sprintf(c.apiConfig.FilterTenantCriteriaPattern, tenant)
	qp2 := map[string]string{"$filter": tenantFilter}
	var systemsByTenantFilter []System

	systemsByTenantFilterFunc := c.getSystemsPagingFunc(ctx, &systemsByTenantFilter, tenant)
	pi2 := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp2, c.apiConfig.PageSize, systemsByTenantFilterFunc)
	if err := pi2.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems with tenant filter for tenant %s", tenant)
	}

	log.C(ctx).Infof("Fetched systems for tenant %s", tenant)
	return append(systems, systemsByTenantFilter...), nil
}

func (c *Client) fetchSystemsForTenant(ctx context.Context, url, tenant string) ([]System, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HTTP request")
	}

	req = req.WithContext(addCredentialsToContext(req.Context(), tenant))

	resp, err := c.httpClient.Do(req)
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

func (c *Client) getSystemsPagingFunc(ctx context.Context, systems *[]System, tenant string) func(string) (uint64, error) {
	return func(url string) (uint64, error) {
		currentSystems, err := c.fetchSystemsForTenant(ctx, url, tenant)
		if err != nil {
			return 0, err
		}
		log.C(ctx).Infof("Fetched page of systems for URL %s", url)
		*systems = append(*systems, currentSystems...)
		return uint64(len(currentSystems)), nil
	}
}

func addCredentialsToContext(ctx context.Context, tenant string) context.Context {
	return auth.SaveToContext(ctx, &auth.MtlsOAuthCredentials{Tenant: tenant})
}
