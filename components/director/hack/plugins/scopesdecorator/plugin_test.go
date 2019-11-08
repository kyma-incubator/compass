package scopesdecorator_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/hack/plugins/scopesdecorator"

	"github.com/99designs/gqlgen/codegen/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMutateConfig(t *testing.T) {
	// GIVEN

	t.Run("Success", func(t *testing.T) {
		cfg, err := config.LoadConfig("testdata/config.yaml")
		require.NoError(t, err)
		testOutputFile := "testdata/test_output.graphql"
		sut := scopesdecorator.NewPlugin(testOutputFile)
		err = sut.MutateConfig(cfg)
		require.NoError(t, err)

		actual, err := ioutil.ReadFile(testOutputFile)
		require.NoError(t, err)

		expected, err := ioutil.ReadFile("testdata/expected.graphql")
		require.NoError(t, err)
		assert.Equal(t, string(expected), string(actual))
		err = os.Remove(testOutputFile)
		require.NoError(t, err)
	})
}
