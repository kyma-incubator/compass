package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/require"
)

func TestExist(t *testing.T) {
	givenID := uuidA()
	givenTenant := uuidB()
	sut := repo.NewExistQuerier(UserType, "users", "tenant_id")

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(defaultExpectedExistQuery()).WithArgs(givenTenant, givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(defaultExpectedExistQuery()).WithArgs(givenTenant, givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.Exists(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.Exists(ctx, givenTenant, repo.Conditions{})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE tenant_id = $1 AND first_name = $2 AND last_name = $3")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenTenant, "john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.Exists(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(defaultExpectedExistQuery()).WithArgs(givenTenant, givenID).WillReturnError(someError())
		// WHEN
		_, err := sut.Exists(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")

	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.Exists(ctx, givenTenant, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

func TestExistGlobal(t *testing.T) {
	givenID := uuidA()
	sut := repo.NewExistQuerierGlobal(UserType, "users")

	t.Run("success when exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when does not exist", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.NoError(t, err)
		require.False(t, ex)
	})

	t.Run("success when no conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WillReturnRows(testdb.RowWhenObjectExist())
		// WHEN
		ex, err := sut.ExistsGlobal(ctx, repo.Conditions{})
		// THEN
		require.NoError(t, err)
		require.True(t, ex)
	})

	t.Run("success when more conditions", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE first_name = $1 AND last_name = $2")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs("john", "doe").WillReturnRows(testdb.RowWhenObjectDoesNotExist())
		// WHEN
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("first_name", "john"), repo.NewEqualCondition("last_name", "doe")})
		// THEN
		require.NoError(t, err)
	})

	t.Run("returns error when operation on db failed", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT 1 FROM users WHERE id_col = $1")
		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)
		mock.ExpectQuery(expectedQuery).WithArgs(givenID).WillReturnError(someError())
		// WHEN
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		// THEN
		require.EqualError(t, err, "Internal Server Error: Unexpected error while executing SQL query")

	})

	t.Run("returns error if missing persistence context", func(t *testing.T) {
		ctx := context.TODO()
		_, err := sut.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id_col", givenID)})
		require.EqualError(t, err, apperrors.NewInternalError("unable to fetch database from context").Error())
	})
}

func defaultExpectedExistQuery() string {
	givenQuery := "SELECT 1 FROM users WHERE tenant_id = $1 AND id_col = $2"
	return regexp.QuoteMeta(givenQuery)
}
