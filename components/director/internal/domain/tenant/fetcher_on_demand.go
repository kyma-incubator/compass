package tenant

import (
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

type fetchOnDemandService struct {
	client           *http.Client
	tenantFetcherURL string
}

// NewFetchOnDemandService returns object responsible for fetching tenants
func NewFetchOnDemandService(client *http.Client, url string) *fetchOnDemandService {
	return &fetchOnDemandService{
		client:           client,
		tenantFetcherURL: url,
	}
}

// FetchOnDemand calls an API which fetch details for the given tenant and store them in the database
func (s *fetchOnDemandService) FetchOnDemand(tenant string) error {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", s.tenantFetcherURL, tenant), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Errorf("while calling tenant-on-demand API %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received status code %d when trying to fetch tenant with ID %s", resp.StatusCode, tenant)
	}
	return nil
}
