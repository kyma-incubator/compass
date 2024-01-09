package product_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	// GIVEN
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Product",
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
				Query:       `^INSERT INTO public.products \(.+\) VALUES \(.+\)$`,
				Args:        fixProductRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       product.NewRepository,
		ModelEntity:               fixProductModelForApp(),
		DBEntity:                  fixEntityProductForApp(),
		NilModelEntity:            fixNilModelProduct(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	// GIVEN
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Global Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.products \(.+\) VALUES \(.+\)$`,
				Args:        fixProductRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       product.NewRepository,
		ModelEntity:               fixProductModelForApp(),
		DBEntity:                  fixEntityProductForApp(),
		NilModelEntity:            fixNilModelProduct(),
		MethodName:                "CreateGlobal",
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityProduct()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.products SET title = ?, short_description = ?, description = ?, vendor = ?, parent = ?, labels = ?, correlation_ids = ?, tags = ?, documentation_labels = ? WHERE id = ? AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixProductUpdateArgs(), entity.ID, tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       product.NewRepository,
		ModelEntity:               fixProductModelForApp(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelProduct(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	entity := fixEntityProduct()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Product Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.products SET title = ?, short_description = ?, description = ?, vendor = ?, parent = ?, labels = ?, correlation_ids = ?, tags = ?, documentation_labels = ? WHERE id = ?`),
				Args:          append(fixProductUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       product.NewRepository,
		ModelEntity:               fixProductModelForApp(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelProduct(),
		DisableConverterErrorTest: true,
		UpdateMethodName:          "UpdateGlobal",
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Product Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.products WHERE id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{productID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: product.NewRepository,
		MethodArgs:          []interface{}{tenantID, productID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Product Delete Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.products WHERE id = $1`),
				Args:          []driver.Value{productID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: product.NewRepository,
		MethodName:          "DeleteGlobal",
		MethodArgs:          []interface{}{productID},
		IsGlobal:            true,
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Product Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.products WHERE id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{productID, tenantID},
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
		RepoConstructorFunc: product.NewRepository,
		TargetID:            productID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, productID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, title, short_description, description, vendor, parent, labels, correlation_ids, id, tags, documentation_labels FROM public.products WHERE id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{productID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: product.NewRepository,
		ExpectedModelEntity: fixProductModelForApp(),
		ExpectedDBEntity:    fixEntityProductForApp(),
		MethodArgs:          []interface{}{tenantID, productID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Product Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, title, short_description, description, vendor, parent, labels, correlation_ids, id, tags, documentation_labels FROM public.products WHERE id = $1`),
				Args:     []driver.Value{productID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: product.NewRepository,
		ExpectedModelEntity: fixProductModelForApp(),
		ExpectedDBEntity:    fixEntityProductForApp(),
		MethodName:          "GetByIDGlobal",
		MethodArgs:          []interface{}{productID},
	}

	suite.Run(t)
}

func TestPgRepository_ListByResourceID(t *testing.T) {
	suiteForApp := testdb.RepoListTestSuite{
		Name: "List Products for Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, title, short_description, description, vendor, parent, labels, correlation_ids, id, tags, documentation_labels FROM public.products WHERE app_id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRowWithTitleForApp("title1")...).AddRow(fixProductRowWithTitleForApp("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   product.NewRepository,
		ExpectedModelEntities: []interface{}{fixProductModelWithTitleForApp("title1"), fixProductModelWithTitleForApp("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityProductWithTitleForApp("title1"), fixEntityProductWithTitleForApp("title2")},
		MethodArgs:            []interface{}{tenantID, appID, resource.Application},
		MethodName:            "ListByResourceID",
	}

	suiteForAppTemplateVersion := testdb.RepoListTestSuite{
		Name: "List Products for Application Template Version",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, title, short_description, description, vendor, parent, labels, correlation_ids, id, tags, documentation_labels FROM public.products WHERE app_template_version_id = $1 FOR UPDATE`),
				Args:     []driver.Value{appTemplateVersionID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRowWithTitleForAppTemplateVersion("title1")...).AddRow(fixProductRowWithTitleForAppTemplateVersion("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   product.NewRepository,
		ExpectedModelEntities: []interface{}{fixProductModelWithTitleForAppTemplateVersion("title1"), fixProductModelWithTitleForAppTemplateVersion("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityProductWithTitleForAppTemplateVersion("title1"), fixEntityProductWithTitleForAppTemplateVersion("title2")},
		MethodArgs:            []interface{}{tenantID, appTemplateVersionID, resource.ApplicationTemplateVersion},
		MethodName:            "ListByResourceID",
	}

	suiteForApp.Run(t)
	suiteForAppTemplateVersion.Run(t)
}

func TestPgRepository_ListGlobal(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Global",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, app_template_version_id, title, short_description, description, vendor, parent, labels, correlation_ids, id, tags, documentation_labels FROM public.products WHERE app_id IS NULL FOR UPDATE`),
				Args:     []driver.Value{},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRowWithTitleForApp("title1")...).AddRow(fixProductRowWithTitleForApp("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   product.NewRepository,
		ExpectedModelEntities: []interface{}{fixProductModelWithTitleForApp("title1"), fixProductModelWithTitleForApp("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityProductWithTitleForApp("title1"), fixEntityProductWithTitleForApp("title2")},
		MethodArgs:            []interface{}{},
		MethodName:            "ListGlobal",
	}

	suite.Run(t)
}
