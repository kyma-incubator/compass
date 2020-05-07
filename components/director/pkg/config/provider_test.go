package config_test

import (
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvider_Load(t *testing.T) {
	t.Run("returns error when file not found", func(t *testing.T) {
		// GIVEN
		sut := config.NewProvider("not_existing_file.yaml")
		// WHEN
		err := sut.Load()
		// THEN
		require.Error(t, err)
		assert.True(t, strings.HasPrefix(err.Error(), "while reading file not_existing_file.yaml"))
	})

	t.Run("returns error on invalid yaml", func(t *testing.T) {
		// GIVEN
		sut := config.NewProvider("testdata/invalid.yaml")
		// WHEN
		err := sut.Load()
		// THEN
		require.EqualError(t, err, "while unmarshalling YAML: error converting YAML to JSON: yaml: found unexpected end of stream")
	})
}
