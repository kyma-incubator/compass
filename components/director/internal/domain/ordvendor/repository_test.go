package ordvendor_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	vendorModel := fixVendorModel()
	vendorEntity := fixEntityVendor()
	insertQuery := `^INSERT INTO public.vendors \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixVendorRow()...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", vendorModel).Return(vendorEntity, nil).Once()
		pgRepository := ordvendor.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, vendorModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		pgRepository := ordvendor.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.vendors SET title = ?, type = ?, labels = ? WHERE tenant_id = ? AND ord_id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		vendorModel := fixVendorModel()
		entity := fixEntityVendor()

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", vendorModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(append(fixVendorUpdateArgs(), tenantID, entity.OrdID)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := ordvendor.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, vendorModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := ordvendor.NewRepository(convMock)
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
	deleteQuery := `^DELETE FROM public.vendors WHERE tenant_id = \$1 AND ord_id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, ordID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := ordvendor.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, ordID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.vendors WHERE tenant_id = $1 AND ord_id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, ordID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := ordvendor.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, ordID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	vendorEntity := fixEntityVendor()

	selectQuery := `^SELECT (.+) FROM public.vendors WHERE tenant_id = \$1 AND ord_id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixVendorColumns()).
			AddRow(fixVendorRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, ordID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", vendorEntity).Return(&model.Vendor{OrdID: ordID, TenantID: tenantID}, nil).Once()
		pgRepository := ordvendor.NewRepository(convMock)
		// WHEN
		modelBndl, err := pgRepository.GetByID(ctx, tenantID, ordID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, ordID, modelBndl.OrdID)
		assert.Equal(t, tenantID, modelBndl.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := ordvendor.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, ordID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelBndl, err := repo.GetByID(ctx, tenantID, ordID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelBndl)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixVendorColumns()).
			AddRow(fixVendorRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, ordID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", vendorEntity).Return(&model.Vendor{}, testError).Once()
		pgRepository := ordvendor.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, ordID)
		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})
}
