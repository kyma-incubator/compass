package bundleinstanceauth_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Create(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
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

		// WHEN
		err := repo.Create(ctx, biaModel)

		// THEN
		assert.NoError(t, err)
	})

	t.Run("Error when item is nil", func(t *testing.T) {
		// GIVEN

		repo := bundleinstanceauth.NewRepository(nil)

		// WHEN
		err := repo.Create(context.TODO(), nil)

		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})

	t.Run("DB Error", func(t *testing.T) {
		// GIVEN
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

		// WHEN
		err := repo.Create(ctx, biaModel)

		// THEN
		expectedError := fmt.Sprintf("while saving entity with id %s to db: Internal Server Error: Unexpected error while executing SQL query", testID)
		require.EqualError(t, err, expectedError)
	})

	t.Run("Converter Error", func(t *testing.T) {
		// GIVEN
		biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
		mockConverter := &automock.EntityConverter{}
		mockConverter.On("ToEntity", biaModel).Return(&bundleinstanceauth.Entity{}, testError)
		defer mockConverter.AssertExpectations(t)

		repo := bundleinstanceauth.NewRepository(mockConverter)

		// WHEN
		err := repo.Create(context.TODO(), biaModel)

		// THEN
		require.EqualError(t, err, "while converting BundleInstanceAuth model to entity: test")
	})
}

func TestRepository_GetByID(t *testing.T) {
	biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
	biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

	suite := testdb.RepoGetTestSuite{
		Name: "Get BIA",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, owner_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE id = $1 AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = $2) OR owner_id = $3)`),
				Args:     []driver.Value{testID, testTenant, testTenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)}),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						fixSQLRows([]sqlRow{}),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundleinstanceauth.NewRepository,
		ExpectedModelEntity: biaModel,
		ExpectedDBEntity:    biaEntity,
		MethodArgs:          []interface{}{testTenant, testID},
	}

	suite.Run(t)
}

func TestRepository_GetForBundle(t *testing.T) {
	biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
	biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

	suite := testdb.RepoGetTestSuite{
		Name: "Get BIA For Bundle",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, owner_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE id = $1 AND bundle_id = $2 AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = $3) OR owner_id = $4)`),
				Args:     []driver.Value{testID, testBundleID, testTenant, testTenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						fixSQLRows([]sqlRow{fixSQLRowFromEntity(*biaEntity)}),
					}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{
						fixSQLRows([]sqlRow{}),
					}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundleinstanceauth.NewRepository,
		ExpectedModelEntity: biaModel,
		ExpectedDBEntity:    biaEntity,
		MethodArgs:          []interface{}{testTenant, testID, testBundleID},
		MethodName:          "GetForBundle",
	}

	suite.Run(t)
}

func TestRepository_ListByBundleID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List BIA by BundleID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, owner_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE bundle_id = $1 AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = $2) OR owner_id = $3)`),
				Args:     []driver.Value{testBundleID, testTenant, testTenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{fixSQLRows([]sqlRow{
						fixSQLRowFromEntity(*fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)),
						fixSQLRowFromEntity(*fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)),
					})}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(testTableColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundleinstanceauth.NewRepository,
		ExpectedModelEntities: []interface{}{fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil), fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)},
		ExpectedDBEntities:    []interface{}{fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil), fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)},
		MethodArgs:            []interface{}{testTenant, testBundleID},
		MethodName:            "ListByBundleID",
	}

	suite.Run(t)
}

func TestRepository_ListByRuntimeID(t *testing.T) {
	suite := testdb.RepoListTestSuite{
		Name: "List BIA by RuntimeID",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:    regexp.QuoteMeta(`SELECT id, owner_id, bundle_id, context, input_params, auth_value, status_condition, status_timestamp, status_message, status_reason, runtime_id, runtime_context_id FROM public.bundle_instance_auths WHERE runtime_id = $1 AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = $2) OR owner_id = $3)`),
				Args:     []driver.Value{testRuntimeID, testTenant, testTenant},
				IsSelect: true,
				ValidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{fixSQLRows([]sqlRow{
						fixSQLRowFromEntity(*fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)),
						fixSQLRowFromEntity(*fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)),
					})}
				},
				InvalidRowsProvider: func() []*sqlmock.Rows {
					return []*sqlmock.Rows{sqlmock.NewRows(testTableColumns)}
				},
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc:   bundleinstanceauth.NewRepository,
		ExpectedModelEntities: []interface{}{fixModelBundleInstanceAuth("foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil), fixModelBundleInstanceAuth("bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)},
		ExpectedDBEntities:    []interface{}{fixEntityBundleInstanceAuth(t, "foo", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil), fixEntityBundleInstanceAuth(t, "bar", testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)},
		MethodArgs:            []interface{}{testTenant, testRuntimeID},
		MethodName:            "ListByRuntimeID",
	}

	suite.Run(t)
}

func TestRepository_Update(t *testing.T) {
	updateStmt := regexp.QuoteMeta(`UPDATE public.bundle_instance_auths SET context = ?, input_params = ?, auth_value = ?, status_condition = ?, status_timestamp = ?, status_message = ?, status_reason = ? WHERE id = ? AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = ? AND owner = true) OR owner_id = ?)`)

	var nilBiaModel *model.BundleInstanceAuth
	biaModel := fixModelBundleInstanceAuth(testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)
	biaEntity := fixEntityBundleInstanceAuth(t, testID, testBundleID, testTenant, fixModelAuth(), fixModelStatusSucceeded(), nil)

	suite := testdb.RepoUpdateTestSuite{
		Name: "Update BIA",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         updateStmt,
				Args:          []driver.Value{biaEntity.Context, biaEntity.InputParams, biaEntity.AuthValue, biaEntity.StatusCondition, biaEntity.StatusTimestamp, biaEntity.StatusMessage, biaEntity.StatusReason, testID, testTenant, testTenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 0),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundleinstanceauth.NewRepository,
		ModelEntity:         biaModel,
		DBEntity:            biaEntity,
		NilModelEntity:      nilBiaModel,
		TenantID:            testTenant,
	}

	suite.Run(t)
}

func TestRepository_Delete(t *testing.T) {
	suite := testdb.RepoDeleteTestSuite{
		Name: "Delete BIA",
		SQLQueryDetails: []testdb.SQLQueryDetails{
			{
				Query:         regexp.QuoteMeta(`DELETE FROM public.bundle_instance_auths WHERE id = $1 AND (id IN (SELECT id FROM bundle_instance_auths_tenants WHERE tenant_id = $2 AND owner = true) OR owner_id = $3)`),
				Args:          []driver.Value{testID, testTenant, testTenant},
				ValidResult:   sqlmock.NewResult(-1, 1),
				InvalidResult: sqlmock.NewResult(-1, 2),
			},
		},
		ConverterMockProvider: func() testdb.Mock {
			return &automock.EntityConverter{}
		},
		RepoConstructorFunc: bundleinstanceauth.NewRepository,
		MethodArgs:          []interface{}{testTenant, testID},
	}

	suite.Run(t)
}
