package webhook_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testCaseSuccess                  = "success"
	testCaseSuccessWithAuth          = "success with auth"
	testCaseErrorOnConvertingObjects = "got error on converting object"
	testCaseErrorOnDBCommunication   = "got error on db communication"
)

func TestRepositoryGetByID(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", givenEntity()).Return(givenModel(), nil)

		sut := webhook.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, tenant_id, app_id, type, url, auth FROM public.webhooks WHERE tenant_id = $1 AND id = $2")).
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

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", givenAuthAsAString(t))

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, tenant_id, app_id, type, url, auth FROM public.webhooks WHERE tenant_id = $1 AND id = $2")).
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
		mockConverter.On("FromEntity", givenEntity()).Return(model.Webhook{}, givenError())

		sut := webhook.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"}).AddRow(
			givenID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil)

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
		require.EqualError(t, err, "while getting object from DB: some error")
	})

}

func TestRepositoryCreate(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(givenEntity(), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("INSERT INTO public.webhooks ( id, tenant_id, app_id, type, url, auth ) VALUES ( ?, ?, ?, ?, ?, ? )")).WithArgs(
			givenID(), givenTenant(), givenApplicationID(), string(model.WebhookTypeConfigurationChanged), "http://kyma.io", nil).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, ptr(givenModel()))
		// THEN
		require.NoError(t, err)
	})

	t.Run(testCaseSuccessWithAuth, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModelWithAuth()).Return(givenEntityWithAuth(t), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("INSERT INTO public.webhooks ( id, tenant_id, app_id, type, url, auth ) VALUES ( ?, ?, ?, ?, ?, ? )")).WithArgs(
			givenID(), givenTenant(), givenApplicationID(), string(model.WebhookTypeConfigurationChanged), "http://kyma.io", givenAuthAsAString(t)).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, ptr(givenModelWithAuth()))
		// THEN
		require.NoError(t, err)
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(givenEntity(), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(ctx, ptr(givenModel()))
		// THEN
		require.EqualError(t, err, "while inserting row to 'public.webhooks' table: some error")
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(webhook.Entity{}, givenError())

		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Create(context.TODO(), ptr(givenModel()))
		// THEN
		require.EqualError(t, err, "while converting model to entity: some error")
	})
}

func TestRepositoryCreateMany(t *testing.T) {
	const expectedInsert = "INSERT INTO public.webhooks ( id, tenant_id, app_id, type, url, auth ) VALUES ( ?, ?, ?, ?, ?, ? )"
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
			"one", "", "", "", "", nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(
			"two", "", "", "", "", nil).WillReturnResult(sqlmock.NewResult(-1, 1))
		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WithArgs(
			"three", "", "", "", "", nil).WillReturnResult(sqlmock.NewResult(-1, 1))

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

		given := []*model.Webhook{{ID: "one", URL: "unlucky", Type: model.WebhookTypeConfigurationChanged}, {ID: "two"}, {ID: "three"}}
		mockConverter.On("ToEntity", *given[0]).Return(webhook.Entity{}, givenError())

		ctx := persistence.SaveToContext(context.TODO(), nil)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating many webhooks: while converting model to entity: some error")
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)

		given := []*model.Webhook{{ID: "one", URL: "unlucky", Type: model.WebhookTypeConfigurationChanged}, {ID: "two"}, {ID: "three"}}
		mockConverter.On("ToEntity", *given[0]).Return(webhook.Entity{ID: "one"}, nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(expectedInsert)).WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.CreateMany(ctx, given)
		// THEN
		require.EqualError(t, err, "while creating many webhooks: while inserting row to 'public.webhooks' table: some error")
	})
}

func TestRepositoryUpdate(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(givenEntity(), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("UPDATE public.webhooks SET type = ?, url = ?, auth = ? WHERE tenant_id = ? AND id = ? AND app_id = ?")).WithArgs(
			string(model.WebhookTypeConfigurationChanged), "http://kyma.io", nil, givenTenant(), givenID(), givenApplicationID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Update(ctx, ptr(givenModel()))
		// THEN
		require.NoError(t, err)
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(webhook.Entity{}, givenError())

		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Update(context.TODO(), ptr(givenModel()))
		// THEN
		require.EqualError(t, err, "while converting model to entity: some error")
	})

	t.Run(testCaseErrorOnDBCommunication, func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.EntityConverter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("ToEntity", givenModel()).Return(givenEntity(), nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("UPDATE .*").WillReturnError(givenError())

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(mockConverter)
		// WHEN
		err := sut.Update(ctx, ptr(givenModel()))
		// THEN
		require.EqualError(t, err, "while updating single entity: some error")
	})
}

func TestRepositoryDelete(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.webhooks WHERE tenant_id = $1 AND id = $2")).WithArgs(
			givenTenant(), givenID()).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		sut := webhook.NewRepository(nil)
		// WHEN
		err := sut.Delete(ctx, givenTenant(), givenID())
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
		sut := webhook.NewRepository(nil)
		// WHEN
		err := sut.Delete(ctx, givenTenant(), givenID())
		// THEN
		require.EqualError(t, err, "while deleting from database: some error")
	})
}

func TestRepositoryDeleteAllByApplicationID(t *testing.T) {
	sut := webhook.NewRepository(nil)
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.webhooks WHERE tenant_id = $1 AND app_id = $2")).WithArgs(
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
		require.EqualError(t, err, "while deleting from database: some error")
	})
}

func TestRepositoryListByApplicationID(t *testing.T) {
	t.Run(testCaseSuccess, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity",
			webhook.Entity{ID: givenID(),
				TenantID: givenTenant(),
				AppID:    givenApplicationID(),
				Type:     string(model.WebhookTypeConfigurationChanged),
				URL:      "http://kyma.io"}).
			Return(model.Webhook{
				ID: givenID(),
			}, nil)

		mockConv.On("FromEntity",
			webhook.Entity{ID: anotherID(),
				TenantID: givenTenant(),
				AppID:    givenApplicationID(),
				Type:     string(model.WebhookTypeConfigurationChanged),
				URL:      "http://kyma2.io"}).
			Return(model.Webhook{ID: anotherID()}, nil)

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"}).
			AddRow(givenID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil).
			AddRow(anotherID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma2.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT id, tenant_id, app_id, type, url, auth FROM public.webhooks WHERE tenant_id=$1 AND app_id = 'cccccccc-cccc-cccc-cccc-cccccccccccc'")).WithArgs(givenTenant()).WillReturnRows(rows)
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

		noRows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"})

		dbMock.ExpectQuery("SELECT").WithArgs(givenTenant()).WillReturnRows(noRows)
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
		require.EqualError(t, err, "while fetching list of objects from DB: some error")
	})

	t.Run(testCaseErrorOnConvertingObjects, func(t *testing.T) {
		// GIVEN
		mockConv := &automock.EntityConverter{}
		defer mockConv.AssertExpectations(t)
		mockConv.On("FromEntity", mock.Anything).Return(model.Webhook{}, givenError())

		sut := webhook.NewRepository(mockConv)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "app_id", "type", "url", "auth"}).
			AddRow(givenID(), givenTenant(), givenApplicationID(), model.WebhookTypeConfigurationChanged, "http://kyma.io", nil)

		dbMock.ExpectQuery(regexp.QuoteMeta("SELECT")).WithArgs(givenTenant()).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		// WHEN
		_, err := sut.ListByApplicationID(ctx, givenTenant(), givenApplicationID())
		// THEN
		require.EqualError(t, err, "while converting Webhook to model: some error")
	})
}

func givenID() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func anotherID() string {
	return "dddddddd-dddd-dddd-dddd-dddddddddddd"

}

func givenTenant() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func givenApplicationID() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}

func givenEntity() webhook.Entity {
	return webhook.Entity{
		ID:       givenID(),
		TenantID: givenTenant(),
		AppID:    givenApplicationID(),
		Type:     string(model.WebhookTypeConfigurationChanged),
		URL:      "http://kyma.io",
	}
}

func givenEntityWithAuth(t *testing.T) webhook.Entity {
	e := givenEntity()
	e.Auth = sql.NullString{Valid: true, String: givenAuthAsAString(t)}
	return e
}

func givenAuthAsAString(t *testing.T) string {
	b, err := json.Marshal(givenBasicAuth())
	require.NoError(t, err)
	return string(b)
}

func givenModel() model.Webhook {
	return model.Webhook{
		ID:            givenID(),
		Tenant:        givenTenant(),
		ApplicationID: givenApplicationID(),
		Type:          model.WebhookTypeConfigurationChanged,
		URL:           "http://kyma.io",
	}
}

func givenModelWithAuth() model.Webhook {
	m := givenModel()
	m.Auth = givenBasicAuth()
	return m
}

func ptr(in model.Webhook) *model.Webhook {
	return &in
}

func givenError() error {
	return errors.New("some error")
}
