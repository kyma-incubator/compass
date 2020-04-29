package config_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_GetApplicationHideSelectors(t *testing.T) {
	t.Run("requires Load", func(t *testing.T) {
		sut := config.NewProvider("anything")
		_, err := sut.GetApplicationHideSelectors()
		require.Error(t, err, "required selectors configuration not loaded")
	})

	// GIVEN
	sut := config.NewProvider("testdata/valid.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns app hide selectors", func(t *testing.T) {
		expectedMap := map[string][]string{
			"applicationType": {"Test/App", "Work In Progress"},
			"second":          {"Single"},
		}
		// WHEN
		actual, err := sut.GetApplicationHideSelectors()
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedMap, actual)
	})

	sut = config.NewProvider("testdata/valid-hide-selectors-empty.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns app hide selectors as empty map when none specified", func(t *testing.T) {
		// WHEN
		actual, err := sut.GetApplicationHideSelectors()
		// THEN
		require.NoError(t, err)
		assert.Nil(t, actual)
	})

	sut = config.NewProvider("testdata/invalid-hide-selectors-invalid-format.yaml")
	require.NoError(t, sut.Load())

	t.Run("returns error when app hide selectors in invalid format", func(t *testing.T) {
		// WHEN
		_, err := sut.GetApplicationHideSelectors()
		// THEN
		require.Error(t, err)
	})
}
