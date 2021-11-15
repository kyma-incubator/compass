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

type RepoListTestSuite struct {
	Name                      string
	SqlQueryDetails           []SqlQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ExpectedModelEntities     []interface{}
	ExpectedDBEntities        []interface{}
	MethodArgs                []interface{}
	AdditionalConverterArgs   []interface{}
	DisableConverterErrorTest bool
	MethodName                string
}

func (suite *RepoListTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		panic("missing method name")
	}
	
	for _, queryDetails := range suite.SqlQueryDetails {
		if !queryDetails.IsSelect {
			panic("list suite should expect only select SQL statements")
		}
	}

	if len(suite.ExpectedDBEntities) != len(suite.ExpectedModelEntities) {
		panic("for each DB entity a corresponding model Entity is expected")
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
			for i := range suite.ExpectedDBEntities {
				convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntities[i]}, suite.AdditionalConverterArgs...)...).Return(suite.ExpectedModelEntities[i], nil).Once()
			}
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			//THEN
			require.NoError(t, err)
			require.ElementsMatch(t, suite.ExpectedModelEntities, res)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns empty slice when no rows", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
			}

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			//THEN
			require.NoError(t, err)
			require.Empty(t, res)
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
				res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)

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
				convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntities[0]}, suite.AdditionalConverterArgs...)...).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				//WHEN
				res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
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

func callList(repo interface{}, ctx context.Context, methodName string, args []interface{}) (interface{}, error) {
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
