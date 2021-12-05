package repo_test

import (
	"context"
	"database/sql/driver"
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

func TestList(t *testing.T) {
	sut := repo.NewLister(appTableName, appColumns)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription).
			AddRow(appID2, appName2, appDescription2)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		err := sut.List(ctx, resourceType, tenantID, &dest)
		require.NoError(t, err)
		assert.Len(t, dest, 2)
		assert.Contains(t, dest, *fixApp)
		assert.Contains(t, dest, *fixApp2)
	})

	t.Run("lists all items successfully with additional parameters", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE name = $1 AND description != $2 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))).
			WithArgs(appName, appDescription2, tenantID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("name", appName),
			repo.NewNotEqualCondition("description", appDescription2),
		}

		err := sut.List(ctx, resourceType, tenantID, &dest, conditions...)
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.List(ctx, resourceType, tenantID, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if empty tenant", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.List(ctx, resourceType, "", nil)
		require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, resourceType, tenantID, &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		var dest UserCollection
		err := sut.List(ctx, resourceType, tenantID, &dest)
		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})
}

func TestListWithEmbeddedTenant(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewListerWithEmbeddedTenant(userTableName, "tenant_id", []string{"id", "first_name", "last_name", "age"})

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1`)).
			WithArgs(tenantID).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, UserType, tenantID, &dest)
		require.NoError(t, err)
		assert.Len(t, dest, 2)
		assert.Contains(t, dest, peter)
		assert.Contains(t, dest, homer)
	})

	t.Run("lists all items successfully with additional parameters", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND age != $3")).
			WithArgs(tenantID, "Peter", 18).WillReturnRows(rows)
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("first_name", "Peter"),
			repo.NewNotEqualCondition("age", 18),
		}

		err := sut.List(ctx, UserType, tenantID, &dest, conditions...)
		require.NoError(t, err)
		assert.Len(t, dest, 1)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.List(ctx, UserType, tenantID, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		err := sut.List(ctx, UserType, tenantID, &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		var dest UserCollection
		err := sut.List(ctx, UserType, tenantID, &dest)
		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})
}

func TestListGlobal(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewListerGlobal(UserType, "users", []string{"id", "first_name", "last_name", "age"})

	t.Run("lists all items successfully", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users`)).
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

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, first_name, last_name, age FROM users WHERE first_name = $1 AND age != $2")).
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

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		var dest UserCollection
		err := sut.ListGlobal(ctx, &dest)
		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})
}
