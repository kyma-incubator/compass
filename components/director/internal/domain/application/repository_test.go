package application_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
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
	t.Run("Success", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.applications WHERE tenant_id = $1 AND id = $2")).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := application.NewRepository(nil)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		// then
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := application.NewRepository(nil)

		// when
		err := repo.Delete(ctx, givenTenant(), givenID())

		// then
		require.EqualError(t, err, "while deleting from database: some error")
	})
}

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.applications ( id, tenant_id, name, description, status_condition, status_timestamp, healthcheck_url, integration_system_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(givenID(), givenTenant(), appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, appModel)

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

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, appModel)

		// then
		require.EqualError(t, err, "while inserting row to 'public.applications' table: some error")
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
	updateStmt := `UPDATE public\.applications SET name = \?, description = \?, status_condition = \?, status_timestamp = \?, healthcheck_url = \?, integration_system_id = \? WHERE tenant_id = \? AND id = \?`

	t.Run("Success", func(t *testing.T) {
		// given
		appModel := fixDetailedModelApplication(t, givenID(), givenTenant(), "Test app", "Test app description")
		appEntity := fixDetailedEntityApplication(t, givenID(), givenTenant(), "Test app", "Test app description")

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", appModel).Return(appEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WithArgs(appModel.Name, appModel.Description, appModel.Status.Condition, appModel.Status.Timestamp, appModel.HealthCheckURL, appModel.IntegrationSystemID, givenTenant(), givenID()).
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
		repo := application.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, appModel)

		// then
		require.EqualError(t, err, "while updating single entity: some error")
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

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}).
			AddRow(givenID(), givenTenant(), appEntity.Name, appEntity.Description, appEntity.StatusCondition, appEntity.StatusTimestamp, appEntity.HealthCheckURL, appEntity.IntegrationSystemID)

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
		repo := application.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repo.GetByID(ctx, givenTenant(), givenID())

		// then
		require.EqualError(t, err, "while getting object from DB: some error")
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

	pageableQuery := `^SELECT (.+) FROM public\.applications WHERE tenant_id=\$1 ORDER BY id LIMIT %d OFFSET %d$`
	countQuery := `SELECT COUNT\(\*\) FROM public\.applications WHERE tenant_id=\$1`

	t.Run("Success", func(t *testing.T) {
		// given
		rows := sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}).
			AddRow(appEntity1.ID, appEntity1.TenantID, appEntity1.Name, appEntity1.Description, appEntity1.StatusCondition, appEntity1.StatusTimestamp, appEntity1.HealthCheckURL, appEntity1.IntegrationSystemID).
			AddRow(appEntity2.ID, appEntity2.TenantID, appEntity2.Name, appEntity2.Description, appEntity2.StatusCondition, appEntity2.StatusTimestamp, appEntity2.HealthCheckURL, appEntity2.IntegrationSystemID)

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

func TestPgRepository_ListByRuntimeScenarios(t *testing.T) {
	tenantID := uuid.New()
	app1ID := uuid.New()
	app2ID := uuid.New()
	pageSize := 5
	cursor := ""
	timestamp, err := time.Parse(time.RFC3339, "2002-10-02T10:00:00-05:00")
	assert.NoError(t, err)

	runtimeScenarios := []string{"Java", "Go", "Elixir"}
	scenariosQuery := fmt.Sprintf(`SELECT "app_id" FROM public.labels
					WHERE "app_id" IS NOT NULL AND "tenant_id" = '%s'
						AND "key" = 'scenarios' AND "value" @> '["Java"]'
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND "tenant_id" = '%s'
						AND "key" = 'scenarios' AND "value" @> '["Go"]'
					UNION SELECT "app_id" FROM public.labels
						WHERE "app_id" IS NOT NULL AND "tenant_id" = '%s'
						AND "key" = 'scenarios' AND "value" @> '["Elixir"]'`, tenantID, tenantID, tenantID)
	applicationScenarioQuery := regexp.QuoteMeta(scenariosQuery)

	pagableQuery := fmt.Sprintf(`SELECT (.+) FROM public\.applications WHERE tenant_id=\$1 AND "id" IN \(%s\) ORDER BY id LIMIT %d OFFSET %d`,
		applicationScenarioQuery,
		pageSize,
		0)

	countQuery := fmt.Sprintf(`SELECT COUNT\(\*\) FROM public\.applications WHERE tenant_id=\$1 AND "id" IN \(%s\)$`, applicationScenarioQuery)

	conv := application.NewConverter(nil, nil, nil, nil)

	testCases := []struct {
		Name                    string
		ExpectedApplicationRows *sqlmock.Rows
		TotalCount              int
		ExpectedError           error
	}{
		{
			Name: "Success",
			ExpectedApplicationRows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}).
				AddRow(app1ID, tenantID, "App ABC", "Description for application ABC", "INITIAL", timestamp, "http://domain.local/app1", "test").
				AddRow(app2ID, tenantID, "App XYZ", "Description for application XYZ", "INITIAL", timestamp, "http://domain.local/app2", "test"),
			TotalCount:    2,
			ExpectedError: nil,
		},
		{
			Name:                    "Return empty page when no application match",
			ExpectedApplicationRows: sqlmock.NewRows([]string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "healthcheck_url", "integration_system_id"}),
			TotalCount:              0,
			ExpectedError:           nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			sqlxDB, sqlMock := testdb.MockDatabase(t)
			if testCase.ExpectedApplicationRows != nil {
				sqlMock.ExpectQuery(pagableQuery).
					WithArgs(tenantID).
					WillReturnRows(testCase.ExpectedApplicationRows)

				countRow := sqlMock.NewRows([]string{"count"}).AddRow(testCase.TotalCount)
				sqlMock.ExpectQuery(countQuery).
					WithArgs(tenantID).
					WillReturnRows(countRow)
			}
			repository := application.NewRepository(conv)

			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			//WHEN
			page, err := repository.ListByScenarios(ctx, tenantID, runtimeScenarios, pageSize, cursor)

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
