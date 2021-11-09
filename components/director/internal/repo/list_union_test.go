package repo_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestUnionList(t *testing.T) {
	sut := repo.NewUnionLister(appTableName, appColumns)
	resourceType := resource.Application
	m2mTable, ok := resourceType.TenantAccessTable()
	require.True(t, ok)

	t.Run("success", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription).
			AddRow(appID2, appName2, appDescription2)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("(SELECT id, name, description FROM %s WHERE %s AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, name, description FROM %s WHERE %s AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1"), appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$5")))).
			WithArgs(tenantID, appID, 10, 0, tenantID, appID2, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id AS id, COUNT(*) AS total_count FROM %s WHERE %s GROUP BY id ORDER BY id ASC",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1") ))).
			WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(appID, 1).AddRow(appID2, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		counts, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, len(counts))
		assert.Equal(t, 1, counts[appID])
		assert.Equal(t, 1, counts[appID2])
		assert.Len(t, dest, 2)
		assert.Equal(t, *fixApp, dest[0])
		assert.Equal(t, *fixApp2, dest[1])
	})

	t.Run("success with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("(SELECT id, name, description FROM %s WHERE name = $1 AND %s AND id = $3 ORDER BY id ASC LIMIT $4 OFFSET $5) UNION (SELECT id, name, description FROM %s WHERE name = $6 AND %s AND id = $8 ORDER BY id ASC LIMIT $9 OFFSET $10)",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2"), appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$7")))).
			WithArgs(appName, tenantID, appID, 10, 0, appName, tenantID, appID2, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id AS id, COUNT(*) AS total_count FROM %s WHERE name = $1 AND %s GROUP BY id ORDER BY id ASC",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$2")))).
			WithArgs(appName, tenantID).WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(appID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		counts, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest, repo.NewEqualCondition("name", appName))
		require.NoError(t, err)
		assert.Equal(t, 1, len(counts))
		assert.Equal(t, 1, counts[appID])
		assert.Len(t, dest, 1)
		assert.Equal(t, *fixApp, dest[0])
	})

	t.Run("error when union list fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("(SELECT id, name, description FROM %s WHERE %s AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, name, description FROM %s WHERE %s AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1"), appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$5")))).
			WithArgs(tenantID, appID, 10, 0, tenantID, appID2, 10, 0).WillReturnError(someError())
		var dest AppCollection

		counts, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})

	t.Run("error when count fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows(appColumns).
			AddRow(appID, appName, appDescription).
			AddRow(appID2, appName2, appDescription2)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("(SELECT id, name, description FROM %s WHERE %s AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, name, description FROM %s WHERE %s AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1"), appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$5")))).
			WithArgs(tenantID, appID, 10, 0, tenantID, appID2, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id AS id, COUNT(*) AS total_count FROM %s WHERE %s GROUP BY id ORDER BY id ASC",
			appTableName, fmt.Sprintf(tenantIsolationConditionWithoutOwnerCheckFmt, m2mTable, "$1") ))).
			WithArgs(tenantID).WillReturnError(someError())
		var dest AppCollection

		counts, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if empty tenant", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.List(ctx, resourceType, "", []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, apperrors.NewTenantRequiredError().Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "zzz", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest AppCollection

		_, err := sut.List(ctx, resourceType, tenantID, []string{appID, appID2}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestUnionListWithEmbeddedTenant(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, Tenant: tenantID, ID: peterID}
	peterRow := []driver.Value{peterID, tenantID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, Tenant: tenantID, ID: homerID}
	homerRow := []driver.Value{homerID, tenantID, "Homer", "Simpson", 55}

	sut := repo.NewUnionListerWithEmbeddedTenant(userTableName, "tenant_id", []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $5 AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)")).
			WithArgs(tenantID, peterID, 10, 0, tenantID, homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users WHERE tenant_id = $1 GROUP BY id ORDER BY id ASC")).
			WithArgs(tenantID).WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(peterID, 1).AddRow(homerID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, len(counts))
		assert.Equal(t, 1, counts[peterID])
		assert.Equal(t, 1, counts[homerID])
		assert.Len(t, dest, 2)
		assert.Equal(t, peter, dest[0])
		assert.Equal(t, homer, dest[1])
	})

	t.Run("success with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND first_name = $2 AND id = $3 ORDER BY id ASC LIMIT $4 OFFSET $5) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $6 AND first_name = $7 AND id = $8 ORDER BY id ASC LIMIT $9 OFFSET $10)")).
			WithArgs(tenantID, "Peter", peterID, 10, 0, tenantID, "Peter", homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users WHERE tenant_id = $1 AND first_name = $2 GROUP BY id ORDER BY id ASC")).
			WithArgs(tenantID, "Peter").WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(peterID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest, repo.NewEqualCondition("first_name", "Peter"))
		require.NoError(t, err)
		assert.Equal(t, 1, len(counts))
		assert.Equal(t, 1, counts[peterID])
		assert.Len(t, dest, 1)
		assert.Equal(t, peter, dest[0])
	})

	t.Run("error when union list fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $5 AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)")).
			WithArgs(tenantID, peterID, 10, 0, tenantID, homerID, 10, 0).WillReturnError(someError())
		var dest UserCollection

		counts, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})

	t.Run("error when count fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $1 AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE tenant_id = $5 AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)")).
			WithArgs(tenantID, peterID, 10, 0, tenantID, homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users WHERE tenant_id = $1 GROUP BY id ORDER BY id ASC")).
			WithArgs(tenantID).WillReturnError(someError())
		var dest UserCollection

		counts, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "zzz", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, err := sut.List(ctx, UserType, tenantID, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}

func TestUnionListGlobal(t *testing.T) {
	peterID := "peterID"
	homerID := "homerID"
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, Tenant: tenantID, ID: peterID}
	peterRow := []driver.Value{peterID, tenantID, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, Tenant: tenantID, ID: homerID}
	homerRow := []driver.Value{homerID, tenantID, "Homer", "Simpson", 55}

	sut := repo.NewUnionListerGlobal(UserType, userTableName, []string{"id", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $4 ORDER BY id ASC LIMIT $5 OFFSET $6)")).
			WithArgs(peterID, 10, 0, homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users GROUP BY id ORDER BY id ASC")).
			WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(peterID, 1).AddRow(homerID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, len(counts))
		assert.Equal(t, 1, counts[peterID])
		assert.Equal(t, 1, counts[homerID])
		assert.Len(t, dest, 2)
		assert.Equal(t, peter, dest[0])
		assert.Equal(t, homer, dest[1])
	})

	t.Run("success with additional conditions", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name = $1 AND id = $2 ORDER BY id ASC LIMIT $3 OFFSET $4) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE first_name = $5 AND id = $6 ORDER BY id ASC LIMIT $7 OFFSET $8)")).
			WithArgs("Peter", peterID, 10, 0, "Peter", homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users WHERE first_name = $1 GROUP BY id ORDER BY id ASC")).
			WithArgs("Peter").WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(peterID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest, repo.NewEqualCondition("first_name", "Peter"))
		require.NoError(t, err)
		assert.Equal(t, 1, len(counts))
		assert.Equal(t, 1, counts[peterID])
		assert.Len(t, dest, 1)
		assert.Equal(t, peter, dest[0])
	})

	t.Run("error when union list fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $4 ORDER BY id ASC LIMIT $5 OFFSET $6)")).
			WithArgs(peterID, 10, 0, homerID, 10, 0).WillReturnError(someError())
		var dest UserCollection

		counts, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})


	t.Run("error when count fails", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta("(SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $1 ORDER BY id ASC LIMIT $2 OFFSET $3) UNION (SELECT id, tenant_id, first_name, last_name, age FROM users WHERE id = $4 ORDER BY id ASC LIMIT $5 OFFSET $6)")).
			WithArgs(peterID, 10, 0, homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id AS id, COUNT(*) AS total_count FROM users GROUP BY id ORDER BY id ASC")).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)
		require.Error(t, err)
		require.EqualError(t, err,"Internal Server Error: Unexpected error while executing SQL query")
		require.Nil(t, counts)
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "zzz", repo.OrderByParams{repo.NewAscOrderBy("id")}, nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, err := sut.ListGlobal(ctx, []string{peterID, homerID}, "id", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id")}, &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
