package broker

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Currently only azure is supported

func TestInputBuilderFactoryForAzurePlan(t *testing.T) {
	// given
	var (
		inputComponentList  = fixKymaComponentList()
		mappedComponentList = mapToGQLComponentConfigurationInput(inputComponentList)
		toDisableComponents = []string{"kiali"}
		smOverrides         = internal.ServiceManagerEntryDTO{URL: "http://sm-pico-bello-url.com"}
	)

	optComponentsSvc := &automock.OptionalComponentService{}
	defer optComponentsSvc.AssertExpectations(t)
	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return(toDisableComponents)
	optComponentsSvc.On("ExecuteDisablers", mappedComponentList, toDisableComponents[0]).Return(mappedComponentList, nil)

	factory := NewInputBuilderFactory(optComponentsSvc, inputComponentList, "1.10.0", internal.ServiceManagerOverride{})

	// when
	builder, found := factory.ForPlan(azurePlanID)

	// then
	require.True(t, found)

	// when
	input, err := builder.
		SetProvisioningParameters(internal.ProvisioningParametersDTO{
			Name: "azure-cluster",
		}).
		SetERSContext(internal.ERSContext{
			ServiceManager: smOverrides,
		}).
		SetProvisioningConfig(ProvisioningConfig{
			AzureSecretName: "azure-secret",
		}).
		Build()

	// then
	require.NoError(t, err)
	assert.Equal(t, "azure", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "azure-secret", input.ClusterConfig.GardenerConfig.TargetSecret)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)

	assertServiceManagerOverrides(t, input.KymaConfig.Components, smOverrides)
}

func assertServiceManagerOverrides(t *testing.T, components internal.ComponentConfigurationInputList, overrides internal.ServiceManagerEntryDTO) {
	smComponent, found := find(components, serviceManagerComponentName)
	require.True(t, found)
	assert.Equal(t, overrides.URL, smComponent.Configuration[0].Value)
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
		{Name: serviceManagerComponentName, Namespace: "kyma-system"},
	}
}
