package repo_test

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
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

	sut := repo.NewLister("users", "tenant_col",
		[]string{"id_col", "tenant_col", "first_name", "last_name", "age"})

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "tenant_col", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_col, tenant_col, first_name, last_name, age FROM users WHERE tenant_col=$1`)).
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

		rows := sqlmock.NewRows([]string{"id_col", "tenant_col", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_col, tenant_col, first_name, last_name, age FROM users WHERE tenant_col=$1 AND first_name='Peter' AND age > 18`)).
			WithArgs(givenTenant).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, givenTenant, &dest, "first_name='Peter'", "age > 18")
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.List(ctx, givenTenant, nil)
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, givenTenant, &dest)
		require.EqualError(t, err, "while fetching list of objects from DB: some error")
	})
}

func TestListGlobal(t *testing.T) {
	peterID := uuidA()
	homerID := uuidC()
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewListerGlobal("users",
		[]string{"id_col", "first_name", "last_name", "age"})

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
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id_col, first_name, last_name, age FROM users WHERE first_name='Peter' AND age > 18`)).
			WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.ListGlobal(ctx, &dest, "first_name='Peter'", "age > 18")
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.ListGlobal(ctx, nil)
		require.EqualError(t, err, "unable to fetch database from context")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.ListGlobal(ctx, &dest)
		require.EqualError(t, err, "while fetching list of objects from DB: some error")
	})
}
