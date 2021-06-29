package product_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	productModel := fixProductModel()
	productEntity := fixEntityProduct()
	insertQuery := `^INSERT INTO public.products \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixProductRow()...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", productModel).Return(productEntity, nil).Once()
		pgRepository := product.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, productModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		pgRepository := product.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.products SET title = ?, short_description = ?, vendor = ?, parent = ?, labels = ?, correlation_ids = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		productModel := fixProductModel()
		entity := fixEntityProduct()

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", productModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(append(fixProductUpdateArgs(), tenantID, entity.ID)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := product.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, productModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := product.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, nil)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := `^DELETE FROM public.products WHERE tenant_id = \$1 AND id = \$2$`

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

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.products WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, productID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := product.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, productID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	productEntity := fixEntityProduct()

	selectQuery := `^SELECT (.+) FROM public.products WHERE tenant_id = \$1 AND id = \$2$`

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

	selectQuery := `^SELECT (.+) FROM public.products 
		WHERE tenant_id = \$1 AND app_id = \$2`

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
