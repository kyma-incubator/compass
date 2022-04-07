package tenantfetchersvc

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
)

type fetcher struct {
	svc tenantfetcher.SubaccountOnDemandService
}

// NewTenantFetcher creates new fetcher
func NewTenantFetcher(svc tenantfetcher.SubaccountOnDemandService) *fetcher {
	return &fetcher{
		svc: svc,
	}
}

// FetchTenantOnDemand fetches creation events for a subaccount and creates a tenant for the subaccount in case it doesn't exist
func (f *fetcher) FetchTenantOnDemand(ctx context.Context, tenantID string) error {
	return f.svc.SyncTenant(ctx, tenantID)
}
