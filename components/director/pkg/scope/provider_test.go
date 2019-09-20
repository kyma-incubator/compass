package scope_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("returns error on invalid yaml", func(t *testing.T) {
		// GIVEN
		sut := scope.NewProvider("testdata/invalid.yaml")
		// WHEN
		err := sut.Load()
		// THEN
		require.EqualError(t, err, "while converting YAML to JSON: yaml: found unexpected end of stream")
	})
}

func TestScopeProviderGetRequiredScopes(t *testing.T) {
	t.Run("GetRequiredScopes requires Load", func(t *testing.T) {
		sut := scope.NewProvider("anything")
		_, err := sut.GetRequiredScopes("queries.runtime")
		require.Error(t, err, "required scopes configuration not loaded")
	})

	// GIVEN
	sut := scope.NewProvider("testdata/valid.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns single scope", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetRequiredScopes("queries.runtime")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, actual, []string{"runtime:get"})
	})

	t.Run("returns many scopes", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetRequiredScopes("mutations.createApplication")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, actual, []string{"application:create", "global:create"})
	})

	t.Run("returns error if required scopes are empty", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("mutations.empty")
		// THEN
		require.Equal(t, scope.RequiredScopesNotDefinedError, err)
	})

	t.Run("returns error if path not found", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("does.not.exist")
		// THEN
		require.EqualError(t, err, "while searching configuration using path $.does.not.exist: key error: does not found in object")
	})

	t.Run("return error if path is invalid", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("...queries")
		// THEN
		require.Error(t, err, "while searching configuration using path $....queries: expression don't support in filter")
	})

	t.Run("returns error if path points to invalid type", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("queries")
		// THEN
		require.EqualError(t, err, "unexpected scopes definition, should be string or list of strings, but was map[string]interface {}")

	})

	t.Run("returns error if path points to list with invalid types", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("mutations.updateApplication")
		// THEN
		require.EqualError(t, err, "unexpected scope value in a list, should be string but was float64")

	})

}
