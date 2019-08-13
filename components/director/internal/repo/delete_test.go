package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

// create table users(id uuid primary key, tenant uuid, first_name varchar(100),last_name varchar(100), age int);
// insert into users(id,tenant,first_name,last_name,age) values('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa','bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb','Adam','Sze',33);
// delete from users where id='aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' and tenant='bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb';

func TestDelete(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	expectedQuery := regexp.QuoteMeta("DELETE FROM users WHERE tenant_col=$1 AND id_col=$2")
	sut := repo.NewDeleter("users", "tenant_col", "id_col")

	t.Run("success", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(expectedQuery).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(-1, 1))
		// WHEN
		err := sut.Delete(ctx, givenTenant, givenID)
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error on db operation", func(t *testing.T) {
		// GIVEN
		givenErr := errors.New("some err")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(expectedQuery).WithArgs(givenTenant, givenID).WillReturnError(givenErr)
		// WHEN
		err := sut.Delete(ctx, givenTenant, givenID)
		// THEN
		require.EqualError(t, err, "while deleting from database: some err")
	})

	t.Run("returns error when object not found", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(expectedQuery).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(0, 12))
		// WHEN
		err := sut.Delete(ctx, givenTenant, givenID)
		// THEN
		require.EqualError(t, err, "delete should remove single row, but removed 12 rows")
	})

	t.Run("returns error when removed more than one object", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectExec(expectedQuery).WithArgs(givenTenant, givenID).WillReturnResult(sqlmock.NewResult(0, 0))
		// WHEN
		err := sut.Delete(ctx, givenTenant, givenID)
		// THEN
		require.EqualError(t, err, "delete should remove single row, but removed 0 rows")
	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		err := sut.Delete(ctx, givenTenant, givenID)
		require.EqualError(t, err, "unable to fetch database from context")
	})

}

func uuidA() string {
	return "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}

func uuidB() string {
	return "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
}

func uuidC() string {
	return "cccccccc-cccc-cccc-cccc-cccccccccccc"
}
