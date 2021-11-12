package bundleinstanceauth_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
		biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", biaModel).Return(biaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.bundle_instance_auths ( id, owner_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixCreateArgs(*biaEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, biaModel)

		// then
		assert.NoError(t, err)
	})

	t.Run("Error when item is nil", func(t *testing.T) {
		// given

		repo := bundleinstanceauth.NewRepository(nil)

		// when
		err := repo.Create(context.TODO(), nil)

		// then
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
		biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", biaModel).Return(biaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, biaModel)

		// then
		expectedError := fmt.Sprintf("while saving entity with id %s to db: Internal Server Error: Unexpected error while executing SQL query", testID)
		require.EqualError(t, err, expectedError)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", biaModel).Return(&bundleinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(context.TODO(), biaModel)

		// then
		require.EqualError(t, err, "while converting BundleInstanceAuth model to entity: test")
	})
}

/*
func TestRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
		biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *biaEntity).Return(*biaModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)})

		dbMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.bundle_instance_auths WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
			WithArgs(testTenant, testID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetByID(ctx, testTenant, testID)

		// then
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, biaModel, actual)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *biaEntity).Return(model.BundleInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)})

		dbMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.bundle_instance_auths WHERE %s AND id = \$2$`, fixTenantIsolationSubquery())).
			WithArgs(testTenant, testID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetByID(ctx, testTenant, testID)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, actual)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := bundleinstanceauth.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(testTenant, testID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repo.GetByID(ctx, testTenant, testID)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestRepository_GetForBundle(t *testing.T) {
	// given
	biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
	biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

	selectQuery := fmt.Sprintf(`^SELECT (.+) FROM public.bundle_instance_auths WHERE %s AND id = \$2 AND bundle_id = \$3`, fixTenantIsolationSubquery())

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)})

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(testTenant, testID, testBundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", *biaEntity).Return(*biaModel, nil).Once()
		repo := bundleinstanceauth.NewRepository(convMock)

		// WHEN
		actual, err := repo.GetForBundle(ctx, testTenant, testID, testBundleID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, biaModel, actual)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := bundleinstanceauth.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(testTenant, testID, testBundleID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repo.GetForBundle(ctx, testTenant, testID, testBundleID)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		// given
		biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, nil, fixModelStatusPending(), nil)

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *biaEntity).Return(model.BundleInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)})

		dbMock.ExpectQuery(fmt.Sprintf(`^SELECT (.+) FROM public.bundle_instance_auths WHERE %s AND id = \$2 AND bundle_id = \$3`, fixTenantIsolationSubquery())).
			WithArgs(testTenant, testID, testBundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetForBundle(ctx, testTenant, testID, testBundleID)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, actual)
	})
}

func TestRepository_ListByBundleID(t *testing.T) {
	//GIVEN
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		biaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
		}
		biaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND bundle_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testBundleID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*biaEntities[0]),
				fixSQLRowFromEntity(*biaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *biaEntities[0]).Return(*biaModels[0], nil).Once()
		convMock.On("FromEntity", *biaEntities[1]).Return(*biaModels[1], nil).Once()
		pgRepository := bundleinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByBundleID(ctx, testTenant, testBundleID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		biaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
		}
		biaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND bundle_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testBundleID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*biaEntities[0]),
				fixSQLRowFromEntity(*biaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *biaEntities[0]).Return(*biaModels[0], nil).Once()
		convMock.On("FromEntity", *biaEntities[1]).Return(*biaModels[1], testError).Once()
		pgRepository := bundleinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByBundleID(ctx, testTenant, testBundleID)

		//THEN
		require.EqualError(t, err, "while creating BundleInstanceAuth model from entity: test")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND bundle_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testBundleID).
			WillReturnError(testError)

		pgRepository := bundleinstanceauth.NewRepository(nil)

		//WHEN
		result, err := pgRepository.ListByBundleID(ctx, testTenant, testBundleID)

		//THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
	})
}

func TestRepository_ListByRuntimeID(t *testing.T) {
	//GIVEN
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		biaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID),
		}
		biaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND runtime_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testRuntimeID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*biaEntities[0]),
				fixSQLRowFromEntity(*biaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *biaEntities[0]).Return(*biaModels[0], nil).Once()
		convMock.On("FromEntity", *biaEntities[1]).Return(*biaModels[1], nil).Once()
		pgRepository := bundleinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByRuntimeID(ctx, testTenant, testRuntimeID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		biaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID),
		}
		biaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), &testRuntimeID),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND runtime_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testRuntimeID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*biaEntities[0]),
				fixSQLRowFromEntity(*biaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *biaEntities[0]).Return(*biaModels[0], nil).Once()
		convMock.On("FromEntity", *biaEntities[1]).Return(*biaModels[1], testError).Once()
		pgRepository := bundleinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByRuntimeID(ctx, testTenant, testRuntimeID)

		//THEN
		require.EqualError(t, err, "while creating BundleInstanceAuth model from entity: test")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := fmt.Sprintf(`SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE %s AND runtime_id = $2`, fixUnescapedTenantIsolationSubquery())
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testRuntimeID).
			WillReturnError(testError)

		pgRepository := bundleinstanceauth.NewRepository(nil)

		//WHEN
		result, err := pgRepository.ListByRuntimeID(ctx, testTenant, testRuntimeID)

		//THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
	})
}

*/

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(fmt.Sprintf(`UPDATE public.bundle_instance_auths SET auth_value = ?, status_condition = ?, status_timestamp = ?, status_message = ?, status_reason = ? WHERE id = ? AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = '%s' AND owner = true) OR owner_id = ?)`, testTenant))

	var nilBiaModel *model.BundleInstanceAuth
	biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
	biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update BIA",
		SqlQueryDetails: []testdb.SqlQueryDetails{
			{
				Query:       updateStmt,
				Args:        []driver.Value{biaEntity.AuthValue, biaEntity.StatusCondition, biaEntity.StatusTimestamp, biaEntity.StatusMessage, biaEntity.StatusReason, testID, testTenant},
				ValidResult: sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:       bundleinstanceauth.NewRepository,
		ModelEntity:               biaModel,
		DBEntity:                  biaEntity,
		NilModelEntity:            nilBiaModel,
		TenantID:                  testTenant,
	}

	suite.Run(t)
}

/*
func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM public.bundle_instance_auths WHERE %s AND id = $2", fixUnescapedTenantIsolationSubquery()))).WithArgs(
			testTenant, testID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(nil)

		// when
		err := repo.Delete(ctx, testTenant, testID)

		// then
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("DELETE FROM .*").WithArgs(
			testTenant, testID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(nil)

		// when
		err := repo.Delete(ctx, testTenant, testID)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
*/
