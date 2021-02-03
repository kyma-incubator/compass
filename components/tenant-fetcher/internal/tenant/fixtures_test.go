package tenant_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/tenant-fetcher/internal/tenant"
	"github.com/pkg/errors"
)

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.ProviderName, ent.Status}
}

const (
	testID           = "foo"
	testName         = "bar"
	createQuery      = "INSERT INTO public.business_tenant_mappings ( id, external_name, external_tenant, provider_name, status ) VALUES ( ?, ?, ?, ?, ? )"
	deleteQuery      = "DELETE FROM public.business_tenant_mappings WHERE external_tenant = $1"
	testProviderName = "saas-manager"
)

var testError = errors.New("test error")

var createQueryArgs = fixTenantMappingCreateArgs(tenant.Entity{
	ID:             testID,
	Name:           testID,
	ExternalTenant: testID,
	Status:         tenant.Active,
	ProviderName:   testProviderName,
})
