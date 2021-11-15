package ordvendor_test

import (
	"database/sql/driver"
	"fmt"
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
		SqlQueryDetails: []testdb.SqlQueryDetails{
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

func TestPgRepository_Update(t *testing.T) {
	entity := fixEntityVendor()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Vendor",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.vendors SET title = ?, labels = ?, partners = ? WHERE id = ? AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
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
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Vendor Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Vendor Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Vendor",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, title, labels, partners, id FROM public.vendors WHERE id = $1 AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = $2))`),
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

/*
func TestPgRepository_ListByApplicationID(t *testing.T) {
	// GIVEN
	totalCount := 2
	firstVendorEntity := fixEntityVendor()
	secondVendorEntity := fixEntityVendor()

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.vendors WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixVendorColumns()).
			AddRow(fixVendorRow()...).
			AddRow(fixVendorRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstVendorEntity).Return(&model.Vendor{ID: firstVendorEntity.ID}, nil)
		convMock.On("FromEntity", secondVendorEntity).Return(&model.Vendor{ID: secondVendorEntity.ID}, nil)
		pgRepository := ordvendor.NewRepository(convMock)
		// WHEN
		modelVendor, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelVendor, totalCount)
		assert.Equal(t, firstVendorEntity.ID, modelVendor[0].ID)
		assert.Equal(t, secondVendorEntity.ID, modelVendor[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/
