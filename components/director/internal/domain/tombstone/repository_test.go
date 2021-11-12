package tombstone_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"

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

*/

func TestPgRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Tombstone Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{tombstoneID, tenantID},
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
		RepoConstructorFunc: tombstone.NewRepository,
		TargetID:            tombstoneID,
		TenantID:            tenantID,
	}

	suite.Run(t)
}

func TestPgRepository_GetByID(t *testing.T) {
	suite := testdb.RepoGetTestSuite{
		Name: "Get Tombstone",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT ord_id, app_id, removal_date, id FROM public.tombstones WHERE id = $1 AND (id IN (SELECT id FROM tombstones_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{tombstoneID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns()).AddRow(fixTombstoneRow()...)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixTombstoneColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: tombstone.NewRepository,
		ExpectedModelEntity: fixTombstoneModel(),
		ExpectedDBEntity:    fixEntityTombstone(),
		MethodArgs:          []interface{}{tenantID, tombstoneID},
	}

	suite.Run(t)
}

/*
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
