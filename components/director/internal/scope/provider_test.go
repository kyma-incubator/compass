package scope_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestScopeProviderLoad(t *testing.T) {

	t.Run("returns error when file not found", func(t *testing.T) {
		// GIVEN
		sut := scope.NewProvider("not_existing_file.yaml")
		// WHEN
		err := sut.Load()
		// THEN
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "while reading file not_existing_file.yaml"))
	})

	t.Run("returns error when file is invalid YAML", func(t *testing.T) {
		// GIVEN
		sut := scope.NewProvider("testdata/invalid.yaml")
		// WHEN
		err := sut.Load()
		// THEN
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "while converting YAML to JSON"))

	})
}

func TestScopeProviderGetRequiredScopes(t *testing.T) {
	// GIVEN
	sut := scope.NewProvider("testdata/valid.yaml")
	require.NoError(t, sut.Load())
	t.Run("returns error if preceding Load failed", func(t *testing.T) {
		// WHEN
		// THEN
	})

	t.Run("returns single scope", func(t *testing.T) {
		actual, err := sut.GetRequiredScopes("queries.runtime")
		require.NoError(t, err)
		fmt.Println(actual)
	})

	t.Run("returns many scopes", func(t *testing.T) {
		actual, err := sut.GetRequiredScopes("mutations.createApplication")
		require.NoError(t, err)
		fmt.Println(actual)
	})

	t.Run("returns no scopes", func(t *testing.T) {

	})
}
