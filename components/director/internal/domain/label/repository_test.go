package label_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Upsert(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"

		labelModel := model.Label{
			ID:         "foo",
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
		}
		labelEntity := label.Entity{
			Key:      key,
			Value:    value,
			TenantID: tnt,
			ID:       "foo",
			AppID:    sql.NullString{},
			RuntimeID: sql.NullString{
				String: objID,
				Valid:  true,
			},
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, key, value ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), key ) DO UPDATE SET value=EXCLUDED.value`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.Key, labelEntity.Value).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"

		labelModel := model.Label{
			ID:         "foo",
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
		}
		labelEntity := label.Entity{
			Key:       key,
			Value:     value,
			TenantID:  tnt,
			ID:        "foo",
			RuntimeID: sql.NullString{},
			AppID: sql.NullString{
				String: objID,
				Valid:  true,
			},
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, key, value ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), key ) DO UPDATE SET value=EXCLUDED.value`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.Key, labelEntity.Value).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - Operation", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		testErr := errors.New("Test error")

		labelModel := model.Label{
			ID:         "foo",
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
		}
		labelEntity := label.Entity{
			Key:       key,
			Value:     value,
			TenantID:  tnt,
			ID:        "foo",
			RuntimeID: sql.NullString{},
			AppID: sql.NullString{
				String: objID,
				Valid:  true,
			},
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, key, value ) VALUES ( ?, ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_id, coalesce(app_id, '00000000-0000-0000-0000-000000000000'), coalesce(runtime_id, '00000000-0000-0000-0000-000000000000'), key ) DO UPDATE SET value=EXCLUDED.value`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.Key, labelEntity.Value).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
}

func TestRepository_GetByKey(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "lorem ipsum"
		tnt := "tenant"
		id := "id"

		expected := &model.Label{
			ObjectType: objType,
			ObjectID:   objID,
			Value:      value,
			Key:        key,
			Tenant:     tnt,
			ID:         id,
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity",
			label.Entity{ID: id, TenantID: tnt, Key: key, Value: value, RuntimeID: sql.NullString{Valid: true, String: objID}}).Return(
			model.Label{
				ID:         id,
				Tenant:     tnt,
				Key:        key,
				ObjectID:   objID,
				ObjectType: objType,
				Value:      value}, nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND runtime_id = $2 AND tenant_id = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).AddRow(id, tnt, key, value, nil, objID)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.GetByKey(ctx, tnt, objType, objID, key)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		assert.Equal(t, value, actual.Value)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		value := "lorem ipsum"
		tnt := "tenant"
		id := "id"

		mockConverter := &automock.Converter{}
		expected := &model.Label{
			ObjectType: objType,
			ObjectID:   objID,
			Value:      value,
			Key:        key,
			Tenant:     tnt,
			ID:         id,
		}
		mockConverter.On("FromEntity",
			label.Entity{ID: id, TenantID: tnt, Key: key, Value: value, AppID: sql.NullString{Valid: true, String: objID}}).Return(
			model.Label{
				ID:         id,
				Tenant:     tnt,
				Key:        key,
				ObjectID:   objID,
				ObjectType: objType,
				Value:      value}, nil).Once()

		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND app_id = $2 AND tenant_id = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).AddRow(id, tnt, key, value, objID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.GetByKey(ctx, tnt, objType, objID, key)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
		assert.Equal(t, value, actual.Value)
	})

	t.Run("Error - Doesn't exist", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		tnt := "tenant"
		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND app_id = $2 AND tenant_id = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.GetByKey(ctx, tnt, objType, objID, key)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		tnt := "tenant"

		labelRepo := label.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND app_id = $2 AND tenant_id = $3`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.GetByKey(ctx, tnt, objType, objID, key)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		_, err := labelRepo.GetByKey(context.TODO(), "tenant", objType, objID, "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_ListForObject(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		tnt := "tenant"

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", RuntimeID: sql.NullString{Valid: true, String: objID}},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", RuntimeID: sql.NullString{Valid: true, String: objID}},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE runtime_id = $1 AND tenant_id = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, "foo", "test1", nil, objID).
			AddRow("2", tnt, "bar", "test2", nil, objID)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", AppID: sql.NullString{Valid: true, String: objID}},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", AppID: sql.NullString{Valid: true, String: objID}},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE app_id = $1 AND tenant_id = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, "foo", "test1", objID, nil).
			AddRow("2", tnt, "bar", "test2", objID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error - Doesn't exist", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE app_id = $1 AND tenant_id = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"

		labelRepo := label.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE app_id = $1 AND tenant_id = $2`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		_, err := labelRepo.ListForObject(context.TODO(), "tenant", objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_ListByKey(t *testing.T) {
	t.Run("Success - Label for Application and Runtime", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		labelKey := "foo"
		objType := model.RuntimeLabelableObject
		rtmObjID := "foo"
		appObjID := "bar"
		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: labelKey, Value: "test1", RuntimeID: sql.NullString{Valid: true, String: rtmObjID}},
			{ID: "2", TenantID: tnt, Key: labelKey, Value: "test2", AppID: sql.NullString{Valid: true, String: appObjID}},
		}
		expected := []*model.Label{
			{ID: "1", Tenant: tnt, Key: labelKey, Value: "test1", ObjectType: objType, ObjectID: rtmObjID},
			{ID: "2", Tenant: tnt, Key: labelKey, Value: "test2", ObjectType: model.ApplicationLabelableObject, ObjectID: appObjID},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", inputItems[0]).Return(*expected[0], nil).Once()
		mockConverter.On("FromEntity", inputItems[1]).Return(*expected[1], nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, labelKey, "test1", nil, rtmObjID).
			AddRow("2", tnt, labelKey, "test2", appObjID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListByKey(ctx, tnt, labelKey)
		// THEN
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, actual)
	})

	t.Run("Empty label", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs("key", tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListByKey(ctx, tnt, "key")
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"

		labelRepo := label.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		dbMock.ExpectQuery(escapedQuery).WithArgs("key", tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.ListByKey(ctx, tnt, "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)

		// WHEN
		_, err := labelRepo.ListByKey(context.TODO(), "tenant", "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND runtime_id = $2 AND tenant_id = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tnt, objType, objID, key)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND app_id = $2 AND tenant_id = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tnt, objType, objID, key)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - Operation", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		tnt := "tenant"
		testErr := errors.New("Test err")

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND app_id = $2 AND tenant_id = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tnt, objType, objID, key)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		err := labelRepo.Delete(context.TODO(), "tenant", objType, objID, "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_DeleteAll(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE runtime_id = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success - Label for Application", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE app_id = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - Operation", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"
		testErr := errors.New("Test err")

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE app_id = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tnt, objType, objID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		err := labelRepo.DeleteAll(context.TODO(), "tenant", objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}
func TestRepository_DeleteByKey(t *testing.T) {
	tenant := "tenant"
	key := "key"

	t.Run("Success - Deleted labels", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, tenant).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKey(ctx, tenant, key)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - can't fetch persistence from context", func(t *testing.T) {
		labelRepo := label.NewRepository(nil)
		// WHEN
		err := labelRepo.DeleteByKey(context.TODO(), tenant, key)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})

	t.Run("Error - Operation", func(t *testing.T) {
		// GIVEN
		testErr := errors.New("Test err")

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, tenant).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKey(ctx, tenant, key)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		labelRepo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		err := labelRepo.DeleteAll(context.TODO(), "tenant", objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})

	t.Run("No rows were affected - should succeed", func(t *testing.T) {
		// GIVEN
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND tenant_id = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, tenant).WillReturnResult(sqlmock.NewResult(1, 0))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKey(ctx, tenant, key)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_GetRuntimeScenariosWhereRuntimesLabelsMatchSelector(t *testing.T) {
	query := regexp.QuoteMeta(`SELECT * FROM LABELS AS L WHERE l."key"='scenarios' AND l.tenant_id=$3 AND l.runtime_id in 
					(
				SELECT LA.runtime_id FROM LABELS AS LA WHERE LA."key"=$1 AND value ?| array[$2] AND LA.tenant_id=$3 AND LA.runtime_ID IS NOT NULL
			);`)
	tnt := "tenant"
	selectorKey := "KEY"
	selectorValue := "VALUE"
	labelValue, err := json.Marshal([]string{selectorValue})
	require.NoError(t, err)
	rtmID := "651038e0-e4b6-4036-a32f-f6e9846003f4"
	testErr := errors.New("test error")
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID)

		dbMock.ExpectQuery(query).
			WithArgs(selectorKey, selectorValue, tnt).WillReturnRows(mockedRows)
		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		//WHEN
		_, err = labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, tnt, selectorKey, selectorValue)

		//THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error, while fetch scenarios from database", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		dbMock.ExpectQuery(query).WithArgs(selectorKey, selectorValue, tnt).WillReturnError(testErr)
		labelRepo := label.NewRepository(nil)
		//WHEN
		_, err = labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, tnt, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		dbMock.AssertExpectations(t)
	})

	t.Run("Error, while converting entity to model", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID)

		dbMock.ExpectQuery(query).WithArgs(selectorKey, selectorValue, tnt).WillReturnRows(mockedRows)
		conv := &automock.Converter{}
		conv.On("FromEntity", label.Entity{
			ID:        "id",
			TenantID:  tnt,
			Key:       selectorKey,
			RuntimeID: repo.NewNullableString(&rtmID),
			Value:     string(labelValue),
		}).Return(model.Label{}, nil).Once()
		conv.On("FromEntity", mock.Anything).Return(model.Label{}, testErr).Once()
		labelRepo := label.NewRepository(conv)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		//WHEN
		_, err = labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, tnt, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		dbMock.AssertExpectations(t)
		conv.AssertExpectations(t)
		dbMock.AssertExpectations(t)
	})

	t.Run("Error , no persistence in context", func(t *testing.T) {
		//GIVEN
		labelRepo := label.NewRepository(nil)
		//WHEN
		_, err := labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(context.TODO(), "", "", "")

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "while fetching persistence from context: unable to fetch database from context")
	})
}

func TestRepository_GetRuntimesIDsWhereLabelsMatchSelector(t *testing.T) {
	tenantID := "3c9e9c37-8623-44e2-98c8-5040a94bac63"
	selectorKey := "KEY"
	selectorValue := "VALUE"
	query := regexp.QuoteMeta(`SELECT LA.runtime_id FROM LABELS AS LA WHERE LA."key"=$1 AND value ?| array[$2] AND LA.tenant_id=$3 AND LA.runtime_ID IS NOT NULL;`)
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		rtm1ID := "fd1a54dc-828e-4097-a4cb-40e7e46eb28a"
		rtm2ID := "6c3311a7-339c-4283-955b-ca90eaf5f7b5"
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"runtime_id"}).
			AddRow(rtm1ID).
			AddRow(rtm2ID)

		dbMock.ExpectQuery(query).WithArgs(selectorKey, selectorValue, tenantID).WillReturnRows(mockedRows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		rtmIDs, err := labelRepo.GetRuntimesIDsByStringLabel(ctx, tenantID, selectorKey, selectorValue)

		//THEN
		require.NoError(t, err)
		dbMock.AssertExpectations(t)
		assert.ElementsMatch(t, rtmIDs, []string{rtm1ID, rtm2ID})
	})

	t.Run("Query return error", func(t *testing.T) {
		//GIVEN
		testErr := errors.New("test err")
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(query).WithArgs(selectorKey, selectorValue, tenantID).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetRuntimesIDsByStringLabel(ctx, tenantID, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		dbMock.AssertExpectations(t)
	})

	t.Run("Return error when no persistance in context", func(t *testing.T) {
		labelRepo := label.NewRepository(nil)
		//WHEN
		_, err := labelRepo.GetRuntimesIDsByStringLabel(context.TODO(), tenantID, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while fetching persistence from context")
	})
}

func TestRepository_GetScenarioLabelsForRuntimes(t *testing.T) {
	tenantID := "3c9e9c37-8623-44e2-98c8-5040a94bac63"
	rtm1ID := "fd1a54dc-828e-4097-a4cb-40e7e46eb28a"
	rtm2ID := "6c3311a7-339c-4283-955b-ca90eaf5f7b5"
	rtmIDs := []string{rtm1ID, rtm2ID}
	testErr := errors.New("test error")

	query := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, key, value FROM public.labels WHERE tenant_id=$1 AND key = '%s' AND runtime_id IN ('%s', '%s')`, model.ScenariosKey, rtm1ID, rtm2ID))
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm1ID).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm2ID)

		dbMock.ExpectQuery(query).WithArgs(tenantID).WillReturnRows(mockedRows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		labels, err := labelRepo.GetScenarioLabelsForRuntimes(ctx, tenantID, rtmIDs)

		//THEN
		require.NoError(t, err)
		require.Len(t, labels, 2)
		dbMock.AssertExpectations(t)
	})

	t.Run("Converter returns error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm1ID).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm2ID)

		dbMock.ExpectQuery(query).WithArgs(tenantID).WillReturnRows(mockedRows)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := &automock.Converter{}
		conv.On("FromEntity", mock.Anything).Return(model.Label{}, testErr)
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetScenarioLabelsForRuntimes(ctx, tenantID, rtmIDs)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		dbMock.AssertExpectations(t)
		conv.AssertExpectations(t)
	})

	t.Run("Database returns error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		dbMock.ExpectQuery(query).WithArgs(tenantID).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetScenarioLabelsForRuntimes(ctx, tenantID, rtmIDs)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
		dbMock.AssertExpectations(t)
	})

	t.Run("Database returns error, when runtimesIDs size is 0", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetScenarioLabelsForRuntimes(ctx, tenantID, []string{})

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Cannot execute query without runtimesIDs")
		dbMock.AssertExpectations(t)
	})
}
