package tenant

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// FetchOnDemandAPIConfig is the configuration needed for tenant on demand fetch API
type FetchOnDemandAPIConfig struct {
	TenantOnDemandURL string `envconfig:"optional,APP_FETCH_TENANT_URL"`
	IsDisabled        bool   `envconfig:"default=false,APP_DISABLE_TENANT_ON_DEMAND_MODE"`
}

// Fetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=Fetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type Fetcher interface {
	FetchOnDemand(tenant, parentTenant string) error
}

// Client is responsible for making HTTP requests.
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type fetchOnDemandService struct {
	client           Client
	tenantFetcherURL string
}

type noopOnDemandService struct{}

// NewFetchOnDemandService returns object responsible for fetching tenants
func NewFetchOnDemandService(client Client, config FetchOnDemandAPIConfig) Fetcher {
	if config.IsDisabled {
		return &noopOnDemandService{}
	}
	return &fetchOnDemandService{
		client:           client,
		tenantFetcherURL: config.TenantOnDemandURL,
	}
}

// FetchOnDemand calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
func (s *fetchOnDemandService) FetchOnDemand(tenant, parentTenant string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/%s", s.tenantFetcherURL, parentTenant, tenant), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Wrapf(err, "while calling tenant-on-demand API")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status code %d when trying to fetch tenant with ID %s", resp.StatusCode, tenant)
	}
	return nil
}

func (s *noopOnDemandService) FetchOnDemand(tenant, parentTenant string) error {
	return nil
}
