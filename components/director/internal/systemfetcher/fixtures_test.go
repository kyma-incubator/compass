package systemfetcher_test

import (
	"database/sql"
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	testExternal      = "external"
	testProvider      = "Compass"
	initializedColumn = "initialized"
)

var (
	testTableColumns = []string{"id", "external_name", "external_tenant", "parent", "type", "provider_name", "status"}
)

func newModelBusinessTenantMapping(id, name string) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         "",
		Type:           tenant.Account,
		Provider:       testProvider,
		Status:         tenant.Active,
	}
}

func newModelBusinessTenantMappingWithComputedValues(id, name string, initialized *bool) *model.BusinessTenantMapping {
	tenantModel := newModelBusinessTenantMapping(id, name)
	tenantModel.Initialized = initialized
	return tenantModel
}

func newEntityBusinessTenantMapping(id, name string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         sql.NullString{},
		Type:           tenant.Account,
		ProviderName:   testProvider,
		Status:         tenant.Active,
	}
}

func newEntityBusinessTenantMappingWithComputedValues(id, name string, initialized *bool) *tenant.Entity {
	tenantEntity := newEntityBusinessTenantMapping(id, name)
	tenantEntity.Initialized = initialized
	return tenantEntity
}

type sqlRow struct {
	id             string
	name           string
	externalTenant string
	parent         sql.NullString
	typeRow        string
	provider       string
	status         tenant.Status
}

type sqlRowWithComputedValues struct {
	sqlRow
	initialized *bool
}

func fixSQLRowsWithComputedValues(rows []sqlRowWithComputedValues) *sqlmock.Rows {
	columns := append(testTableColumns, initializedColumn)
	out := sqlmock.NewRows(columns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.parent, row.typeRow, row.provider, row.status, row.initialized)
	}
	return out
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.parent, row.typeRow, row.provider, row.status)
	}
	return out
}

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.Parent, ent.Type, ent.ProviderName, ent.Status}
}

func newModelBusinessTenantMappingInput(name string) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         "",
		Type:           string(tenant.Account),
		Provider:       testProvider,
	}
}

func newGraphQLTenant(id, internalID, name string) *graphql.Tenant {
	return &graphql.Tenant{
		ID:         id,
		InternalID: internalID,
		Name:       str.Ptr(name),
	}
}
