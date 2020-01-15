package tenant_test

import (
	"database/sql/driver"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

const (
	testExternal = "external"
	testID       = "foo"
	testName     = "bar"
	testPageSize = 3
	testCursor   = ""
	testInternal = "internal"
)

var (
	testError        = errors.New("test error")
	testDescription  = "bazz"
	testTableColumns = []string{"id", "name", "external_tenant", "internal_tenant", "provider_name", "status"}
)

func fixModelTenantMapping(id, name string) *model.TenantMapping {
	return &model.TenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		InternalTenant: testInternal,
		Provider:       "Compass",
		Status:         model.Active,
	}
}

func fixInactiveModelTenantMapping(id, name string) *model.TenantMapping {
	return &model.TenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		InternalTenant: testInternal,
		Provider:       "Compass",
		Status:         model.Inactive,
	}
}

func fixEntityTenantMapping(id, name string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		InternalTenant: testInternal,
		ProviderName:   "Compass",
		Status:         tenant.Active,
	}
}

func fixInactiveEntityTenantMapping(id, name string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		InternalTenant: testInternal,
		ProviderName:   "Compass",
		Status:         tenant.Inactive,
	}
}

type sqlRow struct {
	id             string
	name           string
	externalTenant string
	internalTenant string
	provider       string
	status         tenant.TenantStatus
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.internalTenant, row.provider, row.status)
	}
	return out
}

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.InternalTenant, ent.ProviderName, ent.Status}
}

func fixModelTenantMapingPage(tenants []*model.TenantMapping) model.TenantMappingPage {
	return model.TenantMappingPage{
		Data: tenants,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(tenants),
	}
}
