package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/lib/pq"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestUpsert(t *testing.T) {
	expectedQuery := regexp.QuoteMeta(`INSERT INTO users ( id_col, tenant_col, first_name, last_name, age ) 
		VALUES ( ?, ?, ?, ?, ? ) ON CONFLICT ( tenant_col, first_name, last_name ) DO UPDATE SET age=EXCLUDED.age`)
	sut := repo.NewUpserter("users", []string{"id_col", "tenant_col", "first_name", "last_name", "age"}, []string{"tenant_col", "first_name", "last_name"}, []string{"age"})
	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{
			ID:        "given_id",
			Tenant:    "given_tenant",
			FirstName: "given_first_name",
			LastName:  "given_last_name",
			Age:       55,
		}

		mock.ExpectExec(expectedQuery).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sut.Upsert(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(expectedQuery).
			WillReturnError(someError())
		// WHEN
		err := sut.Upsert(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "while upserting row to 'users' table: some error")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(expectedQuery).
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.Upsert(ctx, givenUser)
		// THEN
		require.True(t, apperrors.IsNotUnique(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.Upsert(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		// WHEN
		err := sut.Upsert(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, "item cannot be nil")
	})
}

func TestUpsertWhenWrongConfiguration(t *testing.T) {
	sut := repo.NewUpserter("users", []string{"id_col", "tenant_col", "column_does_not_exist"}, []string{"id_col", "tenant_col"}, []string{"column_does_not_exist"})
	// GIVEN
	db, mock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	defer mock.AssertExpectations(t)
	// WHEN
	err := sut.Upsert(ctx, User{})
	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not find name column_does_not_exist")
}
