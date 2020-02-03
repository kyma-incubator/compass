package broker

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
)

const serviceManagerComponentName = "service-manager-proxy"

//go:generate mockery -name=OptionalComponentService -output=automock -outpkg=automock -case=underscore

// Defines dependency
type (
	OptionalComponentService interface {
		ExecuteDisablers(components internal.ComponentConfigurationInputList, names ...string) (internal.ComponentConfigurationInputList, error)
		ComputeComponentsToDisable(optComponentsToKeep []string) []string
	}

	hyperscalerInputProvider interface {
		Defaults() *gqlschema.ClusterConfigInput
		ApplyParameters(input *gqlschema.ClusterConfigInput, params internal.ProvisioningParametersDTO)
	}
)

// Defines API
type (
	ConcreteInputBuilder interface {
		SetProvisioningParameters(params internal.ProvisioningParametersDTO) ConcreteInputBuilder
		SetERSContext(ersCtx internal.ERSContext) ConcreteInputBuilder
		SetProvisioningConfig(brokerConfig ProvisioningConfig) ConcreteInputBuilder
		Build() (gqlschema.ProvisionRuntimeInput, error)
	}

	InputBuilderForPlan interface {
		ForPlan(planID string) (ConcreteInputBuilder, bool)
	}
)

type InputBuilderFactory struct {
	kymaVersion        string
	optComponentsSvc   OptionalComponentService
	serviceManager     internal.ServiceManagerOverride
	fullComponentsList internal.ComponentConfigurationInputList
}

func NewInputBuilderFactory(optComponentsSvc OptionalComponentService, fullComponentsList []v1alpha1.KymaComponent, kymaVersion string, smOverride internal.ServiceManagerOverride) InputBuilderForPlan {
	return &InputBuilderFactory{
		kymaVersion:        kymaVersion,
		serviceManager:     smOverride,
		optComponentsSvc:   optComponentsSvc,
		fullComponentsList: mapToGQLComponentConfigurationInput(fullComponentsList),
	}
}

func (f *InputBuilderFactory) ForPlan(planID string) (ConcreteInputBuilder, bool) {
	var provider hyperscalerInputProvider
	switch planID {
	case gcpPlanID:
		provider = &gcpInputProvider{}
	case azurePlanID:
		provider = &azureInputProvider{}
	// insert cases for other providers like AWS or GCP
	default:
		return nil, false
	}

	return &InputBuilder{
		planID:                    planID,
		kymaVersion:               f.kymaVersion,
		serviceManager:            f.serviceManager,
		hyperscalerInputProvider:  provider,
		optionalComponentsService: f.optComponentsSvc,
		fullRuntimeComponentList:  f.fullComponentsList,
	}, true
}

type InputBuilder struct {
	planID                    string
	kymaVersion               string
	serviceManager            internal.ServiceManagerOverride
	hyperscalerInputProvider  hyperscalerInputProvider
	optionalComponentsService OptionalComponentService
	fullRuntimeComponentList  internal.ComponentConfigurationInputList

	ersCtx                 internal.ERSContext
	provisioningConfig     ProvisioningConfig
	provisioningParameters internal.ProvisioningParametersDTO
}

func (b *InputBuilder) SetProvisioningParameters(params internal.ProvisioningParametersDTO) ConcreteInputBuilder {
	b.provisioningParameters = params
	return b
}

func (b *InputBuilder) SetERSContext(ersCtx internal.ERSContext) ConcreteInputBuilder {
	b.ersCtx = ersCtx
	return b
}

func (b *InputBuilder) SetProvisioningConfig(brokerConfig ProvisioningConfig) ConcreteInputBuilder {
	b.provisioningConfig = brokerConfig
	return b
}

func (b *InputBuilder) applyProvisioningParameters(in *gqlschema.ProvisionRuntimeInput) error {
	updateString(&in.RuntimeInput.Name, &b.provisioningParameters.Name)

	updateInt(&in.ClusterConfig.GardenerConfig.NodeCount, b.provisioningParameters.NodeCount)
	updateInt(&in.ClusterConfig.GardenerConfig.MaxUnavailable, b.provisioningParameters.MaxUnavailable)
	updateInt(&in.ClusterConfig.GardenerConfig.MaxSurge, b.provisioningParameters.MaxSurge)
	updateInt(&in.ClusterConfig.GardenerConfig.AutoScalerMin, b.provisioningParameters.AutoScalerMin)
	updateInt(&in.ClusterConfig.GardenerConfig.AutoScalerMax, b.provisioningParameters.AutoScalerMax)
	updateInt(&in.ClusterConfig.GardenerConfig.VolumeSizeGb, b.provisioningParameters.VolumeSizeGb)
	updateString(&in.ClusterConfig.GardenerConfig.Region, b.provisioningParameters.Region)
	updateString(&in.ClusterConfig.GardenerConfig.MachineType, b.provisioningParameters.MachineType)

	b.hyperscalerInputProvider.ApplyParameters(in.ClusterConfig, b.provisioningParameters)

	return nil
}

func (b *InputBuilder) disableNotSelectedComponents(in *gqlschema.ProvisionRuntimeInput) error {
	toDisable := b.optionalComponentsService.ComputeComponentsToDisable(b.provisioningParameters.OptionalComponentsToInstall)

	filterOut, err := b.optionalComponentsService.ExecuteDisablers(in.KymaConfig.Components, toDisable...)
	if err != nil {
		return errors.Wrapf(err, "while disabling components %v", toDisable)
	}

	in.KymaConfig.Components = filterOut

	return nil
}

func (b *InputBuilder) applyServiceManagerOverrides(in *gqlschema.ProvisionRuntimeInput) error {
	var smOverrides []*gqlschema.ConfigEntryInput
	if b.serviceManager.CredentialsOverride {
		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: b.serviceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: b.serviceManager.Username,
			},
			{
				Key:    "sm.password",
				Value:  b.serviceManager.Password,
				Secret: ptr.Bool(true),
			},
		}
	} else {
		smOverrides = []*gqlschema.ConfigEntryInput{
			{
				Key:   "config.sm.url",
				Value: b.ersCtx.ServiceManager.URL,
			},
			{
				Key:   "sm.user",
				Value: b.ersCtx.ServiceManager.Credentials.BasicAuth.Username,
			},
			{
				Key:    "sm.password",
				Value:  b.ersCtx.ServiceManager.Credentials.BasicAuth.Password,
				Secret: ptr.Bool(true),
			},
		}
	}

	for i := range in.KymaConfig.Components {
		if in.KymaConfig.Components[i].Component == serviceManagerComponentName {
			in.KymaConfig.Components[i].Configuration = append(in.KymaConfig.Components[i].Configuration, smOverrides...)
		}
	}

	return nil
}

// applyTemporaryCustomization applies some additional information. This is only a temporary solution for MVP scenario.
// Will be removed and refactored in near future.
func (b *InputBuilder) applyTemporaryCustomization(in *gqlschema.ProvisionRuntimeInput) error {
	switch b.planID {
	case azurePlanID:
		in.ClusterConfig.GardenerConfig.TargetSecret = b.provisioningConfig.AzureSecretName
	case gcpPlanID:
		in.ClusterConfig.GardenerConfig.TargetSecret = b.provisioningConfig.GCPSecretName
	default:
		return fmt.Errorf("unknown Plan ID %s", b.planID)
	}

	return nil
}

func (b *InputBuilder) initInput() gqlschema.ProvisionRuntimeInput {
	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  &gqlschema.RuntimeInput{},
		ClusterConfig: b.hyperscalerInputProvider.Defaults(),
		KymaConfig: &gqlschema.KymaConfigInput{
			Version:    b.kymaVersion,
			Components: b.fullRuntimeComponentList.DeepCopy(),
		},
	}

}

func (b *InputBuilder) Build() (gqlschema.ProvisionRuntimeInput, error) {
	input := b.initInput()

	for _, step := range []struct {
		name    string
		execute func(in *gqlschema.ProvisionRuntimeInput) error
	}{
		{
			name:    "applying provisioning parameters customization",
			execute: b.applyProvisioningParameters,
		},
		{
			name:    "disabling optional components that were not selected",
			execute: b.disableNotSelectedComponents,
		},
		{
			name:    "applying service manager overrides",
			execute: b.applyServiceManagerOverrides,
		},
		{
			name:    "applying temporary customization",
			execute: b.applyTemporaryCustomization,
		},
	} {
		if err := step.execute(&input); err != nil {
			return gqlschema.ProvisionRuntimeInput{}, errors.Wrapf(err, "while %s", step.name)
		}
	}

	return input, nil
}

func updateString(toUpdate *string, value *string) {
	if value != nil {
		*toUpdate = *value
	}
}

func updateInt(toUpdate *int, value *int) {
	if value != nil {
		*toUpdate = *value
	}
}

func mapToGQLComponentConfigurationInput(kymaComponents []v1alpha1.KymaComponent) internal.ComponentConfigurationInputList {
	var input internal.ComponentConfigurationInputList
	for _, component := range kymaComponents {
		input = append(input, &gqlschema.ComponentConfigurationInput{
			Component: component.Name,
			Namespace: component.Namespace,
		})
	}
	return input
}
