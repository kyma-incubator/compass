package ordpackage_test

import (
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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
				Args:        fixPackageRowForApp(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModelForApp(),
		DBEntity:                  fixEntityPackageForApp(),
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Package Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.packages \(.+\) VALUES \(.+\)$`,
				Args:        fixPackageRowForAppTemplateVersion(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModelForAppTemplateVersion(),
		DBEntity:                  fixEntityPackageForAppTemplateVersion(),
		NilModelEntity:            fixNilModelPackage(),
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		MethodName:                "CreateGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityPackageForApp()

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
		ModelEntity:               fixPackageModelForApp(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelPackage(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	entity := fixEntityPackageForAppTemplateVersion()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Package Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`UPDATE public.packages SET vendor = ?, title = ?, short_description = ?, description = ?, version = ?, package_links = ?, links = ?,
		licence_type = ?, tags = ?, countries = ?, labels = ?, policy_level = ?, custom_policy_level = ?, part_of_products = ?, line_of_business = ?, industry = ?, resource_hash = ?, documentation_labels = ?, support_info = ? WHERE id = ?`),
				Args:          append(fixPackageUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordpackage.NewRepository,
		ModelEntity:               fixPackageModelForAppTemplateVersion(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelPackage(),
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
		UpdateMethodName:          "UpdateGlobal",
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
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{packageID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRowForApp()...)}
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
		ExpectedModelEntity: fixPackageModelForApp(),
		ExpectedDBEntity:    fixEntityPackageForApp(),
		MethodArgs:          []interface{}{tenantID, packageID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Package",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE id = $1`),
				Args:     []driver.Value{packageID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRowForAppTemplateVersion()...)}
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
		ExpectedModelEntity: fixPackageModelForAppTemplateVersion(),
		ExpectedDBEntity:    fixEntityPackageForAppTemplateVersion(),
		MethodArgs:          []interface{}{packageID},
		MethodName:          "GetByIDGlobal",
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	suiteForApp := testdb.RepoListTestSuite{
		Name: "List Packages For Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE app_id = $1 AND (id IN (SELECT id FROM packages_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRowWithTitleForApp("title1")...).AddRow(fixPackageRowWithTitleForApp("title2")...)}
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
		ExpectedModelEntities: []interface{}{fixPackageModelWithTitleForApp("title1"), fixPackageModelWithTitleForApp("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityPackageWithTitleForApp("title1"), fixEntityPackageWithTitleForApp("title2")},
		MethodArgs:            []interface{}{tenantID, appID, resource.Application},
		MethodName:            "ListByResourceID",
	}

	suiteForAppTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Packages for Application Template Version",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_version_id, ord_id, vendor, title, short_description, description, version, package_links, links, licence_type, tags, countries, labels, policy_level, custom_policy_level, part_of_products, line_of_business, industry, resource_hash, documentation_labels, support_info FROM public.packages WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixPackageColumns()).AddRow(fixPackageRowWithTitleForAppTemplateVersion("title1")...).AddRow(fixPackageRowWithTitleForAppTemplateVersion("title2")...)}
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
		ExpectedModelEntities: []interface{}{fixPackageModelWithTitleForAppTemplateVersion("title1"), fixPackageModelWithTitleForAppTemplateVersion("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityPackageWithTitleForAppTemplateVersion("title1"), fixEntityPackageWithTitleForAppTemplateVersion("title2")},
		MethodArgs:            []interface{}{tenantID, appTemplateVersionID, resource.ApplicationTemplateVersion},
		MethodName:            "ListByResourceID",
	}

	suiteForApp.Run(t)
	suiteForAppTemplateVersion.Run(t)
}
