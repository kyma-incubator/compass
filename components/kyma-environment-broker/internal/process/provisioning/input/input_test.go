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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInputBuilderFactoryOverrides(t *testing.T) {
	t.Run("should append overrides for the same components multiple times", func(t *testing.T) {
		// given
		var (
			dummyOptComponentsSvc = dummyOptionalComponentServiceMock(fixKymaComponentList())

			overridesA1 = []*gqlschema.ConfigEntryInput{
				{Key: "key-1", Value: "pico"},
				{Key: "key-2", Value: "bello"},
			}
			overridesA2 = []*gqlschema.ConfigEntryInput{
				{Key: "key-3", Value: "hakuna"},
				{Key: "key-4", Value: "matata", Secret: ptr.Bool(true)},
			}
		)
		componentsProvider := &automock.ComponentListProvider{}
		componentsProvider.On("AllComponents", mock.AnythingOfType("string")).Return(fixKymaComponentList(), nil)

		builder, err := NewInputBuilderFactory(dummyOptComponentsSvc, componentsProvider, Config{}, "not-important")
		assert.NoError(t, err)
		creator, found := builder.ForPlan(broker.AzurePlanID)
		require.True(t, found)

		// when
		creator.
			AppendOverrides("keb", overridesA1).
			AppendOverrides("keb", overridesA2)

		// then
		out, err := creator.Create()
		require.NoError(t, err)

		overriddenComponent, found := find(out.KymaConfig.Components, "keb")
		require.True(t, found)

		assertContainsAllOverrides(t, overriddenComponent.Configuration, overridesA1, overridesA1)
	})

	t.Run("should append global overrides", func(t *testing.T) {
		// given
		var (
			optComponentsSvc = dummyOptionalComponentServiceMock(fixKymaComponentList())

			overridesA1 = []*gqlschema.ConfigEntryInput{
				{Key: "key-1", Value: "pico"},
				{Key: "key-2", Value: "bello"},
			}
			overridesA2 = []*gqlschema.ConfigEntryInput{
				{Key: "key-3", Value: "hakuna"},
				{Key: "key-4", Value: "matata", Secret: ptr.Bool(true)},
			}
		)
		componentsProvider := &automock.ComponentListProvider{}
		componentsProvider.On("AllComponents", mock.AnythingOfType("string")).Return(fixKymaComponentList(), nil)

		builder, err := NewInputBuilderFactory(optComponentsSvc, componentsProvider, Config{}, "not-important")
		assert.NoError(t, err)
		creator, found := builder.ForPlan(broker.AzurePlanID)
		require.True(t, found)

		// when
		creator.
			AppendGlobalOverrides(overridesA1).
			AppendGlobalOverrides(overridesA2)

		// then
		out, err := creator.Create()
		require.NoError(t, err)

		assertContainsAllOverrides(t, out.KymaConfig.Configuration, overridesA1, overridesA1)
	})
}

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
	componentsProvider := &automock.ComponentListProvider{}
	componentsProvider.On("AllComponents", mock.AnythingOfType("string")).Return(inputComponentList, nil)
	defer componentsProvider.AssertExpectations(t)

	factory, err := NewInputBuilderFactory(optComponentsSvc, componentsProvider, config, "1.10.0")
	assert.NoError(t, err)

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

func dummyOptionalComponentServiceMock(inputComponentList []v1alpha1.KymaComponent) *automock.OptionalComponentService {
	mappedComponentList := mapToGQLComponentConfigurationInput(inputComponentList)

	optComponentsSvc := &automock.OptionalComponentService{}
	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return([]string{})
	optComponentsSvc.On("ExecuteDisablers", mappedComponentList).Return(mappedComponentList, nil)
	return optComponentsSvc
}

func assertContainsAllOverrides(t *testing.T, gotOverrides []*gqlschema.ConfigEntryInput, expOverrides ...[]*gqlschema.ConfigEntryInput) {
	var expected []*gqlschema.ConfigEntryInput
	for _, o := range expOverrides {
		expected = append(expected, o...)
	}

	require.Len(t, gotOverrides, len(expected))
	for _, o := range expected {
		assert.Contains(t, gotOverrides, o)
	}
}
