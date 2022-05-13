package tenant

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

// Client is responsible for making HTTP requests.
//go:generate mockery --name=Client --output=automock --outpkg=automock --case=underscore --disable-version-string
type Client interface {
	Do(req *http.Request) (*http.Response, error)
}

type fetchOnDemandService struct {
	client           Client
	tenantFetcherURL string
}

// NewFetchOnDemandService returns object responsible for fetching tenants
func NewFetchOnDemandService(client Client, url string) *fetchOnDemandService {
	return &fetchOnDemandService{
		client:           client,
		tenantFetcherURL: url,
	}
}

// FetchOnDemand calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
func (s *fetchOnDemandService) FetchOnDemand(tenant string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.tenantFetcherURL, tenant), nil)
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
