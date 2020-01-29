package runtime_test

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/runtime"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLokiDisabler(t *testing.T) {
	// given
	sut := runtime.NewLokiDisabler()

	givenComponents := internal.ComponentConfigurationInputList{
		{
			Component: "logging",
			Namespace: "kyma-system",
		},
	}
	expComponents := internal.ComponentConfigurationInputList{
		{
			Component: "logging",
			Namespace: "kyma-system",
			Configuration: []*gqlschema.ConfigEntryInput{
				{
					Key:   "loki.enabled",
					Value: "false",
				},
				{
					Key:   "fluent-bit.conf.Output.loki.enabled",
					Value: "false",
				},
			},
		},
	}

	// when
	modifiedComponents := sut.Disable(givenComponents)

	// then
	assert.EqualValues(t, expComponents, modifiedComponents)
}
