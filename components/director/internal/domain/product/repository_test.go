package product_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
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
		ModelEntity:               fixProductModel(),
		DBEntity:                  fixEntityProduct(),
		NilModelEntity:            fixNilModelProduct(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityProduct()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.products SET title = ?, short_description = ?, vendor = ?, parent = ?, labels = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          append(fixProductUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       product.NewRepository,
		ModelEntity:               fixProductModel(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelProduct(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
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
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Product",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, short_description, vendor, parent, labels, correlation_ids, id FROM public.products WHERE id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2))`),
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
		ExpectedModelEntity: fixProductModel(),
		ExpectedDBEntity:    fixEntityProduct(),
		MethodArgs:          []interface{}{tenantID, productID},
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Products",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, short_description, vendor, parent, labels, correlation_ids, id FROM public.products WHERE app_id = $1 AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixProductColumns()).AddRow(fixProductRowWithTitle("title1")...).AddRow(fixProductRowWithTitle("title2")...)}
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
		ExpectedModelEntities: []interface{}{fixProductModelWithTitle("title1"), fixProductModelWithTitle("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityProductWithTitle("title1"), fixEntityProductWithTitle("title2")},
		MethodArgs:            []interface{}{tenantID, appID},
		MethodName:            "ListByApplicationID",
	}

	suite.Run(t)
}
