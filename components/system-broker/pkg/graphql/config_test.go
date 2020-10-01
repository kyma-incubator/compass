package graphql_test

import (
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	var tests = []struct {
		Msg string
		ConfigProvider func () *graphql.Config
		ExpectValid bool
	} {
		{
			Msg: "Default config should be valid",
			ConfigProvider: func() *graphql.Config {
				return graphql.DefaultConfig()
			},
			ExpectValid: true,
		},
		{
			Msg: "Missing GraphQL endpoint should be invalid",
			ConfigProvider: func() *graphql.Config {
				config := graphql.DefaultConfig()
				config.GraphqlEndpoint = ""
				return config
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Msg, func(t *testing.T) {
			err := test.ConfigProvider().Validate()
			if test.ExpectValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
