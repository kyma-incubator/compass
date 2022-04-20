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

// RepoUpsertTestSuite represents a generic test suite for repository Upsert, TrustedUpsert and UpsertGlobal methods of any entity that has externally managed tenants in m2m table/view.
// This test suite is not suitable for global entities or entities with embedded tenant in them.
type RepoUpsertTestSuite struct {
	Name                      string
	SQLQueryDetails           []SQLQueryDetails
	ConverterMockProvider     func() Mock
	RepoConstructorFunc       interface{}
	ModelEntity               interface{}
	DBEntity                  interface{}
	NilModelEntity            interface{}
	TenantID                  string
	DisableConverterErrorTest bool
	UpsertMethodName          string
}

// Run runs the generic repo upsert test suite
func (suite *RepoUpsertTestSuite) Run(t *testing.T) bool {
	if len(suite.UpsertMethodName) == 0 {
		suite.UpsertMethodName = "Upsert"
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
			err := callUpsert(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpsertMethodName)
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
				err := callUpsert(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpsertMethodName)
				// THEN
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

				convMock := suite.ConverterMockProvider()
				convMock.On("ToEntity", suite.ModelEntity).Return(nil, testErr).Once()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				err := callUpsert(pgRepository, ctx, suite.TenantID, suite.ModelEntity, suite.UpsertMethodName)
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
			err := callUpsert(pgRepository, ctx, suite.TenantID, suite.NilModelEntity, suite.UpsertMethodName)
			// THEN
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Internal Server Error")
			convMock.AssertExpectations(t)
		})
	})
}

func callUpsert(repo interface{}, ctx context.Context, tenant string, modelEntity interface{}, methodName string) error {
	results := reflect.ValueOf(repo).MethodByName(methodName).Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(tenant), reflect.ValueOf(modelEntity)})
	if len(results) != 2 {
		panic("Update should return two arguments")
	}
	result := results[1].Interface()
	if result == nil {
		return nil
	}
	err, ok := result.(error)
	if !ok {
		panic("Expected result to be an error")
	}
	return err
}
