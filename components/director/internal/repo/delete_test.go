package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

type methodToTest = func(ctx context.Context, tenant string, conditions repo.Conditions) error
type methodToTestWithoutTenant = func(ctx context.Context, conditions repo.Conditions) error

var tableName = "users"

func TestDelete(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewDeleter(UserType, tableName, "tenant_id")

	tc := map[string]methodToTest{
		"DeleteMany": sut.DeleteMany,
		"DeleteOne":  sut.DeleteOne,
	}
	for tn, testedMethod := range tc {
		t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when no conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, givenTenant, repo.Conditions{})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when more conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] returns error on db operation", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnError(someError())
			// WHEN
			err := testedMethod(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing persistence context", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})
	}
}

func TestDeleteGlobal(t *testing.T) {
	givenID := uuidA()
	sut := repo.NewDeleterGlobal(UserType, tableName)

	tc := map[string]methodToTestWithoutTenant{
		"DeleteMany": sut.DeleteManyGlobal,
		"DeleteOne":  sut.DeleteOneGlobal,
	}
	for tn, testedMethod := range tc {
		t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id_col = $1", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when no conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, repo.Conditions{})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when more conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE first_name = $1 AND last_name = $2", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs("john", "doe").WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] returns error on db operation", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id_col = $1", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnError(someError())
			// WHEN
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing persistence context", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})
	}
}

func TestDeleteReactsOnNumberOfRemovedObjects(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewDeleter(UserType, tableName, "tenant_id")

	cases := map[string]struct {
		methodToTest      methodToTest
		givenRowsAffected int64
		expectedErrString string
	}{
		"[DeleteOne] returns error when removed more than one object": {
			methodToTest:      sut.DeleteOne,
			givenRowsAffected: 154,
			expectedErrString: "Internal Server Error: delete should remove single row, but removed 154 rows",
		},
		"[DeleteOne] returns error when object not found": {
			methodToTest:      sut.DeleteOne,
			givenRowsAffected: 0,
			expectedErrString: "Internal Server Error: delete should remove single row, but removed 0 rows",
		},
		"[Delete Many] success when removed more than one object": {
			methodToTest:      sut.DeleteMany,
			givenRowsAffected: 154,
			expectedErrString: "",
		},
		"[Delete Many] success when not found objects to remove": {
			methodToTest:      sut.DeleteMany,
			givenRowsAffected: 0,
			expectedErrString: "",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(0, tc.givenRowsAffected))
			// WHEN
			err := tc.methodToTest(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			if tc.expectedErrString != "" {
				require.EqualError(t, err, tc.expectedErrString)
			}
		})
	}
}

func TestDeleteGlobalReactsOnNumberOfRemovedObjects(t *testing.T) {
	givenID := uuidA()
	sut := repo.NewDeleterGlobal(UserType, tableName)

	cases := map[string]struct {
		methodToTest      methodToTestWithoutTenant
		givenRowsAffected int64
		expectedErrString string
	}{
		"[DeleteOne] returns error when removed more than one object": {
			methodToTest:      sut.DeleteOneGlobal,
			givenRowsAffected: 154,
			expectedErrString: "Internal Server Error: delete should remove single row, but removed 154 rows",
		},
		"[DeleteOne] returns error when object not found": {
			methodToTest:      sut.DeleteOneGlobal,
			givenRowsAffected: 0,
			expectedErrString: "Internal Server Error: delete should remove single row, but removed 0 rows",
		},
		"[Delete Many] success when removed more than one object": {
			methodToTest:      sut.DeleteManyGlobal,
			givenRowsAffected: 154,
			expectedErrString: "",
		},
		"[Delete Many] success when not found objects to remove": {
			methodToTest:      sut.DeleteManyGlobal,
			givenRowsAffected: 0,
			expectedErrString: "",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id_col = $1", tableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnResult(sqlmock.NewResult(0, tc.givenRowsAffected))
			// WHEN
			err := tc.methodToTest(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
			// THEN
			if tc.expectedErrString != "" {
				require.EqualError(t, err, tc.expectedErrString)
			}
		})
	}
}

func defaultExpectedDeleteQuery() string {
	return regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND id_col = $2", tableName))
}

func uuidA() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func uuidB() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func uuidC() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}
