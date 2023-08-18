package repo_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/internal/repo/testdb"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdvisoryLock(t *testing.T) {
	sut := repo.NewDBFunction()
	var expectedIdentifier int64 = 1101223

	t.Run("Success with identifier", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT pg_try_advisory_xact_lock($1)")

		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).AddRow(true)
		mock.ExpectQuery(expectedQuery).WithArgs(expectedIdentifier).WillReturnRows(rows)
		// WHEN
		result, err := sut.AdvisoryLock(ctx, expectedIdentifier)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, true)
	})

	t.Run("Error when multiple results are returned", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT pg_try_advisory_xact_lock($1)")

		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"}).AddRow(true).AddRow(true)
		mock.ExpectQuery(expectedQuery).WithArgs(expectedIdentifier).WillReturnRows(rows)
		// WHEN
		result, err := sut.AdvisoryLock(ctx, expectedIdentifier)
		// THEN
		require.Error(t, err)
		assert.Equal(t, result, false)
	})

	t.Run("Error when no results are returned", func(t *testing.T) {
		// GIVEN
		expectedQuery := regexp.QuoteMeta("SELECT pg_try_advisory_xact_lock($1)")

		db, mock := testdb.MockDatabase(t)
		ctx := persistence.SaveToContext(context.TODO(), db)
		defer mock.AssertExpectations(t)

		rows := sqlmock.NewRows([]string{"pg_try_advisory_xact_lock"})
		mock.ExpectQuery(expectedQuery).WithArgs(expectedIdentifier).WillReturnRows(rows)
		// WHEN
		result, err := sut.AdvisoryLock(ctx, expectedIdentifier)
		// THEN
		require.Error(t, err)
		assert.Equal(t, result, false)
	})

}
