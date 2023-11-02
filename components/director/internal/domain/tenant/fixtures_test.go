package tenant_test

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	testLicenseType               = "TESTLICENSE"
	initializedColumn             = "initialized"
	invalidResourceType           = "INVALID"
)

var (
	testCustomerID               = str.Ptr("customerID")
	testError                    = errors.New("test error")
	testTableColumns             = []string{"id", "external_name", "external_tenant", "parent", "type", "provider_name", "status"}
	tenantAccessTestTableColumns = []string{"tenant_id", "id", "owner"}
	tenantAccessInput            = graphql.TenantAccessInput{
		TenantID:     testExternal,
		ResourceType: graphql.TenantAccessObjectTypeApplication,
		ResourceID:   testID,
		Owner:        true,
	}
	tenantAccessInputWithInvalidResourceType = graphql.TenantAccessInput{
		TenantID:     testExternal,
		ResourceType: graphql.TenantAccessObjectType(invalidResourceType),
		ResourceID:   testID,
		Owner:        true,
	}
	tenantAccessGQL = &graphql.TenantAccess{
		TenantID:     testExternal,
		ResourceType: graphql.TenantAccessObjectTypeApplication,
		ResourceID:   testID,
		Owner:        true,
	}
	tenantAccessModel = &model.TenantAccess{
		ExternalTenantID: testExternal,
		InternalTenantID: testInternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
	}
	tenantAccessModelWithoutExternalTenant = &model.TenantAccess{
		InternalTenantID: testInternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
	}
	tenantAccessWithoutInternalTenantModel = &model.TenantAccess{
		ExternalTenantID: testExternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
	}
	tenantAccessEntity = &repo.TenantAccess{
		TenantID:   testInternal,
		ResourceID: testID,
		Owner:      true,
	}
	invalidTenantAccessModel = &model.TenantAccess{
		ResourceType: invalidResourceType,
	}
	expectedTenantModel = &model.BusinessTenantMapping{
		ID:             testExternal,
		Name:           testName,
		ExternalTenant: testExternal,
		Parent:         "",
		Type:           tenant.Account,
		Provider:       testProvider,
		Status:         tenant.Active,
		Initialized:    nil,
	}

	expectedTenantGQL = &graphql.Tenant{
		ID:          testExternal,
		InternalID:  testInternal,
		Name:        str.Ptr(testName),
		Type:        string(tenant.Account),
		ParentID:    "",
		Initialized: nil,
		Labels:      nil,
	}
)

func newModelBusinessTenantMapping(id, name string) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithType(id, name, "", nil, tenant.Account)
}

func newModelBusinessTenantMappingWithLicense(id, name string, licenseType *string) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithType(id, name, "", licenseType, tenant.Account)
}

func newModelBusinessTenantMappingWithType(id, name, parent string, licenseType *string, tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         parent,
		Type:           tenantType,
		Provider:       testProvider,
		Status:         tenant.Active,
		LicenseType:    licenseType,
	}
}

func newModelBusinessTenantMappingWithComputedValues(id, name string, initialized *bool) *model.BusinessTenantMapping {
	tenantModel := newModelBusinessTenantMapping(id, name)
	tenantModel.Initialized = initialized
	return tenantModel
}

func newModelBusinessTenantMappingWithParentAndType(id, name, parent string, licenseType *string, tntType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parent:         parent,
		Type:           tntType,
		Provider:       testProvider,
		Status:         tenant.Active,
		Initialized:    boolToPtr(true),
		LicenseType:    licenseType,
	}
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

func newEntityBusinessTenantMappingWithParentAndAccount(id, name, parent string, tntType tenant.Type) *tenant.Entity {
	tnt := newEntityBusinessTenantMappingWithParent(id, name, parent)
	tnt.Type = tntType

	return tnt
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

func newModelBusinessTenantMappingInput(name, subdomain, region string, licenseType *string) model.BusinessTenantMappingInput {
	return newModelBusinessTenantMappingInputWithType(testExternal, name, "", subdomain, region, licenseType, tenant.Account)
}

func newModelBusinessTenantMappingInputWithCustomerID(name string, customerID *string) model.BusinessTenantMappingInput {
	tnt := newModelBusinessTenantMappingInputWithType(testExternal, name, "", "", "", nil, tenant.Subaccount)
	tnt.CustomerID = customerID
	return tnt
}

func newModelBusinessTenantMappingInputWithType(tenantID, name, parent, subdomain, region string, licenseType *string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: tenantID,
		Subdomain:      subdomain,
		Region:         region,
		Parent:         parent,
		Type:           tenant.TypeToStr(tenantType),
		Provider:       testProvider,
		LicenseType:    licenseType,
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

func boolToPtr(in bool) *bool {
	return &in
}

func unusedConverter() *automock.BusinessTenantMappingConverter {
	return &automock.BusinessTenantMappingConverter{}
}

func unusedDBMock(t *testing.T) (*sqlx.DB, testdb.DBMock) {
	return testdb.MockDatabase(t)
}

func unusedTenantConverter() *automock.BusinessTenantMappingConverter {
	return &automock.BusinessTenantMappingConverter{}
}

func unusedTenantService() *automock.BusinessTenantMappingService {
	return &automock.BusinessTenantMappingService{}
}

func unusedFetcherService() *automock.Fetcher {
	return &automock.Fetcher{}
}
