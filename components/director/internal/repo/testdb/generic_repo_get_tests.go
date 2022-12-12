package testdb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// RepoGetTestSuite represents a generic test suite for repository Get method of any global entity or entity that has externally managed tenants in m2m table/view.
// This test suite is not suitable entities with embedded tenant in them.
type RepoGetTestSuite struct {
	Name                                  string
	SQLQueryDetails                       []SQLQueryDetails
	ConverterMockProvider                 func() Mock
	RepoConstructorFunc                   interface{}
	ExpectedModelEntity                   interface{}
	ExpectedDBEntity                      interface{}
	MethodArgs                            []interface{}
	AdditionalConverterArgs               []interface{}
	DisableConverterErrorTest             bool
	MethodName                            string
	ExpectNotFoundError                   bool
	AfterNotFoundErrorSQLQueryDetails     []SQLQueryDetails
	AfterNotFoundErrorExpectedModelEntity interface{}
	AfterNotFoundErrorExpectedDBEntity    interface{}
}

// Run runs the generic repo get test suite
func (suite *RepoGetTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		suite.MethodName = "GetByID"
	}

	for _, queryDetails := range suite.SQLQueryDetails {
		if !queryDetails.IsSelect {
			panic("get suite should expect only select SQL statements")
		}
	}

	return t.Run(suite.Name, func(t *testing.T) {
		testErr := errors.New("test error")

		t.Run("success", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureValidSQLQueries(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntity}, suite.AdditionalConverterArgs...)...).Return(suite.ExpectedModelEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			// THEN
			require.NoError(t, err)
			require.Equal(t, suite.ExpectedModelEntity, res)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns not found error when no rows", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureInvalidSelect(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			if suite.ExpectNotFoundError {
				convMock.On("FromEntity", append([]interface{}{suite.AfterNotFoundErrorExpectedDBEntity}, suite.AdditionalConverterArgs...)...).Return(suite.AfterNotFoundErrorExpectedModelEntity, nil).Once()
				configureValidSQLQueries(sqlMock, suite.AfterNotFoundErrorSQLQueryDetails)
			}
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			// THEN
			if !suite.ExpectNotFoundError {
				require.Error(t, err)
				require.Equal(t, apperrors.NotFound, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), apperrors.NotFoundMsg)
				require.Nil(t, res)
				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			} else {
				require.NoError(t, err)
				require.Equal(t, suite.AfterNotFoundErrorExpectedModelEntity, res)
				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			}

		})

		for i := range suite.SQLQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureFailureForSQLQueryOnIndex(sqlMock, suite.SQLQueryDetails, i, testErr)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)

				// WHEN
				res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)

				// THEN
				require.Nil(t, res)

				require.Error(t, err)
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		if !suite.DisableConverterErrorTest {
			t.Run("error when conversion fail", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureValidSQLQueries(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntity}, suite.AdditionalConverterArgs...)...).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				// THEN
				require.Nil(t, res)

				require.Error(t, err)
				require.Contains(t, err.Error(), testErr.Error())

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}
	})
}

func callGet(repo interface{}, ctx context.Context, methodName string, args []interface{}) (interface{}, error) {
	argsVals := make([]reflect.Value, 1, len(args))
	argsVals[0] = reflect.ValueOf(ctx)
	for _, arg := range args {
		argsVals = append(argsVals, reflect.ValueOf(arg))
	}
	results := reflect.ValueOf(repo).MethodByName(methodName).Call(argsVals)
	if len(results) != 2 {
		panic("Get should return two argument")
	}

	errResult := results[1].Interface()
	if errResult == nil {
		return results[0].Interface(), nil
	}
	err, ok := errResult.(error)
	if !ok {
		panic("Expected result to be an error")
	}
	return results[0].Interface(), err
}
