package bundleinstanceauth_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.bundle_instance_auths ( id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixCreateArgs(*piaEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, piaModel)

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
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, piaModel)

		// then
		expectedError := fmt.Sprintf("while saving entity with id %s to db: Internal Server Error: Unexpected error while executing SQL query", testID)
		require.EqualError(t, err, expectedError)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(bundleinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(context.TODO(), piaModel)

		// then
		require.EqualError(t, err, "while converting BundleInstanceAuth model to entity: test")
	})
}

func TestRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(*piaModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.bundle_instance_auths WHERE tenant_id = \$1 AND id = \$2$`).
			WithArgs(testTenant, testID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetByID(ctx, testTenant, testID)

		// then
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, piaModel, actual)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, nil, fixModelStatusPending())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(model.BundleInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.bundle_instance_auths WHERE tenant_id = \$1 AND id = \$2$`).
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
	piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
	piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

	selectQuery := `^SELECT (.+) FROM public.bundle_instance_auths WHERE tenant_id = \$1 AND id = \$2 AND bundle_id = \$3`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(testTenant, testID, testBundleID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntity).Return(*piaModel, nil).Once()
		repo := bundleinstanceauth.NewRepository(convMock)

		// WHEN
		actual, err := repo.GetForBundle(ctx, testTenant, testID, testBundleID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, piaModel, actual)
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
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, nil, fixModelStatusPending())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(model.BundleInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.bundle_instance_auths WHERE tenant_id = \$1 AND id = \$2 AND bundle_id = \$3`).
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

		piaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}
		piaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}

		query := `SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.bundle_instance_auths WHERE tenant_id = $1 AND bundle_id = $2`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testBundleID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*piaEntities[0]),
				fixSQLRowFromEntity(*piaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntities[0]).Return(*piaModels[0], nil).Once()
		convMock.On("FromEntity", *piaEntities[1]).Return(*piaModels[1], nil).Once()
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

		piaModels := []*model.BundleInstanceAuth{
			fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}
		piaEntities := []*bundleinstanceauth.Entity{
			fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}

		query := `SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.bundle_instance_auths WHERE tenant_id = $1 AND bundle_id = $2`
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant, testBundleID).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*piaEntities[0]),
				fixSQLRowFromEntity(*piaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntities[0]).Return(*piaModels[0], nil).Once()
		convMock.On("FromEntity", *piaEntities[1]).Return(*piaModels[1], testError).Once()
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

		query := `SELECT id, tenant_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.bundle_instance_auths WHERE tenant_id = $1 AND bundle_id = $2`
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

func TestRepository_Update(t *testing.T) {
	updateStmt := `UPDATE public\.bundle_instance_auths SET auth_value = \?, status_condition = \?, status_timestamp = \?, status_message = \?, status_reason = \? WHERE tenant_id = \? AND id = \?`

	t.Run("Success", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WithArgs().
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, piaModel)

		// then
		assert.NoError(t, err)
	})

	t.Run("Error when item is nil", func(t *testing.T) {
		// given

		repo := bundleinstanceauth.NewRepository(nil)

		// when
		err := repo.Update(context.TODO(), nil)

		// then
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, piaModel)

		// then
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		piaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(bundleinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Update(context.TODO(), piaModel)

		// then
		require.EqualError(t, err, "while converting model to entity: test")
	})
}

func TestRepository_Delete(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.bundle_instance_auths WHERE tenant_id = $1 AND id = $2")).WithArgs(
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
