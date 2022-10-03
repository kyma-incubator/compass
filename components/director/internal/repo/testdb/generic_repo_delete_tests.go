package testdb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// RepoDeleteTestSuite represents a generic test suite for repository Delete method of any global entity or entity that has externally managed tenants in m2m table/view.
type RepoDeleteTestSuite struct {
	Name                  string
	SQLQueryDetails       []SQLQueryDetails
	ConverterMockProvider func() Mock
	RepoConstructorFunc   interface{}
	MethodName            string
	MethodArgs            []interface{}
	IsDeleteMany          bool
	IsGlobal              bool
}

// Run runs the generic repo delete test suite
func (suite *RepoDeleteTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		suite.MethodName = "Delete"
	}

	return t.Run(suite.Name, func(t *testing.T) {
		testErr := errors.New("test error")

		t.Run("success", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureValidSQLQueries(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			// THEN
			require.NoError(t, err)

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		if !suite.IsDeleteMany { // Single delete requires exactly one row to be deleted
			t.Run("returns error if no entity matches criteria", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureNoEntityDeleted(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				// THEN
				require.Error(t, err)

				if suite.IsGlobal {
					require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), "delete should remove single row, but removed 0 rows")
				} else {
					require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)
				}

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})

			t.Run("returns error if more than one entity matches criteria", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureMoreThanOneEntityDeleted(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "delete should remove single row, but removed")

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
				err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}
	})
}

func callDelete(repo interface{}, ctx context.Context, methodName string, args []interface{}) error {
	argsVals := make([]reflect.Value, 1, len(args))
	argsVals[0] = reflect.ValueOf(ctx)
	for _, arg := range args {
		argsVals = append(argsVals, reflect.ValueOf(arg))
	}
	results := reflect.ValueOf(repo).MethodByName(methodName).Call(argsVals)
	if len(results) != 1 {
		panic("Create should return one argument")
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

func configureMoreThanOneEntityDeleted(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
		} else {
			sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.InvalidResult)
			break
		}
	}
}

func configureNoEntityDeleted(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
		} else {
			sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlmock.NewResult(-1, 0))
			break
		}
	}
}
