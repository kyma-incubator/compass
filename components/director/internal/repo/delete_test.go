package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewDeleter("users", "tenant_col")

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
		// WHEN
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		givenTenant := uuidB()
		expectedQuery := regexp.QuoteMeta("DELETE FROM users WHERE tenant_col = $1 AND first_name = $2 AND last_name = $3")
		sut := repo.NewDeleter("users", "tenant_col")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnResult(sqlmock.NewResult(-1, 1))
		// WHEN
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "first_name", Val: "john"}, {Field: "last_name", Val: "doe"}})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnError(someError())
		// WHEN
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.EqualError(t, err, "while deleting from database: some error")
	})

	t.Run("returns error when removed more than one object", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(0, 12))
		// WHEN
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.EqualError(t, err, "delete should remove single row, but removed 12 rows")
	})

	t.Run("returns error when object not found", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(defaultExpectedDeleteQuery()).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(0, 0))
		// WHEN
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.EqualError(t, err, "delete should remove single row, but removed 0 rows")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Delete(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		require.EqualError(t, err, "unable to fetch database from context")
	})

}

func defaultExpectedDeleteQuery() string {
	return regexp.QuoteMeta("DELETE FROM users WHERE tenant_col = $1 AND id_col = $2")
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
