package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/avast/retry-go/v4"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/paging"
	"github.com/pkg/errors"
)

// APIClient missing godoc
//
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
	SystemRPSLimit  uint64        `envconfig:"default=15,APP_SYSTEM_INFORMATION_RPS_LIMIT"`
}

// Client missing godoc
type Client struct {
	apiConfig      APIConfig
	httpClient     APIClient
	jwtTokenClient APIClient
}

// NewClient missing godoc
func NewClient(apiConfig APIConfig, client APIClient, tokenClient APIClient) *Client {
	return &Client{
		apiConfig:      apiConfig,
		httpClient:     client,
		jwtTokenClient: tokenClient,
	}
}

var currentRPS uint64

// FetchSystemsForTenant fetches systems from the service
func (c *Client) FetchSystemsForTenant(ctx context.Context, tenant *model.BusinessTenantMapping, systemSynchronizationTimestamps map[string]SystemSynchronizationTimestamp) ([]System, error) {
	qp := c.buildFilter(systemSynchronizationTimestamps)
	log.C(ctx).Infof("Fetching systems for tenant %s of type %s with query: %s", tenant.ExternalTenant, tenant.Type, qp)

	var systems []System

	systemsFunc := c.getSystemsPagingFunc(ctx, &systems, tenant)
	pi := paging.NewPageIterator(c.apiConfig.Endpoint, c.apiConfig.PagingSkipParam, c.apiConfig.PagingSizeParam, qp, c.apiConfig.PageSize, systemsFunc)
	if err := pi.FetchAll(); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch systems for tenant %s", tenant.ExternalTenant)
	}

	log.C(ctx).Infof("Fetched systems for tenant %s", tenant.ExternalTenant)
	return systems, nil
}

func (c *Client) fetchSystemsForTenant(ctx context.Context, url string, tenant *model.BusinessTenantMapping) ([]System, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new HTTP request")
	}

	var (
		resp *http.Response
	)
	if tenant.Type == tenantEntity.Customer {
		resp, err = c.jwtTokenClient.Do(req, tenant.ExternalTenant)
	} else {
		resp, err = c.httpClient.Do(req, tenant.ExternalTenant)
	}
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

	for idx, system := range systems {
		templatedSystem, err := system.EnhanceWithTemplateID()
		if err != nil {
			return nil, errors.Wrap(err, "failed to map systems with Application Template ID")
		}
		systems[idx] = templatedSystem
	}

	return systems, nil
}

func (c *Client) getSystemsPagingFunc(ctx context.Context, systems *[]System, tenant *model.BusinessTenantMapping) func(string) (uint64, error) {
	return func(url string) (uint64, error) {
		err := retry.Do(
			func() error {
				if atomic.LoadUint64(&currentRPS) >= c.apiConfig.SystemRPSLimit {
					return errors.New("RPS limit reached")
				} else {
					atomic.AddUint64(&currentRPS, 1)
					return nil
				}
			},
			retry.Attempts(0),
			retry.Delay(time.Millisecond*100),
		)

		if err != nil {
			return 0, err
		}

		var currentSystems []System
		err = retry.Do(
			func() error {
				currentSystems, err = c.fetchSystemsForTenant(ctx, url, tenant)
				if err != nil && err.Error() == "unexpected status code: expected: 200, but got: 401" {
					return retry.Unrecoverable(err)
				}
				return err
			},
			retry.Attempts(3),
			retry.Delay(time.Second),
			retry.OnRetry(func(n uint, err error) {
				log.C(ctx).Infof("Retrying request attempt (%d) after error %v", n, err)
			}),
		)

		atomic.AddUint64(&currentRPS, ^uint64(0))

		if err != nil {
			return 0, err
		}
		log.C(ctx).Infof("Fetched page of systems for URL %s", url)
		*systems = append(*systems, currentSystems...)
		return uint64(len(currentSystems)), nil
	}
}

func (c *Client) buildFilter(systemSynchronizationTimestamps map[string]SystemSynchronizationTimestamp) map[string]string {
	var filterBuilder FilterBuilder

	usedSystemRoles := make(map[string]bool)

	for _, key := range SortedTemplateMappingKeys {
		templateMapping := ApplicationTemplates[key]

		appTemplateLblFilter, ok := templateMapping.Labels[ApplicationTemplateLabelFilter]
		if !ok {
			continue
		}

		appTemplateLblFilterArr, ok := appTemplateLblFilter.Value.([]interface{})
		if !ok {
			continue
		}

		for _, lbl := range appTemplateLblFilterArr {
			appTemplateLblStr, ok := lbl.(string)
			if !ok {
				continue
			}

			if _, exists := usedSystemRoles[appTemplateLblStr]; exists {
				continue
			}

			usedSystemRoles[appTemplateLblStr] = true

			expr1 := filterBuilder.NewExpression(SystemSourceKey, "eq", appTemplateLblStr)

			lblExists := false
			minTime := time.Now()

			if timestamp, ok := systemSynchronizationTimestamps[appTemplateLblStr]; ok {
				lblExists = true
				if timestamp.LastSyncTimestamp.Before(minTime) {
					minTime = timestamp.LastSyncTimestamp
				}
			}
			if lblExists {
				expr2 := filterBuilder.NewExpression("lastChangeDateTime", "gt", minTime.String())
				filterBuilder.addFilter(expr1, expr2)
			} else {
				filterBuilder.addFilter(expr1)
			}
		}
	}
	result := map[string]string{"fetchAcrossZones": "true"}

	if len(c.apiConfig.FilterCriteria) > 0 {
		result["$filter"] = fmt.Sprintf(c.apiConfig.FilterCriteria, filterBuilder.buildFilterQuery())
	}

	selectFilter := strings.Join(SelectFilter, ",")
	if len(selectFilter) > 0 {
		result["$select"] = selectFilter
	}
	return result
}
