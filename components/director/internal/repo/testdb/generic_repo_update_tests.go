package testdb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RepoUpdateTestSuite represents a generic test suite for repository Update and UpdateWithVersion methods of any global entity or entity that has externally managed tenants in m2m table/view.
type RepoUpdateTestSuite struct {
	Name                      string
	SQLQueryDetails           []SQLQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ModelEntity               interface{}
	DBEntity                  interface{}
	NilModelEntity            interface{}
	TenantID                  string
	DisableConverterErrorTest bool
	UpdateMethodName          string
	IsGlobal                  bool
}

// Run runs the generic repo update test suite
func (suite *RepoUpdateTestSuite) Run(t *testing.T) bool {
	if len(suite.UpdateMethodName) == 0 {
		suite.UpdateMethodName = "Update"
	}

	return t.Run(suite.Name, func(t *testing.T) {
		testErr := errors.New("test error")

		t.Run("success", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureValidSQLQueries(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
			// THEN
			require.NoError(t, err)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		for i := range suite.SQLQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureFailureForSQLQueryOnIndex(sqlMock, suite.SQLQueryDetails, i, testErr)

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		if suite.UpdateMethodName == "UpdateWithVersion" { // We have an exists check
			t.Run("error when entity is missing", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureInvalidSelect(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.InvalidOperation, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "entity does not exist or caller tenant does not have owner access")

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		if !suite.DisableConverterErrorTest {
			t.Run("error when conversion fail", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				// THEN
				require.Error(t, err)
				require.Contains(t, err.Error(), testErr.Error())

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		t.Run("error if 0 affected rows", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureInvalidUpdateSQLQuery(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
			// THEN
			require.Error(t, err)
			if suite.UpdateMethodName == "UpdateWithVersion" {
				require.Equal(t, apperrors.ConcurrentUpdate, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), apperrors.ConcurrentUpdateMsg)
			} else if suite.IsGlobal {
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), fmt.Sprintf(apperrors.ShouldUpdateSingleRowButUpdatedMsgF, 0))
			} else {
				require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)
			}

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns error when item is nil", func(t *testing.T) {
			ctx := context.TODO()
			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callUpdate(pgRepository, ctx, suite.TenantID, suite.NilModelEntity, suite.UpdateMethodName)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Internal Server Error")
			convMock.AssertExpectations(t)
		})
	})
}

func callUpdate(repo interface{}, ctx context.Context, tenant string, modelEntity interface{}, methodName string) error {
	args := []reflect.Value{reflect.ValueOf(ctx)}
	if len(tenant) > 0 {
		args = append(args, reflect.ValueOf(tenant))
	}
	args = append(args, reflect.ValueOf(modelEntity))
	results := reflect.ValueOf(repo).MethodByName(methodName).Call(args)
	if len(results) != 1 {
		panic("Update should return one argument")
	}
	result := results[0].Interface()
	if result == nil {
		return nil
	}
	err, ok := result.(error)
	if !ok {
		panic("Expected result to be an error")
	}
	return err
}

func configureInvalidUpdateSQLQuery(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
		} else {
			sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.InvalidResult)
		}
	}
}
