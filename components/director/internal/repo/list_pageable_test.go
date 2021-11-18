package repo_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListPageable(t *testing.T) {
	sut := repo.NewPageableQuerier(appTableName, appColumns)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("returns first page and there are no more pages", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription).
			AddRow(appID2, appName2, appDescription2)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 10 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(2))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		actualPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 10, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, actualTotal)
		assert.Len(t, dest, 2)
		assert.Equal(t, *fixApp, dest[0])
		assert.Equal(t, *fixApp2, dest[1])
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns full page and has next page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription).
			AddRow(appID2, appName2, appDescription2)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 2 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		actualPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 2)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns many pages and I can traverse it using cursor", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rowsForPage1 := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription)
		rowsForPage2 := sqlmock.NewRows(appColumns).
			AddRow(appID2, appName2, appDescription2)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 1 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rowsForPage1)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 1 OFFSET 1", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rowsForPage2)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))

		ctx := persistence.SaveToContext(context.TODO(), db)
		var first AppCollection

		actualFirstPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 1, "", "id", &first)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, first, 1)
		assert.True(t, actualFirstPage.HasNextPage)
		assert.NotEmpty(t, actualFirstPage.EndCursor)

		var second AppCollection
		actualSecondPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 1, actualFirstPage.EndCursor, "id", &second)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, second, 1)
		assert.True(t, actualSecondPage.HasNextPage)
		assert.NotEmpty(t, actualSecondPage.EndCursor)
	})

	t.Run("returns page without conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 2 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		actualPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns page with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE name = $1 AND description != $2 AND %s ORDER BY id LIMIT 2 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))).
			WithArgs(appName, appDescription2, tenantID).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE name = $1 AND description != $2 AND %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$3")))).
			WithArgs(appName, appDescription2, tenantID).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("name", appName),
			repo.NewNotEqualCondition("description", appDescription2),
		}

		actualPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest, conditions...)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns empty page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 2 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(0))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		actualPage, actualTotal, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 0, actualTotal)
		assert.Empty(t, dest)
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, _, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if empty tenant", func(t *testing.T) {
		ctx := context.TODO()
		_, _, err := sut.List(ctx, resourceType, "", 2, "", "id", nil)
		require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.List(ctx, resourceType, tenantID, 2, "zzz", "", nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error if wrong pagination attributes", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.List(ctx, resourceType, tenantID, -3, "", "id", nil)
		require.EqualError(t, err, "while converting offset and limit to cursor: Invalid data [reason=page size cannot be smaller than 1]")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		_, _, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("returns error on calculating total count", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id, name, description FROM %s WHERE %s ORDER BY id LIMIT 2 OFFSET 0", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1")))).
			WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		_, _, err := sut.List(ctx, resourceType, tenantID, 2, "", "id", &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestListPageableWithEmbeddedTenant(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewPageableQuerierWithEmbeddedTenant(userTableName, "tenant_id", []string{"id", "first_name", "last_name", "age"})

	t.Run("returns first page and there are no more pages", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 10 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(2))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 10, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, actualTotal)
		assert.Len(t, dest, 2)
		assert.Equal(t, peter, dest[0])
		assert.Equal(t, homer, dest[1])
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns full page and has next page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 2 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 2)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns many pages and I can traverse it using cursor", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rowsForPage1 := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		rowsForPage2 := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 1 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rowsForPage1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 1 OFFSET 1`)).WithArgs(tenantID).WillReturnRows(rowsForPage2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))

		ctx := persistence.SaveToContext(context.TODO(), db)
		var first UserCollection

		actualFirstPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 1, "", "id", &first)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, first, 1)
		assert.True(t, actualFirstPage.HasNextPage)
		assert.NotEmpty(t, actualFirstPage.EndCursor)

		var second UserCollection
		actualSecondPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 1, actualFirstPage.EndCursor, "id", &second)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, second, 1)
		assert.True(t, actualSecondPage.HasNextPage)
		assert.NotEmpty(t, actualSecondPage.EndCursor)
	})

	t.Run("returns page without conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 2 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns page with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND age != $3 ORDER BY id LIMIT 2 OFFSET 0")).
			WithArgs(tenantID, "Peter", 18).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND first_name = $2 AND age != $3")).
			WithArgs(tenantID, "Peter", 18).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("first_name", "Peter"),
			repo.NewNotEqualCondition("age", 18),
		}

		actualPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest, conditions...)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns empty page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 2 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users WHERE tenant_id = $1`)).WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(0))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 0, actualTotal)
		assert.Empty(t, dest)
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, _, err := sut.List(ctx, UserType, tenantID, 2, "", "id", nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.List(ctx, UserType, tenantID, 2, "zzz", "", nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error if wrong pagination attributes", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.List(ctx, UserType, tenantID, -3, "", "id", nil)
		require.EqualError(t, err, "while converting offset and limit to cursor: Invalid data [reason=page size cannot be smaller than 1]")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, _, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		var dest UserCollection

		_, _, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error on calculating total count", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users WHERE tenant_id = $1 ORDER BY id LIMIT 2 OFFSET 0`)).WithArgs(tenantID).WillReturnRows(rows)
		mock.ExpectQuery(`SELECT COUNT\(\*\).*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, _, err := sut.List(ctx, UserType, tenantID, 2, "", "id", &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestListPageableGlobal(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, ID: peterID}
	peterRow := []driver.Value{peterID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, ID: homerID}
	homerRow := []driver.Value{homerID, "Homer", "Simpson", 55}

	sut := repo.NewPageableQuerierGlobal("UserType", "users",
		[]string{"id", "first_name", "last_name", "age"})

	t.Run("returns first page and there are no more pages", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 10 OFFSET 0`)).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(2))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.ListGlobal(ctx, 10, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, actualTotal)
		assert.Len(t, dest, 2)
		assert.Equal(t, peter, dest[0])
		assert.Equal(t, homer, dest[1])
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns full page and has next page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 2 OFFSET 0`)).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.ListGlobal(ctx, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 2)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns many pages and I can traverse it using cursor", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rowsForPage1 := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		rowsForPage2 := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 1 OFFSET 0`)).WillReturnRows(rowsForPage1)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 1 OFFSET 1`)).WillReturnRows(rowsForPage2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))

		ctx := persistence.SaveToContext(context.TODO(), db)
		var first UserCollection

		actualFirstPage, actualTotal, err := sut.ListGlobal(ctx, 1, "", "id", &first)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, first, 1)
		assert.True(t, actualFirstPage.HasNextPage)
		assert.NotEmpty(t, actualFirstPage.EndCursor)

		var second UserCollection
		actualSecondPage, actualTotal, err := sut.ListGlobal(ctx, 1, actualFirstPage.EndCursor, "id", &second)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, second, 1)
		assert.True(t, actualSecondPage.HasNextPage)
		assert.NotEmpty(t, actualSecondPage.EndCursor)
	})

	t.Run("returns page without conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 2 OFFSET 0`)).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.ListGlobal(ctx, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns page with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id, first_name, last_name, age FROM users WHERE first_name = $1 AND age != $2 ORDER BY id LIMIT 2 OFFSET 0")).
			WithArgs("Peter", 18).
			WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM users WHERE first_name = $1 AND age != $2")).
			WithArgs("Peter", 18).
			WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(100))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		conditions := repo.Conditions{
			repo.NewEqualCondition("first_name", "Peter"),
			repo.NewNotEqualCondition("age", 18),
		}

		actualPage, actualTotal, err := sut.ListGlobal(ctx, 2, "", "id", &dest, conditions...)
		require.NoError(t, err)
		assert.Equal(t, 100, actualTotal)
		assert.Len(t, dest, 1)
		assert.True(t, actualPage.HasNextPage)
		assert.NotEmpty(t, actualPage.EndCursor)
	})

	t.Run("returns empty page", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 2 OFFSET 0`)).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM users`)).WillReturnRows(sqlmock.NewRows([]string{""}).AddRow(0))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		actualPage, actualTotal, err := sut.ListGlobal(ctx, 2, "", "id", &dest)
		require.NoError(t, err)
		assert.Equal(t, 0, actualTotal)
		assert.Empty(t, dest)
		assert.False(t, actualPage.HasNextPage)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, _, err := sut.ListGlobal(ctx, 2, "", "id", nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.ListGlobal(ctx, 2, "zzz", "", nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error if wrong pagination attributes", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, _, err := sut.ListGlobal(ctx, -3, "", "id", nil)
		require.EqualError(t, err, "while converting offset and limit to cursor: Invalid data [reason=page size cannot be smaller than 1]")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, _, err := sut.ListGlobal(ctx, 2, "", "id", &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})

	t.Run("context properly canceled", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		ctx = persistence.SaveToContext(ctx, db)
		var dest UserCollection

		_, _, err := sut.ListGlobal(ctx, 2, "", "id", &dest)

		require.EqualError(t, err, "Internal Server Error: Maximum processing timeout reached")
	})

	t.Run("returns error on calculating total count", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "first_name", "last_name", "age"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id, first_name, last_name, age FROM users ORDER BY id LIMIT 2 OFFSET 0`)).WillReturnRows(rows)
		mock.ExpectQuery(`SELECT COUNT\(\*\).*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, _, err := sut.ListGlobal(ctx, 2, "", "id", &dest)
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
