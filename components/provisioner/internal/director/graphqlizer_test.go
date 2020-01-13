package director

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphqlizer_RuntimeInputToGraphQL(t *testing.T) {
	t.Run("Should return valid graphqlized runtime input", func(t *testing.T) {
		//given
		runtimeInput := gqlschema.RuntimeInput{
			Name:        "test runtime",
			Description: util.StringPtr("wow, this is nice description!"),
			Labels:      &gqlschema.Labels{"Label": []string{"yup", "indeed"}},
		}

		expectedGraphlizedInput := `{
		name: "test runtime",
		description: "wow, this is nice description!",
		labels: {
			Label: ["yup","indeed"],
	},
	}`

		var graph graphqlizer

		//when
		actual, err := graph.RuntimeInputToGraphQL(runtimeInput)
		require.NoError(t, err)

		//then
		assert.Equal(t, expectedGraphlizedInput, actual)
	})

	t.Run("Should return valid graphqlized runtime input if optional fields are empty", func(t *testing.T) {
		//given
		runtimeInput := gqlschema.RuntimeInput{
			Name: "test runtime",
		}

		expectedGraphlizedInput := `{
		name: "test runtime",
	}`

		var graph graphqlizer

		//when
		actual, err := graph.RuntimeInputToGraphQL(runtimeInput)
		require.NoError(t, err)

		//then
		assert.Equal(t, expectedGraphlizedInput, actual)
	})
}
