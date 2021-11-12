package tombstone_test

import (
	"database/sql/driver"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestPgRepository_Create(t *testing.T) {
	//GIVEN
	var nilTsModel *model.Tombstone
	tombstoneModel := fixTombstoneModel()
	tombstoneEntity := fixEntityTombstone()
	suite := testdb.RepoCreateTestSuite{
		Name: "Create Tombstone",
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
				Query:       `^INSERT INTO public.tombstones \(.+\) VALUES \(.+\)$`,
				Args:        fixTombstoneRow(),
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tombstone.NewRepository,
		ModelEntity:               tombstoneModel,
		DBEntity:                  tombstoneEntity,
		NilModelEntity:            nilTsModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_Update(t *testing.T) {
	var nilTsModel *model.Tombstone
	entity := fixEntityTombstone()

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Tombstone",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.tombstones SET removal_date = ? WHERE id = ? AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          append(fixTombstoneUpdateArgs(), entity.ID),
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       tombstone.NewRepository,
		ModelEntity:               fixTombstoneModel(),
		DBEntity:                  entity,
		NilModelEntity:            nilTsModel,
		TenantID:                  tenantID,
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
func TestPgRepository_Delete(t *testing.T) {
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
	deleteQuery := fmt.Sprintf(`^DELETE FROM public.tombstones WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

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
	existQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT 1 FROM public.tombstones WHERE %s AND id = $2`, fixUnescapedTenantIsolationSubquery()))

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

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.tombstones WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())

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

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.tombstones WHERE %s AND app_id = \$2`, fixTenantIsolationSubquery())

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
*/
