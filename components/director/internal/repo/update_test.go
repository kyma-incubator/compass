package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestUpdateSingle(t *testing.T) {
	sut := repo.NewUpdater("users", []string{"first_name", "last_name", "age"}, "tenant_col", []string{"id_col"})
	givenUser := User{
		ID:        "given_id",
		Tenant:    "given_tenant",
		FirstName: "given_first_name",
		LastName:  "given_last_name",
		Age:       55,
	}
	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_col = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error on db operation failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec("UPDATE users .*").
			WillReturnError(errors.New("some error"))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "while updating single entity: some error")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec("UPDATE users .*").
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.True(t, repo.IsNotUnique(err))
	})

	t.Run("returns error if modified more than one row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_col = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 157))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "should update single row, but updated 157 rows")
	})

	t.Run("returns error if does not modified any row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_col = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 0))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "should update single row, but updated 0 rows")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingle(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error if entity is nil", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingle(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, "item cannot be nil")
	})
}
