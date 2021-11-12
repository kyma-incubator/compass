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

type RepoExistTestSuite struct {
	Name                  string
	SqlQueryDetails       []SqlQueryDetails
	ConverterMockProvider func() Mock
	RepoConstructorFunc   interface{}
	TargetID              string
	TenantID              string
	RefEntity             interface{}
}

func (suite *RepoExistTestSuite) Run(t *testing.T) bool {
	for _, queryDetails := range suite.SqlQueryDetails {
		if !queryDetails.IsSelect {
			panic("exist suite should expect only select SQL statements")
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
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity)
			//THEN
			require.NoError(t, err)
			require.True(t, found)

			sqlMock.AssertExpectations(t)
			convMock.AssertExpectations(t)
		})

		t.Run("returns false if entity does not exist", func(t *testing.T) {
			sqlxDB, sqlMock := MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), sqlxDB)

			for _, sqlDetails := range suite.SqlQueryDetails {
				sqlMock.ExpectQuery(sqlDetails.Query).WithArgs(sqlDetails.Args...).WillReturnRows(sqlDetails.InvalidRowsProvider()...)
			}

			convMock := suite.ConverterMockProvider()
			pgRepository := createRepo(suite.RepoConstructorFunc, convMock)
			//WHEN
			found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity)
			//THEN
			require.NoError(t, err)
			require.False(t, found)

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
				found, err := callExists(pgRepository, ctx, suite.TenantID, suite.TargetID, suite.RefEntity)
				//THEN
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

func callExists(repo interface{}, ctx context.Context, tenant string, id string, refEntity interface{}) (bool, error) {
	args := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(tenant), reflect.ValueOf(id)}
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
