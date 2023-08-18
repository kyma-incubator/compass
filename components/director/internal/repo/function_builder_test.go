package repo_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAdvisoryLockGlobal(t *testing.T) {
	sut := repo.NewFunctionBuilder()
	var expectedIdentifier int64 = 1101223

	t.Run("success with rebuild and identifier", func(t *testing.T) {
		// GIVEN
		expectedQuery := "SELECT pg_try_advisory_xact_lock($1)"

		// WHEN
		query, args, err := sut.BuildAdvisoryLock(expectedIdentifier)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, 1, len(args))
		assert.Equal(t, expectedIdentifier, args[0])
		assert.Equal(t, expectedQuery, removeWhitespace(query))
	})
}
