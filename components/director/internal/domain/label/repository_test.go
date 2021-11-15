package label_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	var nilLabelModel *model.Label
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create Application Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_applications WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{appLabelEntity.ID, appLabelEntity.TenantID, appLabelEntity.AppID, appLabelEntity.RuntimeID, appLabelEntity.RuntimeContextID, appLabelEntity.Key, appLabelEntity.Value, appLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         applabelModel,
		DBEntity:            appLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
	}

	runtimeLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create Runtime Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM tenant_runtimes WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{runtimeLabelEntity.ID, runtimeLabelEntity.TenantID, runtimeLabelEntity.AppID, runtimeLabelEntity.RuntimeID, runtimeLabelEntity.RuntimeContextID, runtimeLabelEntity.Key, runtimeLabelEntity.Value, runtimeLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimelabelModel,
		DBEntity:            runtimeLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
	}

	runtimeCtxLabelSuite := testdb.RepoCreateTestSuite{
		Name: "Create RuntimeCtx Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM runtime_contexts_tenants WHERE tenant_id = $1 AND id = $2 AND owner = $3"),
				Args:     []driver.Value{tenantID, refID, true},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:       regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )"),
				Args:        []driver.Value{runtimeCtxLabelEntity.ID, runtimeCtxLabelEntity.TenantID, runtimeCtxLabelEntity.AppID, runtimeCtxLabelEntity.RuntimeID, runtimeCtxLabelEntity.RuntimeContextID, runtimeCtxLabelEntity.Key, runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.Version},
				ValidResult: sqlmock.NewResult(-1, 1),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimeCtxlabelModel,
		DBEntity:            runtimeCtxLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
	}

	appLabelSuite.Run(t)
	runtimeLabelSuite.Run(t)
	runtimeCtxLabelSuite.Run(t)

	// Additional tests - tenant labels are created globally as the tenant is embedded in the entity.
	t.Run("Success create - Label for Tenant", func(t *testing.T) {
		// GIVEN
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Create(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})
}

/*
func TestRepository_Upsert(t *testing.T) {
	t.Run("Success update - Label for Runtime", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:              key,
			Value:            value,
			TenantID:         tnt,
			ID:               id,
			AppID:            sql.NullString{},
			RuntimeContextID: sql.NullString{},
			RuntimeID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).AddRow(id, tnt, key, value, nil, objID, nil, version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedUpdateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE %s AND id = ?`, fixUpdateTenantIsolationSubquery()))
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.TenantID, labelEntity.ID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Runtime Context", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeContextLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:       key,
			Value:     value,
			TenantID:  tnt,
			ID:        id,
			AppID:     sql.NullString{},
			RuntimeID: sql.NullString{},
			RuntimeContextID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_context_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).AddRow(id, tnt, key, value, nil, nil, objID, version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedUpdateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE %s AND id = ?`, fixUpdateTenantIsolationSubquery()))
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.TenantID, labelEntity.ID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success update - Label for Application", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:              key,
			Value:            value,
			TenantID:         tnt,
			ID:               id,
			RuntimeID:        sql.NullString{},
			RuntimeContextID: sql.NullString{},
			AppID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		mockConverter.On("FromEntity", labelEntity).Return(labelModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND app_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).AddRow(id, tnt, key, value, objID, nil, nil, version)
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedUpdateQuery := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE %s AND id = ?`, fixUpdateTenantIsolationSubquery()))
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.TenantID, labelEntity.ID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Runtime", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:              key,
			Value:            value,
			TenantID:         tnt,
			ID:               id,
			AppID:            sql.NullString{},
			RuntimeContextID: sql.NullString{},
			RuntimeID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{})
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Runtime Context", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeContextLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:       key,
			Value:     value,
			TenantID:  tnt,
			ID:        id,
			AppID:     sql.NullString{},
			RuntimeID: sql.NullString{},
			RuntimeContextID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_context_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{})
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Success create - Label for Application", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.ApplicationLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:              key,
			Value:            value,
			TenantID:         tnt,
			ID:               id,
			RuntimeID:        sql.NullString{},
			RuntimeContextID: sql.NullString{},
			AppID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND app_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{})
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.NoError(t, err)
	})

	t.Run("Error - GetByKey", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		testErr := errors.New("Test error")
		version := 42

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}

		labelRepo := label.NewRepository(&automock.Converter{})

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id = $3`, fixUnescapedTenantIsolationSubquery()))
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Error - Create", func(t *testing.T) {
		// GIVEN
		id := "foo"
		objType := model.RuntimeLabelableObject
		objID := "foo"
		key := "test"
		value := "test"
		tnt := "tenant"
		version := 42
		testErr := errors.New("Test error")

		labelModel := model.Label{
			ID:         id,
			ObjectType: objType,
			ObjectID:   objID,
			Tenant:     tnt,
			Key:        key,
			Value:      value,
			Version:    version,
		}
		labelEntity := label.Entity{
			Key:              key,
			Value:            value,
			TenantID:         tnt,
			ID:               id,
			AppID:            sql.NullString{},
			RuntimeContextID: sql.NullString{},
			RuntimeID: sql.NullString{
				String: objID,
				Valid:  true,
			},
			Version: version,
		}

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedGetQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id = $3`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{})
		dbMock.ExpectQuery(escapedGetQuery).WithArgs(tnt, key, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		escapedInsertQuery := regexp.QuoteMeta("INSERT INTO public.labels ( id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ? )")
		dbMock.ExpectExec(escapedInsertQuery).WithArgs(labelEntity.ID, labelEntity.TenantID, labelEntity.AppID, labelEntity.RuntimeID, labelEntity.RuntimeContextID, labelEntity.Key, labelEntity.Value, labelEntity.Version).WillReturnError(testErr)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Upsert(ctx, &labelModel)
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

*/
func TestRepository_UpdateWithVersion(t *testing.T) {
	version := 42

	var nilLabelModel *model.Label
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Application Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelId, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          []driver.Value{appLabelEntity.Value, appLabelEntity.ID, version},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         applabelModel,
		DBEntity:            appLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	runtimeLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update Runtime Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelId, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          []driver.Value{runtimeLabelEntity.Value, runtimeLabelEntity.ID, version},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimelabelModel,
		DBEntity:            runtimeLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	runtimeCtxLabelSuite := testdb.RepoUpdateTestSuite{
		Name: "Update RuntimeCtx Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta("SELECT 1 FROM public.labels WHERE id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2 AND owner = true))"),
				Args:     []driver.Value{labelId, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectExist()}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{testdb.RowWhenObjectDoesNotExist()}
				},
			},
			{
				Query:         regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = '%s' AND owner = true))`, tenantID)),
				Args:          []driver.Value{runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.ID, version},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ModelEntity:         runtimeCtxlabelModel,
		DBEntity:            runtimeCtxLabelEntity,
		NilModelEntity:      nilLabelModel,
		TenantID:            tenantID,
		UpdateMethodName:    "UpdateWithVersion",
	}

	appLabelSuite.Run(t)
	runtimeLabelSuite.Run(t)
	runtimeCtxLabelSuite.Run(t)

	// Additional tests - tenant labels are updated globally as the tenant is embedded in the entity.
	t.Run("Success update - Label for Tenant", func(t *testing.T) {
		// GIVEN
		labelModel := fixModelLabel(model.TenantLabelableObject)
		labelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		mockConverter.On("ToEntity", labelModel).Return(labelEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedUpdateQuery := regexp.QuoteMeta(`UPDATE public.labels SET value = ?, version = version+1 WHERE id = ? AND version = ? AND tenant_id = ?`)
		dbMock.ExpectExec(escapedUpdateQuery).WithArgs(labelEntity.Value, labelEntity.ID, version, labelEntity.TenantID).WillReturnResult(sqlmock.NewResult(1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.UpdateWithVersion(ctx, tenantID, labelModel)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_GetByKey(t *testing.T) {
	applabelModel := fixModelLabel(model.ApplicationLabelableObject)
	appLabelEntity := fixEntityLabel(model.ApplicationLabelableObject)
	runtimelabelModel := fixModelLabel(model.RuntimeLabelableObject)
	runtimeLabelEntity := fixEntityLabel(model.RuntimeLabelableObject)
	runtimeCtxlabelModel := fixModelLabel(model.RuntimeContextLabelableObject)
	runtimeCtxLabelEntity := fixEntityLabel(model.RuntimeContextLabelableObject)

	appLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Application Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(appLabelEntity.ID, appLabelEntity.TenantID, appLabelEntity.AppID, appLabelEntity.RuntimeID, appLabelEntity.RuntimeContextID, appLabelEntity.Key, appLabelEntity.Value, appLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: applabelModel,
		ExpectedDBEntity:    appLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	rtLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(runtimeLabelEntity.ID, runtimeLabelEntity.TenantID, runtimeLabelEntity.AppID, runtimeLabelEntity.RuntimeID, runtimeLabelEntity.RuntimeContextID, runtimeLabelEntity.Key, runtimeLabelEntity.Value, runtimeLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: runtimelabelModel,
		ExpectedDBEntity:    runtimeLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	rtCtxLabelSuite := testdb.RepoGetTestSuite{
		Name: "Get Runtime Context Label",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3))`),
				Args:     []driver.Value{key, refID, tenantID},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns).
						AddRow(runtimeCtxLabelEntity.ID, runtimeCtxLabelEntity.TenantID, runtimeCtxLabelEntity.AppID, runtimeCtxLabelEntity.RuntimeID, runtimeCtxLabelEntity.RuntimeContextID, runtimeCtxLabelEntity.Key, runtimeCtxLabelEntity.Value, runtimeCtxLabelEntity.Version)}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(fixColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		ExpectedModelEntity: runtimeCtxlabelModel,
		ExpectedDBEntity:    runtimeCtxLabelEntity,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, key},
		MethodName:          "GetByKey",
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		tenantlabelModel := fixModelLabel(model.TenantLabelableObject)
		tenantLabelEntity := fixEntityLabel(model.TenantLabelableObject)

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", tenantLabelEntity).Return(tenantlabelModel, nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE tenant_id = $1 AND key = $2`))
		mockedRows := sqlmock.NewRows(fixColumns).AddRow(tenantLabelEntity.ID, tenantLabelEntity.TenantID, tenantLabelEntity.AppID, tenantLabelEntity.RuntimeID, tenantLabelEntity.RuntimeContextID, tenantLabelEntity.Key, tenantLabelEntity.Value, tenantLabelEntity.Version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tenantID, key).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.GetByKey(ctx, tenantID, model.TenantLabelableObject, tenantID, key)
		// THEN
		require.NoError(t, err)
		require.Equal(t, tenantlabelModel, actual)
		require.Equal(t, value, actual.Value)
	})
}

/*
func TestRepository_ListForObject(t *testing.T) {
	t.Run("Success - Label for Runtime", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeLabelableObject
		objID := "foo"
		tnt := "tenant"
		version := 42

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", RuntimeID: sql.NullString{Valid: true, String: objID}, Version: version},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", RuntimeID: sql.NullString{Valid: true, String: objID}, Version: version},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID, Version: version},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID, Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND runtime_id = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tnt, "foo", "test1", nil, objID, nil, version).
			AddRow("2", tnt, "bar", "test2", nil, objID, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Runtime Context", func(t *testing.T) {
		// GIVEN
		objType := model.RuntimeContextLabelableObject
		objID := "foo"
		tnt := "tenant"
		version := 42

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", RuntimeContextID: sql.NullString{Valid: true, String: objID}, Version: version},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", RuntimeContextID: sql.NullString{Valid: true, String: objID}, Version: version},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID, Version: version},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID, Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND runtime_context_id = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tnt, "foo", "test1", nil, nil, objID, version).
			AddRow("2", tnt, "bar", "test2", nil, nil, objID, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

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
		version := 42

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", AppID: sql.NullString{Valid: true, String: objID}, Version: version},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", AppID: sql.NullString{Valid: true, String: objID}, Version: version},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID, Version: version},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID, Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND app_id = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tnt, "foo", "test1", objID, nil, nil, version).
			AddRow("2", tnt, "bar", "test2", objID, nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		// GIVEN
		objType := model.TenantLabelableObject
		objID := "foo"
		tnt := "tenant"
		version := 42

		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: "foo", Value: "test1", Version: version},
			{ID: "2", TenantID: tnt, Key: "bar", Value: "test2", Version: version},
		}
		expected := map[string]*model.Label{
			"foo": {ID: "1", Tenant: tnt, Key: "foo", Value: "test1", ObjectType: objType, ObjectID: objID, Version: version},
			"bar": {ID: "2", Tenant: tnt, Key: "bar", Value: "test2", ObjectType: objType, ObjectID: objID, Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			mockConverter.On("FromEntity", entity).Return(*expected[entity.Key], nil).Once()
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND tenant_id = $2 AND app_id IS NULL AND runtime_context_id IS NULL AND runtime_id IS NULL`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tnt, "foo", "test1", nil, nil, nil, version).
			AddRow("2", tnt, "bar", "test2", nil, nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

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

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND app_id = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnRows(mockedRows)

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

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND app_id = $2`, fixUnescapedTenantIsolationSubquery()))
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, sql.NullString{Valid: true, String: objID}).WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.ListForObject(ctx, tnt, objType, objID)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
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
	t.Run("Success - Label for Application, Runtime and Runtime Context", func(t *testing.T) {
		// GIVEN
		tnt := "tenant"
		labelKey := "foo"
		objType := model.RuntimeLabelableObject
		rtmObjID := "foo"
		rtmCtxObjID := "foo"
		appObjID := "bar"
		version := 42
		inputItems := []label.Entity{
			{ID: "1", TenantID: tnt, Key: labelKey, Value: "test1", RuntimeID: sql.NullString{Valid: true, String: rtmObjID}, Version: version},
			{ID: "2", TenantID: tnt, Key: labelKey, Value: "test2", AppID: sql.NullString{Valid: true, String: appObjID}, Version: version},
			{ID: "3", TenantID: tnt, Key: labelKey, Value: "test3", RuntimeContextID: sql.NullString{Valid: true, String: rtmCtxObjID}, Version: version},
		}
		expected := []*model.Label{
			{ID: "1", Tenant: tnt, Key: labelKey, Value: "test1", ObjectType: objType, ObjectID: rtmObjID, Version: version},
			{ID: "2", Tenant: tnt, Key: labelKey, Value: "test2", ObjectType: model.ApplicationLabelableObject, ObjectID: appObjID, Version: version},
			{ID: "3", Tenant: tnt, Key: labelKey, Value: "test3", ObjectType: model.RuntimeContextLabelableObject, ObjectID: rtmCtxObjID, Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		mockConverter.On("FromEntity", inputItems[0]).Return(*expected[0], nil).Once()
		mockConverter.On("FromEntity", inputItems[1]).Return(*expected[1], nil).Once()
		mockConverter.On("FromEntity", inputItems[2]).Return(*expected[2], nil).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tnt, labelKey, "test1", nil, rtmObjID, nil, version).
			AddRow("2", tnt, labelKey, "test2", appObjID, nil, nil, version).
			AddRow("3", tnt, labelKey, "test3", nil, nil, rtmCtxObjID, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, labelKey).WillReturnRows(mockedRows)

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

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2`, fixUnescapedTenantIsolationSubquery()))
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"})
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, "key").WillReturnRows(mockedRows)

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

		escapedQuery := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2`, fixUnescapedTenantIsolationSubquery()))
		dbMock.ExpectQuery(escapedQuery).WithArgs(tnt, "key").WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		_, err := labelRepo.ListByKey(ctx, tnt, "key")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
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

func TestRepository_ListGlobalByKeyAndObjects(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		tenantIDs := []string{"tenant1", "tenant2", "tenant3"}
		labelKey := "key"
		version := 42

		inputItems := []label.Entity{
			{ID: "1", TenantID: tenantIDs[0], Key: labelKey, Value: "test1", AppID: sql.NullString{Valid: true, String: objIDs[0]}, Version: version},
			{ID: "2", TenantID: tenantIDs[1], Key: labelKey, Value: "test2", AppID: sql.NullString{Valid: true, String: objIDs[1]}, Version: version},
			{ID: "3", TenantID: tenantIDs[2], Key: labelKey, Value: "test3", AppID: sql.NullString{Valid: true, String: objIDs[2]}, Version: version},
		}
		expected := []*model.Label{
			{ID: "1", Tenant: tenantIDs[0], Key: labelKey, Value: "test1", ObjectType: objType, ObjectID: objIDs[0], Version: version},
			{ID: "2", Tenant: tenantIDs[1], Key: labelKey, Value: "test2", ObjectType: objType, ObjectID: objIDs[1], Version: version},
			{ID: "3", Tenant: tenantIDs[2], Key: labelKey, Value: "test3", ObjectType: objType, ObjectID: objIDs[2], Version: version},
		}

		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)
		for _, entity := range inputItems {
			for i := range expected {
				if expected[i].ObjectID == entity.AppID.String {
					mockConverter.On("FromEntity", entity).Return(*expected[i], nil).Once()
				}
			}
		}

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4)`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tenantIDs[0], labelKey, "test1", objIDs[0], nil, nil, version).
			AddRow("2", tenantIDs[1], labelKey, "test2", objIDs[1], nil, nil, version).
			AddRow("3", tenantIDs[2], labelKey, "test3", objIDs[2], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		actual, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("Error - Select error", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		labelKey := "key"

		labelRepo := label.NewRepository(nil)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4)`)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).
			WillReturnError(errors.New("persistence error"))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "Unexpected error while executing SQL query")
	})

	t.Run("Error - Converting entity", func(t *testing.T) {
		// GIVEN
		objType := model.ApplicationLabelableObject
		objIDs := []string{"foo", "bar", "baz"}
		tenantIDs := []string{"tenant1", "tenant2", "tenant3"}
		labelKey := "key"
		testErr := errors.New("test error")
		version := 42

		mockConverter := &automock.Converter{}
		mockConverter.On("FromEntity", mock.Anything).Return(model.Label{}, testErr).Once()

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND app_id IN ($2, $3, $4)`)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("1", tenantIDs[0], labelKey, "test1", objIDs[0], nil, nil, version)
		dbMock.ExpectQuery(escapedQuery).WithArgs(labelKey, sql.NullString{Valid: true, String: objIDs[0]}, sql.NullString{Valid: true, String: objIDs[1]}, sql.NullString{Valid: true, String: objIDs[2]}).WillReturnRows(mockedRows)

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)

		// WHEN
		_, err := labelRepo.ListGlobalByKeyAndObjects(ctx, objType, objIDs, labelKey)
		// THEN
		require.Error(t, err)
	})
}

*/
func TestRepository_Delete(t *testing.T) {
	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label Delete",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{key, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, key},
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1 AND key = $2`)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID, key).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.Delete(ctx, tenantID, model.TenantLabelableObject, tenantID, key)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_DeleteAll(t *testing.T) {
	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label Delete All",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE app_id = $1 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label Delete All",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE runtime_id = $1 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label Delete All",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE runtime_context_id = $1 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID},
		MethodName:          "DeleteAll",
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1`)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteAll(ctx, tenantID, model.TenantLabelableObject, tenantID)
		// THEN
		require.NoError(t, err)
	})
}

func TestRepository_DeleteByKey(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Label Delete By Key",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE key = $1 AND (id IN (SELECT id FROM labels_tenants WHERE tenant_id = $2 AND owner = true))`),
				Args:          []driver.Value{key, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, key},
		MethodName:          "DeleteByKey",
		IsDeleteMany:        true,
	}

	suite.Run(t)
}

func TestRepository_DeleteByKeyNegationPattern(t *testing.T) {
	pattern := "pattern"

	appLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "App Label DeleteByKeyNegationPattern",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND app_id = $2 AND (id IN (SELECT id FROM application_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.ApplicationLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	rtLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Label DeleteByKeyNegationPattern",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND runtime_id = $2 AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	rtCtxLabelSuite := testdb.RepoDeleteTestSuite{
		Name: "Runtime Context Label DeleteByKeyNegationPattern",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.labels WHERE NOT key ~ $1 AND runtime_context_id = $2 AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $3 AND owner = true))`),
				Args:          []driver.Value{pattern, refID, tenantID},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.Converter{}
		},
		RepoConstructorFunc: label.NewRepository,
		MethodArgs:          []interface{}{tenantID, model.RuntimeContextLabelableObject, refID, pattern},
		MethodName:          "DeleteByKeyNegationPattern",
		IsDeleteMany:        true,
	}

	appLabelSuite.Run(t)
	rtLabelSuite.Run(t)
	rtCtxLabelSuite.Run(t)

	t.Run("Success - Label for Tenant", func(t *testing.T) {
		mockConverter := &automock.Converter{}
		defer mockConverter.AssertExpectations(t)

		labelRepo := label.NewRepository(mockConverter)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		escapedQuery := regexp.QuoteMeta(`DELETE FROM public.labels WHERE tenant_id = $1 AND NOT key ~ $2 `)
		dbMock.ExpectExec(escapedQuery).WithArgs(tenantID, pattern).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := context.TODO()
		ctx = persistence.SaveToContext(ctx, db)
		// WHEN
		err := labelRepo.DeleteByKeyNegationPattern(ctx, tenantID, model.TenantLabelableObject, tenantID, pattern)
		// THEN
		require.NoError(t, err)
	})
}

/*
func TestRepository_GetRuntimeScenariosWhereLabelsMatchSelector(t *testing.T) {

	Query:    regexp.QuoteMeta(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE key = $1 AND runtime_id IN
						(SELECT runtime_id FROM public.labels WHERE key = $2 AND value ?| array[$3] AND runtime_id IS NOT NULL AND (id IN (SELECT id FROM runtime_labels_tenants WHERE tenant_id = $4)))
						AND (id IN (SELECT id FROM runtime_contexts_labels_tenants WHERE tenant_id = $5`),
					Args:     []driver.Value{key, key, value, tenantID, tenantID},

	query := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id IN
					(SELECT runtime_id FROM public.labels WHERE %s AND key = $4 AND value ?| array[$5] AND runtime_id IS NOT NULL)`, fixUnescapedTenantIsolationSubqueryWithArg(1), fixUnescapedTenantIsolationSubqueryWithArg(3)))
	tnt := "tenant"
	selectorKey := "KEY"
	selectorValue := "VALUE"
	labelValue, err := json.Marshal([]string{selectorValue})
	require.NoError(t, err)
	rtmID := "651038e0-e4b6-4036-a32f-f6e9846003f4"
	version := 42
	testErr := errors.New("test error")
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID, nil, version).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID, nil, version)

		dbMock.ExpectQuery(query).
			WithArgs(tnt, "scenarios", tnt, selectorKey, selectorValue).WillReturnRows(mockedRows)
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
		dbMock.ExpectQuery(query).WithArgs(tnt, "scenarios", tnt, selectorKey, selectorValue).WillReturnError(testErr)
		labelRepo := label.NewRepository(nil)
		//WHEN
		_, err = labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(ctx, tnt, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
	})

	t.Run("Error, while converting entity to model", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID, nil, version).
			AddRow("id", tnt, selectorKey, labelValue, nil, rtmID, nil, version)

		dbMock.ExpectQuery(query).WithArgs(tnt, "scenarios", tnt, selectorKey, selectorValue).WillReturnRows(mockedRows)
		conv := &automock.Converter{}
		conv.On("FromEntity", label.Entity{
			ID:        "id",
			TenantID:  tnt,
			Key:       selectorKey,
			RuntimeID: repo.NewNullableString(&rtmID),
			Value:     string(labelValue),
			Version:   version,
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
		_, err := labelRepo.GetRuntimeScenariosWhereLabelsMatchSelector(context.TODO(), "tenant", "", "")

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Internal Server Error: unable to fetch database from context")
	})
}

func TestRepository_GetRuntimesIDsWhereLabelsMatchSelector(t *testing.T) {
	tenantID := "3c9e9c37-8623-44e2-98c8-5040a94bac63"
	selectorKey := "KEY"
	selectorValue := "VALUE"
	query := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND value ?| array[$3] AND runtime_id IS NOT NULL`, fixUnescapedTenantIsolationSubquery()))
	t.Run("Success", func(t *testing.T) {
		//GIVEN
		rtm1ID := "fd1a54dc-828e-4097-a4cb-40e7e46eb28a"
		rtm2ID := "6c3311a7-339c-4283-955b-ca90eaf5f7b5"
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"runtime_id"}).
			AddRow(rtm1ID).
			AddRow(rtm2ID)

		dbMock.ExpectQuery(query).WithArgs(tenantID, selectorKey, selectorValue).WillReturnRows(mockedRows)
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
		dbMock.ExpectQuery(query).WithArgs(tenantID, selectorKey, selectorValue).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetRuntimesIDsByStringLabel(ctx, tenantID, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
		dbMock.AssertExpectations(t)
	})

	t.Run("Return error when no persistence in context", func(t *testing.T) {
		labelRepo := label.NewRepository(nil)
		//WHEN
		_, err := labelRepo.GetRuntimesIDsByStringLabel(context.TODO(), tenantID, selectorKey, selectorValue)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to fetch database from context")
	})
}

func TestRepository_GetScenarioLabelsForRuntimes(t *testing.T) {
	tenantID := "3c9e9c37-8623-44e2-98c8-5040a94bac63"
	rtm1ID := "fd1a54dc-828e-4097-a4cb-40e7e46eb28a"
	rtm2ID := "6c3311a7-339c-4283-955b-ca90eaf5f7b5"
	rtmIDs := []string{rtm1ID, rtm2ID}
	testErr := errors.New("test error")
	version := 42

	query := regexp.QuoteMeta(fmt.Sprintf(`SELECT id, tenant_id, app_id, runtime_id, runtime_context_id, key, value, version FROM public.labels WHERE %s AND key = $2 AND runtime_id IN ($3, $4)`, fixUnescapedTenantIsolationSubquery()))
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm1ID, nil, version).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm2ID, nil, version)

		dbMock.ExpectQuery(query).WithArgs(tenantID, model.ScenariosKey, rtm1ID, rtm2ID).WillReturnRows(mockedRows)
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
		mockedRows := sqlmock.NewRows([]string{"id", "tenant_id", "key", "value", "app_id", "runtime_id", "runtime_context_id", "version"}).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm1ID, nil, version).
			AddRow("id", tenantID, model.ScenariosKey, `["DEFAULT","FOO"]`, nil, rtm2ID, nil, version)

		dbMock.ExpectQuery(query).WithArgs(tenantID, model.ScenariosKey, rtm1ID, rtm2ID).WillReturnRows(mockedRows)
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
		dbMock.ExpectQuery(query).WithArgs(tenantID, model.ScenariosKey, rtm1ID, rtm2ID).WillReturnError(testErr)
		ctx := persistence.SaveToContext(context.TODO(), db)

		conv := label.NewConverter()
		labelRepo := label.NewRepository(conv)
		//WHEN
		_, err := labelRepo.GetScenarioLabelsForRuntimes(ctx, tenantID, rtmIDs)

		//THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
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
		assert.EqualError(t, err, "Invalid data [reason=cannot execute query without runtimeIDs]")
		dbMock.AssertExpectations(t)
	})
}
*/
