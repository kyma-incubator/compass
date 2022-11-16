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

// RepoListTestSuite represents a generic test suite for repository List without paging method of any global entity or entity that has externally managed tenants in m2m table/view.
type RepoListTestSuite struct {
	Name                      string
	SQLQueryDetails           []SQLQueryDetails
	ConverterMockProvider     func() Mock
	ShouldSkipMockFromEntity  bool
	RepoConstructorFunc       interface{}
	ExpectedModelEntities     []interface{}
	ExpectedDBEntities        []interface{}
	MethodArgs                []interface{}
	AdditionalConverterArgs   []interface{}
	DisableConverterErrorTest bool
	MethodName                string
	DisableEmptySliceTest     bool
}

// Run runs the generic repo list without paging test suite
func (suite *RepoListTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		panic("missing method name")
	}

	for _, queryDetails := range suite.SQLQueryDetails {
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

			configureValidSQLQueries(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			if !suite.ShouldSkipMockFromEntity {
				for i := range suite.ExpectedDBEntities {
					convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntities[i]}, suite.AdditionalConverterArgs...)...).Return(suite.ExpectedModelEntities[i], nil).Once()
				}
			}
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			// THEN
			require.NoError(t, err)
			assertEqualElements(t, suite.ExpectedModelEntities, res)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		if !suite.DisableEmptySliceTest {
			t.Run("returns empty slice when no rows", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureInvalidAllSelectQueries(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				// THEN
				require.NoError(t, err)
				require.Empty(t, res)
				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		for i := range suite.SQLQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureFailureForSQLQueryOnIndex(sqlMock, suite.SQLQueryDetails, i, testErr)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)

				// WHEN
				res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)

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
				convMock.On("FromEntity", append([]interface{}{suite.ExpectedDBEntities[0]}, suite.AdditionalConverterArgs...)...).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				res, err := callList(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
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

func callList(repo interface{}, ctx context.Context, methodName string, args []interface{}) (interface{}, error) {
	argsVals := make([]reflect.Value, 1, len(args)+1)
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

// assertEqualElements checks that elements of two slices are equal.
// It expects the first slice to contain pointer elements and converts each element of the second slice to pointer if it is a struct value before comparison.
func assertEqualElements(t *testing.T, arr1, arr2 interface{}) {
	array1 := reflect.ValueOf(arr1)
	array2 := reflect.ValueOf(arr2)

	require.Equal(t, array2.Len(), array1.Len())
	for i := 0; i < array1.Len(); i++ {
		actual := array2.Index(i).Interface()
		if array2.Index(i).Kind() != reflect.Ptr {
			actual = array2.Index(i).Addr().Interface()
		}
		require.EqualValues(t, array1.Index(i).Interface(), actual)
	}
}

func configureInvalidAllSelectQueries(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
		}
	}
}
