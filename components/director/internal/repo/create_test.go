package repo_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/stretchr/testify/assert"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	sut := repo.NewCreator(UserType, tableName, []string{"id_col", "tenant_id", "first_name", "last_name", "age"})
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

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id_col, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", tableName))).
			WithArgs("given_id", "given_tenant", "given_first_name", "given_last_name", 55).WillReturnResult(sqlmock.NewResult(1, 1))
		// WHEN
		err := sut.Create(ctx, givenUser)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id_col, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", tableName))).
			WillReturnError(someError())
		// WHEN
		err := sut.Create(ctx, givenUser)
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns non unique error", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		givenUser := User{}

		mock.ExpectExec(regexp.QuoteMeta(fmt.Sprintf("INSERT INTO %s ( id_col, tenant_id, first_name, last_name, age ) VALUES ( ?, ?, ?, ?, ? )", tableName))).
			WillReturnError(&pq.Error{Code: persistence.UniqueViolation})
		// WHEN
		err := sut.Create(ctx, givenUser)
		// THEN
		require.True(t, apperrors.IsNotUniqueError(err))
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		// WHEN
		err := sut.Create(context.TODO(), User{})
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if destination is nil", func(t *testing.T) {
		// WHEN
		err := sut.Create(context.TODO(), nil)
		// THEN
		require.EqualError(t, err, "Internal Server Error: item cannot be nil")
	})
}

func TestCreateWhenWrongConfiguration(t *testing.T) {
	sut := repo.NewCreator(UserType, tableName, []string{"id_col", "tenant_id", "column_does_not_exist"})
	// GIVEN
	db, mock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	defer mock.AssertExpectations(t)
	// WHEN
	err := sut.Create(ctx, User{})
	// THEN
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Unexpected error while executing SQL query")
}
