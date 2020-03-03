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
		inputComponentList  = fixKymaComponentList()
		mappedComponentList = mapToGQLComponentConfigurationInput(inputComponentList)
		toDisableComponents = []string{"kiali"}
		smOverrides         = internal.ServiceManagerEntryDTO{URL: "http://sm-pico-bello-url.com"}
		fixID               = "fix-id"
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
			Name:      "azure-cluster",
			NodeCount: ptr.Integer(4),
		}).
		SetRuntimeLabels(fixID, fixID).
		SetGardenerTargetSecretName("azure-secret").
		SetOverrides(ServiceManagerComponentName, []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: smOverrides.URL,
			},
			{
				Key:   "sm.user",
				Value: smOverrides.Credentials.BasicAuth.Username,
			},
			{
				Key:    "sm.password",
				Value:  smOverrides.Credentials.BasicAuth.Password,
				Secret: ptr.Bool(true),
			},
		}).Create()

	// then
	require.NoError(t, err)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)
	assert.Equal(t, "azure-cluster", input.RuntimeInput.Name)
	assert.Equal(t, "azure", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "azure-secret", input.ClusterConfig.GardenerConfig.TargetSecret)
	assert.Equal(t, 4, input.ClusterConfig.GardenerConfig.NodeCount)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)
	assert.Equal(t, &gqlschema.Labels{
		brokerKeyPrefix + "instance_id":   []string{fixID},
		globalKeyPrefix + "subaccount_id": []string{fixID},
	}, input.RuntimeInput.Labels)

	assertServiceManagerOverrides(t, input.KymaConfig.Components, smOverrides)
}

func assertServiceManagerOverrides(t *testing.T, components internal.ComponentConfigurationInputList, overrides internal.ServiceManagerEntryDTO) {
	smComponent, found := find(components, ServiceManagerComponentName)
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
		{Name: ServiceManagerComponentName, Namespace: "kyma-system"},
	}
}
