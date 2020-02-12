package runtime_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/stretchr/testify/assert"
)

func TestGenericComponentDisabler(t *testing.T) {
	type toDisable struct {
		Name      string
		Namespace string
	}
	tests := []struct {
		name            string
		givenComponents internal.ComponentConfigurationInputList
		expComponents   internal.ComponentConfigurationInputList
		toDisable       toDisable
	}{
		{
			name: "Disable component if the name and namespace match with predicate",
			toDisable: toDisable{
				Name:      "ory",
				Namespace: "ory-system",
			},
			givenComponents: internal.ComponentConfigurationInputList{
				{Component: "dex", Namespace: "kyma-system"},
				{Component: "ory", Namespace: "ory-system"},
			},
			expComponents: internal.ComponentConfigurationInputList{
				{Component: "dex", Namespace: "kyma-system"},
			},
		},
		{
			name: "Disable component if only name match with predicate but namespace not",
			toDisable: toDisable{
				Name:      "ory",
				Namespace: "wrong-namespace-name",
			},
			givenComponents: internal.ComponentConfigurationInputList{
				{Component: "dex", Namespace: "kyma-system"},
				{Component: "ory", Namespace: "ory-system"},
			},
			expComponents: internal.ComponentConfigurationInputList{
				{Component: "dex", Namespace: "kyma-system"},
				{Component: "ory", Namespace: "ory-system"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// given
			sut := runtime.NewGenericComponentDisabler(test.toDisable.Name, test.toDisable.Namespace)

			// when
			modifiedComponents := sut.Disable(test.givenComponents)

			// then
			assert.EqualValues(t, test.expComponents, modifiedComponents)
		})
	}
}
