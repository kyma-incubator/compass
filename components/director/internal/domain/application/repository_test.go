package application_test

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Exists(t *testing.T) {
	suite := testdb.RepoExistTestSuite{
		Name: "Application Exists",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT 1 FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
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
		RepoConstructorFunc: application.NewRepository,
		TargetID:            givenID(),
		TenantID:            givenTenant(),
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Application Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2 AND owner = true))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id"}).AddRow(givenID())}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"id"}).AddRow(givenID()).AddRow("secondID")}
				},
			},
			{
				Query:       regexp.QuoteMeta(`DELETE FROM tenant_applications WHERE id IN ($1)`),
				Args:        []driver.Value{givenID()},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: application.NewRepository,
		MethodArgs:          []interface{}{givenTenant(), givenID()},
		IsTopLeveEntity:     true,
	}

	suite.Run(t)

	// Additional tests - async deletion
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
		appModel.DeletedAt = &deletedAt
		entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(fixAppColumns()).
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.Labels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = '%s' AND owner = true))`, givenTenant()))

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.Labels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, givenID()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		// then
		assert.NoError(t, err)
		assert.Empty(t, appModel.Error)
		assert.False(t, appModel.Ready)
	})

	t.Run("Success when operation mode is set to async explicitly and operation is in the context, and previous error exists", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeDelete,
			OperationCategory: "unregisterApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		deletedAt := time.Now()

		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appModel.DeletedAt = &deletedAt
		appModel.Error = str.Ptr("error")
		entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(fixAppColumns()).
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.Labels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = '%s' AND owner = true))`, givenTenant()))

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.Labels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, givenID()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())
		// then
		assert.NoError(t, err)
		assert.Empty(t, appModel.Error)
		assert.False(t, appModel.Ready)
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
			WithArgs(givenID(), givenTenant()).WillReturnError(givenError())

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
		entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows(fixAppColumns()).
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.Labels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel, nil).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = '%s' AND owner = true))`, givenTenant()))

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.Labels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, givenID()).
			WillReturnError(givenError())

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_Create(t *testing.T) {
	var nilAppModel *model.Application
	appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

	suite := testdb.RepoCreateTestSuite{
		Name: "Generic Create Application",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:       regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number,  name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`),
				Args:        []driver.Value{givenID(), nil, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.Labels), appModel.Ready, appModel.CreatedAt, appModel.UpdatedAt, appModel.DeletedAt, appModel.Error, repo.NewNullableStringFromJSONRawMessage(appModel.CorrelationIDs)},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
			{
				Query:       regexp.QuoteMeta(`INSERT INTO tenant_applications ( tenant_id, id, owner ) VALUES ( ?, ?, ? )`),
				Args:        []driver.Value{givenTenant(), givenID(), true},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: application.NewRepository,
		ModelEntity:         appModel,
		DBEntity:            appEntity,
		NilModelEntity:      nilAppModel,
		TenantID:            givenTenant(),
		IsTopLevelEntity:    true,
	}

	suite.Run(t)

	// Additional tests

	t.Run("Success when operation mode is set to async explicitly and operation is in the context", func(t *testing.T) {
		ctx := context.Background()
		ctx = operation.SaveModeToContext(ctx, graphql.OperationModeAsync)

		op := &operation.Operation{
			OperationType:     operation.OperationTypeCreate,
			OperationCategory: "registerApplication",
		}
		ctx = operation.SaveToContext(ctx, &[]*operation.Operation{op})

		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		appModel.Ready = false
		appEntity.Ready = false

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number,  name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(givenID(), nil, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.Labels), appModel.Ready, appModel.CreatedAt, appModel.UpdatedAt, appModel.DeletedAt, appModel.Error, repo.NewNullableStringFromJSONRawMessage(appModel.CorrelationIDs)).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO tenant_applications ( tenant_id, id, owner ) VALUES ( ?, ?, ? )`)).
			WithArgs(givenTenant(), givenID(), true).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, givenTenant(), appModel)

		// then
		assert.NoError(t, err)
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = '%s' AND owner = true))`, givenTenant()))

	var nilAppModel *model.Application
	appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity.UpdatedAt = &fixedTimestamp
	appEntity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Application",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.Labels), appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error, appEntity.CorrelationIDs, givenID()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: application.NewRepository,
		ModelEntity:         appModel,
		DBEntity:            appEntity,
		NilModelEntity:      nilAppModel,
		TenantID:            givenTenant(),
	}

	suite.Run(t)
}

func TestRepository_GetByID(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.Labels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		ExpectedModelEntity:       fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description"),
		ExpectedDBEntity:          entity,
		MethodArgs:                []interface{}{givenTenant(), givenID()},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

/*
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

	pageableQuery := `^SELECT (.+) FROM public\.applications WHERE %s ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := fmt.Sprintf(`SELECT COUNT\(\*\) FROM public\.applications WHERE %s`, fixTenantIsolationSubquery())

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.Labels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.Labels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, fixTenantIsolationSubquery(), inputPageSize, 0)).
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

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, fixTenantIsolationSubquery(), inputPageSize, 0)).
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
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_ListGlobal(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	inputPageSize := 3
	inputCursor := ""
	totalCount := 2

	pageableQuery := `^SELECT (.+) FROM public\.applications ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := `SELECT COUNT\(\*\) FROM public\.applications`

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.Labels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.Labels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs)

		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, inputPageSize, 0)).
			WithArgs().
			WillReturnRows(rows)

		sqlMock.ExpectQuery(countQuery).
			WithArgs().
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

		conv := &automock.EntityConverter{}
		conv.On("FromEntity", appEntity2).Return(appModel2).Once()
		conv.On("FromEntity", appEntity1).Return(appModel1).Once()
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		modelApp, err := pgRepository.ListGlobal(ctx, inputPageSize, inputCursor)

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
			WithArgs().
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.EntityConverter{}
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// when
		_, err := pgRepository.ListGlobal(ctx, inputPageSize, inputCursor)

		//then
		require.Error(t, err)
		require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestPgRepository_ListAll(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	listQuery := fmt.Sprintf(`^SELECT (.+) FROM public\.applications WHERE %s`, fixTenantIsolationSubquery())

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.Labels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.Labels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs)

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
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
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
	scenariosQuery := fmt.Sprintf(`SELECT "app_id" FROM public.labels
					WHERE "app_id" IS NOT NULL AND %s
						AND "key" = $3 AND "value" ?| array[$4]
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND %s
						AND "key" = $6 AND "value" ?| array[$7]
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND %s
						AND "key" = $9 AND "value" ?| array[$10]`, fixUnescapedTenantIsolationSubqueryWithArg(2), fixUnescapedTenantIsolationSubqueryWithArg(5), fixUnescapedTenantIsolationSubqueryWithArg(8))
	applicationScenarioQuery := regexp.QuoteMeta(scenariosQuery)

	applicationScenarioQueryWithHidingSelectors := regexp.QuoteMeta(
		fmt.Sprintf(`%s EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND %s AND "key" = $12 AND "value" @> $13 EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND %s AND "key" = $15 AND "value" @> $16`, scenariosQuery, fixUnescapedTenantIsolationSubqueryWithArg(11), fixUnescapedTenantIsolationSubqueryWithArg(14)),
	)

	pageableQueryRegex := `SELECT (.+) FROM public\.applications WHERE %s AND id IN \(%s\) ORDER BY id LIMIT %d OFFSET %d`
	pageableQuery := fmt.Sprintf(pageableQueryRegex, fixTenantIsolationSubquery(), applicationScenarioQuery, pageSize, 0)

	pageableQueryWithHidingSelectors := fmt.Sprintf(pageableQueryRegex, fixTenantIsolationSubquery(), applicationScenarioQueryWithHidingSelectors, pageSize, 0)

	countQueryRegex := `SELECT COUNT\(\*\) FROM public\.applications WHERE %s AND id IN \(%s\)$`
	countQuery := fmt.Sprintf(countQueryRegex, fixTenantIsolationSubquery(), applicationScenarioQuery)
	countQueryWithHidingSelectors := fmt.Sprintf(countQueryRegex, fixTenantIsolationSubquery(), applicationScenarioQueryWithHidingSelectors)

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
*/

func TestPgRepository_GetByNameAndSystemNumber(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), appName, "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application By Name and System Number",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids FROM public.applications WHERE name = $1 AND system_number = $2 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $3))`),
				Args:     []driver.Value{appName, systemNumber, givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.Labels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		ExpectedModelEntity:       fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description"),
		ExpectedDBEntity:          entity,
		MethodName:                "GetByNameAndSystemNumber",
		MethodArgs:                []interface{}{givenTenant(), appName, systemNumber},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func givenError() error {
	return errors.New("some error")
}
