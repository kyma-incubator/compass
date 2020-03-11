package input

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputBuilderFactoryForAzurePlan(t *testing.T) {
	// given
	var (
		fixID               = "fix-id"
		inputComponentList  = fixKymaComponentList()
		mappedComponentList = mapToGQLComponentConfigurationInput(inputComponentList)
		toDisableComponents = []string{"kiali"}
		kebOverrides        = []*gqlschema.ConfigEntryInput{
			{Key: "key-1", Value: "pico"},
			{Key: "key-2", Value: "bello", Secret: ptr.Bool(true)},
		}
	)

	optComponentsSvc := &automock.OptionalComponentService{}
	defer optComponentsSvc.AssertExpectations(t)
	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return(toDisableComponents)
	optComponentsSvc.On("ExecuteDisablers", mappedComponentList, toDisableComponents[0]).Return(mappedComponentList, nil)

	config := Config{
		URL: "",
	}
	factory := NewInputBuilderFactory(optComponentsSvc, inputComponentList, config, "1.10.0")

	// when
	builder, found := factory.ForPlan(broker.AzurePlanID)

	// then
	require.True(t, found)

	// when
	input, err := builder.
		SetProvisioningParameters(internal.ProvisioningParametersDTO{
			Name:         "azure-cluster",
			TargetSecret: ptr.String("azure-secret"),
		}).
		SetRuntimeLabels(fixID, fixID).
		SetOverrides("keb", kebOverrides).Create()

	// then
	require.NoError(t, err)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)
	assert.Equal(t, "azure-cluster", input.RuntimeInput.Name)
	assert.Equal(t, "azure", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "azure-secret", input.ClusterConfig.GardenerConfig.TargetSecret)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)
	assert.Equal(t, &gqlschema.Labels{
		brokerKeyPrefix + "instance_id":   []string{fixID},
		globalKeyPrefix + "subaccount_id": []string{fixID},
	}, input.RuntimeInput.Labels)

	assertOverrides(t, "keb", input.KymaConfig.Components, kebOverrides)
}

func assertOverrides(t *testing.T, componentName string, components internal.ComponentConfigurationInputList, overrides []*gqlschema.ConfigEntryInput) {
	overriddenComponent, found := find(components, componentName)
	require.True(t, found)

	assert.Equal(t, overriddenComponent.Configuration, overrides)
}

func find(in internal.ComponentConfigurationInputList, name string) (*gqlschema.ComponentConfigurationInput, bool) {
	for _, c := range in {
		if c.Component == name {
			return c, true
		}
	}
	return nil, false
}

func fixKymaComponentList() []v1alpha1.KymaComponent {
	return []v1alpha1.KymaComponent{
		{Name: "dex", Namespace: "kyma-system"},
		{Name: "ory", Namespace: "kyma-system"},
		{Name: "keb", Namespace: "kyma-system"},
	}
}
