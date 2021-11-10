package testdb

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

// Mock represents a mockery Mock.
type Mock interface {
	AssertExpectations(t mock.TestingT) bool
	On(methodName string, arguments ...interface{}) *mock.Call
}

// SqlQueryDetails represent an SQL expected query details to provide to the DB mock.
type SqlQueryDetails struct {
	Query               string
	IsSelect            bool
	Args                []driver.Value
	ValidResult         driver.Result
	InvalidResult       driver.Result
	ValidRowsProvider   func() []*sqlmock.Rows
	InvalidRowsProvider func() []*sqlmock.Rows
}

// RepoCreateTestSuite represents a generic test suite for all entity repo create methods.
type RepoCreateTestSuite struct {
	Name                      string
	SqlQueryDetails           []SqlQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ModelEntity               interface{}
	DBEntity                  interface{}
	NilModelEntity            interface{}
	TenantID                  string
	DisableConverterErrorTest bool
	IsTopLevelEntity          bool
}

// Run runs the generic repo create test suite
func (suite *RepoCreateTestSuite) Run(t *testing.T) bool {
	return t.Run(suite.Name, func(t *testing.T) {
		testErr := errors.New("test error")

		t.Run("success", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				if sqlDetails.IsSelect {
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				} else {
					sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
				}
			}

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := suite.createRepo(convMock)
			//WHEN
			err := callCreate(pgRepository, ctx, suite.TenantID, suite.ModelEntity)
			//THEN
			require.NoError(t, err)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		if !suite.IsTopLevelEntity {
			t.Run("error when parent access is missing", func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				for _, sqlDetails := range suite.SqlQueryDetails {
					if sqlDetails.IsSelect {
						sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
						break
					}
				}

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := suite.createRepo(convMock)
				//WHEN
				err := callCreate(pgRepository, ctx, suite.TenantID, suite.ModelEntity)
				//THEN
				require.Error(t, err)
				require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), fmt.Sprintf("Tenant %s does not have access to the parent", suite.TenantID))

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		for i := range suite.SqlQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				for _, sqlDetails := range suite.SqlQueryDetails {
					if sqlDetails.IsSelect {
						if sqlDetails.Query == suite.SqlQueryDetails[i].Query {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(testErr)
							break
						} else {
							sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
						}
					} else {
						if sqlDetails.Query == suite.SqlQueryDetails[i].Query {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnError(testErr)
							break
						} else {
							sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
						}
					}
				}

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := suite.createRepo(convMock)
				//WHEN
				err := callCreate(pgRepository, ctx, suite.TenantID, suite.ModelEntity)
				//THEN
				require.Error(t, err)
				if suite.SqlQueryDetails[i].IsSelect {
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
				pgRepository := suite.createRepo(convMock)
				//WHEN
				err := callCreate(pgRepository, ctx, suite.TenantID, suite.ModelEntity)
				//THEN
				require.Error(t, err)
				require.Contains(t, err.Error(), testErr.Error())

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		t.Run("returns error when item is nil", func(t *testing.T) {
			ctx := context.TODO()
			convMock := suite.ConverterMockProvider()
			pgRepository := suite.createRepo(convMock)
			// WHEN
			err := callCreate(pgRepository, ctx, suite.TenantID, suite.NilModelEntity)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Internal Server Error")
			convMock.AssertExpectations(t)
		})
	})
}

// createRepo creates a new repository by the provided constructor func.
// In order to do this for all the different repository implementations we need to do it via reflection.
func (suite *RepoCreateTestSuite) createRepo(convMock interface{}) interface{} {
	v := reflect.ValueOf(suite.RepoConstructorFunc)
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

// callCreate calls the Create method of the given repository.
// In order to do this for all the different repository implementations we need to do it via reflection.
func callCreate(repo interface{}, ctx context.Context, tenant string, modelEntity interface{}) error {
	results := reflect.ValueOf(repo).MethodByName("Create").Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(tenant), reflect.ValueOf(modelEntity)})
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
