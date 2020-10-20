package integrationsystem_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPgRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelIntegrationSystem(testID, testName)
		intSysEntity := fixEntityIntegrationSystem(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", intSysModel).Return(intSysEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.integration_systems ( id, name, description ) VALUES ( ?, ?, ? )`)).
			WithArgs(fixIntegrationSystemCreateArgs(*intSysEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		err := intSysRepo.Create(ctx, *intSysModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when creating", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelIntegrationSystem(testID, testName)
		intSysEntity := fixEntityIntegrationSystem(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", intSysModel).Return(intSysEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.integration_systems ( id, name, description ) VALUES ( ?, ?, ? )`)).
			WithArgs(fixIntegrationSystemCreateArgs(*intSysEntity)...).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		err := intSysRepo.Create(ctx, *intSysModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_Get(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelIntegrationSystem(testID, testName)
		intSysEntity := fixEntityIntegrationSystem(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", intSysEntity).Return(intSysModel).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: testID, name: testName, description: &testDescription},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(rowsToReturn)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		result, err := intSysRepo.Get(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, intSysModel, result)
	})

	t.Run("Error when getting", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		result, err := intSysRepo.Get(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
	})
}

func TestPgRepository_Exists(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnRows(testdb.RowWhenObjectExist())

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(nil)

		// WHEN
		result, err := intSysRepo.Exists(ctx, testID)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result)
	})

	t.Run("Error when checking existence", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT 1 FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(nil)

		// WHEN
		result, err := intSysRepo.Exists(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
		assert.False(t, result)
	})
}

func TestPgRepository_List(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		intSysModels := []*model.IntegrationSystem{
			fixModelIntegrationSystem("id1", "name1"),
			fixModelIntegrationSystem("id2", "name2"),
			fixModelIntegrationSystem("id3", "name3"),
		}

		intSysEntities := []*integrationsystem.Entity{
			fixEntityIntegrationSystem("id1", "name1"),
			fixEntityIntegrationSystem("id2", "name2"),
			fixEntityIntegrationSystem("id3", "name3"),
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", intSysEntities[0]).Return(intSysModels[0]).Once()
		mockConverter.On("FromEntity", intSysEntities[1]).Return(intSysModels[1]).Once()
		mockConverter.On("FromEntity", intSysEntities[2]).Return(intSysModels[2]).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		rowsToReturn := fixSQLRows([]sqlRow{
			{id: "id1", name: "name1", description: &testDescription},
			{id: "id2", name: "name2", description: &testDescription},
			{id: "id3", name: "name3", description: &testDescription},
		})
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description FROM public.integration_systems ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnRows(rowsToReturn)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM public.integration_systems`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		result, err := intSysRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, intSysModels, result.Data)
	})

	t.Run("Error when listing", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, name, description FROM public.integration_systems ORDER BY id LIMIT 3 OFFSET 0`)).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		result, err := intSysRepo.List(ctx, testPageSize, testCursor)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, result.Data)
	})
}

func TestPgRepository_Update(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelIntegrationSystem(testID, testName)
		intSysEntity := fixEntityIntegrationSystem(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", intSysModel).Return(intSysEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.integration_systems SET name = ?, description = ? WHERE id = ?`)).
			WithArgs(testName, testDescription, testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		err := intSysRepo.Update(ctx, *intSysModel)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when updating", func(t *testing.T) {
		// GIVEN
		intSysModel := fixModelIntegrationSystem(testID, testName)
		intSysEntity := fixEntityIntegrationSystem(testID, testName)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", intSysModel).Return(intSysEntity).Once()
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.integration_systems SET name = ?, description = ? WHERE id = ?`)).
			WithArgs(testName, testDescription, testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(mockConverter)

		// WHEN
		err := intSysRepo.Update(ctx, *intSysModel)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(nil)

		// WHEN
		err := intSysRepo.Delete(ctx, testID)

		// THEN
		require.NoError(t, err)
	})

	t.Run("Error when deleting", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)
		dbMock.ExpectExec(regexp.QuoteMeta(`DELETE FROM public.integration_systems WHERE id = $1`)).
			WithArgs(testID).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		intSysRepo := integrationsystem.NewRepository(nil)

		// WHEN
		err := intSysRepo.Delete(ctx, testID)

		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
