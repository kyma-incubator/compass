package packageinstanceauth_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth/automock"
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
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(regexp.QuoteMeta(`INSERT INTO public.package_instance_auths ( id, tenant_id, package_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason ) VALUES ( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? )`)).
			WithArgs(fixCreateArgs(*piaEntity)...).
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := packageinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, piaModel)

		// then
		assert.NoError(t, err)
	})

	t.Run("Error when item is nil", func(t *testing.T) {
		// given

		repo := packageinstanceauth.NewRepository(nil)

		// when
		err := repo.Create(context.Background(), nil)

		// then
		require.EqualError(t, err, "item cannot be nil")
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec("INSERT INTO .*").WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := packageinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(ctx, piaModel)

		// then
		require.EqualError(t, err, "while saving entity to db: while inserting row to 'public.package_instance_auths' table: test")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(packageinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := packageinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Create(context.TODO(), piaModel)

		// then
		require.EqualError(t, err, "while converting PackageInstanceAuth model to entity: test")
	})
}

func TestRepository_GetByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(*piaModel, nil).Once()
		defer mockConverter.AssertExpectations(t)

		repo := packageinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.package_instance_auths WHERE tenant_id = \$1 AND id = \$2$`).
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
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, nil, fixModelStatusPending())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(model.PackageInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := packageinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.package_instance_auths WHERE tenant_id = \$1 AND id = \$2$`).
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
		repo := packageinstanceauth.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(testTenant, testID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repo.GetByID(ctx, testTenant, testID)

		// then
		require.EqualError(t, err, "while getting object from DB: test")
	})
}

func TestRepository_GetForPackage(t *testing.T) {
	// given
	piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
	piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

	selectQuery := `^SELECT (.+) FROM public.package_instance_auths WHERE tenant_id = \$1 AND id = \$2 AND package_id = \$3`

	t.Run("success", func(t *testing.T) {
		sqlxDB, sqlMock := testdb.MockDatabase(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		sqlMock.ExpectQuery(selectQuery).
			WithArgs(testTenant, testID, testPackageID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), sqlxDB)
		convMock := &automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntity).Return(*piaModel, nil).Once()
		repo := packageinstanceauth.NewRepository(convMock)

		// WHEN
		actual, err := repo.GetForPackage(ctx, testTenant, testID, testPackageID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, actual)
		assert.Equal(t, piaModel, actual)
		convMock.AssertExpectations(t)
		sqlMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		repo := packageinstanceauth.NewRepository(nil)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectQuery("SELECT .*").
			WithArgs(testTenant, testID, testPackageID).WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		_, err := repo.GetForPackage(ctx, testTenant, testID, testPackageID)

		// then
		require.EqualError(t, err, "while getting object from DB: test")
	})

	t.Run("returns error when conversion failed", func(t *testing.T) {
		// given
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, nil, fixModelStatusPending())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("FromEntity", *piaEntity).Return(model.PackageInstanceAuth{}, testError).Once()
		defer mockConverter.AssertExpectations(t)

		repo := packageinstanceauth.NewRepository(mockConverter)
		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		rows := fixSQLRows([]sqlRow{fixSQLRowFromEntity(*piaEntity)})

		dbMock.ExpectQuery(`^SELECT (.+) FROM public.package_instance_auths WHERE tenant_id = \$1 AND id = \$2 AND package_id = \$3`).
			WithArgs(testTenant, testID, testPackageID).
			WillReturnRows(rows)

		ctx := persistence.SaveToContext(context.TODO(), db)

		// when
		actual, err := repo.GetForPackage(ctx, testTenant, testID, testPackageID)

		// then
		require.Error(t, err)
		assert.Contains(t, err.Error(), testError.Error())
		require.Nil(t, actual)
	})
}

func TestRepository_ListByPackageID(t *testing.T) {
	//GIVEN
	t.Run("Success", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		piaModels := []*model.PackageInstanceAuth{
			fixModelPackageInstanceAuth("foo", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixModelPackageInstanceAuth("bar", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}
		piaEntities := []*packageinstanceauth.Entity{
			fixEntityPackageInstanceAuth(t, "foo", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixEntityPackageInstanceAuth(t, "bar", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, package_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.package_instance_auths WHERE tenant_id=$1 AND package_id = '%s'`, testPackageID)
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*piaEntities[0]),
				fixSQLRowFromEntity(*piaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntities[0]).Return(*piaModels[0], nil).Once()
		convMock.On("FromEntity", *piaEntities[1]).Return(*piaModels[1], nil).Once()
		pgRepository := packageinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByPackageID(ctx, testTenant, testPackageID)

		//THEN
		require.NoError(t, err)
		require.NotNil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("Converter Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		piaModels := []*model.PackageInstanceAuth{
			fixModelPackageInstanceAuth("foo", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixModelPackageInstanceAuth("bar", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}
		piaEntities := []*packageinstanceauth.Entity{
			fixEntityPackageInstanceAuth(t, "foo", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
			fixEntityPackageInstanceAuth(t, "bar", testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded()),
		}

		query := fmt.Sprintf(`SELECT id, tenant_id, package_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.package_instance_auths WHERE tenant_id=$1 AND package_id = '%s'`, testPackageID)
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant).
			WillReturnRows(fixSQLRows([]sqlRow{
				fixSQLRowFromEntity(*piaEntities[0]),
				fixSQLRowFromEntity(*piaEntities[1]),
			}))

		convMock := automock.EntityConverter{}
		convMock.On("FromEntity", *piaEntities[0]).Return(*piaModels[0], nil).Once()
		convMock.On("FromEntity", *piaEntities[1]).Return(*piaModels[1], testError).Once()
		pgRepository := packageinstanceauth.NewRepository(&convMock)

		//WHEN
		result, err := pgRepository.ListByPackageID(ctx, testTenant, testPackageID)

		//THEN
		require.EqualError(t, err, "while creating PackageInstanceAuth model from entity: test")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
		convMock.AssertExpectations(t)
	})

	t.Run("DB Error", func(t *testing.T) {
		db, dbMock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)

		query := fmt.Sprintf(`SELECT id, tenant_id, package_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason FROM public.package_instance_auths WHERE tenant_id=$1 AND package_id = '%s'`, testPackageID)
		dbMock.ExpectQuery(regexp.QuoteMeta(query)).
			WithArgs(testTenant).
			WillReturnError(testError)

		pgRepository := packageinstanceauth.NewRepository(nil)

		//WHEN
		result, err := pgRepository.ListByPackageID(ctx, testTenant, testPackageID)

		//THEN
		require.EqualError(t, err, "while fetching list of objects from DB: test")
		require.Nil(t, result)
		dbMock.AssertExpectations(t)
	})
}

func TestRepository_Update(t *testing.T) {
	updateStmt := `UPDATE public\.package_instance_auths SET auth_value = \?, status_condition = \?, status_timestamp = \?, status_message = \?, status_reason = \? WHERE tenant_id = \? AND id = \?`

	t.Run("Success", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WithArgs().
			WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := packageinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, piaModel)

		// then
		assert.NoError(t, err)
	})

	t.Run("Error when item is nil", func(t *testing.T) {
		// given

		repo := packageinstanceauth.NewRepository(nil)

		// when
		err := repo.Update(context.Background(), nil)

		// then
		require.EqualError(t, err, "item cannot be nil")
	})

	t.Run("DB Error", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		piaEntity := fixEntityPackageInstanceAuth(t, testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())

		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(*piaEntity, nil).Once()
		defer mockConverter.AssertExpectations(t)

		db, dbMock := testdb.MockDatabase(t)
		defer dbMock.AssertExpectations(t)

		dbMock.ExpectExec(updateStmt).
			WillReturnError(testError)

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := packageinstanceauth.NewRepository(mockConverter)

		// when
		err := repo.Update(ctx, piaModel)

		// then
		require.EqualError(t, err, "while updating single entity: test")
	})

	t.Run("Converter Error", func(t *testing.T) {
		// given
		piaModel := fixModelPackageInstanceAuth(testID, testPackageID, testTenant, fixModelAuth(), fixModelStatusSucceeded())
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", *piaModel).Return(packageinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := packageinstanceauth.NewRepository(mockConverter)

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

		dbMock.ExpectExec(regexp.QuoteMeta("DELETE FROM public.package_instance_auths WHERE tenant_id = $1 AND id = $2")).WithArgs(
			testTenant, testID).WillReturnResult(sqlmock.NewResult(-1, 1))

		ctx := persistence.SaveToContext(context.TODO(), db)
		repo := packageinstanceauth.NewRepository(nil)

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
		repo := packageinstanceauth.NewRepository(nil)

		// when
		err := repo.Delete(ctx, testTenant, testID)

		// then
		require.EqualError(t, err, "while deleting from database: test")
	})
}
