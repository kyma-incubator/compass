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
				Query:       regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.vendors SET title = ?, labels = ?, partners = ? WHERE id = ? AND (id IN (SELECT id FROM vendors_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:        append(fixVendorUpdateArgs(), entity.ID),
				ValidResult: sqlmock.NewResult(-1, 1),
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

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM public.vendors WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, vendorID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := ordvendor.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, vendorID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT 1 FROM public.vendors WHERE %s AND id = $2`, fixUnescapedTenantIsolationSubquery()))

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, vendorID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := ordvendor.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, vendorID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	vendorEntity := fixEntityVendor()

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.vendors WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixVendorColumns()).
			AddRow(fixVendorRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, vendorID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", vendorEntity).Return(&model.Vendor{ID: vendorID, TenantID: tenantID}, nil).Once()
		pgRepository := ordvendor.NewRepository(convMock)
		// WHEN
		modelVendor, err := pgRepository.GetByID(ctx, tenantID, vendorID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, vendorID, modelVendor.ID)
		assert.Equal(t, tenantID, modelVendor.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := ordvendor.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, vendorID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelVendor, err := repo.GetByID(ctx, tenantID, vendorID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelVendor)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixVendorColumns()).
			AddRow(fixVendorRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, vendorID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", vendorEntity).Return(&model.Vendor{}, testError).Once()
		pgRepository := ordvendor.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, vendorID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}

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
