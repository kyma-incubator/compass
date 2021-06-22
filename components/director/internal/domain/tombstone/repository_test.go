package tombstone_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	tombstoneModel := fixTombstoneModel()
	tombstoneEntity := fixEntityTombstone()
	insertQuery := `^INSERT INTO public.tombstones \(.+\) VALUES \(.+\)$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		sqlMock.ExpectExec(insertQuery).
			WithArgs(fixTombstoneRow()...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := automock.EntityConverter{}
		convMock.On("ToEntity", tombstoneModel).Return(tombstoneEntity, nil).Once()
		pgRepository := tombstone.NewRepository(&convMock)
		//WHEN
		err := pgRepository.Create(ctx, tombstoneModel)
		//THEN
		require.NoError(t, err)
		sqlMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("returns error when item is nil", func(t *testing.T) {
		ctx := context.TODO()
		convMock := automock.EntityConverter{}
		pgRepository := tombstone.NewRepository(&convMock)
		// WHEN
		err := pgRepository.Create(ctx, nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "model can not be nil")
		convMock.AssertExpectations(t)
	})
}

func TestPgRepository_Update(t *testing.T) {
	updateQuery := regexp.QuoteMeta(`UPDATE public.tombstones SET removal_date = ? WHERE tenant_id = ? AND id = ?`)

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		tombstoneModel := fixTombstoneModel()
		entity := fixEntityTombstone()

		convMock := &automock.EntityConverter{}
		convMock.On("ToEntity", tombstoneModel).Return(entity, nil)
		sqlMock.ExpectExec(updateQuery).
			WithArgs(append(fixTombstoneUpdateArgs(), tenantID, entity.ID)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		pgRepository := tombstone.NewRepository(convMock)
		//WHEN
		err := pgRepository.Update(ctx, tombstoneModel)
		//THEN
		require.NoError(t, err)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("returns error when model is nil", func(t *testing.T) {
		sqlxDB, _ := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		pgRepository := tombstone.NewRepository(convMock)
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
	deleteQuery := `^DELETE FROM public.tombstones WHERE tenant_id = \$1 AND id = \$2$`

	sqlMock.ExpectExec(deleteQuery).WithArgs(tenantID, tombstoneID).WillReturnResult(sqlmock.NewResult(-1, 1))
	convMock := &automock.EntityConverter{}
	pgRepository := tombstone.NewRepository(convMock)
	//WHEN
	err := pgRepository.Delete(ctx, tenantID, tombstoneID)
	//THEN
	require.NoError(t, err)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_Exists(t *testing.T) {
	//GIVEN
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	existQuery := regexp.QuoteMeta(`SELECT 1 FROM public.tombstones WHERE tenant_id = $1 AND id = $2`)

	sqlMock.ExpectQuery(existQuery).WithArgs(tenantID, tombstoneID).WillReturnRows(testdb.RowWhenObjectExist())
	convMock := &automock.EntityConverter{}
	pgRepository := tombstone.NewRepository(convMock)
	//WHEN
	found, err := pgRepository.Exists(ctx, tenantID, tombstoneID)
	//THEN
	require.NoError(t, err)
	assert.True(t, found)
	sqlMock.AssertExpectations(t)
	convMock.AssertExpectations(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	// given
	tombstoneEntity := fixEntityTombstone()

	selectQuery := `^SELECT (.+) FROM public.tombstones WHERE tenant_id = \$1 AND id = \$2$`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixTombstoneColumns()).
			AddRow(fixTombstoneRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, tombstoneID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", tombstoneEntity).Return(&model.Tombstone{ID: tombstoneID, TenantID: tenantID}, nil).Once()
		pgRepository := tombstone.NewRepository(convMock)
		// WHEN
		modelTombstone, err := pgRepository.GetByID(ctx, tenantID, tombstoneID)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, tombstoneID, modelTombstone.ID)
		assert.Equal(t, tenantID, modelTombstone.TenantID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := tombstone.NewRepository(nil)
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, tombstoneID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		// when
		modelTombstone, err := repo.GetByID(ctx, tenantID, tombstoneID)
		// then

		sqlMock.AssertExpectations(t)
		assert.Nil(t, modelTombstone)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		testError := errors.New("test error")
		rows := sqlmock.NewRows(fixTombstoneColumns()).
			AddRow(fixTombstoneRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, tombstoneID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", tombstoneEntity).Return(&model.Tombstone{}, testError).Once()
		pgRepository := tombstone.NewRepository(convMock)
		// WHEN
		_, err := pgRepository.GetByID(ctx, tenantID, tombstoneID)
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
	firstTombstoneEntity := fixEntityTombstone()
	secondTombstoneEntity := fixEntityTombstone()

	selectQuery := `^SELECT (.+) FROM public.tombstones
		WHERE tenant_id = \$1 AND app_id = \$2`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		rows := sqlmock.NewRows(fixTombstoneColumns()).
			AddRow(fixTombstoneRow()...).
			AddRow(fixTombstoneRow()...)

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(tenantID, appID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", firstTombstoneEntity).Return(&model.Tombstone{ID: firstTombstoneEntity.ID}, nil)
		convMock.On("FromEntity", secondTombstoneEntity).Return(&model.Tombstone{ID: secondTombstoneEntity.ID}, nil)
		pgRepository := tombstone.NewRepository(convMock)
		// WHEN
		modelTombstone, err := pgRepository.ListByApplicationID(ctx, tenantID, appID)
		//THEN
		require.NoError(t, err)
		require.Len(t, modelTombstone, totalCount)
		assert.Equal(t, firstTombstoneEntity.ID, modelTombstone[0].ID)
		assert.Equal(t, secondTombstoneEntity.ID, modelTombstone[1].ID)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})
}
