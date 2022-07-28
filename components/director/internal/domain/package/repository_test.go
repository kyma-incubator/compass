package ordpackage_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/package/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Package",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, appID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       `^INSERT INTO public.packages \(.+\) VALUES \(.+\)$`,
				Args:        fixPackageRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModel(),
		DBEntity:                  fixEntityPackage(),
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityPackage()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Package",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`UPDATE public.packages SET vendor = ?, title = ?, short_description = ?, description = ?, version = ?, package_links = ?, links = ?,
		licence_type = ?, tags = ?, countries = ?, labels = ?, policy_level = ?, custom_policy_level = ?, part_of_products = ?, line_of_business = ?, industry = ?, resource_hash = ?, documentation_labels = ?, support_info = ? WHERE id = ? AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixPackageUpdateArgs(), entity.ID, tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModel(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Package Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.packages WHERE id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{packageID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordpackage.NewRepository,
		MethodArgs:          []interface{}{tenantID, packageID},
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Package Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.packages WHERE id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{packageID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordpackage.NewRepository,
		TargetID:            packageID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, packageID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Package",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{packageID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordpackage.NewRepository,
		ExpectedModelEntity: fixPackageModel(),
		ExpectedDBEntity:    fixEntityPackage(),
		MethodArgs:          []interface{}{tenantID, packageID},
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Packages",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE app_id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRowWithTitle("title1")...).AddRow(fixPackageRowWithTitle("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   ordpackage.NewRepository,
		ExpectedModelEntities: []interface{}{fixPackageModelWithTitle("title1"), fixPackageModelWithTitle("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityPackageWithTitle("title1"), fixEntityPackageWithTitle("title2")},
		MethodArgs:            []interface{}{tenantID, appID},
		MethodName:            "ListByApplicationID",
	}

	suite.Run(t)
}
