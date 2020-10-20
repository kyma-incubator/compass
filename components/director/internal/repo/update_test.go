package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestUpdateSingle(t *testing.T) {
	sut := repo.NewUpdater(UserType, "users", []string{"first_name", "last_name", "age"}, "tenant_id", []string{"id_col"})
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

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_id = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when no id column", func(t *testing.T) {
		// GIVEN
		sut := repo.NewUpdater(UserType, "users", []string{"first_name", "last_name", "age"}, "tenant_id", []string{})
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_id = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant").WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec("UPDATE users .*").
			WillReturnError(someError())
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
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
		require.True(t, apperrors.IsNotUniqueError(err))
	})

	t.Run("returns error if modified more than one row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_id = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 157))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "should update single row, but updated 157 rows")
	})

	t.Run("returns error if does not modified any row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE tenant_id = ? AND id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_tenant", "given_id").WillReturnResult(sqlmock.NewResult(0, 0))
		// WHEN
		err := sut.UpdateSingle(ctx, givenUser)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "should update single row, but updated 0 rows")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingle(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if entity is nil", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingle(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}

func TestUpdateSingleGlobal(t *testing.T) {
	sut := repo.NewUpdaterGlobal(UserType, "users", []string{"first_name", "last_name", "age"}, []string{"id_col"})
	givenUser := User{
		ID:        "given_id",
		FirstName: "given_first_name",
		LastName:  "given_last_name",
		Age:       55,
	}

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("success when no id column", func(t *testing.T) {
		// GIVEN
		sut := repo.NewUpdaterGlobal(UserType, "users", []string{"first_name", "last_name", "age"}, []string{})
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ?")).
			WithArgs("given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(0, 1))
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec("UPDATE users .*").
			WillReturnError(someError())
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec("UPDATE users .*").
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.True(t, apperrors.IsNotUniqueError(err))
	})

	t.Run("returns error if modified more than one row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 157))
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "should update single row, but updated 157 rows")
	})

	t.Run("returns error if does not modified any row", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectExec(regexp.QuoteMeta("UPDATE users SET first_name = ?, last_name = ?, age = ? WHERE id_col = ?")).
			WithArgs("given_first_name", "given_last_name", 55, "given_id").WillReturnResult(sqlmock.NewResult(0, 0))
		// WHEN
		err := sut.UpdateSingleGlobal(ctx, givenUser)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "should update single row, but updated 0 rows")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingleGlobal(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if entity is nil", func(t *testing.T) {
		// WHEN
		err := sut.UpdateSingleGlobal(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("item cannot be nil").Error())
	})
}
