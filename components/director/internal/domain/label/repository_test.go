package label_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`INSERT INTO "public"."labels" (id, tenant_id, key, value, app_id, runtime_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
    		value = EXCLUDED.value
		`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.Key, labelEntity.Value, labelEntity.AppID, labelEntity.RuntimeID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Upsert(ctx, &labelModel)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`INSERT INTO "public"."labels" (id, tenant_id, key, value, app_id, runtime_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
    		value = EXCLUDED.value
		`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.Key, labelEntity.Value, labelEntity.AppID, labelEntity.RuntimeID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Upsert(ctx, &labelModel)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`INSERT INTO "public"."labels" (id, tenant_id, key, value, app_id, runtime_id) VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET
    		value = EXCLUDED.value
		`)
		dbMock.ExpectExec(escapedQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.Key, labelEntity.Value, labelEntity.AppID, labelEntity.RuntimeID).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Upsert(ctx, &labelModel)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})
	//})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)

		// WHEN
		err := repo.Upsert(context.TODO(), &model.Label{})
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "key" = $1 AND "runtime_id" = $2 AND "tenant_id" = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).AddRow(id, tnt, key, value, nil, objID)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.GetByKey(ctx, tnt, objType, objID, key)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "key" = $1 AND "app_id" = $2 AND "tenant_id" = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).AddRow(id, tnt, key, value, objID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.GetByKey(ctx, tnt, objType, objID, key)
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
		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "key" = $1 AND "app_id" = $2 AND "tenant_id" = $3`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := repo.GetByKey(ctx, tnt, objType, objID, key)
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

		repo := label.NewRepository(nil)
		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "key" = $1 AND "app_id" = $2 AND "tenant_id" = $3`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(key, sql.NullString{Valid: true, String: objID}, tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := repo.GetByKey(ctx, tnt, objType, objID, key)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		_, err := repo.GetByKey(context.TODO(), "tenant", objType, objID, "key")
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "runtime_id" = $1 AND "tenant_id" = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, "foo", "test1", nil, objID).
			AddRow("2", tnt, "bar", "test2", nil, objID)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.ListForObject(ctx, tnt, objType, objID)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE  "app_id" = $1 AND "tenant_id" = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, "foo", "test1", objID, nil).
			AddRow("2", tnt, "bar", "test2", objID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := repo.ListForObject(ctx, tnt, objType, objID)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE  "app_id" = $1 AND "tenant_id" = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objID := "foo"
		tnt := "tenant"

		repo := label.NewRepository(nil)
		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE  "app_id" = $1 AND "tenant_id" = $2`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(sql.NullString{Valid: true, String: objID}, tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := repo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		_, err := repo.ListForObject(context.TODO(), "tenant", objType, objID)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE "key" = $1 AND "tenant_id" = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"}).
			AddRow("1", tnt, labelKey, "test1", nil, rtmObjID).
			AddRow("2", tnt, labelKey, "test2", appObjID, nil)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.ListByKey(ctx, tnt, labelKey)
		// THEN
		require.NoError(t, err)
		assert.ElementsMatch(t, expected, actual)
	})

	t.Run("Error - Doesn't exist", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE  "key" = $1 AND "tenant_id" = $2`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id"})
		dbMock.ExpectQuery(escapedQuery).WithArgs("key", tnt).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := repo.ListByKey(ctx, tnt, "key")
		// THEN
		require.NoError(t, err)
		assert.Empty(t, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"

		repo := label.NewRepository(nil)
		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`SELECT "id", "tenant_id", "key", "value", "app_id", "runtime_id" FROM "public"."labels" WHERE  "key" = $1 AND "tenant_id" = $2`)
		dbMock.ExpectQuery(escapedQuery).WithArgs("key", tnt).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := repo.ListByKey(ctx, tnt, "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "persistence error")
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)

		// WHEN
		_, err := repo.ListByKey(context.TODO(), "tenant", "key")
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "key" = $1 AND "runtime_id" = $2 AND "tenant_id" = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Delete(ctx, tnt, objType, objID, key)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "key" = $1 AND "app_id" = $2 AND "tenant_id" = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Delete(ctx, tnt, objType, objID, key)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "key" = $1 AND "app_id" = $2 AND "tenant_id" = $3`)
		dbMock.ExpectExec(escapedQuery).WithArgs(key, objID, tnt).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.Delete(ctx, tnt, objType, objID, key)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		err := repo.Delete(context.TODO(), "tenant", objType, objID, "key")
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "runtime_id" = $1 AND "tenant_id" = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.DeleteAll(ctx, tnt, objType, objID)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "app_id" = $1 AND "tenant_id" = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.DeleteAll(ctx, tnt, objType, objID)
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

		repo := label.NewRepository(mockConverter)

		db, dbMock := mockDatabase(t)
		defer func() {
			require.NoError(t, dbMock.ExpectationsWereMet())
		}()

		escapedQuery := regexp.QuoteMeta(`DELETE FROM "public"."labels" WHERE "app_id" = $1 AND "tenant_id" = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(objID, tnt).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := repo.DeleteAll(ctx, tnt, objType, objID)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), testErr.Error())
	})

	t.Run("Error - Missing persistence", func(t *testing.T) {
		// GIVEN
		repo := label.NewRepository(nil)
		objType := model.RuntimeLabelableObject
		objID := "foo"

		// WHEN
		err := repo.DeleteAll(context.TODO(), "tenant", objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func mockDatabase(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock) {
	sqlDB, sqlMock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(sqlDB, "sqlmock")
	return sqlxDB, sqlMock
}
