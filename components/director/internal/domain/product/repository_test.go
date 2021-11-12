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
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:       regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.products SET title = ?, short_description = ?, vendor = ?, parent = ?, labels = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM products_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:        append(fixProductUpdateArgs(), entity.ID),
				ValidResult: sqlmock.NewResult(-1, 1),
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
/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM public.products WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, productID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := product.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, productID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

*/
func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name:                "Product Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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

/*
func TestPgRepository_GetByID(t *testing.T) {
	// given
	productEntity := fixEntityProduct()

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.products WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixProductColumns()).
			AddRow(fixProductRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, productID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", productEntity).Return(&model.Product{ID: productID, TenantID: tenantID}, nil).Once()
		pgRepository := product.NewRepository(convMock)
		// WHEN
		modelProduct, err := pgRepository.GetByID(ctx, tenantID, productID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, productID, modelProduct.ID)
		assert.Equal(t, tenantID, modelProduct.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := product.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, productID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelProduct, err := repo.GetByID(ctx, tenantID, productID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelProduct)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixProductColumns()).
			AddRow(fixProductRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, productID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", productEntity).Return(&model.Product{}, testError).Once()
		pgRepository := product.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, productID)
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
	firstProductEntity := fixEntityProduct()
	secondProductEntity := fixEntityProduct()

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.products WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixProductColumns()).
			AddRow(fixProductRow()...).
			AddRow(fixProductRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstProductEntity).Return(&model.Product{ID: firstProductEntity.ID}, nil)
		convMock.On("FromEntity", secondProductEntity).Return(&model.Product{ID: secondProductEntity.ID}, nil)
		pgRepository := product.NewRepository(convMock)
		// WHEN
		modelProduct, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelProduct, totalCount)
		assert.Equal(t, firstProductEntity.ID, modelProduct[0].ID)
		assert.Equal(t, secondProductEntity.ID, modelProduct[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
*/