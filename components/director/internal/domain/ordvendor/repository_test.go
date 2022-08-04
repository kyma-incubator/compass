package ordvendor_test

import (
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Vendor",
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
				Query:       `^INSERT INTO public.vendors \(.+\) VALUES \(.+\)$`,
				Args:        fixVendorRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordvendor.NewRepository,
		ModelEntity:               fixVendorModel(),
		DBEntity:                  fixEntityVendor(),
		NilModelEntity:            fixNilModelVendor(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_CreateGlobal(t *testing.T) {
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Global Vendor",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       `^INSERT INTO public.vendors \(.+\) VALUES \(.+\)$`,
				Args:        fixVendorRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordvendor.NewRepository,
		ModelEntity:               fixVendorModel(),
		DBEntity:                  fixEntityVendor(),
		NilModelEntity:            fixNilModelVendor(),
		MethodName:                "CreateGlobal",
		DisableConverterErrorTest: true,
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityVendor()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Vendor",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.vendors SET title = ?, labels = ?, partners = ?, documentation_labels = ? WHERE id = ? AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          append(fixVendorUpdateArgs(), entity.ID, tenantID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordvendor.NewRepository,
		ModelEntity:               fixVendorModel(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelVendor(),
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_UpdateGlobal(t *testing.T) {
	entity := fixEntityVendor()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Global Vendor",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.vendors SET title = ?, labels = ?, partners = ?, documentation_labels = ? WHERE id = ?`),
				Args:          append(fixVendorUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       ordvendor.NewRepository,
		ModelEntity:               fixVendorModel(),
		DBEntity:                  entity,
		NilModelEntity:            fixNilModelVendor(),
		DisableConverterErrorTest: true,
		UpdateMethodName:          "UpdateGlobal",
		IsGlobal:                  true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Vendor Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.vendors WHERE id = $1 AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{vendorID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordvendor.NewRepository,
		MethodArgs:          []interface{}{tenantID, vendorID},
	}

	suite.Run(t)
}

func TestPgRepository_DeleteGlobal(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Vendor Delete",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.vendors WHERE id = $1`),
				Args:          []driver.Value{vendorID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordvendor.NewRepository,
		MethodName:          "DeleteGlobal",
		MethodArgs:          []interface{}{vendorID},
		IsGlobal:            true,
	}

	suite.Run(t)
}

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Vendor Exists",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.vendors WHERE id = $1 AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{vendorID, tenantID},
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
		RepoConstructorFunc: ordvendor.NewRepository,
		TargetID:            vendorID,
		TenantID:            tenantID,
		MethodName:          "Exists",
		MethodArgs:          []interface{}{tenantID, vendorID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Vendor",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, labels, partners, id, documentation_labels FROM public.vendors WHERE id = $1 AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{vendorID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns()).AddRow(fixVendorRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordvendor.NewRepository,
		ExpectedModelEntity: fixVendorModel(),
		ExpectedDBEntity:    fixEntityVendor(),
		MethodArgs:          []interface{}{tenantID, vendorID},
	}

	suite.Run(t)
}

func TestPgRepository_GetByIDGlobal(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Global Vendor",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, labels, partners, id, documentation_labels FROM public.vendors WHERE id = $1`),
				Args:     []driver.Value{vendorID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns()).AddRow(fixVendorRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: ordvendor.NewRepository,
		ExpectedModelEntity: fixVendorModel(),
		ExpectedDBEntity:    fixEntityVendor(),
		MethodName:          "GetByIDGlobal",
		MethodArgs:          []interface{}{vendorID},
	}

	suite.Run(t)
}

func TestPgRepository_ListByApplicationID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Vendors",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, labels, partners, id, documentation_labels FROM public.vendors WHERE app_id = $1 AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{appID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns()).AddRow(fixVendorRowWithTitle("title1")...).AddRow(fixVendorRowWithTitle("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   ordvendor.NewRepository,
		ExpectedModelEntities: []interface{}{fixVendorModelWithTitle("title1"), fixVendorModelWithTitle("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityVendorWithTitle("title1"), fixEntityVendorWithTitle("title2")},
		MethodArgs:            []interface{}{tenantID, appID},
		MethodName:            "ListByApplicationID",
	}

	suite.Run(t)
}

func TestPgRepository_ListGlobal(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List Global Vendors",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, labels, partners, id, documentation_labels FROM public.vendors WHERE app_id IS NULL FOR UPDATE`),
				Args:     []driver.Value{},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns()).AddRow(fixVendorRowWithTitle("title1")...).AddRow(fixVendorRowWithTitle("title2")...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixVendorColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   ordvendor.NewRepository,
		ExpectedModelEntities: []interface{}{fixVendorModelWithTitle("title1"), fixVendorModelWithTitle("title2")},
		ExpectedDBEntities:    []interface{}{fixEntityVendorWithTitle("title1"), fixEntityVendorWithTitle("title2")},
		MethodArgs:            []interface{}{},
		MethodName:            "ListGlobal",
	}

	suite.Run(t)
}
