package repo_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestExist(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	givenQuery := "select 1 from users where id_col=$1 and tenant_col=$2"
	escapedQuery := regexp.QuoteMeta(givenQuery)
	sut := repo.NewExistQuerier("users","id_col","tenant_col")

	t.Run("when exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{""}).AddRow("1")
		mock.ExpectQuery(escapedQuery).WithArgs(givenID, givenTenant).WillReturnRows(rows)
		// WHEN
		ex, err := sut.Exists(ctx, givenID, givenTenant)
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("when does not exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		rows := sqlmock.NewRows([]string{""})
		mock.ExpectQuery(escapedQuery).WithArgs(givenID, givenTenant).WillReturnRows(rows)
		// WHEN
		ex, err := sut.Exists(ctx, givenID, givenTenant)
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("returns error on db operation failed", func(t *testing.T) {
		// GIVEN
		givenErr := errors.New("some error")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(escapedQuery).WithArgs(givenID, givenTenant).WillReturnError(givenErr)
		// WHEN
		_, err := sut.Exists(ctx, givenID, givenTenant)
		// THEN
		require.EqualError(t, err, "while getting object from DB: some error")

	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.Exists(ctx, givenID, givenTenant)
		require.EqualError(t, err, "unable to fetch database from context")
	})
}
