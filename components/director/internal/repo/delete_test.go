package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type methodToTest = func(ctx context.Context, resourceType resource.Type, tenant string, conditions repo.Conditions) error
type methodToTestWithoutTenant = func(ctx context.Context, conditions repo.Conditions) error

func TestDelete(t *testing.T) {
	deleter := repo.NewDeleter(bundlesTableName)
	resourceType := resource.Bundle
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	tc := map[string]methodToTest{
		"DeleteMany": deleter.DeleteMany,
		"DeleteOne":  deleter.DeleteOne,
	}
	for tn, testedMethod := range tc {
		t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$2")))).
				WithArgs(bundleID, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

			// WHEN
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", bundleID)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when no conditions", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$1")))).
				WithArgs(tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

			// WHEN
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when more conditions", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE description = $1 AND name = $2 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$3")))).
				WithArgs(bundleDescription, bundleName, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

			// WHEN
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("description", bundleDescription), repo.NewEqualCondition("name", bundleName)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] fail when 0 entities match conditions", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE description = $1 AND name = $2 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$3")))).
				WithArgs(bundleDescription, bundleName, tenantID).WillReturnResult(sqlmock.NewResult(-1, 0))

			// WHEN
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("description", bundleDescription), repo.NewEqualCondition("name", bundleName)})
			// THEN
			if tn == "DeleteMany" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), apperrors.ShouldBeOwnerMsg)
			}
		})

		t.Run(fmt.Sprintf("[%s] returns error when delete operation returns error", tn), func(t *testing.T) {
			// GIVEN
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE description = $1 AND name = $2 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$3")))).
				WithArgs(bundleDescription, bundleName, tenantID).WillReturnError(someError())

			// WHEN
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("description", bundleDescription), repo.NewEqualCondition("name", bundleName)})
			// THEN
			require.Error(t, err)
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run(fmt.Sprintf("[%s] context properly canceled", tn), func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			defer mock.AssertExpectations(t)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			ctx = persistence.SaveToContext(ctx, db)

			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", bundleID)})

			require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing persistence context", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", bundleID)})
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing tenant id", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, resourceType, "", repo.Conditions{repo.NewEqualCondition("id", bundleID)})
			require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
		})

		t.Run(fmt.Sprintf("[%s] returns error if entity does not have access table", tn), func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)

			err := testedMethod(ctx, resource.Type("unknown"), tenantID, repo.Conditions{repo.NewEqualCondition("id", bundleID)})

			require.Error(t, err)
			assert.Contains(t, err.Error(), "entity unknown does not have access table")
		})
	}

	t.Run("BIA", func(t *testing.T) {
		deleter := repo.NewDeleter(biaTableName)
		resourceType := resource.BundleInstanceAuth
		m2mTable, ok := resourceType.TenantAccessTable()
		require.True(t, ok)

		tc := map[string]methodToTest{
			"DeleteMany": deleter.DeleteMany,
			"DeleteOne":  deleter.DeleteOne,
		}
		for tn, testedMethod := range tc {
			t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
				// GIVEN
				db, mock := testdb.MockDatabase(t)
				ctx := persistence.SaveToContext(context.TODO(), db)
				defer mock.AssertExpectations(t)

				mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id = $1 AND %s", biaTableName, fmt.Sprintf(tenantIsolationConditionForBIA, m2mTable, "$2", "$3")))).
					WithArgs(biaID, tenantID, tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))

				// WHEN
				err := testedMethod(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", biaID)})
				// THEN
				require.NoError(t, err)
			})
		}
	})

	t.Run("[DeleteMany] success when more than one resource matches conditions", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE description = $1 AND name = $2 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$3")))).
			WithArgs(bundleDescription, bundleName, tenantID).WillReturnResult(sqlmock.NewResult(-1, 2))

		// WHEN
		err := deleter.DeleteMany(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("description", bundleDescription), repo.NewEqualCondition("name", bundleName)})
		// THEN
		require.NoError(t, err)
	})

	t.Run("[DeleteOne] fail when more than one resource matches conditions", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE description = $1 AND name = $2 AND %s", bundlesTableName, fmt.Sprintf(tenantIsolationConditionWithOwnerCheckFmt, m2mTable, "$3")))).
			WithArgs(bundleDescription, bundleName, tenantID).WillReturnResult(sqlmock.NewResult(-1, 2))

		// WHEN
		err := deleter.DeleteOne(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("description", bundleDescription), repo.NewEqualCondition("name", bundleName)})
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "delete should remove single row, but removed 2 rows")
	})
}

func TestDeleteGlobal(t *testing.T) {
	givenID := "id"
	sut := repo.NewDeleterGlobal(UserType, userTableName)

	tc := map[string]methodToTestWithoutTenant{
		"DeleteMany": sut.DeleteManyGlobal,
		"DeleteOne":  sut.DeleteOneGlobal,
	}
	for tn, testedMethod := range tc {
		t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id = $1", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when no conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s", userTableName))
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
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE first_name = $1 AND last_name = $2", userTableName))
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
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id = $1", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnError(someError())
			// WHEN
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run("context properly canceled", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			defer mock.AssertExpectations(t)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			ctx = persistence.SaveToContext(ctx, db)

			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})

			require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing persistence context", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})
	}
}

func TestDeleteWithEmbeddedTenant(t *testing.T) {
	givenID := "id"
	sut := repo.NewDeleterWithEmbeddedTenant(userTableName, "tenant_id")

	tc := map[string]methodToTest{
		"DeleteMany": sut.DeleteMany,
		"DeleteOne":  sut.DeleteOne,
	}
	for tn, testedMethod := range tc {
		t.Run(fmt.Sprintf("[%s] success", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND id = $2", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(tenantID, givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when no conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(tenantID).WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] success when more conditions", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(tenantID, "john", "doe").WillReturnResult(sqlmock.NewResult(-1, 1))
			// WHEN
			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
			// THEN
			require.NoError(t, err)
		})

		t.Run(fmt.Sprintf("[%s] returns error on db operation", tn), func(t *testing.T) {
			// GIVEN
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE tenant_id = $1 AND id = $2", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(tenantID, givenID).WillReturnError(someError())
			// WHEN
			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			// THEN
			require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
		})

		t.Run("context properly canceled", func(t *testing.T) {
			db, mock := testdb.MockDatabase(t)
			defer mock.AssertExpectations(t)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
			defer cancel()

			ctx = persistence.SaveToContext(ctx, db)

			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})

			require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
		})

		t.Run(fmt.Sprintf("[%s] returns error if missing persistence context", tn), func(t *testing.T) {
			ctx := context.TODO()
			err := testedMethod(ctx, resource.AutomaticScenarioAssigment, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
		})
	}
}

func TestDeleteConditionTreeWithEmbeddedTenant(t *testing.T) {
	sut := repo.NewDeleteConditionTreeWithEmbeddedTenant(userTableName, "tenant_id")

	t.Run("deletes all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE (tenant_id = $1 AND first_name IN ($2, $3))`)).
			WithArgs(tenantID, "Joe", "Smith").WillReturnResult(sqlmock.NewResult(0, 2))
		ctx := persistence.SaveToContext(context.TODO(), db)

		err := sut.DeleteConditionTree(ctx, UserType, tenantID, &repo.ConditionTree{Operand: repo.NewInConditionForStringValues("first_name", []string{"Joe", "Smith"})})
		require.NoError(t, err)
	})

	t.Run("deletes all items successfully with additional parameters", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("DELETE FROM users WHERE (tenant_id = $1 AND (first_name = $2 OR age != $3))")).
			WithArgs(tenantID, "Joe", 18).WillReturnResult(sqlmock.NewResult(0, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)

		conditions := repo.Or(repo.ConditionTreesFromConditions([]repo.Condition{
			repo.NewEqualCondition("first_name", "Joe"),
			repo.NewNotEqualCondition("age", 18),
		})...)

		err := sut.DeleteConditionTree(ctx, UserType, tenantID, conditions)
		require.NoError(t, err)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.DeleteConditionTree(ctx, UserType, tenantID, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(`DELETE .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)

		err := sut.DeleteConditionTree(ctx, UserType, tenantID, &repo.ConditionTree{Operand: repo.NewInConditionForStringValues("first_name", []string{"Peter", "Homer"})})
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		err := sut.DeleteConditionTree(ctx, UserType, tenantID, &repo.ConditionTree{Operand: repo.NewInConditionForStringValues("first_name", []string{"Peter", "Homer"})})
		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})
}

func TestDeleteGlobalReactsOnNumberOfRemovedObjects(t *testing.T) {
	givenID := "id"
	sut := repo.NewDeleterGlobal(UserType, userTableName)

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
			expectedQuery := regexp.QuoteMeta(fmt.Sprintf("DELETE FROM %s WHERE id = $1", userTableName))
			db, mock := testdb.MockDatabase(t)
			ctx := persistence.SaveToContext(context.TODO(), db)
			defer mock.AssertExpectations(t)
			mock.ExpectExec(expectedQuery).WithArgs(givenID).WillReturnResult(sqlmock.NewResult(0, tc.givenRowsAffected))
			// WHEN
			err := tc.methodToTest(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
			// THEN
			if tc.expectedErrString != "" {
				require.EqualError(t, err, tc.expectedErrString)
			}
		})
	}
}
