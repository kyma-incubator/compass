package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestExist(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	givenQuery := "SELECT 1 FROM users WHERE tenant_col = $1 AND id_col = $2"
	escapedQuery := regexp.QuoteMeta(givenQuery)
	sut := repo.NewExistQuerier("users", "tenant_col")

	t.Run("when exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(escapedQuery).WithArgs(givenTenant, givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("when does not exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(escapedQuery).WithArgs(givenTenant, givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.Exists(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
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
		mock.ExpectQuery(escapedQuery).WithArgs(givenTenant, givenID).WillReturnError(givenErr)
		// WHEN
		_, err := sut.Exists(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		// THEN
		require.EqualError(t, err, "while getting object from DB: some error")

	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.Exists(ctx, givenTenant, repo.Conditions{{Field: "id_col", Val: givenID}})
		require.EqualError(t, err, "unable to fetch database from context")
	})
}
