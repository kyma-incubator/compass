package webhook_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
)

const (
	testCaseSuccess                  = "success"
	testCaseSuccessWithAuth          = "success with auth"
	testCaseErrorOnConvertingObjects = "got error on converting object"
	testCaseErrorOnDBCommunication   = "got error on db communication"
)
/*
func TestRepositoryGetByID(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", fixEntity()).Return(givenModel(), nil)

		sut := webhook.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth",
			"runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil, nil, nil, model.WebhookModeSync, nil, nil, nil, "{}", "{}", "{}", "{}", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, tenant_id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, givenModel(), *actual)
	})

	t.Run(testCaseSuccessWithAuth, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", givenEntityWithAuth(t)).Return(givenModelWithAuth(), nil)

		sut := webhook.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth",
			"runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", fixAuthAsAString(t), nil, nil, model.WebhookModeSync, nil, nil, nil, "{}", "{}", "{}", "{}", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, tenant_id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, givenModelWithAuth(), *actual)
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", fixEntity()).Return(model.Webhook{}, givenError())

		sut := webhook.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth",
			"runtime_id", "integration_system_id", "mode", "correlation_id_key", "retry_interval", "timeout", "url_template", "input_template", "header_template", "output_template", "status_template"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil, nil, nil, model.WebhookModeSync, nil, nil, nil, "{}", "{}", "{}", "{}", nil)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while converting from entity to model: some error")
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		sut := webhook.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.GetByID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
*/
func TestRepositoryCreate(t *testing.T) {
	var nilWebhookModel *model.Webhook

	suite := testdb.RepoCreateTestSuite{
		Name: "Create Application webhook",
		SqlQueryDetails: []testdb.SqlQueryDetails{
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
		RepoConstructorFunc:       webhook.NewRepository,
		ModelEntity:               fixApplicationModelWebhook(givenID(), givenApplicationID(), givenTenant(), "http://kyma.io"),
		DBEntity:                  fixApplicationWebhookEntity(t),
		NilModelEntity:            nilWebhookModel,
		TenantID:                  givenTenant(),
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
/*
func TestRepositoryCreateMany(t *testing.T) {
	const expectedInsert = "INSERT INTO public.webhooks ( id, tenant_id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )"
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{{ID: "one"}, {ID: "two"}, {ID: "three"}}
		mockConverter.On("ToEntity", *given[0]).Return(webhook.Entity{ID: "one"}, nil)
		mockConverter.On("ToEntity", *given[1]).Return(webhook.Entity{ID: "two"}, nil)
		mockConverter.On("ToEntity", *given[2]).Return(webhook.Entity{ID: "three"}, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(
			"one", nil, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(
			"two", nil, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(
			"three", nil, nil, nil, "", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, given)
		// THEN
		require.NoError(t, err)
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{{ID: "one", URL: stringPtr("unlucky"), Type: model.WebhookTypeConfigurationChanged}, {ID: "two"}, {ID: "three"}}
		mockConverter.On("ToEntity", *given[0]).Return(webhook.Entity{}, givenError())

		ctx := persistence.SaveToContext(context.TODO(), nil)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, given)
		// THEN
		expectedErr := fmt.Sprintf("while creating Webhook with type %s and id %s for %s: while converting model to entity: some error", model.WebhookTypeConfigurationChanged, "one", webhook.PrintOwnerInfo(given[0]))
		require.EqualError(t, err, expectedErr)
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{{ID: "one", URL: stringPtr("unlucky"), Type: model.WebhookTypeConfigurationChanged}, {ID: "two"}, {ID: "three"}}
		mockConverter.On("ToEntity", *given[0]).Return(webhook.Entity{ID: "one"}, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, given)
		// THEN
		expectedErr := fmt.Sprintf("while creating Webhook with type %s and id %s for %s: Internal Server Error: Unexpected error while executing SQL query", model.WebhookTypeConfigurationChanged, "one", webhook.PrintOwnerInfo(given[0]))
		require.EqualError(t, err, expectedErr)
	})
}

func TestRepositoryUpdate(t *testing.T) {
	applicaitonModel := givenModel()
	applicaitonEntity := fixEntity()
	applicaitonModel.ApplicationTemplateID = nil
	applicaitonEntity.ApplicationTemplateID = repo.NewValidNullableString("")

	applicaitonTemplateModel := givenModel()
	applicaitonTemplatEntity := fixEntity()
	applicaitonTemplateModel.ApplicationID = nil
	applicaitonTemplatEntity.ApplicationID = repo.NewValidNullableString("")
	applicaitonTemplateModel.TenantID = nil
	applicaitonTemplatEntity.TenantID = repo.NewValidNullableString("")

	tests := []struct {
		name                string
		mockConverterSetter func(*automock.EntityConverter)
		dbMockSetter        func() (*sqlx.DB, testdb.DBMock)
		model               model.Webhook
		expectedError       error
	}{
		{
			name: "Success for application Webhook",
			mockConverterSetter: func(converter *automock.EntityConverter) {
				converter.On("ToEntity", applicaitonModel).Return(applicaitonEntity, nil)
			},
			dbMockSetter: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("UPDATE public.webhooks SET type = ?, url = ?, auth = ?, mode = ?, retry_interval = ?, timeout = ?, url_template = ?, input_template = ?, header_template = ?, output_template = ?, status_template = ? WHERE %s AND id = ? AND app_id = ?", fixUpdateTenantIsolationSubquery()))).WithArgs(
					string(model.WebhookTypeConfigurationChanged), "http://kyma.io", nil, model.WebhookModeSync, nil, nil, "{}", "{}", "{}", "{}", nil, givenTenant(), givenID(), givenApplicationID()).WillReturnResult(sqlmock.NewResult(-1, 1))
				return db, dbMock
			},
			model:         applicaitonModel,
			expectedError: nil,
		},
		{
			name: "Success for applicationTemplate Webhook",
			mockConverterSetter: func(converter *automock.EntityConverter) {
				converter.On("ToEntity", applicaitonTemplateModel).Return(applicaitonTemplatEntity, nil)
			},
			dbMockSetter: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec(regexp.QuoteMeta("UPDATE public.webhooks SET type = ?, url = ?, auth = ?, mode = ?, retry_interval = ?, timeout = ?, url_template = ?, input_template = ?, header_template = ?, output_template = ?, status_template = ? WHERE id = ?")).WithArgs(
					string(model.WebhookTypeConfigurationChanged), "http://kyma.io", nil, model.WebhookModeSync, nil, nil, "{}", "{}", "{}", "{}", nil, givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))
				return db, dbMock
			},
			model:         applicaitonTemplateModel,
			expectedError: nil,
		},
		{
			name: testCaseErrorOnConvertingObjects,
			mockConverterSetter: func(converter *automock.EntityConverter) {
				converter.On("ToEntity", applicaitonModel).Return(webhook.Entity{}, givenError())
			},
			dbMockSetter:  nil,
			model:         applicaitonModel,
			expectedError: errors.New("while converting model to entity: some error"),
		},
		{
			name: testCaseErrorOnDBCommunication,
			mockConverterSetter: func(converter *automock.EntityConverter) {
				converter.On("ToEntity", applicaitonModel).Return(applicaitonEntity, nil)
			},
			dbMockSetter: func() (*sqlx.DB, testdb.DBMock) {
				db, dbMock := testdb.MockDatabase(t)
				dbMock.ExpectExec("UPDATE .*").WillReturnError(givenError())
				return db, dbMock
			},
			model:         applicaitonModel,
			expectedError: errors.New("Internal Server Error: Unexpected error while executing SQL query"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockConverter := &automock.EntityConverter{}
			defer mockConverter.AssertExpectations(t)

			var db *sqlx.DB
			var dbmock testdb.DBMock
			ctx := context.TODO()

			test.mockConverterSetter(mockConverter)

			if test.dbMockSetter != nil {
				db, dbmock = test.dbMockSetter()
				defer dbmock.AssertExpectations(t)
				ctx = persistence.SaveToContext(context.TODO(), db)
			}

			sut := webhook.NewRepository(mockConverter)
			// WHEN
			err := sut.Update(ctx, ptr(test.model))

			if test.expectedError != nil {
				require.EqualError(t, err, test.expectedError.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRepositoryDelete(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
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

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
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
	sut := webhook.NewRepository(nil)
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.webhooks WHERE %s AND app_id = $2", fixUnescapedTenantIsolationSubquery()))).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 123))

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		err := sut.DeleteAllByApplicationID(ctx, givenTenant(), givenID())
		// THEN
		require.NoError(t, err)
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			givenTenant(), givenID()).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		err := sut.DeleteAllByApplicationID(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepositoryListByApplicationID(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity",
			webhook.Entity{ID: givenID(),
				TenantID:      repo.NewValidNullableString(givenTenant()),
				ApplicationID: repo.NewValidNullableString(givenApplicationID()),
				Type:          string(model.WebhookTypeConfigurationChanged),
				URL:           repo.NewValidNullableString("http://kyma.io")}).
			Return(model.Webhook{
				ID: givenID(),
			}, nil)

		mockConv.On("FromEntity",
			webhook.Entity{ID: anotherID(),
				TenantID:      repo.NewValidNullableString(givenTenant()),
				ApplicationID: repo.NewValidNullableString(givenApplicationID()),
				Type:          string(model.WebhookTypeConfigurationChanged),
				URL:           repo.NewValidNullableString("http://kyma2.io")}).
			Return(model.Webhook{ID: anotherID()}, nil)

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth"}).
			AddRow(givenID(), givenTenant(), givenApplicationID(), nil, model.WebhookTypeConfigurationChanged, "http://kyma.io", nil).
			AddRow(anotherID(), givenTenant(), givenApplicationID(), nil, model.WebhookTypeConfigurationChanged, "http://kyma2.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, tenant_id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE %s AND app_id = $2", fixUnescapedTenantIsolationSubquery()))).
			WithArgs(givenTenant(), givenApplicationID()).
			WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.ListByApplicationID(ctx, givenTenant(), givenApplicationID())
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

		noRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth"})

		dbMock.ExpectQuery("SELECT").WithArgs(givenTenant(), givenApplicationID()).WillReturnRows(noRows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.ListByApplicationID(ctx, givenTenant(), givenApplicationID())
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		sut := webhook.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT").WillReturnError(givenError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.ListByApplicationID(ctx, givenTenant(), givenApplicationID())
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity", mock.Anything).Return(model.Webhook{}, givenError())

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth"}).
			AddRow(givenID(), givenTenant(), givenApplicationID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT")).WithArgs(givenTenant(), givenApplicationID()).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.ListByApplicationID(ctx, givenTenant(), givenApplicationID())
		// THEN
		require.EqualError(t, err, "while converting Webhook to model: some error")
	})
}

func TestRepositoryListByApplicationTemplateID(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity",
			webhook.Entity{ID: givenID(),
				ApplicationTemplateID: repo.NewValidNullableString(givenApplicationTemplateID()),
				Type:                  string(model.WebhookTypeConfigurationChanged),
				URL:                   repo.NewValidNullableString("http://kyma.io")}).
			Return(model.Webhook{
				ID: givenID(),
			}, nil)

		mockConv.On("FromEntity",
			webhook.Entity{ID: anotherID(),
				ApplicationTemplateID: repo.NewValidNullableString(givenApplicationTemplateID()),
				Type:                  string(model.WebhookTypeConfigurationChanged),
				URL:                   repo.NewValidNullableString("http://kyma2.io")}).
			Return(model.Webhook{ID: anotherID()}, nil)

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "app_template_id", "type", "url", "auth"}).
			AddRow(givenID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil).
			AddRow(anotherID(), givenApplicationTemplateID(), model.WebhookTypeConfigurationChanged, "http://kyma2.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, tenant_id, app_id, app_template_id, type, url, auth, runtime_id, integration_system_id, mode, correlation_id_key, retry_interval, timeout, url_template, input_template, header_template, output_template, status_template FROM public.webhooks WHERE app_template_id = $1")).
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

		noRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "app_template_id", "type", "url", "auth"})

		dbMock.ExpectQuery("SELECT").WithArgs(givenApplicationTemplateID()).WillReturnRows(noRows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		actual, err := sut.ListByApplicationTemplateID(ctx, givenApplicationTemplateID())
		// THEN
		require.NoError(t, err)
		require.Empty(t, actual)
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
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

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity", mock.Anything).Return(model.Webhook{}, givenError())

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
*/
