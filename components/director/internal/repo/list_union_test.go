package repo_test

import (
	"context"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnionList(t *testing.T) {
	givenTenant := uuidB()
	peterID := uuidA()
	homerID := uuidC()
	peter := User{FirstName: "Peter", LastName: "Griffin", Age: 40, Tenant: givenTenant, ID: peterID}
	peterRow := []driver.Value{peterID, givenTenant, "Peter", "Griffin", 40}
	homer := User{FirstName: "Homer", LastName: "Simpson", Age: 55, Tenant: givenTenant, ID: homerID}
	homerRow := []driver.Value{homerID, givenTenant, "Homer", "Simpson", 55}

	sut := repo.NewUnionLister("UserType", "users", "tenant_id",
		[]string{"id_col", "tenant_id", "first_name", "last_name", "age"})

	t.Run("success", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"id_col", "tenant_id", "first_name", "last_name", "age"}).
			AddRow(peterRow...).
			AddRow(homerRow...)

		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("(SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND id_col = $2 ORDER BY id_col ASC LIMIT $3 OFFSET $4) UNION (SELECT id_col, tenant_id, first_name, last_name, age FROM users WHERE %s AND id_col = $6 ORDER BY id_col ASC LIMIT $7 OFFSET $8)", fixTenantIsolationSubqueryWithArgPosition(1), fixTenantIsolationSubqueryWithArgPosition(5)))).WithArgs(givenTenant, peterID, 10, 0, givenTenant, homerID, 10, 0).WillReturnRows(rows)
		mock.ExpectQuery(regexp.QuoteMeta(fmt.Sprintf("SELECT id_col AS id, COUNT(*) AS total_count FROM users WHERE %s GROUP BY id_col ORDER BY id_col ASC", fixTenantIsolationSubquery()))).WithArgs(givenTenant).WillReturnRows(sqlmock.NewRows([]string{"id", "total_count"}).AddRow(peterID, 1).AddRow(homerID, 1))
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		counts, err := sut.List(ctx, givenTenant, []string{peterID, homerID}, "id_col", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id_col")}, &dest)
		require.NoError(t, err)
		assert.Equal(t, 2, len(counts))
		assert.Equal(t, 1, counts[peterID])
		assert.Equal(t, 1, counts[homerID])
		assert.Len(t, dest, 2)
		assert.Equal(t, peter, dest[0])
		assert.Equal(t, homer, dest[1])
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.List(ctx, givenTenant, []string{peterID, homerID}, "id_col", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id_col")}, nil)
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})

	t.Run("returns error if wrong cursor", func(t *testing.T) {
		ctx := persistence.SaveToContext(context.TODO(), &sqlx.Tx{})
		_, err := sut.List(ctx, givenTenant, []string{peterID, homerID}, "id_col", 10, "zzz", repo.OrderByParams{repo.NewAscOrderBy("id_col")}, nil)
		require.EqualError(t, err, "while decoding page cursor: cursor is not correct: illegal base64 data at input byte 0")
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		db, mock := testdb.MockDatabase(t)
		defer mock.AssertExpectations(t)

		mock.ExpectQuery(`SELECT .*`).WillReturnError(someError())
		ctx := persistence.SaveToContext(context.TODO(), db)
		var dest UserCollection

		_, err := sut.List(ctx, givenTenant, []string{peterID, homerID}, "id_col", 10, "", repo.OrderByParams{repo.NewAscOrderBy("id_col")}, &dest)

		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")
	})
}
