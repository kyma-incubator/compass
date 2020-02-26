package broker

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	cloudProvider "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provider"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
)

const (
	brokerKeyPrefix             = "broker_"
	globalKeyPrefix             = "global_"
	serviceManagerComponentName = "service-manager-proxy"
)

//go:generate mockery -name=OptionalComponentService -output=automock -outpkg=automock -case=underscore
//go:generate mockery -name=InputBuilderForPlan -output=automock -outpkg=automock -case=underscore

// Defines dependency
type (
	OptionalComponentService interface {
		ExecuteDisablers(components internal.ComponentConfigurationInputList, names ...string) (internal.ComponentConfigurationInputList, error)
		ComputeComponentsToDisable(optComponentsToKeep []string) []string
	}

	HyperscalerInputProvider interface {
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
		SetInstanceID(instanceID string) ConcreteInputBuilder
		Build() (gqlschema.ProvisionRuntimeInput, error)
	}

	InputBuilderForPlan interface {
		ForPlan(planID string) (ConcreteInputBuilder, bool)
	}
)

type InputBuilderFactory struct {
	kymaVersion                string
	optComponentsSvc           OptionalComponentService
	serviceManager             internal.ServiceManagerOverride
	hyperscalerAccountProvider hyperscaler.AccountProvider
	fullComponentsList         internal.ComponentConfigurationInputList
	directorURL                string
}

func NewInputBuilderFactory(optComponentsSvc OptionalComponentService, fullComponentsList []v1alpha1.KymaComponent, kymaVersion string, smOverride internal.ServiceManagerOverride, directorURL string, hyperscalerAccountProvider hyperscaler.AccountProvider) InputBuilderForPlan {
	return &InputBuilderFactory{
		kymaVersion:                kymaVersion,
		serviceManager:             smOverride,
		optComponentsSvc:           optComponentsSvc,
		hyperscalerAccountProvider: hyperscalerAccountProvider,
		fullComponentsList:         mapToGQLComponentConfigurationInput(fullComponentsList),
		directorURL:                directorURL,
	}
}

func (f *InputBuilderFactory) ForPlan(planID string) (ConcreteInputBuilder, bool) {
	var provider HyperscalerInputProvider
	switch planID {
	case gcpPlanID:
		provider = &cloudProvider.GcpInput{}
	case azurePlanID:
		provider = &cloudProvider.AzureInput{}
	// insert cases for other providers like AWS or GCP
	default:
		return nil, false
	}

	return &InputBuilder{
		planID:                     planID,
		kymaVersion:                f.kymaVersion,
		serviceManager:             f.serviceManager,
		hyperscalerInputProvider:   provider,
		hyperscalerAccountProvider: f.hyperscalerAccountProvider,
		optionalComponentsService:  f.optComponentsSvc,
		fullRuntimeComponentList:   f.fullComponentsList,
		directorURL:                f.directorURL,
	}, true
}

type InputBuilder struct {
	planID                     string
	instanceID                 string
	kymaVersion                string
	serviceManager             internal.ServiceManagerOverride
	hyperscalerInputProvider   HyperscalerInputProvider
	hyperscalerAccountProvider hyperscaler.AccountProvider
	optionalComponentsService  OptionalComponentService
	fullRuntimeComponentList   internal.ComponentConfigurationInputList
	directorURL                string

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

func (b *InputBuilder) SetInstanceID(instanceID string) ConcreteInputBuilder {
	b.instanceID = instanceID
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

func applySingleOverride(in *gqlschema.ProvisionRuntimeInput, componentName, key, value string) {
	override := gqlschema.ConfigEntryInput{
		Key:   key,
		Value: value,
	}

	for i := range in.KymaConfig.Components {
		if in.KymaConfig.Components[i].Component == componentName {
			in.KymaConfig.Components[i].Configuration = append(in.KymaConfig.Components[i].Configuration, &override)
		}
	}
}

func (b *InputBuilder) applyManagementPlaneOverrides(in *gqlschema.ProvisionRuntimeInput) error {

	// core override
	applySingleOverride(in, "core", "console.managementPlane.url", b.directorURL)

	// compass runtime agent override
	applySingleOverride(in, "compass-runtime-agent", "managementPlane.url", b.directorURL)

	return nil
}

// applyTemporaryCustomization applies some additional information. This is only a temporary solution for MVP scenario.
// Will be removed and refactored in near future.
func (b *InputBuilder) applyTemporaryCustomization(in *gqlschema.ProvisionRuntimeInput) error {
	// old implementation
	//switch b.planID {
	//case azurePlanID:
	//	in.ClusterConfig.GardenerConfig.TargetSecret = b.provisioningConfig.AzureSecretName
	//case gcpPlanID:
	//	in.ClusterConfig.GardenerConfig.TargetSecret = b.provisioningConfig.GCPSecretName
	//default:
	//	return fmt.Errorf("unknown Plan ID %s", b.planID)
	//}

	if in.ClusterConfig.GardenerConfig == nil {
		return fmt.Errorf("gardener config for cluster is empty (nil)")
	}

	targetSecretName, err := b.hyperscalerAccountProvider.GardenerSecretName(in.ClusterConfig.GardenerConfig, b.planID)

	if err != nil {
		return err
	}

	in.ClusterConfig.GardenerConfig.TargetSecret = targetSecretName

	return nil
}

func (b *InputBuilder) applyRuntimeLabels(in *gqlschema.ProvisionRuntimeInput) error {
	in.RuntimeInput.Labels = &gqlschema.Labels{
		brokerKeyPrefix + "instance_id":   []string{b.instanceID},
		globalKeyPrefix + "subaccount_id": []string{b.ersCtx.SubAccountID},
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
			name:    "applying management plane overrides",
			execute: b.applyManagementPlaneOverrides,
		},
		{
			name:    "applying temporary customization",
			execute: b.applyTemporaryCustomization,
		},
		{
			name:    "applying labels to runtime",
			execute: b.applyRuntimeLabels,
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
