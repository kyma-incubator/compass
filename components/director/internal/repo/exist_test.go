package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/require"
)

func TestExist(t *testing.T) {
	sut := repo.NewExistQuerier(appTableName)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE first_name = $1 AND last_name = $2 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe", tenantID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta(fmt.Sprintf("SELECT 1 FROM %s WHERE id = $1 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(appID, tenantID).WillReturnError(someError())
		// WHEN
		_, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		_, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)})

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error if empty tenant", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		_, err := sut.Exists(ctx, resourceType, "", repo.Conditions{repo.NewEqualCondition("id", appID)})
		require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.Exists(ctx, resourceType, tenantID, repo.Conditions{repo.NewEqualCondition("id", appID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

func TestExistWithEmbeddedTenant(t *testing.T) {
	givenID := "id"
	sut := repo.NewExistQuerierWithEmbeddedTenant(userTableName, "tenant_id")

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, "john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1 AND id = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(tenantID, givenID).WillReturnError(someError())
		// WHEN
		_, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		_, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.Exists(ctx, UserType, tenantID, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

func TestExistGlobal(t *testing.T) {
	givenID := "id"
	sut := repo.NewExistQuerierGlobal(UserType, "users")

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE first_name = $1 AND last_name = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnError(someError())
		// WHEN
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", givenID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

func TestExistsGlobalWithConditionTree(t *testing.T) {
	givenID := "id"
	sut := repo.NewExistsQuerierGlobalWithConditionTree(UserType, "users")

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobalWithConditionTree(ctx, &repo.ConditionTree{Operand: repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.ExistsGlobalWithConditionTree(ctx, &repo.ConditionTree{Operand: repo.NewEqualCondition("id", givenID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobalWithConditionTree(ctx, nil)
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE (first_name = $1 AND last_name = $2)")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.ExistsGlobalWithConditionTree(ctx, repo.And(&repo.ConditionTree{Operand: repo.NewEqualCondition("first_name", "john")}, &repo.ConditionTree{Operand: repo.NewEqualCondition("last_name", "doe")}))
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnError(someError())
		// WHEN
		_, err := sut.ExistsGlobalWithConditionTree(ctx, &repo.ConditionTree{Operand: repo.NewEqualCondition("id", givenID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)

		_, err := sut.ExistsGlobalWithConditionTree(ctx, &repo.ConditionTree{Operand: repo.NewEqualCondition("id", givenID)})

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.ExistsGlobalWithConditionTree(ctx, &repo.ConditionTree{Operand: repo.NewEqualCondition("id", givenID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}
