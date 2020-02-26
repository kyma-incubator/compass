package broker

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker/automock"

	hyperscalerMocks "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
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
		fixID               = "fix-id"
	)

	optComponentsSvc := &automock.OptionalComponentService{}
	defer optComponentsSvc.AssertExpectations(t)
	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return(toDisableComponents)
	optComponentsSvc.On("ExecuteDisablers", mappedComponentList, toDisableComponents[0]).Return(mappedComponentList, nil)

	accountProviderMock := &hyperscalerMocks.AccountProvider{}

	accountProviderMock.On("GardenerSecretName", mock.MatchedBy(getGardenerRuntimeInputMatcherForAzure()), azurePlanID).Return("gardener-secret-azurePlanID", nil)

	factory := NewInputBuilderFactory(optComponentsSvc, inputComponentList, "1.10.0", internal.ServiceManagerOverride{}, "https://compass-gateway-auth-oauth.kyma.local/director/graphql", accountProviderMock)

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
			SubAccountID:   fixID,
		}).
		//SetProvisioningConfig(ProvisioningConfig{
		//	AzureSecretName: "azure-secret",
		//}).
		SetInstanceID(fixID).
		Build()

	// then
	require.NoError(t, err)
	assert.EqualValues(t, mappedComponentList, input.KymaConfig.Components)
	assert.Equal(t, "azure-cluster", input.RuntimeInput.Name)
	assert.Equal(t, "azure", input.ClusterConfig.GardenerConfig.Provider)
	assert.Equal(t, "gardener-secret-azurePlanID", input.ClusterConfig.GardenerConfig.TargetSecret)
	assert.Equal(t, &gqlschema.Labels{
		brokerKeyPrefix + "instance_id":   []string{fixID},
		globalKeyPrefix + "subaccount_id": []string{fixID},
	}, input.RuntimeInput.Labels)

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

func getGardenerRuntimeInputMatcherForAzure() func(*gqlschema.GardenerConfigInput) bool {
	return func(input *gqlschema.GardenerConfigInput) bool {
		return input.ProviderSpecificConfig.AzureConfig != nil
	}
}

func getGardenerRuntimeInputMatcherForGCP() func(*gqlschema.GardenerConfigInput) bool {
	return func(input *gqlschema.GardenerConfigInput) bool {
		return input.ProviderSpecificConfig.GcpConfig != nil
	}
}
