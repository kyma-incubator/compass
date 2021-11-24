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

// RepoDeleteTestSuite represents a generic test suite for repository Delete method of any entity that has externally managed tenants in m2m table/view.
// This test suite is not suitable for global entities or entities with embedded tenant in them.
type RepoDeleteTestSuite struct {
	Name                  string
	SQLQueryDetails       []SQLQueryDetails
	ConverterMockProvider func() Mock
	RepoConstructorFunc   interface{}
	MethodName            string
	MethodArgs            []interface{}
	IsTopLeveEntity       bool
	IsDeleteMany          bool
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

			for _, sqlDetails := range suite.SQLQueryDetails {
				if sqlDetails.IsSelect {
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				} else {
					sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
				}
			}

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
			// THEN
			require.NoError(t, err)

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		if !suite.IsDeleteMany {
			if suite.IsTopLeveEntity {
				t.Run("returns unauthorized if no entity matches criteria", func(t *testing.T) {
					sqlxDB, sqlMock := MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

					for _, sqlDetails := range suite.SQLQueryDetails {
						if sqlDetails.IsSelect {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlmock.NewRows([]string{}))
							break
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
						}
					}

					convMock := suite.ConverterMockProvider()
					pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
					// WHEN
					err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
					// THEN
					require.Error(t, err)
					require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)

					sqlMock.AssertExpectations(t)
					convMock.AssertExpectations(t)
				})
			} else {
				t.Run("returns unauthorized if no entity matches criteria", func(t *testing.T) {
					sqlxDB, sqlMock := MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

					for _, sqlDetails := range suite.SQLQueryDetails {
						if sqlDetails.IsSelect {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlmock.NewResult(-1, 0))
							break
						}
					}

					convMock := suite.ConverterMockProvider()
					pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
					// WHEN
					err := callDelete(pgRepository, ctx, suite.MethodName, suite.MethodArgs)
					// THEN
					require.Error(t, err)
					require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)

					sqlMock.AssertExpectations(t)
					convMock.AssertExpectations(t)
				})
			}

			if suite.IsTopLeveEntity {
				t.Run("returns error if more than one entity matches criteria", func(t *testing.T) {
					sqlxDB, sqlMock := MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

					for _, sqlDetails := range suite.SQLQueryDetails {
						if sqlDetails.IsSelect {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
							break
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
						}
					}

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
			} else {
				t.Run("returns error if more than one entity matches criteria", func(t *testing.T) {
					sqlxDB, sqlMock := MockDatabase(t)
					ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

					for _, sqlDetails := range suite.SQLQueryDetails {
						if sqlDetails.IsSelect {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.InvalidResult)
							break
						}
					}

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
		}

		for i := range suite.SQLQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				for _, sqlDetails := range suite.SQLQueryDetails {
					if sqlDetails.IsSelect {
						if sqlDetails.Query == suite.SQLQueryDetails[i].Query {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(testErr)
							break
						} else {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
						}
					} else {
						if sqlDetails.Query == suite.SQLQueryDetails[i].Query {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(testErr)
							break
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
						}
					}
				}

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
