package tenant_test

import (
	"database/sql/driver"
	"errors"
	"regexp"
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
	testParent2External           = "externalParent2"
	testParent3External           = "externalParent3"
	testInternal                  = "internalID"
	testID                        = "foo"
	testID2                       = "foo2"
	testName                      = "bar"
	testParent2Name               = "parent2"
	testParent3Name               = "parent3"
	testParentID                  = "parent"
	testParentID2                 = "parent2"
	testParentID3                 = "parent3"
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
	testCustomerID                = str.Ptr("0000customerID")
	testCustomerIDTrimmed         = str.Ptr("customerID")
	testError                     = errors.New("test error")
	testTableColumns              = []string{"id", "external_name", "external_tenant", "type", "provider_name", "status"}
	tenantAccessTestTableColumns  = []string{"tenant_id", "id", "owner", "source"}
	testTenantParentsTableColumns = []string{"tenant_id", "parent_id"}
	testRootParents               = []*model.BusinessTenantMapping{{ID: testParentID}, {ID: testParentID2}}
	tenantGAModel                 = &model.BusinessTenantMapping{
		ID:             testInternal,
		ExternalTenant: testExternal,
		Type:           tenant.Account,
	}
	tenantFolderModel = &model.BusinessTenantMapping{
		ID:             testInternal,
		ExternalTenant: testExternal,
		Type:           tenant.Folder,
	}
	tenantAccessInput = graphql.TenantAccessInput{
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
		Source:           testInternal,
	}
	tenantAccessModelWithSource = &model.TenantAccess{
		ExternalTenantID: testExternal,
		InternalTenantID: testInternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
		Source:           testInternal,
	}
	tenantAccessModelWithoutExternalTenant = &model.TenantAccess{
		InternalTenantID: testInternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
		Source:           testInternal,
	}
	tenantAccessWithoutInternalTenantModel = &model.TenantAccess{
		ExternalTenantID: testExternal,
		ResourceType:     resource.Application,
		ResourceID:       testID,
		Owner:            true,
		Source:           testInternal,
	}
	tenantAccessEntity = &repo.TenantAccess{
		TenantID:   testInternal,
		ResourceID: testID,
		Owner:      true,
		Source:     testInternal,
	}
	invalidTenantAccessModel = &model.TenantAccess{
		ResourceType: invalidResourceType,
	}
	expectedTenantModel = &model.BusinessTenantMapping{
		ID:             testExternal,
		Name:           testName,
		ExternalTenant: testExternal,
		Parents:        []string{},
		Type:           tenant.Account,
		Provider:       testProvider,
		Status:         tenant.Active,
		Initialized:    nil,
	}
	expectedTenantModels = []*model.BusinessTenantMapping{
		{
			ID:             testExternal,
			Name:           testName,
			ExternalTenant: testExternal,
			Parents:        []string{},
			Type:           tenant.Account,
			Provider:       testProvider,
			Status:         tenant.Active,
			Initialized:    nil,
		},
	}

	expectedTenantGQL = &graphql.Tenant{
		ID:          testExternal,
		InternalID:  testInternal,
		Name:        str.Ptr(testName),
		Type:        string(tenant.Account),
		Parents:     []string{},
		Initialized: nil,
		Labels:      nil,
	}

	expectedTenantGQLs = []*graphql.Tenant{
		{
			ID:          testExternal,
			InternalID:  testInternal,
			Name:        str.Ptr(testName),
			Type:        string(tenant.Account),
			Parents:     []string{},
			Initialized: nil,
			Labels:      nil,
		},
	}
)

func newModelBusinessTenantMapping(id, name string, parents []string) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithType(id, name, parents, nil, tenant.Account)
}

func newModelBusinessTenantMappingWithLicense(id, name string, licenseType *string) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithType(id, name, []string{}, licenseType, tenant.Account)
}

func newModelBusinessTenantMappingWithTypeAndExternalID(id, externalID, name string, parents []string, licenseType *string, tenantType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: externalID,
		Parents:        parents,
		Type:           tenantType,
		Provider:       testProvider,
		Status:         tenant.Active,
		LicenseType:    licenseType,
	}
}

func newModelBusinessTenantMappingWithType(id, name string, parents []string, licenseType *string, tenantType tenant.Type) *model.BusinessTenantMapping {
	return newModelBusinessTenantMappingWithTypeAndExternalID(id, testExternal, name, parents, licenseType, tenantType)
}

func newModelBusinessTenantMappingWithComputedValues(id, name string, initialized *bool, parents []string) *model.BusinessTenantMapping {
	tenantModel := newModelBusinessTenantMapping(id, name, parents)
	tenantModel.Initialized = initialized
	return tenantModel
}

func newModelBusinessTenantMappingWithParentAndType(id, name string, parents []string, licenseType *string, tntType tenant.Type) *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             id,
		Name:           name,
		ExternalTenant: testExternal,
		Parents:        parents,
		Type:           tntType,
		Provider:       testProvider,
		Status:         tenant.Active,
		Initialized:    boolToPtr(true),
		LicenseType:    licenseType,
	}
}

func newEntityBusinessTenantMappingWithExternalID(id, externalID, name string) *tenant.Entity {
	return &tenant.Entity{
		ID:             id,
		Name:           name,
		ExternalTenant: externalID,
		Type:           tenant.Account,
		ProviderName:   testProvider,
		Status:         tenant.Active,
	}
}

func newEntityBusinessTenantMapping(id, name string) *tenant.Entity {
	return newEntityBusinessTenantMappingWithExternalID(id, testExternal, name)
}

func newEntityBusinessTenantMappingWithParentAndAccount(id, name string, tntType tenant.Type) *tenant.Entity {
	tnt := newEntityBusinessTenantMapping(id, name)
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
	typeRow        string
	provider       string
	status         tenant.Status
}

type sqlTenantParentsRow struct {
	tenantID string
	parentID string
}

type sqlRowWithComputedValues struct {
	sqlRow
	initialized *bool
}

func fixSQLRowsWithComputedValues(rows []sqlRowWithComputedValues) *sqlmock.Rows {
	columns := append(testTableColumns, initializedColumn)
	out := sqlmock.NewRows(columns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.typeRow, row.provider, row.status, row.initialized)
	}
	return out
}

func fixSQLRows(rows []sqlRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTableColumns)
	for _, row := range rows {
		out.AddRow(row.id, row.name, row.externalTenant, row.typeRow, row.provider, row.status)
	}
	return out
}

func fixSQLTenantParentsRows(rows []sqlTenantParentsRow) *sqlmock.Rows {
	out := sqlmock.NewRows(testTenantParentsTableColumns)
	for _, row := range rows {
		out.AddRow(row.tenantID, row.parentID)
	}
	return out
}

func fixTenantMappingCreateArgs(ent tenant.Entity) []driver.Value {
	return []driver.Value{ent.ID, ent.Name, ent.ExternalTenant, ent.Type, ent.ProviderName, ent.Status}
}

func fixTenantParentCreateArgs(tenantID, parentID string) []driver.Value {
	return []driver.Value{tenantID, parentID}
}

func newModelBusinessTenantMappingInput(name, subdomain, region string, licenseType *string) model.BusinessTenantMappingInput {
	return newModelBusinessTenantMappingInputWithType(testExternal, name, []string{}, subdomain, region, licenseType, tenant.Account)
}

func newModelBusinessTenantMappingInputWithAdditionalFields(name, subdomain, region string, licenseType, additionalFields *string) model.BusinessTenantMappingInput {
	tnt := newModelBusinessTenantMappingInputWithType(testExternal, name, []string{}, subdomain, region, licenseType, tenant.Account)
	tnt.AdditionalFields = additionalFields

	return tnt
}

func newModelBusinessTenantMappingInputWithCustomerID(name string, customerID *string) model.BusinessTenantMappingInput {
	tnt := newModelBusinessTenantMappingInputWithType(testExternal, name, []string{}, "", "", nil, tenant.Subaccount)
	tnt.CustomerID = customerID
	return tnt
}

func newModelBusinessTenantMappingInputWithType(tenantID, name string, parents []string, subdomain, region string, licenseType *string, tenantType tenant.Type) model.BusinessTenantMappingInput {
	return model.BusinessTenantMappingInput{
		Name:           name,
		ExternalTenant: tenantID,
		Subdomain:      subdomain,
		Region:         region,
		Parents:        parents,
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
			Source:     testID,
		},
	}
}

func fixTenantAccessesRow() []driver.Value {
	return []driver.Value{testID, "resourceID", true, "source"}
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

func unusedTenantMappingRepo() *automock.TenantMappingRepository {
	return &automock.TenantMappingRepository{}
}

func unusedFetcherService() *automock.Fetcher {
	return &automock.Fetcher{}
}

func fixDeleteTenantAccessesQuery() string {
	return regexp.QuoteMeta(`WITH RECURSIVE
    parents AS
        (SELECT t1.id,
                t1.type,
                tp1.parent_id,
                0                                                    AS depth,
                CAST( $1 AS uuid) AS child_id
         FROM business_tenant_mappings t1
                  LEFT JOIN tenant_parents tp1 ON t1.id = tp1.tenant_id
         WHERE id = $2
         UNION ALL
         SELECT t2.id, t2.type, tp2.parent_id, p.depth + 1, p.id AS child_id
         FROM business_tenant_mappings t2
                  LEFT JOIN tenant_parents tp2 ON t2.id = tp2.tenant_id
                  INNER JOIN parents p ON p.parent_id = t2.id),
    parent_access_records_count AS (SELECT pp.id    AS tenant_id,
                                           act.id   AS obj_id,
                                           pp.depth,
                                           COUNT(1) AS access_records_count
                                    FROM `) + `(.+)` + regexp.QuoteMeta(`act
                                             JOIN parents pp ON act.tenant_id = pp.id
                                    WHERE act.id IN ($3)
                                    GROUP BY pp.id, pp.depth, act.id),
    anchor AS (SELECT par.*
               FROM parent_access_records_count par
                        LEFT JOIN
                    parent_access_records_count par2 ON par.obj_id = par2.obj_id
                        AND par.depth > par2.depth AND par2.access_records_count > 1
               WHERE par.access_records_count > 1
                 AND par2.tenant_id IS NULL

               UNION ALL

               SELECT par.*
               FROM parent_access_records_count par
                        LEFT JOIN
                    parent_access_records_count par2 ON par.obj_id = par2.obj_id
                        AND par.depth > par2.depth AND par2.access_records_count > 1
                        LEFT JOIN
                    parent_access_records_count par3 ON par.obj_id = par3.obj_id
                        AND par.depth < par3.depth
               WHERE par.access_records_count = 1
                 AND par2.tenant_id IS NULL
                 AND par3.tenant_id IS NULL)
DELETE
FROM `) + `(.+)` + regexp.QuoteMeta(`act
WHERE act.id IN ($4)
  AND EXISTS (SELECT id
              FROM parents
              WHERE tenant_id = parents.id
                AND source = parents.child_id
                AND parents.depth <= ALL (SELECT a.depth FROM anchor a WHERE a.obj_id = act.id));`)
}

func fixInsertTenantAccessesQuery() string {
	return regexp.QuoteMeta(`WITH RECURSIVE parents AS
                  (SELECT t1.id, t1.type, tp1.parent_id, 0 AS depth, CAST(? AS uuid) AS child_id
                   FROM business_tenant_mappings t1 LEFT JOIN tenant_parents tp1 on t1.id = tp1.tenant_id
                   WHERE id=?
                   UNION ALL
                   SELECT t2.id, t2.type, tp2.parent_id, p.depth+ 1, p.id AS child_id
                   FROM business_tenant_mappings t2 LEFT JOIN tenant_parents tp2 on t2.id = tp2.tenant_id
                                                    INNER JOIN parents p on p.parent_id = t2.id)
			INSERT INTO `) + `(.+)` + regexp.QuoteMeta(` ( tenant_id, id, owner, source )  (SELECT parents.id AS tenant_id, ? as id, ? AS owner, parents.child_id as source FROM parents WHERE type != 'cost-object'
                                                                                                                OR (type = 'cost-object' AND depth = (SELECT MIN(depth) FROM parents WHERE type = 'cost-object'))
					)
			ON CONFLICT ( tenant_id, id, source ) DO NOTHING`)
}

func fixDeleteTenantAccessesFromDirective() string {
	return regexp.QuoteMeta(`
DELETE FROM  `) + `(.+)` + regexp.QuoteMeta(` a 
WHERE id IN ($1) AND source IN ($2, $3) AND NOT EXISTS
		(SELECT 1 FROM `) + `(.+)` + regexp.QuoteMeta(` ta WHERE ta.tenant_id = a.source AND ta.id = a.id);
`)
}
