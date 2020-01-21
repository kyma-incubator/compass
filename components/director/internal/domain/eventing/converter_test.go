package eventing

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
)

func Test_RuntimeEventingConfigurationToGraphQL(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.RuntimeEventingConfiguration
		Expected *graphql.RuntimeEventingConfiguration
	}{
		{
			Name:  "Valid input model",
			Input: fixRuntimeEventngCfgWithURL(t, dummyEventingURL),
			Expected: &graphql.RuntimeEventingConfiguration{
				DefaultURL: dummyEventingURL,
			},
		}, {
			Name:     "Nil input model",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			eventingCfgGQL := RuntimeEventingConfigurationToGraphQL(testCase.Input)

			require.Equal(t, testCase.Expected, eventingCfgGQL)
		})
	}
}

func Test_ApplicationEventingConfigurationToGraphQL(t *testing.T) {
	validURL := fixValidURL(t, dummyEventingURL)

	testCases := []struct {
		Name     string
		Input    *model.ApplicationEventingConfiguration
		Expected *graphql.ApplicationEventingConfiguration
	}{
		{
			Name: "Valid input model",
			Input: &model.ApplicationEventingConfiguration{
				EventingConfiguration: model.EventingConfiguration{
					DefaultURL: validURL,
				},
			},
			Expected: &graphql.ApplicationEventingConfiguration{
				DefaultURL: dummyEventingURL,
			},
		}, {
			Name:     "Nil input model",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			eventingCfgGQL := ApplicationEventingConfigurationToGraphQL(testCase.Input)

			require.Equal(t, testCase.Expected, eventingCfgGQL)
		})
	}
}
