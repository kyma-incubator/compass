package repo_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	givenTenant := uuidB()
	peterID := uuidA()
	homerID := uuidC()
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, Tenant: givenTenant, ID: peterID}
	peterRow := []driver.Value{peterID, givenTenant, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, Tenant: givenTenant, ID: homerID}
	homerRow := []driver.Value{homerID, givenTenant, "Homer", "Simpson", 55}

	sut := repo.NewLister("UserType", "users", "tenant_id",
		[]string{"id_col", "tenant_id", "first_name", "last_name", "age"})

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1")).
			WithArgs(givenTenant).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, givenTenant, &dest)
		require.NoError(t, err)
		assert.Len(t, dest, 2)
		assert.Contains(t, dest, peter)
		assert.Contains(t, dest, homer)
	})

	t.Run("lists all items successfully with additional parameters", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND age != $3`)).
			WithArgs(givenTenant, "Peter", 18).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("first_name", "Peter"),
			repo.NewNotEqualCondition("age", 18),
		}

		err := sut.List(ctx, givenTenant, &dest, conditions...)
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.List(ctx, givenTenant, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, givenTenant, &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestListGlobal(t *testing.T) {
	peterID := uuidA()
	homerID := uuidC()
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewListerGlobal(UserType, "users", []string{"id_col", "first_name", "last_name", "age"})

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_col, first_name, last_name, age FROM users`)).
			WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.ListGlobal(ctx, &dest)
		require.NoError(t, err)
		assert.Len(t, dest, 2)
		assert.Contains(t, dest, peter)
		assert.Contains(t, dest, homer)
	})

	t.Run("lists all items successfully with additional parameters", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id_col, first_name, last_name, age FROM users WHERE first_name = $1 AND age != $2")).
			WithArgs("Peter", 18).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("first_name", "Peter"),
			repo.NewNotEqualCondition("age", 18),
		}

		err := sut.ListGlobal(ctx, &dest, conditions...)
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.ListGlobal(ctx, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.ListGlobal(ctx, &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
