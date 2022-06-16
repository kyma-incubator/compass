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

// RepoExistTestSuite represents a generic test suite for repository Exists method of any entity that has externally managed tenants in m2m table/view.
// This test suite is not suitable for global entities or entities with embedded tenant in them.
type RepoExistTestSuite struct {
	Name                  string
	SQLQueryDetails       []SQLQueryDetails
	ConverterMockProvider func() Mock
	RepoConstructorFunc   interface{}
	TargetID              string
	TenantID              string
	RefEntity             interface{}
	IsGlobal              bool
}

// Run runs the generic repo exists test suite
func (suite *RepoExistTestSuite) Run(t *testing.T) bool {
	for _, queryDetails := range suite.SQLQueryDetails {
		if !queryDetails.IsSelect {
			panic("exist suite should expect only select SQL statements")
		}
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
			found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity, suite.IsGlobal)
			// THEN
			require.NoError(t, err)
			require.True(t, found)

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns false if entity does not exist", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			configureInvalidSelect(sqlMock, suite.SQLQueryDetails)

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			// WHEN
			found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity, suite.IsGlobal)
			// THEN
			require.NoError(t, err)
			require.False(t, found)

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		for i := range suite.SQLQueryDetails {
			t.Run(fmt.Sprintf("error if SQL query %d fail", i), func(t *testing.T) {
				sqlxDB, sqlMock := MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

				configureFailureForSQLQueryOnIndex(sqlMock, suite.SQLQueryDetails, i, testErr)

				convMock := suite.ConverterMockProvider()
				pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
				// WHEN
				found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity, suite.IsGlobal)
				// THEN
				require.Error(t, err)
				require.Equal(t, apperrors.InternalError, apperrors.ErrorCode(err))
				require.Contains(t, err.Error(), "Internal Server Error: Unexpected error while executing SQL query")

				require.False(t, found)

				sqlMock.AssertExpectations(t)
				convMock.AssertExpectations(t)
			})
		}
	})
}

func callExists(repo interface{}, ctx context.Context, tenant string, id string, refEntity interface{}, isGlobal bool) (bool, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(tenant), reflect.ValueOf(id)}
	if isGlobal {
		args = []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(id)}
	}
	if refEntity != nil {
		args = append(args, reflect.ValueOf(refEntity))
	}
	results := reflect.ValueOf(repo).MethodByName("Exists").Call(args)
	if len(results) != 2 {
		panic("Exists should return two arguments")
	}

	found := results[0].Bool()
	errResult := results[1].Interface()
	if errResult == nil {
		return found, nil
	}
	err, ok := errResult.(error)
	if !ok {
		panic("Expected result to be an error")
	}
	return found, err
}
