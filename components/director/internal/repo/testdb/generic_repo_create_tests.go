package testdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock represents a mockery Mock.
type Mock interface {
	AssertExpectations(t mock.TestingT) bool
	On(methodName string, arguments ...interface{}) *mock.Call
}

// SQLQueryDetails represent an SQL expected query details to provide to the DB mock.
type SQLQueryDetails struct {
	Query               string
	IsSelect            bool
	Args                []driver.Value
	ValidResult         driver.Result
	InvalidResult       driver.Result
	ValidRowsProvider   func() []*sqlmock.Rows
	InvalidRowsProvider func() []*sqlmock.Rows
}

// RepoCreateTestSuite represents a generic test suite for repository Create method of any global entity or entity that has externally managed tenants in m2m table/view.
type RepoCreateTestSuite struct {
	Name                      string
	SQLQueryDetails           []SQLQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ModelEntity               interface{}
	DBEntity                  interface{}
	NilModelEntity            interface{}
	TenantID                  string
	DisableConverterErrorTest bool
	MethodName                string
	IsTopLevelEntity          bool
	IsGlobal                  bool
}

// Run runs the generic repo create test suite
func (suite *RepoCreateTestSuite) Run(t *testing.T) bool {
	if len(suite.MethodName) == 0 {
		suite.MethodName = "Create"
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
			err := callCreate(pgRepository, suite.MethodName, ctx, suite.TenantID, suite.ModelEntity)

			// THEN
			require.NoError(t, err)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		if !suite.IsTopLevelEntity && !suite.IsGlobal {
			t.Run("error when parent access is missing", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureInvalidSelect(sqlMock, suite.SQLQueryDetails)

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callCreate(pgRepository, suite.MethodName, ctx, suite.TenantID, suite.ModelEntity)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), fmt.Sprintf("Tenant %s does not have access to the parent", suite.TenantID))

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
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callCreate(pgRepository, suite.MethodName, ctx, suite.TenantID, suite.ModelEntity)
				// THEN
				require.Error(t, err)
				if suite.SQLQueryDetails[i].IsSelect {
					require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), fmt.Sprintf("Tenant %s does not have access to the parent", suite.TenantID))
				} else {
					require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
					require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")
				}
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
				err := callCreate(pgRepository, suite.MethodName, ctx, suite.TenantID, suite.ModelEntity)
				// THEN
				require.Error(t, err)
				require.Contains(t, err.Error(), testErr.Error())

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		t.Run("returns error when item is nil", func(t *testing.T) {
			ctx := context.TODO()
			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			err := callCreate(pgRepository, suite.MethodName, ctx, suite.TenantID, suite.NilModelEntity)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Internal Server Error")
			convMock.AssertExpectations(t)
		})
	})
}

// callCreate calls the Create method of the given repository.
// In order to do this for all the different repository implementations we need to do it via reflection.
func callCreate(repo interface{}, methodName string, ctx context.Context, tenant string, modelEntity interface{}) error {
	args := []reflect.Value{reflect.ValueOf(ctx)}
	if len(tenant) > 0 {
		args = append(args, reflect.ValueOf(tenant))
	}
	args = append(args, reflect.ValueOf(modelEntity))
	results := reflect.ValueOf(repo).MethodByName(methodName).Call(args)
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

// createRepo creates a new repository by the provided constructor func.
// In order to do this for all the different repository implementations we need to do it via reflection.
func createRepo(repoConstructorFunc interface{}, convMock interface{}) interface{} {
	v := reflect.ValueOf(repoConstructorFunc)
	if v.Kind() != reflect.Func {
		panic("Repo constructor should be a function")
	}
	t := v.Type()

	if t.NumOut() != 1 {
		panic("Repo constructor should return only one argument")
	}

	if t.NumIn() == 0 {
		return v.Call(nil)[0].Interface()
	}

	if t.NumIn() != 1 {
		panic("Repo constructor should accept zero or one arguments")
	}

	mockVal := reflect.ValueOf(convMock)
	return v.Call([]reflect.Value{mockVal})[0].Interface()
}

func configureValidSQLQueries(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
		} else {
			sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
		}
	}
}

func configureInvalidSelect(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
			break
		}
	}
}

func configureFailureForSQLQueryOnIndex(sqlMock DBMock, sqlQueryDetails []SQLQueryDetails, i int, expectedErr error) {
	for _, sqlDetails := range sqlQueryDetails {
		if sqlDetails.IsSelect {
			if sqlDetails.Query == sqlQueryDetails[i].Query {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(expectedErr)
				break
			} else {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
			}
		} else {
			if sqlDetails.Query == sqlQueryDetails[i].Query {
				sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(expectedErr)
				break
			} else {
				sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
			}
		}
	}
}
