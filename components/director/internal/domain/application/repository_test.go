package application_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"

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
		SQLQueryDetails: []testdb.SQLQueryDetails{
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
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenID(), givenTenant()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: application.NewRepository,
		MethodArgs:          []interface{}{givenTenant(), givenID()},
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
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, system_status = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, documentation_labels = ?, system_number = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = ? AND owner = true))`)

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.SystemStatus, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.OrdLabels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, appEntityWithDeletedTimestamp.DocumentationLabels, appEntityWithDeletedTimestamp.SystemNumber, givenID(), givenTenant()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// WHEN
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
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, system_status = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, documentation_labels = ?, system_number = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = ? AND owner = true))`)

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.SystemStatus, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.OrdLabels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, appEntityWithDeletedTimestamp.DocumentationLabels, appEntityWithDeletedTimestamp.SystemNumber, givenID(), givenTenant()).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// WHEN
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

		// WHEN
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
			AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels)

		dbMock.ExpectQuery(regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`)).
			WithArgs(givenID(), givenTenant()).
			WillReturnRows(rows)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", entity).Return(appModel, nil).Once()

		appEntityWithDeletedTimestamp := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntityWithDeletedTimestamp.Ready = false
		appEntityWithDeletedTimestamp.DeletedAt = &deletedAt
		mockConverter.On("ToEntity", appModel).Return(appEntityWithDeletedTimestamp, nil).Once()
		defer mockConverter.AssertExpectations(t)

		updateStmt := regexp.QuoteMeta(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, system_status = ?, healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, documentation_labels = ?, system_number = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = ? AND owner = true))`)

		dbMock.ExpectExec(updateStmt).
			WithArgs(appEntityWithDeletedTimestamp.Name, appEntityWithDeletedTimestamp.Description, appEntityWithDeletedTimestamp.StatusCondition, appEntityWithDeletedTimestamp.StatusTimestamp, appEntityWithDeletedTimestamp.SystemStatus, appEntityWithDeletedTimestamp.HealthCheckURL, appEntityWithDeletedTimestamp.IntegrationSystemID, appEntityWithDeletedTimestamp.ProviderName, appEntityWithDeletedTimestamp.BaseURL, appEntityWithDeletedTimestamp.OrdLabels, appEntityWithDeletedTimestamp.Ready, appEntityWithDeletedTimestamp.CreatedAt, appEntityWithDeletedTimestamp.UpdatedAt, appEntityWithDeletedTimestamp.DeletedAt, appEntityWithDeletedTimestamp.Error, appEntityWithDeletedTimestamp.CorrelationIDs, appEntityWithDeletedTimestamp.DocumentationLabels, appEntityWithDeletedTimestamp.SystemNumber, givenID(), givenTenant()).
			WillReturnError(givenError())

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// WHEN
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
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:       regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number,  name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`),
				Args:        []driver.Value{givenID(), nil, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.SystemStatus, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.OrdLabels), appModel.Ready, appModel.CreatedAt, appModel.UpdatedAt, appModel.DeletedAt, appModel.Error, repo.NewNullableStringFromJSONRawMessage(appModel.CorrelationIDs), repo.NewNullableStringFromJSONRawMessage(appModel.DocumentationLabels)},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
			{
				Query:       regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO tenant_applications ( tenant_id, id, owner ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)`),
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

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number,  name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(givenID(), nil, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.SystemStatus, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.OrdLabels), appModel.Ready, appModel.CreatedAt, appModel.UpdatedAt, appModel.DeletedAt, appModel.Error, repo.NewNullableStringFromJSONRawMessage(appModel.CorrelationIDs), repo.NewNullableStringFromJSONRawMessage(appModel.DocumentationLabels)).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		dbMock.ExpectExec(regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO tenant_applications ( tenant_id, id, owner ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)`)).
			WithArgs(givenTenant(), givenID(), true).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx = persistence.SaveToContext(ctx, db)

		repo := application.NewRepository(mockConverter)

		// WHEN
		err := repo.Create(ctx, givenTenant(), appModel)

		// then
		assert.NoError(t, err)
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.applications SET name = ?, description = ?, status_condition = ?, status_timestamp = ?, system_status = ?,  healthcheck_url = ?, integration_system_id = ?, provider_name = ?, base_url = ?, labels = ?, ready = ?, created_at = ?, updated_at = ?, deleted_at = ?, error = ?, correlation_ids = ?, documentation_labels = ?, system_number = ? WHERE id = ? AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = ? AND owner = true))`)

	var nilAppModel *model.Application
	appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity.UpdatedAt = &fixedTimestamp
	appEntity.DeletedAt = &fixedTimestamp // This is needed as workaround so that updatedAt timestamp is not updated

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.SystemStatus, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.OrdLabels), appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error, appEntity.CorrelationIDs, repo.NewNullableStringFromJSONRawMessage(appModel.DocumentationLabels), appEntity.SystemNumber, givenID(), givenTenant()},
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

func TestRepository_Upsert(t *testing.T) {
	upsertStmt := regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels ) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20 ) ON CONFLICT ( system_number ) DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description, status_condition=EXCLUDED.status_condition, system_status=EXCLUDED.system_status, provider_name=EXCLUDED.provider_name, base_url=EXCLUDED.base_url, labels=EXCLUDED.labels WHERE (public.applications.id IN (SELECT id FROM tenant_applications WHERE tenant_id = $21 AND owner = true)) RETURNING id;`)

	var nilAppModel *model.Application
	appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

	suite := testdb.RepoUpsertTestSuite{
		Name: "Upsert Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: upsertStmt,
				Args:  []driver.Value{givenID(), sql.NullString{}, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.SystemStatus, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.OrdLabels), appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error, appEntity.CorrelationIDs, appEntity.DocumentationLabels, givenTenant()},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows([]string{"id"}).
							AddRow(givenID()),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows([]string{"id"}),
					}
				},
				IsSelect: true,
			},
			{
				Query:       regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO tenant_applications ( tenant_id, id, owner ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)`),
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
	}

	suite.Run(t)
}

func TestRepository_TrustedUpsert(t *testing.T) {
	upsertStmt := regexp.QuoteMeta(`INSERT INTO public.applications ( id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels ) VALUES ( $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20 ) ON CONFLICT ( system_number ) DO UPDATE SET name=EXCLUDED.name, description=EXCLUDED.description, status_condition=EXCLUDED.status_condition, system_status=EXCLUDED.system_status, provider_name=EXCLUDED.provider_name, base_url=EXCLUDED.base_url, labels=EXCLUDED.labels RETURNING id;`)

	var nilAppModel *model.Application
	appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

	suite := testdb.RepoUpsertTestSuite{
		Name: "Trusted Upsert Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: upsertStmt,
				Args:  []driver.Value{givenID(), sql.NullString{}, appModel.SystemNumber, appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.SystemStatus, appModel.HealthCheckURL, appModel.IntegrationSystemID, appModel.ProviderName, appModel.BaseURL, repo.NewNullableStringFromJSONRawMessage(appModel.OrdLabels), appEntity.Ready, appEntity.CreatedAt, appEntity.UpdatedAt, appEntity.DeletedAt, appEntity.Error, appEntity.CorrelationIDs, appEntity.DocumentationLabels},
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows([]string{"id"}).
							AddRow(givenID()),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows([]string{"id"}),
					}
				},
				IsSelect: true,
			},
			{
				Query:       regexp.QuoteMeta(`WITH RECURSIVE parents AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN parents t on t2.id = t.parent) INSERT INTO tenant_applications ( tenant_id, id, owner ) (SELECT parents.id AS tenant_id, ? as id, ? AS owner FROM parents)`),
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
		UpsertMethodName:    "TrustedUpsert",
	}

	suite.Run(t)
}

func TestRepository_GetByID(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels),
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

func TestRepository_GetGlobalByIDForUpdate(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id = $1 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $2)) FOR UPDATE`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition,
								entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL,
								entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels),
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
		MethodName:                "GetByIDForUpdate",
	}

	suite.Run(t)
}

func TestPgRepository_List(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Applications",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query: regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications
												WHERE (id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3])
												AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $4))) ORDER BY id LIMIT 2 OFFSET 0`),
				Args:     []driver.Value{givenTenant(), model.ScenariosKey, "scenario", givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns()).
						AddRow(appEntity1.ID, appEntity1.ApplicationTemplateID, appEntity1.SystemNumber, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.SystemStatus, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.OrdLabels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs, appEntity1.DocumentationLabels).
						AddRow(appEntity2.ID, appEntity2.ApplicationTemplateID, appEntity2.SystemNumber, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.SystemStatus, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.OrdLabels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs, appEntity2.DocumentationLabels),
					}
				},
			},
			{
				Query: regexp.QuoteMeta(`SELECT COUNT(*) FROM public.applications
												WHERE (id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3])
												AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $4)))`),
				Args:     []driver.Value{givenTenant(), model.ScenariosKey, "scenario", givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{appModel1, appModel2},
				ExpectedDBEntities:    []interface{}{appEntity1, appEntity2},
				ExpectedPage: &model.ApplicationPage{
					Data: []*model.Application{appModel1, appModel2},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		MethodArgs:                []interface{}{givenTenant(), []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(model.ScenariosKey, `$[*] ? ( @ == "scenario" )`)}, 2, ""},
		MethodName:                "List",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
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

	pageableQuery := `SELECT (.+) FROM public\.applications ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := `SELECT COUNT\(\*\) FROM public\.applications`

	t.Run("Success", func(t *testing.T) {
		// GIVEN
		rows := sqlmock.NewRows([]string{"id", "name", "system_number", "description", "status_condition", "status_timestamp", "system_status", "healthcheck_url", "integration_system_id", "provider_name", "base_url", "labels", "ready", "created_at", "updated_at", "deleted_at", "error", "correlation_ids", "documentation_labels"}).
			AddRow(appEntity1.ID, appEntity1.Name, appEntity1.SystemNumber, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.SystemStatus, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.OrdLabels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs, appEntity1.DocumentationLabels).
			AddRow(appEntity2.ID, appEntity2.Name, appEntity2.SystemNumber, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.SystemStatus, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.OrdLabels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs, appEntity2.DocumentationLabels)

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

		// WHEN
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
		// GIVEN
		sqlxDB, sqlMock := testdb.MockDatabase(t)
		defer sqlMock.AssertExpectations(t)

		sqlMock.ExpectQuery(fmt.Sprintf(pageableQuery, inputPageSize, 0)).
			WithArgs().
			WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		conv := &automock.EntityConverter{}
		defer conv.AssertExpectations(t)

		pgRepository := application.NewRepository(conv)

		// WHEN
		_, err := pgRepository.ListGlobal(ctx, inputPageSize, inputCursor)

		// THEN
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

	suite := testdb.RepoListTestSuite{
		Name: "List Applications",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $1))`),
				Args:     []driver.Value{givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns()).
						AddRow(appEntity1.ID, appEntity1.ApplicationTemplateID, appEntity1.SystemNumber, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.SystemStatus, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.OrdLabels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs, appEntity1.DocumentationLabels).
						AddRow(appEntity2.ID, appEntity2.ApplicationTemplateID, appEntity2.SystemNumber, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.SystemStatus, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.OrdLabels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs, appEntity2.DocumentationLabels),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		ExpectedModelEntities:     []interface{}{appModel1, appModel2},
		ExpectedDBEntities:        []interface{}{appEntity1, appEntity2},
		MethodArgs:                []interface{}{givenTenant()},
		MethodName:                "ListAll",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListByRuntimeScenarios(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	hidingSelectors := map[string][]string{"foo": {"bar", "baz"}}

	suite := testdb.RepoListPageableTestSuite{
		Name: "List Applications By Scenarios",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{

				//SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE (id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3] UNION SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $4)) AND "key" = $5 AND "value" ?| array[$6] UNION SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $7)) AND "key" = $8 AND "value" ?| array[$9] EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $10)) AND "key" = $11 AND "value" @> $12 EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $13)) AND "key" = $14 AND "value" @> $15) AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $16))) ORDER BY id LIMIT 2 OFFSET 0
				//SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE (id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3] UNION SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $4)) AND "key" = $5 AND "value" ?| array[$6] UNION SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $7)) AND "key" = $8 AND "value" ?| array[$9] EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $10)) AND "key" = $11 AND "value" @> $12 EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $13)) AND "key" = $14 AND "value" @> $15) AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $16))) ORDER BY id LIMIT 2 OFFSET 0
				Query: regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications
												WHERE (id IN (SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3]
													UNION SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $4)) AND "key" = $5 AND "value" ?| array[$6]
													UNION SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $7)) AND "key" = $8 AND "value" ?| array[$9]
													EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $10)) AND "key" = $11 AND "value" @> $12
													EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $13)) AND "key" = $14 AND "value" @> $15)
												AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $16))) ORDER BY id LIMIT 2 OFFSET 0`),
				Args:     []driver.Value{givenTenant(), model.ScenariosKey, "Java", givenTenant(), model.ScenariosKey, "Go", givenTenant(), model.ScenariosKey, "Elixir", givenTenant(), "foo", strconv.Quote("bar"), givenTenant(), "foo", strconv.Quote("baz"), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns()).
						AddRow(appEntity1.ID, appEntity1.ApplicationTemplateID, appEntity1.SystemNumber, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.SystemStatus, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.OrdLabels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs, appEntity1.DocumentationLabels).
						AddRow(appEntity2.ID, appEntity2.ApplicationTemplateID, appEntity2.SystemNumber, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.SystemStatus, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.OrdLabels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs, appEntity2.DocumentationLabels),
					}
				},
			},
			{
				Query: regexp.QuoteMeta(`SELECT COUNT(*) FROM public.applications
												WHERE (id IN (SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" ?| array[$3]
													UNION SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $4)) AND "key" = $5 AND "value" ?| array[$6]
													UNION SELECT "app_id" FROM public.labels
													WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $7)) AND "key" = $8 AND "value" ?| array[$9]
													EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $10)) AND "key" = $11 AND "value" @> $12
													EXCEPT SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $13)) AND "key" = $14 AND "value" @> $15)
												AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $16)))`),
				Args:     []driver.Value{givenTenant(), model.ScenariosKey, "Java", givenTenant(), model.ScenariosKey, "Go", givenTenant(), model.ScenariosKey, "Elixir", givenTenant(), "foo", strconv.Quote("bar"), givenTenant(), "foo", strconv.Quote("baz"), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows([]string{"count"}).AddRow(2)}
				},
			},
		},
		Pages: []testdb.PageDetails{
			{
				ExpectedModelEntities: []interface{}{appModel1, appModel2},
				ExpectedDBEntities:    []interface{}{appEntity1, appEntity2},
				ExpectedPage: &model.ApplicationPage{
					Data: []*model.Application{appModel1, appModel2},
					PageInfo: &pagination.Page{
						StartCursor: "",
						EndCursor:   "",
						HasNextPage: false,
					},
					TotalCount: 2,
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		MethodArgs:                []interface{}{uuid.MustParse(givenTenant()), []string{"Java", "Go", "Elixir"}, 2, "", hidingSelectors},
		MethodName:                "ListByScenarios",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_GetByNameAndSystemNumber(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), appName, "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application By Name and System Number",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE name = $1 AND system_number = $2 AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $3))`),
				Args:     []driver.Value{appName, systemNumber, givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels),
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

func TestPgRepository_GetByFilter(t *testing.T) {
	entity := fixDetailedEntityApplication(t, givenID(), givenTenant(), appName, "Test app description")
	suite := testdb.RepoGetTestSuite{
		Name: "Get Application By filter",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" @> $3) AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $4))`),
				Args:     []driver.Value{givenTenantAsUUID(), "SCC", "{\"Subaccount\":\"subacc\", \"LocationID\":\"LocationID\", \"Host\":\"Host\"}", givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						sqlmock.NewRows(fixAppColumns()).
							AddRow(entity.ID, entity.ApplicationTemplateID, entity.SystemNumber, entity.Name, entity.Description, entity.StatusCondition, entity.StatusTimestamp, entity.SystemStatus, entity.HealthCheckURL, entity.IntegrationSystemID, entity.ProviderName, entity.BaseURL, entity.OrdLabels, entity.Ready, entity.CreatedAt, entity.UpdatedAt, entity.DeletedAt, entity.Error, entity.CorrelationIDs, entity.DocumentationLabels),
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
		MethodName:                "GetByFilter",
		MethodArgs:                []interface{}{givenTenant(), []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("SCC", "{\"Subaccount\":\"subacc\", \"LocationID\":\"LocationID\", \"Host\":\"Host\"}")}},
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func TestPgRepository_ListAllByFilter(t *testing.T) {
	app1ID := "aec0e9c5-06da-4625-9f8a-bda17ab8c3b9"
	app2ID := "ccdbef8f-b97a-490c-86e2-2bab2862a6e4"
	appEntity1 := fixDetailedEntityApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appEntity2 := fixDetailedEntityApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	appModel1 := fixDetailedModelApplication(t, app1ID, givenTenant(), "App 1", "App desc 1")
	appModel2 := fixDetailedModelApplication(t, app2ID, givenTenant(), "App 2", "App desc 2")

	suite := testdb.RepoListTestSuite{
		Name: "List Applications",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_template_id, system_number, name, description, status_condition, status_timestamp, system_status, healthcheck_url, integration_system_id, provider_name, base_url, labels, ready, created_at, updated_at, deleted_at, error, correlation_ids, documentation_labels FROM public.applications WHERE id IN (SELECT "app_id" FROM public.labels WHERE "app_id" IS NOT NULL AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $1)) AND "key" = $2 AND "value" @> $3) AND (id IN (SELECT id FROM tenant_applications WHERE tenant_id = $4))`),
				Args:     []driver.Value{givenTenantAsUUID(), "scc", "{\"locationId\":\"locationId\"}", givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns()).
						AddRow(appEntity1.ID, appEntity1.ApplicationTemplateID, appEntity1.SystemNumber, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.SystemStatus, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID, appEntity1.ProviderName, appEntity1.BaseURL, appEntity1.OrdLabels, appEntity1.Ready, appEntity1.CreatedAt, appEntity1.UpdatedAt, appEntity1.DeletedAt, appEntity1.Error, appEntity1.CorrelationIDs, appEntity1.DocumentationLabels).
						AddRow(appEntity2.ID, appEntity2.ApplicationTemplateID, appEntity2.SystemNumber, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.SystemStatus, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID, appEntity2.ProviderName, appEntity2.BaseURL, appEntity2.OrdLabels, appEntity2.Ready, appEntity2.CreatedAt, appEntity2.UpdatedAt, appEntity2.DeletedAt, appEntity2.Error, appEntity2.CorrelationIDs, appEntity2.DocumentationLabels),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixAppColumns())}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       application.NewRepository,
		ExpectedModelEntities:     []interface{}{appModel1, appModel2},
		ExpectedDBEntities:        []interface{}{appEntity1, appEntity2},
		MethodArgs:                []interface{}{givenTenant(), []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("scc", "{\"locationId\":\"locationId\"}")}},
		MethodName:                "ListAllByFilter",
		DisableConverterErrorTest: true,
	}

	suite.Run(t)
}

func givenError() error {
	return errors.New("some error")
}
