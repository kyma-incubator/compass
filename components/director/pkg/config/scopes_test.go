package config_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_GetRequiredScopes(t *testing.T) {
	t.Run("requires Load", func(t *testing.T) {
		sut := config.NewProvider("anything")
		_, err := sut.GetRequiredScopes("graphql.query.runtime")
		require.Error(t, err, "required scopes configuration not loaded")
	})

	// GIVEN
	sut := config.NewProvider("testdata/valid.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns single scope", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetRequiredScopes("graphql.query.runtime")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, actual, []string{"runtime:get"})
	})

	t.Run("returns many scopes", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetRequiredScopes("graphql.mutation.createApplication")
		// THEN
		require.NoError(t, err)
		assert.Equal(t, actual, []string{"application:create", "global:create"})
	})

	t.Run("returns error if required scopes are empty", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("graphql.mutation.empty")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "required scopes are not defined")
	})

	t.Run("returns error if path not found", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("does.not.exist")
		// THEN
		require.EqualError(t, err, "while searching configuration using path $.does.not.exist: key error: does not found in object")
	})

	t.Run("return error if path is invalid", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("...graphql")
		// THEN
		require.Error(t, err, "while searching configuration using path $....graphql: expression don't support in filter")
	})

	t.Run("returns error if path points to invalid type", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("graphql.query")
		// THEN
		require.EqualError(t, err, "unexpected scopes definition, should be string or list of strings, but was map[string]interface {}")
	})

	t.Run("returns error if path points to list with invalid types", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("graphql.mutation.updateApplication")
		// THEN
		require.EqualError(t, err, "unexpected float64 value in a string list")
	})
}

func TestProvider_GetRequiredGrantTypes(t *testing.T) {
	const grantTypesPath = "clientCredentialsRegistrationGrantTypes"
	var expectedGrantTypes = []string{"client_credentials"}

	t.Run("requires Load", func(t *testing.T) {
		sut := config.NewProvider("anything")
		_, err := sut.GetRequiredGrantTypes(grantTypesPath)
		require.Error(t, err, "required scopes configuration not loaded")
	})

	// GIVEN
	sut := config.NewProvider("testdata/valid.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns grant types", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetRequiredGrantTypes(grantTypesPath)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, actual, expectedGrantTypes)
	})

	t.Run("returns error when single value is provided instead of a list", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredGrantTypes("graphql.query.runtime")
		// THEN
		require.EqualError(t, err, "unexpected grant_types definition, should be a list of strings, but was string")
	})

	t.Run("returns error if required scopes are empty", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredGrantTypes("graphql.mutation.empty")
		// THEN
		require.Error(t, err)
		assert.EqualError(t, err, "required scopes are not defined")
	})

	t.Run("returns error if path not found", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredGrantTypes("does.not.exist")
		// THEN
		require.EqualError(t, err, "while searching configuration using path $.does.not.exist: key error: does not found in object")
	})

	t.Run("return error if path is invalid", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredGrantTypes("...graphql")
		// THEN
		require.Error(t, err, "while searching configuration using path $....graphql: expression don't support in filter")
	})

	t.Run("returns error if path points to invalid type", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredGrantTypes("graphql.query")
		// THEN
		require.EqualError(t, err, "unexpected grant_types definition, should be a list of strings, but was map[string]interface {}")
	})

	t.Run("returns error if path points to list with invalid types", func(t *testing.T) {
		// WHEN
		_, err := sut.GetRequiredScopes("graphql.mutation.updateApplication")
		// THEN
		require.EqualError(t, err, "unexpected float64 value in a string list")
	})
}
