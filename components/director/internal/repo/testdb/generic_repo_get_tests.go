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

type RepoGetTestSuite struct {
	Name                      string
	SqlQueryDetails           []SqlQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ExpectedModelEntity       interface{}
	ExpectedDBEntity          interface{}
	MethodArgs                []interface{}
	AdditionalConverterArgs   []interface{}
	DisableConverterErrorTest bool
	MethodName                string
}

func (suite *RepoGetTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		suite.MethodName = "GetByID"
	}

	for _, queryDetails := range suite.SqlQueryDetails {
		if !queryDetails.IsSelect {
			panic("get suite should expect only select SQL statements")
		}
	}

	return t.Run(suite.Name, func(t *testing.T) {
		testErr := errors.New("test error")

		t.Run("success", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
			}

			convMock := suite.ConverterMockProvider()
			convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntity}, suite.AdditionalConverterArgs...)...).Return(suite.ExpectedModelEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			//THEN
			require.NoError(t, err)
			require.Equal(t, suite.ExpectedModelEntity, res)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns not found error when no rows", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
			}

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			//THEN
			require.Error(t, err)
			require.Equal(t, apperrors.NotFound, apperrors.ErrorCode(err))
			require.Contains(t, err.Error(), apperrors.NotFoundMsg)
			require.Nil(t, res)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		for i := range suite.SqlQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				for _, sqlDetails := range suite.SqlQueryDetails {
					if sqlDetails.Query == suite.SqlQueryDetails[i].Query {
						sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(testErr)
						break
					} else {
						sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
					}
				}

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)

				//WHEN
				res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)

				//THEN
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

				for _, sqlDetails := range suite.SqlQueryDetails {
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				}

				convMock := suite.ConverterMockProvider()
				convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntity}, suite.AdditionalConverterArgs...)...).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				//WHEN
				res, err := callGet(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				//THEN
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
