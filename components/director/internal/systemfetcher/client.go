package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/paging"
	"github.com/pkg/errors"
)

// APIClient missing godoc
//go:generate mockery --name=APIClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIClient interface {
	Do(*http.Request, string) (*http.Response, error)
}

// APIConfig missing godoc
type APIConfig struct {
	Endpoint        string        `envconfig:"APP_SYSTEM_INFORMATION_ENDPOINT"`
	FilterCriteria  string        `envconfig:"APP_SYSTEM_INFORMATION_FILTER_CRITERIA"`
	Timeout         time.Duration `envconfig:"APP_SYSTEM_INFORMATION_FETCH_TIMEOUT"`
	PageSize        uint64        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SIZE"`
	PagingSkipParam string        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SKIP_PARAM"`
	PagingSizeParam string        `envconfig:"APP_SYSTEM_INFORMATION_PAGE_SIZE_PARAM"`
	SystemSourceKey string        `envconfig:"APP_SYSTEM_INFORMATION_SOURCE_KEY"`
}

// Client missing godoc
type Client struct {
	apiConfig  APIConfig
	httpClient APIClient
}

// NewClient missing godoc
func NewClient(apiConfig APIConfig, client APIClient) *Client {
	return &Client{
		apiConfig:  apiConfig,
		httpClient: client,
	}
}

// FetchSystemsForTenant fetches systems from the service
func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant string) ([]System, error) {
	qp := c.buildFilter()
	log.C(ctx).Infof("Fetching systems for tenant %s with query: %s", tenant, qp)

	var systems []System

	systemsFunc := c.getSystemsPagingFunc(ctx, &systems, tenant)
	pi := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp, c.apiConfig.PageSize, systemsFunc)
	if err := pi.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems for tenant %s", tenant)
	}

	log.C(ctx).Infof("Fetched systems for tenant %s", tenant)
	return systems, nil
}

func (c *Client) fetchSystemsForTenant(ctx context.Context, url, tenant string) ([]System, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HTTP request")
	}

	resp, err := c.httpClient.Do(req, tenant)
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

	respBody, err := io.ReadAll(resp.Body)
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

func (c *Client) buildFilter() map[string]string {
	var queryBuilder strings.Builder

	for idx, at := range ApplicationTemplates {
		lbl, ok := at.Labels[ApplicationTemplateLabelFilter]
		if !ok {
			continue
		}

		queryBuilder.WriteString(fmt.Sprintf(" %s eq '%s' ", c.apiConfig.SystemSourceKey, lbl.Value))

		if idx < len(ApplicationTemplates)-1 {
			queryBuilder.WriteString("or")
		}
	}

	return map[string]string{"$filter": fmt.Sprintf(c.apiConfig.FilterCriteria, queryBuilder.String()), "fetchAcrossZones": "true"}
}
