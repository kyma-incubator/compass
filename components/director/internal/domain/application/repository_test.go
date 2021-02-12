package application_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Exists(t *testing.T) {
	// given
	sqlxDB, sqlMock := testdb.MockDatabase(t)
	defer sqlMock.AssertExpectations(t)

	sqlMock.ExpectQuery(regexp.QuoteMeta("SELECT 1 FROM public.applications WHERE tenant_id = $1 AND id = $2")).WithArgs(
		givenTenant(), givenID()).
		WillReturnRows(testdb.RowWhenObjectExist())

	ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

	repo := application.NewRepository(nil)

	// when
	ex, err := repo.Exists(ctx, givenTenant(), givenID())

	// then
	require.NoError(t, err)
	assert.True(t, ex)
}

func TestRepository_Delete(t *testing.T) {
	var executeDeleteFunc = func(ctx context.Context) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.applications WHERE tenant_id = $1 AND id = $2")).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)
		repo := application.NewRepository(nil)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		// then
		require.NoError(t, err)
	}

	t.Run("Success", func(t *testing.T) {
		executeDeleteFunc(context.Background())
	})

	t.Run("Success when operation mode is set to sync explicitly and no operation in the context", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeSync)

		executeDeleteFunc(ctx)
	})

	t.Run("Success when operation mode is set to async explicitly and operation is in the context", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeDelete,
			OperationCategory: "unregisterApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		deletedAt := time.Now()

		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appModel.Ready = false
		appModel.DeletedAt = &deletedAt
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels",
			"ready", "created_at", "updated_at", "deleted_at", "error"}).
			AddRow(givenID(), givenTenant(), appEntity.Name, appEntity.Description, appEntity.StatusCondition, appEntity.StatusTimestamp, appEntity.HealthCheckURL, appEntity.IntegrationSystemID, appEntity.ProviderName,
				appEntity.BaseURL, appEntity.Labels, appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.applications WHERE tenant_id = \$1 AND id = \$2$`).
			WithArgs(givenTenant(), givenID()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", appEntity).Return(appModel).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := `UPDATE public\.applications SET name = \?, description = \?, status_condition = \?, status_timestamp = \?, healthcheck_url = \?, integration_system_id = \?, provider_name = \?,  base_url = \?, labels = \?, ready = \?, created_at = \?, updated_at = \?, deleted_at = \?, error = \? WHERE tenant_id = \? AND id = \?`

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.Labels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, givenTenant(), givenID()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		assert.NoError(t, err)
	})

	t.Run("Failure when operation mode is set to async, operation is in the context but fetch application fails", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeDelete,
			OperationCategory: "unregisterApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(nil)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Failure when operation mode is set to async, operation is in the context but update application fails", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeDelete,
			OperationCategory: "unregisterApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		deletedAt := time.Now()

		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appModel.Ready = false
		appModel.DeletedAt = &deletedAt
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}).
			AddRow(givenID(), givenTenant(), appEntity.Name, appEntity.Description, appEntity.StatusCondition, appEntity.StatusTimestamp, appEntity.HealthCheckURL, appEntity.IntegrationSystemID, appEntity.ProviderName, appEntity.BaseURL, appEntity.Labels, appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.applications WHERE tenant_id = \$1 AND id = \$2$`).
			WithArgs(givenTenant(), givenID()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", appEntity).Return(appModel, nil).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := `UPDATE public\.applications SET name = \?, description = \?, status_condition = \?, status_timestamp = \?, healthcheck_url = \?, integration_system_id = \?, provider_name = \?,  base_url = \?, labels = \?, ready = \?, created_at = \?, updated_at = \?, deleted_at = \?, error = \? WHERE tenant_id = \? AND id = \?`

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.Labels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, givenTenant(), givenID()).
			WillReturnError(givenError())

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repository := application.NewRepository(nil)

		// when
		err := repository.Delete(ctx, givenTenant(), givenID())

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_Create(t *testing.T) {
	var executeCreateFunc = func(ctx context.Context, appID string, mode graphql.OperationMode, operationError error) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		if mode == graphql.OperationModeAsync {
			appModel.Ready = false
			appEntity.Ready = false
		}

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		if operationError == nil {
			db, dbMock := testdb.MockDatabase(t)
			defer dbMock.AssertExpectations(t)

			dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.applications ( id, tenant_id, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
				WithArgs(givenID(), givenTenant(), appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.Labels), appModel.Ready, appModel.CreatedAt, appModel.UpdatedAt, appModel.DeletedAt, appModel.Error).
				WillReturnResult(sqlmock.NewResult(-1, 1))

			ctx = persistence.SaveToContext(ctx, db)
		}

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, appModel)

		// then
		if operationError != nil {
			assert.Error(t, err)
			assert.Contains(t, err.Error(), operationError.Error())
		} else {
			assert.NoError(t, err)
		}
	}

	t.Run("Success", func(t *testing.T) {
		executeCreateFunc(context.Background(), givenID(), graphql.OperationModeSync, nil)
	})

	t.Run("Success when operation mode is set to sync explicitly and no operation in the context", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeSync)

		executeCreateFunc(ctx, givenID(), graphql.OperationModeSync, nil)
	})

	t.Run("Success when operation mode is set to async explicitly and operation is in the context", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeCreate,
			OperationCategory: "registerApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		appID := givenID()
		executeCreateFunc(ctx, appID, graphql.OperationModeAsync, nil)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repository := application.NewRepository(mockConverter)

		// when
		err := repository.Create(ctx, appModel)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(&application.Entity{}, givenError())
		defer mockConverter.AssertExpectations(t)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Create(context.TODO(), appModel)

		// then
		require.EqualError(t, err, "while converting to Application entity: some error")
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := `UPDATE public\.applications SET name = \?, description = \?, status_condition = \?, status_timestamp = \?, healthcheck_url = \?, integration_system_id = \?, provider_name = \?, base_url = \?, labels = \?, ready = \?, created_at = \?, updated_at = \?, deleted_at = \?, error = \? WHERE tenant_id = \? AND id = \?`

	t.Run("Success", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity.UpdatedAt = &fixedTimestamp
		appEntity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WithArgs(appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.Labels), appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error, givenTenant(), givenID()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, appModel)

		// then
		assert.NoError(t, err)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repository := application.NewRepository(mockConverter)

		// when
		err := repository.Update(ctx, appModel)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(&application.Entity{}, givenError())
		defer mockConverter.AssertExpectations(t)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Update(context.TODO(), appModel)

		// then
		require.EqualError(t, err, "while converting to Application entity: some error")
	})
}

func TestRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", appEntity).Return(appModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		repo := application.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}).
			AddRow(givenID(), givenTenant(), appEntity.Name, appEntity.Description, appEntity.StatusCondition, appEntity.StatusTimestamp, appEntity.HealthCheckURL, appEntity.IntegrationSystemID, appEntity.ProviderName, appEntity.BaseURL, appEntity.Labels, appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error)

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.applications WHERE tenant_id = \$1 AND id = \$2$`).
			WithArgs(givenTenant(), givenID()).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetByID(ctx, givenTenant(), givenID())

		// then
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, appModel, actual)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repository := application.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repository.GetByID(ctx, givenTenant(), givenID())

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_List(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2

	pageableQuery := `^SELECT (.+) FROM public\.applications WHERE tenant_id = \$1 ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := `SELECT COUNT\(\*\) FROM public\.applications WHERE tenant_id = \$1`

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.Labels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.Labels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, inputPageSize, 0)).
			WithArgs(givenTenant()).
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs(givenTenant()).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		conv := &automock.EntityConverter{}
		conv.On("FromEntity", appEntity2).Return(appModel2).Once()
		conv.On("FromEntity", appEntity1).Return(appModel1).Once()
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		modelApp, err := pgRepository.List(ctx, givenTenant(), nil, inputPageSize, inputCursor)

		// then
		require.NoError(t, err)
		require.Len(t, modelApp.Data, 2)
		assert.Equal(t, appEntity1.ID, modelApp.Data[0].ID)
		assert.Equal(t, appEntity2.ID, modelApp.Data[1].ID)
		assert.Equal(t, "", modelApp.PageInfo.StartCursor)
		assert.Equal(t, totalCount, modelApp.TotalCount)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, inputPageSize, 0)).
			WithArgs(givenTenant()).
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.EntityConverter{}
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		_, err := pgRepository.List(ctx, givenTenant(), nil, inputPageSize, inputCursor)

		//then
		require.Error(t, err)
		require.Contains(t, err.Error(), "while fetching list of objects from DB: some error")
	})
}

func TestPgRepository_ListAll(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	listQuery := `^SELECT (.+) FROM public\.applications WHERE tenant_id = \$1`

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.Labels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.Labels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(listQuery).
			WithArgs(givenTenant()).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		conv := &automock.EntityConverter{}
		conv.On("FromEntity", appEntity2).Return(appModel2).Once()
		conv.On("FromEntity", appEntity1).Return(appModel1).Once()
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		modelApp, err := pgRepository.ListAll(ctx, givenTenant())

		// then
		require.NoError(t, err)
		require.Len(t, modelApp, 2)
		assert.Equal(t, appEntity1.ID, modelApp[0].ID)
		assert.Equal(t, appEntity2.ID, modelApp[1].ID)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(listQuery).
			WithArgs(givenTenant()).
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.EntityConverter{}
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		_, err := pgRepository.ListAll(ctx, givenTenant())

		//then
		require.Error(t, err)
		require.Contains(t, err.Error(), "error while executing SQL query")
	})
}

func TestPgRepository_ListByRuntimeScenarios(t *testing.T) {
	tenantID := uuid.New()
	app1ID := uuid.New()
	app2ID := uuid.New()
	pageSize := 5
	cursor := ""
	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)

	runtimeScenarios := []string{"Java", "Go", "Elixir"}
	scenariosKey := "scenarios"
	scenariosQuery := `SELECT "app_id" FROM public.labels
					WHERE "app_id" IS NOT NULL AND "tenant_id" = $2
						AND "key" = $3 AND "value" ?| array[$4]
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND "tenant_id" = $5
						AND "key" = $6 AND "value" ?| array[$7]
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND "tenant_id" = $8
						AND "key" = $9 AND "value" ?| array[$10]`
	applicationScenarioQuery := regexp.QuoteMeta(scenariosQuery)

	applicationScenarioQueryWithHidingSelectors := regexp.QuoteMeta(
		fmt.Sprintf(`%s EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND "tenant_id" = $11 AND "key" = $12 AND "value" @> $13 EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND "tenant_id" = $14 AND "key" = $15 AND "value" @> $16`, scenariosQuery),
	)

	pageableQueryRegex := `SELECT (.+) FROM public\.applications WHERE tenant_id = \$1 AND id IN \(%s\) ORDER BY id LIMIT %d OFFSET %d`
	pageableQuery := fmt.Sprintf(pageableQueryRegex,
		applicationScenarioQuery,
		pageSize,
		0)

	pageableQueryWithHidingSelectors := fmt.Sprintf(pageableQueryRegex,
		applicationScenarioQueryWithHidingSelectors,
		pageSize,
		0)

	countQueryRegex := `SELECT COUNT\(\*\) FROM public\.applications WHERE tenant_id = \$1 AND id IN \(%s\)$`
	countQuery := fmt.Sprintf(countQueryRegex, applicationScenarioQuery)
	countQueryWithHidingSelectors := fmt.Sprintf(countQueryRegex, applicationScenarioQueryWithHidingSelectors)

	conv := application.NewConverter(nil, nil)
	intSysID := repo.NewValidNullableString("iiiiiiiii-iiii-iiii-iiii-iiiiiiiiiiii")

	testCases := []struct {
		Name                     string
		InputHidingSelectors     map[string][]string
		ExpectedPageableQuery    string
		ExpectedCountQuery       string
		ExpectedQueriesInputArgs []driver.Value
		ExpectedApplicationRows  *sqlmock.Rows
		TotalCount               int
		ExpectedError            error
	}{
		{
			Name:                     "Success",
			InputHidingSelectors:     nil,
			ExpectedPageableQuery:    pageableQuery,
			ExpectedCountQuery:       countQuery,
			ExpectedQueriesInputArgs: []driver.Value{tenantID, tenantID, scenariosKey, "Java", tenantID, scenariosKey, "Go", tenantID, scenariosKey, "Elixir"},
			ExpectedApplicationRows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "ready", "created_at", "updated_at", "deleted_at", "error"}).
				AddRow(app1ID, tenantID, "App ABC", "Description for application ABC", "INITIAL", timestamp, "http://domain.local/app1", intSysID, true, fixedTimestamp, fixedTimestamp, time.Time{}, nil).
				AddRow(app2ID, tenantID, "App XYZ", "Description for application XYZ", "INITIAL", timestamp, "http://domain.local/app2", intSysID, true, fixedTimestamp, fixedTimestamp, time.Time{}, nil),
			TotalCount:    2,
			ExpectedError: nil,
		},
		{
			Name: "Success with hiding selectors",
			InputHidingSelectors: map[string][]string{
				"foo": {"bar", "baz"},
			},
			ExpectedPageableQuery:    pageableQueryWithHidingSelectors,
			ExpectedCountQuery:       countQueryWithHidingSelectors,
			ExpectedQueriesInputArgs: []driver.Value{tenantID, tenantID, scenariosKey, "Java", tenantID, scenariosKey, "Go", tenantID, scenariosKey, "Elixir", tenantID, "foo", strconv.Quote("bar"), tenantID, "foo", strconv.Quote("baz")},
			ExpectedApplicationRows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "ready", "created_at", "updated_at", "deleted_at", "error"}).
				AddRow(app1ID, tenantID, "App ABC", "Description for application ABC", "INITIAL", timestamp, "http://domain.local/app1", intSysID, true, fixedTimestamp, fixedTimestamp, time.Time{}, nil).
				AddRow(app2ID, tenantID, "App XYZ", "Description for application XYZ", "INITIAL", timestamp, "http://domain.local/app2", intSysID, true, fixedTimestamp, fixedTimestamp, time.Time{}, nil),
			TotalCount:    2,
			ExpectedError: nil,
		},
		{
			Name:                     "Return empty page when no application match",
			InputHidingSelectors:     nil,
			ExpectedPageableQuery:    pageableQuery,
			ExpectedCountQuery:       countQuery,
			ExpectedQueriesInputArgs: []driver.Value{tenantID, tenantID, scenariosKey, "Java", tenantID, scenariosKey, "Go", tenantID, scenariosKey, "Elixir"},
			ExpectedApplicationRows:  sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "ready", "created_at", "updated_at", "deleted_at", "error"}),
			TotalCount:               0,
			ExpectedError:            nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sqlxDB, sqlMock := testdb.MockDatabase(t)
			if testCase.ExpectedApplicationRows != nil {
				sqlMock.ExpectQuery(testCase.ExpectedPageableQuery).
					WithArgs(testCase.ExpectedQueriesInputArgs...).
					WillReturnRows(testCase.ExpectedApplicationRows)

				countRow := sqlMock.NewRows([]string{"count"}).AddRow(testCase.TotalCount)
				sqlMock.ExpectQuery(testCase.ExpectedCountQuery).
					WithArgs(testCase.ExpectedQueriesInputArgs...).
					WillReturnRows(countRow)
			}
			repository := application.NewRepository(conv)

			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			//WHEN
			page, err := repository.ListByScenarios(ctx, tenantID, runtimeScenarios, pageSize, cursor, testCase.InputHidingSelectors)

			//THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
				assert.NotNil(t, page)
			}
			assert.NoError(t, sqlMock.ExpectationsWereMet())
		})
	}
}

func givenError() error {
	return errors.New("some error")
}
