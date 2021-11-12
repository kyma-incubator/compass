package testdb

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

// RepoUpdateTestSuite represents a generic test suite for repository Update and UpdateWithVersion methods of any entity that has externally managed tenants in m2m table/view.
// This test suite is not suitable for global entities or entities with embedded tenant in them.
type RepoUpdateTestSuite struct {
	Name                      string
	SqlQueryDetails           []SqlQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ModelEntity               interface{}
	DBEntity                  interface{}
	NilModelEntity            interface{}
	TenantID                  string
	DisableConverterErrorTest bool
	UpdateMethodName          string
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

			for _, sqlDetails := range suite.SqlQueryDetails {
				if sqlDetails.IsSelect {
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				} else {
					sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.ValidResult)
				}
			}

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
			//THEN
			require.NoError(t, err)
			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

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
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				//WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				//THEN
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

				for _, sqlDetails := range suite.SqlQueryDetails {
					if sqlDetails.IsSelect {
						sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
						break
					}
				}

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				//WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				//THEN
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
				//WHEN
				err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
				//THEN
				require.Error(t, err)
				require.Contains(t, err.Error(), testErr.Error())

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}

		t.Run("error if 0 affected rows", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				if sqlDetails.IsSelect {
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				} else {
					sqlMock.ExpectExec(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnResult(sqlDetails.InvalidResult)
				}
			}

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			err := callUpdate(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpdateMethodName)
			//THEN
			require.Error(t, err)
			if suite.UpdateMethodName == "UpdateWithVersion" {
				require.Equal(t, apperrors.ConcurrentUpdate, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), apperrors.ConcurrentUpdateMsg)
			} else {
				require.Equal(t, apperrors.Unauthorized, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)
			}

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns error if tenant_id is not uuid", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				if sqlDetails.IsSelect {
					sqlDetails.Args[len(sqlDetails.Args)-1] = "tenantID" // Override the tenantID expected argument in order to use non-uuid value for the test.
					sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.ValidRowsProvider()...)
				}
			}

			convMock := suite.ConverterMockProvider()
			convMock.On("ToEntity", suite.ModelEntity).Return(suite.DBEntity, nil).Once()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			err := callUpdate(pgRepository, ctx, "tenantID", suite.ModelEntity, suite.UpdateMethodName)
			//THEN
			require.Error(t, err)
			require.Contains(t, err.Error(), "tenant_id tenantID should be UUID")

			convMock.AssertExpectations(t)
			sqlMock.AssertExpectations(t)
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
	results := reflect.ValueOf(repo).MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(tenant), reflect.ValueOf(modelEntity)})
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
