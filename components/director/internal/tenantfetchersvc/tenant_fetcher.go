package tenantfetchersvc

import (
	"context"
)

type fetcher struct {
	svc SubaccountOnDemandService
}

// NewTenantFetcher creates new fetcher
func NewTenantFetcher(svc SubaccountOnDemandService) *fetcher {
	return &fetcher{
		svc: svc,
	}
}

// FetchTenantOnDemand fetches creation events for a subaccount and creates a tenant for the subaccount in case it doesn't exist
func (f *fetcher) FetchTenantOnDemand(ctx context.Context, tenantID string, parentTenantID string) error {
	return f.svc.SyncTenant(ctx, tenantID, parentTenantID)
}
