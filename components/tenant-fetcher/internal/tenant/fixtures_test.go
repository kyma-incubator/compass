package tenant_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
)

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.ProviderName, ent.Status}
}

const (
	testID                           = "foo"
	createQuery                      = "INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ? )"
	deleteQuery                      = "DELETE FROM public.business_tenant_mappings WHERE external_tenant = $1"
	customerID                       = "customer-id"
	subdomain                        = "subdomain"
	testProviderName                 = "test-provider"
	tenantProviderTenantIdProperty   = "TenantId"
	tenantProviderCustomerIdProperty = "CustomerId"
	tenantProviderSubdomainProperty  = "Subdomain"
)

var testError = errors.New("test error")

var createQueryArgs = fixTenantMappingCreateArgs(tenant.Entity{
	ID:             testID,
	Name:           testID,
	ExternalTenant: testID,
	Status:         tenant.Active,
	ProviderName:   testProviderName,
})

type errReader int

func (errReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}
