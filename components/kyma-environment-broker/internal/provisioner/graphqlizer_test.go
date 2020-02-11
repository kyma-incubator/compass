package provisioner

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKymaConfigToGraphQLAllParametersProvided(t *testing.T) {
	// given
	fixInput := gqlschema.KymaConfigInput{
		Version: "966",
		Components: []*gqlschema.ComponentConfigurationInput{
			{
				Component: "pico",
				Namespace: "bello",
			},
			{
				Component: "hakuna",
				Namespace: "matata",
				Configuration: []*gqlschema.ConfigEntryInput{
					{
						Key:    "testing-secret-key",
						Value:  "testing-secret-value",
						Secret: ptr.Bool(true),
					},
					{
						Key:   "testing-public-key",
						Value: "testing-public-value",
					},
				},
			},
		},
	}
	expRender := `{
		version: "966"
        components: [
          {
            component: "pico"
            namespace: "bello" 
          }
          {
            component: "hakuna"
            namespace: "matata"
            configuration: [
              {
                key: "testing-secret-key"
                value: "testing-secret-value"
                secret: true
              }
              {
                key: "testing-public-key"
                value: "testing-public-value"
              } 
            ] 
          } 
        ]         
	}`

	sut := Graphqlizer{}

	// when
	gotRender, err := sut.KymaConfigToGraphQL(fixInput)

	// then
	require.NoError(t, err)

	assert.Equal(t, expRender, gotRender)
}

func TestKymaConfigToGraphQLOnlyKymaVersion(t *testing.T) {
	// given
	fixInput := gqlschema.KymaConfigInput{
		Version: "966",
	}
	expRender := `{
		version: "966"         
	}`

	sut := Graphqlizer{}

	// when
	gotRender, err := sut.KymaConfigToGraphQL(fixInput)

	// then
	require.NoError(t, err)

	assert.Equal(t, expRender, gotRender)
}
