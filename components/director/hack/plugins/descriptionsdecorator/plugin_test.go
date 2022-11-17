package descriptionsdecorator_test

import (
	"errors"
	"os"
	"testing"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/kyma-incubator/compass/components/director/hack/plugins/descriptionsdecorator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutateConfig(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		cfg, err := config.LoadConfig("testdata/config.yaml")
		require.NoError(t, err)
		testOutputFile := "testdata/test_output.graphql"
		testExamplesDir := "testdata/examples"
		d := descriptionsdecorator.NewPlugin(testOutputFile, testExamplesDir)
		err = d.MutateConfig(cfg)
		defer func(t *testing.T) {
			err = os.Remove(testOutputFile)
			require.NoError(t, err)
		}(t)
		require.NoError(t, err)

		actual, err := os.ReadFile(testOutputFile)
		require.NoError(t, err)
		expected, err := os.ReadFile("testdata/expected.graphql")
		require.NoError(t, err)
		assert.Equal(t, string(expected), string(actual))
	})

	t.Run("No examples directory", func(t *testing.T) {
		// GIVEN
		cfg, err := config.LoadConfig("testdata/config.yaml")
		require.NoError(t, err)
		testOutputFile := "testdata/test_output.graphql"
		testExamplesDir := "testdata/examples"
		d := descriptionsdecorator.NewPlugin(testOutputFile, testExamplesDir)
		err = d.MutateConfig(cfg)
		defer func(t *testing.T) {
			err = os.Remove(testOutputFile)
			require.NoError(t, err)
		}(t)
		require.Nil(t, err)
	})

	t.Run("No config file", func(t *testing.T) {
		// GIVEN
		_, err := config.LoadConfig("testdata/no_config.yaml")
		testerr := errors.New("unable to read config: open testdata/no_config.yaml: no such file or directory")
		assert.EqualError(t, err, testerr.Error())
	})

	t.Run("Wrong schema in config file", func(t *testing.T) {
		// GIVEN
		cfg, err := config.LoadConfig("testdata/wrong_config.yaml")
		require.NoError(t, err)
		testOutputFile := "testdata/test_output.graphql"
		testExamplesDir := "testdata/examples"
		d := descriptionsdecorator.NewPlugin(testOutputFile, testExamplesDir)
		err = d.MutateConfig(cfg)
		defer func(t *testing.T) {
			err = os.Remove(testOutputFile)
			require.NoError(t, err)
		}(t)
		assert.Nil(t, err)
	})
}
