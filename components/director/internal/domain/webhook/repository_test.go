package webhook_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

func TestRepositoryGetByID(t *testing.T) {
	whModel := fixApplicationModelWebhook(givenID(), givenApplicationID(), givenTenant(), "http://kyma.io")
	whEntity := fixApplicationWebhookEntity(t)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Webhook By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE id = $1 AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(whModel.ID, givenApplicationID(), nil, whModel.Type, whModel.URL, fixAuthAsAString(t), nil, nil, whModel.Mode, whModel.CorrelationIDKey, whModel.RetryInterval, whModel.Timeout, whModel.URLTemplate, whModel.InputTemplate, whModel.HeaderTemplate, whModel.OutputTemplate, whModel.StatusTemplate)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: webhook.NewRepository,
		ExpectedModelEntity: whModel,
		ExpectedDBEntity:    whEntity,
		MethodArgs:          []interface{}{givenTenant(), givenID(), model.ApplicationWebhookReference},
	}

	suite.Run(t)
}

func TestRepositoryGetByIDGlobal(t *testing.T) {
	whType := model.WebhookTypeConfigurationChanged
	whModel := fixApplicationTemplateModelWebhookWithType(givenID(), givenApplicationTemplateID(), "http://kyma.io", whType)
	whEntity := fixApplicationTemplateWebhookEntity(t)

	// GIVEN
	mockConverter := &automock.EntityConverter{}
	defer mockConverter.AssertExpectations(t)
	mockConverter.On("FromEntity", whEntity).Return(whModel, nil)

	db, dbMock := testdb.MockDatabase(t)
	defer dbMock.AssertExpectations(t)

	rows := sqlmock.NewRows(fixColumns).
		AddRow(whModel.ID, nil, givenApplicationTemplateID(), whModel.Type, whModel.URL, fixAuthAsAString(t), nil, nil, whModel.Mode, whModel.CorrelationIDKey, whModel.RetryInterval, whModel.Timeout, whModel.URLTemplate, whModel.InputTemplate, whModel.HeaderTemplate, whModel.OutputTemplate, whModel.StatusTemplate)

	dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE id = $1")).
		WithArgs(givenID()).WillReturnRows(rows)

	ctx := persistence.SaveToContext(context.TODO(), db)
	sut := webhook.NewRepository(mockConverter)
	// WHEN
	wh, err := sut.GetByIDGlobal(ctx, givenID())
	// THEN
	require.NoError(t, err)
	require.Equal(t, whModel, wh)
	mockConverter.AssertExpectations(t)
}

func TestRepositoryCreate(t *testing.T) {
	var nilWebhookModel *model.Webhook

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Application webhook",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{givenTenant(), givenApplicationID(), true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.webhooks ( id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{givenID(), givenApplicationID(), sql.NullString{}, string(model.WebhookTypeConfigurationChanged), "http://kyma.io", fixAuthAsAString(t), nil, nil, model.WebhookModeSync, nil, nil, nil, "{}", "{}", "{}", "{}", nil},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: webhook.NewRepository,
		ModelEntity:         fixApplicationModelWebhook(givenID(), givenApplicationID(), givenTenant(), "http://kyma.io"),
		DBEntity:            fixApplicationWebhookEntity(t),
		NilModelEntity:      nilWebhookModel,
		TenantID:            givenTenant(),
	}

	suite.Run(t)

	// Additional tests for application template webhook -> global create
	t.Run("Create ApplicationTemplate webhook", func(t *testing.T) {
		tmplWhModel := fixApplicationTemplateModelWebhook(givenID(), givenApplicationTemplateID(), "http://kyma.io")
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tmplWhModel).Return(fixApplicationTemplateWebhookEntity(t), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("INSERT INTO public.webhooks ( id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )")).WithArgs(
			givenID(), sql.NullString{}, givenApplicationTemplateID(), string(model.WebhookTypeConfigurationChanged), "http://kyma.io", fixAuthAsAString(t), nil, nil, model.WebhookModeSync, nil, nil, nil, "{}", "{}", "{}", "{}", nil).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, "", tmplWhModel)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepositoryCreateMany(t *testing.T) {
	expectedParentAccess := regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3")
	expectedInsert := regexp.QuoteMeta("INSERT INTO public.webhooks ( id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )")

	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{
			{ID: "one", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "two", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "three", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
		}
		mockConverter.On("ToEntity", given[0]).Return(&webhook.Entity{ID: "one", ApplicationID: repo.NewValidNullableString(givenApplicationID())}, nil)
		mockConverter.On("ToEntity", given[1]).Return(&webhook.Entity{ID: "two", ApplicationID: repo.NewValidNullableString(givenApplicationID())}, nil)
		mockConverter.On("ToEntity", given[2]).Return(&webhook.Entity{ID: "three", ApplicationID: repo.NewValidNullableString(givenApplicationID())}, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(expectedParentAccess).WithArgs(givenTenant(), givenApplicationID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(expectedInsert).WithArgs(
			"one", givenApplicationID(), nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectQuery(expectedParentAccess).WithArgs(givenTenant(), givenApplicationID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(expectedInsert).WithArgs(
			"two", givenApplicationID(), nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectQuery(expectedParentAccess).WithArgs(givenTenant(), givenApplicationID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(expectedInsert).WithArgs(
			"three", givenApplicationID(), nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, givenTenant(), given)
		// THEN
		require.NoError(t, err)
	})

	t.Run("got error on converting object", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{
			{ID: "one", URL: stringPtr("unlucky"), Type: model.WebhookTypeConfigurationChanged, ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "two", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "three", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference}}
		mockConverter.On("ToEntity", given[0]).Return(&webhook.Entity{}, givenError())

		ctx := persistence.SaveToContext(context.TODO(), nil)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, givenTenant(), given)
		// THEN
		expectedErr := fmt.Sprintf("while creating Webhook with type %s and id %s for %s: while converting model to entity: some error", model.WebhookTypeConfigurationChanged, "one", model.ApplicationWebhookReference)
		require.EqualError(t, err, expectedErr)
	})

	t.Run("got error  on db communication", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{
			{ID: "one", URL: stringPtr("unlucky"), Type: model.WebhookTypeConfigurationChanged, ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "two", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference},
			{ID: "three", ObjectID: givenApplicationID(), ObjectType: model.ApplicationWebhookReference}}
		mockConverter.On("ToEntity", given[0]).Return(&webhook.Entity{ID: "one", ApplicationID: repo.NewValidNullableString(givenApplicationID())}, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery(expectedParentAccess).WithArgs(givenTenant(), givenApplicationID(), true).WillReturnRows(testdb.RowWhenObjectExist())
		dbMock.ExpectExec(expectedInsert).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, givenTenant(), given)
		// THEN
		expectedErr := fmt.Sprintf("while creating Webhook with type %s and id %s for %s: Internal Server Error: Unexpected error while executing SQL query", model.WebhookTypeConfigurationChanged, "one", model.ApplicationWebhookReference)
		require.EqualError(t, err, expectedErr)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	var nilWebhookModel *model.Webhook

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update Application webhook",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`UPDATE public.webhooks SET type = ?, url = ?, auth = ?, mode = ?, retry_interval = ?, timeout = ?, url_template = ?, input_template = ?, header_template = ?, output_template = ?, status_template = ? WHERE id = ? AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = ? AND owner = true))`),
				Args:          []driver.Value{string(model.WebhookTypeConfigurationChanged), "http://kyma.io", fixAuthAsAString(t), model.WebhookModeSync, nil, nil, "{}", "{}", "{}", "{}", nil, givenID(), givenTenant()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: webhook.NewRepository,
		ModelEntity:         fixApplicationModelWebhook(givenID(), givenApplicationID(), givenTenant(), "http://kyma.io"),
		DBEntity:            fixApplicationWebhookEntity(t),
		NilModelEntity:      nilWebhookModel,
		TenantID:            givenTenant(),
	}

	suite.Run(t)

	// Additional tests for application template webhook -> global create
	t.Run("Update ApplicationTemplate webhook", func(t *testing.T) {
		tmplWhModel := fixApplicationTemplateModelWebhook(givenID(), givenApplicationTemplateID(), "http://kyma.io")
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", tmplWhModel).Return(fixApplicationTemplateWebhookEntity(t), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`UPDATE public.webhooks SET type = ?, url = ?, auth = ?, mode = ?, retry_interval = ?, timeout = ?, url_template = ?, input_template = ?, header_template = ?, output_template = ?, status_template = ? WHERE id = ? AND app_template_id = ?`)).
			WithArgs(string(model.WebhookTypeConfigurationChanged), "http://kyma.io", fixAuthAsAString(t), model.WebhookModeSync, nil, nil, "{}", "{}", "{}", "{}", nil, givenID(), givenApplicationTemplateID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Update(ctx, "", tmplWhModel)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepositoryDelete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.webhooks WHERE id = $1")).WithArgs(
			givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(nil)
		// WHEN
		err := sut.Delete(ctx, givenID())
		// THEN
		require.NoError(t, err)
	})

	t.Run("got error on db communication", func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(nil)
		// WHEN
		err := sut.Delete(ctx, givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepositoryDeleteAllByApplicationID(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Webhook Delete by ApplicationID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.webhooks WHERE app_id = $1 AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{givenApplicationID(), givenTenant()},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: webhook.NewRepository,
		MethodArgs:          []interface{}{givenTenant(), givenApplicationID()},
		MethodName:          "DeleteAllByApplicationID",
		IsDeleteMany:        true,
	}

	suite.Run(t)
}

func TestRepositoryListByApplicationID(t *testing.T) {
	testListByObjectID(t, "ListByReferenceObjectID", repo.NoLock, []interface{}{givenTenant(), givenApplicationID(), model.ApplicationWebhookReference})
}

func TestRepositoryListByApplicationIDWithSelectForUpdate(t *testing.T) {
	testListByObjectID(t, "ListByApplicationIDWithSelectForUpdate", " "+repo.ForUpdateLock, []interface{}{givenTenant(), givenApplicationID()})
}

func TestRepositoryListByRuntimeID(t *testing.T) {
	whID1 := "whID1"
	whID2 := "whID2"
	whModel1 := fixRuntimeModelWebhook(whID1, givenRuntimeID(), "http://kyma.io")
	whEntity1 := fixRuntimeWebhookEntityWithID(t, whID1)

	whModel2 := fixRuntimeModelWebhook(whID2, givenRuntimeID(), "http://kyma.io")
	whEntity2 := fixRuntimeWebhookEntityWithID(t, whID2)

	suite := testdb.RepoListTestSuite{
		Name: "List Webhooks by Runtime ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE runtime_id = $1 AND (id IN (SELECT id FROM runtime_webhooks_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{givenRuntimeID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(whModel1.ID, nil, nil, whModel1.Type, whModel1.URL, fixAuthAsAString(t), givenRuntimeID(), nil, whModel1.Mode, whModel1.CorrelationIDKey, whModel1.RetryInterval, whModel1.Timeout, whModel1.URLTemplate, whModel1.InputTemplate, whModel1.HeaderTemplate, whModel1.OutputTemplate, whModel1.StatusTemplate).
						AddRow(whModel2.ID, nil, nil, whModel2.Type, whModel2.URL, fixAuthAsAString(t), givenRuntimeID(), nil, whModel2.Mode, whModel2.CorrelationIDKey, whModel2.RetryInterval, whModel2.Timeout, whModel2.URLTemplate, whModel2.InputTemplate, whModel2.HeaderTemplate, whModel2.OutputTemplate, whModel2.StatusTemplate),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   webhook.NewRepository,
		ExpectedModelEntities: []interface{}{whModel1, whModel2},
		ExpectedDBEntities:    []interface{}{whEntity1, whEntity2},
		MethodArgs:            []interface{}{givenTenant(), givenRuntimeID(), model.RuntimeWebhookReference},
		MethodName:            "ListByReferenceObjectID",
	}

	suite.Run(t)

	t.Run("Error when webhookReferenceObjectType is not supported", func(t *testing.T) {
		repository := webhook.NewRepository(nil)
		_, err := repository.ListByReferenceObjectID(context.TODO(), "", "", model.IntegrationSystemWebhookReference)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "referenced object should be one of application and runtime")
	})
}

func testListByObjectID(t *testing.T, methodName string, lockClause string, args []interface{}) {
	whID1 := "whID1"
	whID2 := "whID2"
	whModel1 := fixApplicationModelWebhook(whID1, givenApplicationID(), givenTenant(), "http://kyma.io")
	whEntity1 := fixApplicationWebhookEntityWithID(t, whID1)

	whModel2 := fixApplicationModelWebhook(whID2, givenApplicationID(), givenTenant(), "http://kyma.io")
	whEntity2 := fixApplicationWebhookEntityWithID(t, whID2)

	suite := testdb.RepoListTestSuite{
		Name: "List Webhooks by Application ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE app_id = $1 AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = $2))` + lockClause),
				Args:     []driver.Value{givenApplicationID(), givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(whModel1.ID, givenApplicationID(), nil, whModel1.Type, whModel1.URL, fixAuthAsAString(t), nil, nil, whModel1.Mode, whModel1.CorrelationIDKey, whModel1.RetryInterval, whModel1.Timeout, whModel1.URLTemplate, whModel1.InputTemplate, whModel1.HeaderTemplate, whModel1.OutputTemplate, whModel1.StatusTemplate).
						AddRow(whModel2.ID, givenApplicationID(), nil, whModel2.Type, whModel2.URL, fixAuthAsAString(t), nil, nil, whModel2.Mode, whModel2.CorrelationIDKey, whModel2.RetryInterval, whModel2.Timeout, whModel2.URLTemplate, whModel2.InputTemplate, whModel2.HeaderTemplate, whModel2.OutputTemplate, whModel2.StatusTemplate),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   webhook.NewRepository,
		ExpectedModelEntities: []interface{}{whModel1, whModel2},
		ExpectedDBEntities:    []interface{}{whEntity1, whEntity2},
		MethodArgs:            args,
		MethodName:            methodName,
	}

	suite.Run(t)
}

func TestRepository_ListByWebhookType(t *testing.T) {
	whID := "whID1"
	whType := model.WebhookTypeOpenResourceDiscovery

	whModel := fixApplicationModelWebhook(whID, givenApplicationID(), givenTenant(), "http://kyma.io")
	whModel.Type = whType
	whEntity := fixApplicationWebhookEntityWithIDAndWebhookType(t, whID, whType)

	suite := testdb.RepoListTestSuite{
		Name: "List Webhooks by type",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE type = $1`),
				Args:     []driver.Value{whType},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(whModel.ID, givenApplicationID(), nil, whModel.Type, whModel.URL, fixAuthAsAString(t), nil, nil, whModel.Mode, whModel.CorrelationIDKey, whModel.RetryInterval, whModel.Timeout, whModel.URLTemplate, whModel.InputTemplate, whModel.HeaderTemplate, whModel.OutputTemplate, whModel.StatusTemplate),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   webhook.NewRepository,
		ExpectedModelEntities: []interface{}{whModel},
		ExpectedDBEntities:    []interface{}{whEntity},
		MethodArgs:            []interface{}{whType},
		MethodName:            "ListByWebhookType",
	}

	suite.Run(t)
}

func TestRepositoryListByApplicationTemplateID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity",
			&webhook.Entity{ID: givenID(),
				ApplicationTemplateID: repo.NewValidNullableString(givenApplicationTemplateID()),
				Type:                  string(model.WebhookTypeConfigurationChanged),
				URL:                   repo.NewValidNullableString("http://kyma.io")}).
			Return(&model.Webhook{
				ID: givenID(),
			}, nil)

		mockConv.On("FromEntity",
			&webhook.Entity{ID: anotherID(),
				ApplicationTemplateID: repo.NewValidNullableString(givenApplicationTemplateID()),
				Type:                  string(model.WebhookTypeConfigurationChanged),
				URL:                   repo.NewValidNullableString("http://kyma2.io")}).
			Return(&model.Webhook{ID: anotherID()}, nil)

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "app_template_id", "type", "url", "auth"}).
			AddRow(givenID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil).
			AddRow(anotherID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma2.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE app_template_id = $1")).
			WithArgs(givenApplicationTemplateID()).
			WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.ListByApplicationTemplateID(ctx, givenApplicationTemplateID())
		// THEN
		require.NoError(t, err)
		require.Len(t, actual, 2)
		assert.Equal(t, givenID(), actual[0].ID)
		assert.Equal(t, anotherID(), actual[1].ID)
	})

	t.Run("success if no found", func(t *testing.T) {
		// GIVE
		sut := webhook.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		noRows := sqlmock.NewRows([]string{"id", "app_id", "app_template_id", "type", "url", "auth"})

		dbMock.ExpectQuery("SELECT").WithArgs(givenApplicationTemplateID()).WillReturnRows(noRows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.ListByApplicationTemplateID(ctx, givenApplicationTemplateID())
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run("got error  on db communication", func(t *testing.T) {
		// GIVEN
		sut := webhook.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT").WillReturnError(givenError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.ListByApplicationTemplateID(ctx, givenApplicationTemplateID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("got error on converting object", func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity", mock.Anything).Return(&model.Webhook{}, givenError())

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "app_template_id", "type", "url", "auth"}).
			AddRow(givenID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT")).WithArgs(givenApplicationTemplateID()).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.ListByApplicationTemplateID(ctx, givenApplicationTemplateID())
		// THEN
		require.EqualError(t, err, "while converting Webhook to model: some error")
	})
}

func Test_ListByReferenceObjectTypeAndWebhookType(t *testing.T) {
	whID1 := "whID1"
	whID2 := "whID2"
	whType := model.WebhookTypeConfigurationChanged

	whModel1 := fixApplicationModelWebhook(whID1, givenApplicationID(), givenTenant(), "http://kyma.io")
	whEntity1 := fixApplicationWebhookEntityWithIDAndWebhookType(t, whID1, whType)

	whModel2 := fixApplicationModelWebhook(whID2, givenApplicationID(), givenTenant(), "http://kyma.io")
	whEntity2 := fixApplicationWebhookEntityWithIDAndWebhookType(t, whID2, whType)

	suite := testdb.RepoListTestSuite{
		Name: "List Webhooks by Application ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE app_id IS NOT NULL AND type = $1 AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = $2))`),
				Args:     []driver.Value{whType, givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(whModel1.ID, givenApplicationID(), nil, whModel1.Type, whModel1.URL, fixAuthAsAString(t), nil, nil, whModel1.Mode, whModel1.CorrelationIDKey, whModel1.RetryInterval, whModel1.Timeout, whModel1.URLTemplate, whModel1.InputTemplate, whModel1.HeaderTemplate, whModel1.OutputTemplate, whModel1.StatusTemplate).
						AddRow(whModel2.ID, givenApplicationID(), nil, whModel2.Type, whModel2.URL, fixAuthAsAString(t), nil, nil, whModel2.Mode, whModel2.CorrelationIDKey, whModel2.RetryInterval, whModel2.Timeout, whModel2.URLTemplate, whModel2.InputTemplate, whModel2.HeaderTemplate, whModel2.OutputTemplate, whModel2.StatusTemplate),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   webhook.NewRepository,
		ExpectedModelEntities: []interface{}{whModel1, whModel2},
		ExpectedDBEntities:    []interface{}{whEntity1, whEntity2},
		MethodArgs:            []interface{}{givenTenant(), whType, model.ApplicationWebhookReference},
		MethodName:            "ListByReferenceObjectTypeAndWebhookType",
	}

	suite.Run(t)
}

func TestRepositoryGetByIDAndWebhookType(t *testing.T) {
	whType := model.WebhookTypeConfigurationChanged
	whModel := fixApplicationModelWebhookWithType(givenID(), givenApplicationID(), givenTenant(), "http://kyma.io", whType)
	whEntity := fixApplicationWebhookEntity(t)

	suite := testdb.RepoGetTestSuite{
		Name: "Get Webhook By ID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE app_id = $1 AND type = $2 AND (id IN (SELECT id FROM application_webhooks_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{givenID(), whType, givenTenant()},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).AddRow(whModel.ID, givenApplicationID(), nil, whModel.Type, whModel.URL, fixAuthAsAString(t), nil, nil, whModel.Mode, whModel.CorrelationIDKey, whModel.RetryInterval, whModel.Timeout, whModel.URLTemplate, whModel.InputTemplate, whModel.HeaderTemplate, whModel.OutputTemplate, whModel.StatusTemplate)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: webhook.NewRepository,
		ExpectedModelEntity: whModel,
		ExpectedDBEntity:    whEntity,
		MethodArgs:          []interface{}{givenTenant(), givenID(), model.ApplicationWebhookReference, whType},
		MethodName:          "GetByIDAndWebhookType",
	}

	suite.Run(t)
}
