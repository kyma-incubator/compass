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
	testProvider = "Compass"
)

var (
	testError        = errors.New("test error")
	testTableColumns = []string{"id", "external_name", "external_tenant", "provider_name", "status"}
)

func newModelBusinessTenantMapping(id, name string) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Provider:       testProvider,
		Status:         model.Active,
	}
}

func newEntityBusinessTenantMapping(id, name string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		ProviderName:   testProvider,
		Status:         tenant.Active,
	}
}

type sqlRow struct {
	id             string
	name           string
	externalTenant string
	provider       string
	status         tenant.TenantStatus
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.provider, row.status)
	}
	return out
}

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.ProviderName, ent.Status}
}

func newModelBusinessTenantMapingPage(tenants []*model.BusinessTenantMapping) model.BusinessTenantMappingPage {
	return model.BusinessTenantMappingPage{
		Data: tenants,
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: len(tenants),
	}
}

func newModelBusinessTenantMappingInput(name string) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: testExternal,
		Provider:       testProvider,
	}
}
