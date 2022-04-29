package tenant_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	testExternal                  = "external"
	testInternal                  = "internalID"
	testID                        = "foo"
	testName                      = "bar"
	testParentID                  = "parent"
	testParentID2                 = "parent2"
	testInternalParentID          = "internal-parent"
	testTemporaryInternalParentID = "internal-parent-temp"
	testSubdomain                 = "subdomain"
	testRegion                    = "eu-1"
	testProvider                  = "Compass"
	initializedColumn             = "initialized"
)

var (
	testError        = errors.New("test error")
	testTableColumns = []string{"id", "external_name", "external_tenant", "parent", "type", "provider_name", "status"}
)

func newModelBusinessTenantMapping(id, name string) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithType(id, name, "", tenant.Account)
}

func newModelBusinessTenantMappingWithType(id, name, parent string, tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         parent,
		Type:           tenantType,
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
	return newEntityBusinessTenantMappingWithParent(id, name, "")
}

func newEntityBusinessTenantMappingWithParent(id, name, parent string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         repo.NewValidNullableString(parent),
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

func newModelBusinessTenantMappingInput(name, subdomain, region string) model.BusinessTenantMappingInput {
	return newModelBusinessTenantMappingInputWithType(testExternal, name, "", subdomain, region, tenant.Account)
}

func newModelBusinessTenantMappingInputWithType(tenantID, name, parent, subdomain, region string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: tenantID,
		Subdomain:      subdomain,
		Region:         region,
		Parent:         parent,
		Type:           tenant.TypeToStr(tenantType),
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

func fixTenantAccesses() []repo.TenantAccess {
	return []repo.TenantAccess{
		{
			TenantID:   testID,
			ResourceID: "resourceID",
			Owner:      true,
		},
	}
}

func fixTenantAccessesRow() []driver.Value {
	return []driver.Value{testID, "resourceID", true}
}
