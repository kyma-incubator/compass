package systemssync_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_List(t *testing.T) {
	systemsSyncEntity := fixSystemsSyncEntity(syncID, syncTenantID, syncProductID, lastSyncTime)
	systemsSyncModel := fixSystemsSyncModel(syncID, syncTenantID, syncProductID, lastSyncTime)

	suite := testdb.RepoListTestSuite{
		Name: "List all systems synchronization timestamps",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, product_id, last_sync_timestamp FROM public.systems_sync_timestamps`),
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixSystemsSyncTimestampsColumns()).AddRow(systemsSyncEntity.ID, systemsSyncEntity.TenantID, systemsSyncEntity.ProductID, systemsSyncEntity.LastSyncTimestamp),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixSystemsSyncTimestampsColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       systemssync.NewRepository,
		ExpectedDBEntities:        []interface{}{systemsSyncEntity},
		ExpectedModelEntities:     []interface{}{systemsSyncModel},
		MethodName:                "List",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestRepository_Upsert(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		systemsSyncEntity := fixSystemsSyncEntity(syncID, syncTenantID, syncProductID, lastSyncTime)
		systemsSyncModel := fixSystemsSyncModel(syncID, syncTenantID, syncProductID, lastSyncTime)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		mockConverter.On("ToEntity", systemsSyncModel).Return(systemsSyncEntity).Once()

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.systems_sync_timestamps ( id, tenant_id, product_id, last_sync_timestamp ) VALUES ( ?, ?, ?, ? ) ON CONFLICT ( id ) DO UPDATE SET tenant_id=EXCLUDED.tenant_id, product_id=EXCLUDED.product_id, last_sync_timestamp=EXCLUDED.last_sync_timestamp`)).
			WithArgs(fixSystemsSyncCreateArgs(*systemsSyncEntity)...).
			WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		systemsSyncRepo := systemssync.NewRepository(mockConverter)

		// WHEN
		err := systemsSyncRepo.Upsert(ctx, systemsSyncModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when upserting system sync timestamp", func(t *testing.T) {
		// GIVEN
		systemsSyncEntity := fixSystemsSyncEntity(syncID, syncTenantID, syncProductID, lastSyncTime)
		systemsSyncModel := fixSystemsSyncModel(syncID, syncTenantID, syncProductID, lastSyncTime)

		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		mockConverter.On("ToEntity", systemsSyncModel).Return(systemsSyncEntity).Once()

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.systems_sync_timestamps ( id, tenant_id, product_id, last_sync_timestamp ) VALUES ( ?, ?, ?, ? ) ON CONFLICT ( id ) DO UPDATE SET tenant_id=EXCLUDED.tenant_id, product_id=EXCLUDED.product_id, last_sync_timestamp=EXCLUDED.last_sync_timestamp`)).
			WithArgs(fixSystemsSyncCreateArgs(*systemsSyncEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		systemsSyncRepo := systemssync.NewRepository(mockConverter)

		// WHEN
		err := systemsSyncRepo.Upsert(ctx, systemsSyncModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
