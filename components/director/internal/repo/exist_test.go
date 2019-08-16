package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/stretchr/testify/require"
)

func TestExist(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	givenQuery := "SELECT 1 FROM users WHERE tenant_col = $1 AND id_col = $2"
	escapedQuery := regexp.QuoteMeta(givenQuery)
	sut := repo.NewExistQuerier("users", "tenant_col")

	t.Run("success when exist", func(t *testing.T) {
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

	t.Run("success when does not exist", func(t *testing.T) {
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

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(escapedQuery).WithArgs(givenTenant, givenID).WillReturnError(someError())
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

func TestExistWithManyConditions(t *testing.T) {
	// GIVEN
	givenTenant := uuidB()
	expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_col = $1 AND first_name = $2 AND last_name = $3")
	sut := repo.NewExistQuerier("users", "tenant_col")
	db, mock := testdb.MockDatabase(t)
	ctx := persistence.SaveToContext(context.TODO(), db)
	defer mock.AssertExpectations(t)
	mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
	// WHEN
	_, err := sut.Exists(ctx, givenTenant, repo.Conditions{{Field: "first_name", Val: "john"}, {Field: "last_name", Val: "doe"}})
	// THEN
	require.NoError(t, err)
}
