package input

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	cloudProvider "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provider"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/vburenin/nsync"
)

type (
	OptionalComponentService interface {
		ExecuteDisablers(components internal.ComponentConfigurationInputList, names ...string) (internal.ComponentConfigurationInputList, error)
		ComputeComponentsToDisable(optComponentsToKeep []string) []string
	}

	HyperscalerInputProvider interface {
		Defaults() *gqlschema.ClusterConfigInput
		ApplyParameters(input *gqlschema.ClusterConfigInput, params internal.ProvisioningParametersDTO)
	}

	CreatorForPlan interface {
		IsPlanSupport(planID string) bool
		ForPlan(planID string) (internal.ProvisionInputCreator, bool)
	}
)

type InputBuilderFactory struct {
	kymaVersion        string
	config             Config
	optComponentsSvc   OptionalComponentService
	fullComponentsList internal.ComponentConfigurationInputList
}

func NewInputBuilderFactory(optComponentsSvc OptionalComponentService, fullComponentsList []v1alpha1.KymaComponent, config Config, kymaVersion string) CreatorForPlan {
	return &InputBuilderFactory{
		config:             config,
		kymaVersion:        kymaVersion,
		optComponentsSvc:   optComponentsSvc,
		fullComponentsList: mapToGQLComponentConfigurationInput(fullComponentsList),
	}
}

func (f *InputBuilderFactory) IsPlanSupport(planID string) bool {
	switch planID {
	case broker.GcpPlanID, broker.AzurePlanID:
		return true
	default:
		return false
	}
}

func (f *InputBuilderFactory) ForPlan(planID string) (internal.ProvisionInputCreator, bool) {
	if !f.IsPlanSupport(planID) {
		return nil, false
	}

	var provider HyperscalerInputProvider
	switch planID {
	case broker.GcpPlanID:
		provider = &cloudProvider.GcpInput{}
	case broker.AzurePlanID:
		provider = &cloudProvider.AzureInput{}
	// insert cases for other providers like AWS or GCP
	default:
		return nil, false
	}

	initInput := f.initInput(provider)

	return &RuntimeInput{
		input:                     initInput,
		overrides:                 make(map[string][]*gqlschema.ConfigEntryInput, 0),
		globalOverrides:           make([]*gqlschema.ConfigEntryInput, 0),
		mutex:                     nsync.NewNamedMutex(),
		hyperscalerInputProvider:  provider,
		optionalComponentsService: f.optComponentsSvc,
	}, true
}

func (f *InputBuilderFactory) initInput(provider HyperscalerInputProvider) gqlschema.ProvisionRuntimeInput {
	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  &gqlschema.RuntimeInput{},
		ClusterConfig: provider.Defaults(),
		KymaConfig: &gqlschema.KymaConfigInput{
			Version:    f.kymaVersion,
			Components: f.fullComponentsList.DeepCopy(),
		},
	}
}

func mapToGQLComponentConfigurationInput(kymaComponents []v1alpha1.KymaComponent) internal.ComponentConfigurationInputList {
	var input internal.ComponentConfigurationInputList
	for _, component := range kymaComponents {
		input = append(input, &gqlschema.ComponentConfigurationInput{
			Component: component.Name,
			Namespace: component.Namespace,
			// TODO: Add the source here
		})
	}
	return input
}
