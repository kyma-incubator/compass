package input

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/vburenin/nsync"
)

const (
	brokerKeyPrefix = "broker_"
	globalKeyPrefix = "global_"
)

type Config struct {
	URL string
}

type RuntimeInput struct {
	input           gqlschema.ProvisionRuntimeInput
	mutex           *nsync.NamedMutex
	overrides       map[string][]*gqlschema.ConfigEntryInput
	globalOverrides []*gqlschema.ConfigEntryInput

	hyperscalerInputProvider  HyperscalerInputProvider
	optionalComponentsService OptionalComponentService
	provisioningParameters    internal.ProvisioningParametersDTO
}

func (r *RuntimeInput) SetProvisioningParameters(params internal.ProvisioningParametersDTO) internal.ProvisionInputCreator {
	r.provisioningParameters = params
	return r
}

// AppendOverrides sets the overrides for the given component and discard the previous ones.
//
// Deprecated: use AppendOverrides
func (r *RuntimeInput) SetOverrides(component string, overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	// currently same as in AppendOverrides function, as we working on the same underlying object.
	r.mutex.Lock("AppendOverrides")
	defer r.mutex.Unlock("AppendOverrides")

	r.overrides[component] = overrides
	return r
}

// AppendOverrides appends overrides for the given components, the existing overrides are preserved.
func (r *RuntimeInput) AppendOverrides(component string, overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	r.mutex.Lock("AppendOverrides")
	defer r.mutex.Unlock("AppendOverrides")

	r.overrides[component] = append(r.overrides[component], overrides...)
	return r
}

// AppendAppendGlobalOverrides appends overrides, the existing overrides are preserved.
func (r *RuntimeInput) AppendGlobalOverrides(overrides []*gqlschema.ConfigEntryInput) internal.ProvisionInputCreator {
	r.mutex.Lock("AppendGlobalOverrides")
	defer r.mutex.Unlock("AppendGlobalOverrides")

	r.globalOverrides = append(r.globalOverrides, overrides...)
	return r
}

func (r *RuntimeInput) SetRuntimeLabels(instanceID, subAccountID string) internal.ProvisionInputCreator {
	r.mutex.Lock("SetRuntimeLabels")
	defer r.mutex.Unlock("SetRuntimeLabels")

	r.input.RuntimeInput.Labels = &gqlschema.Labels{
		brokerKeyPrefix + "instance_id":   []string{instanceID},
		globalKeyPrefix + "subaccount_id": []string{subAccountID},
	}

	return r
}

func (r *RuntimeInput) Create() (gqlschema.ProvisionRuntimeInput, error) {
	for _, step := range []struct {
		name    string
		execute func() error
	}{
		{
			name:    "applying provisioning parameters customization",
			execute: r.applyProvisioningParameters,
		},
		{
			name:    "disabling optional components that were not selected",
			execute: r.disableNotSelectedComponents,
		},
		{
			name:    "applying components overrides",
			execute: r.applyOverrides,
		},
		{
			name:    "applying global overrides",
			execute: r.applyGlobalOverrides,
		},
	} {
		if err := step.execute(); err != nil {
			return gqlschema.ProvisionRuntimeInput{}, errors.Wrapf(err, "while %s", step.name)
		}
	}

	return r.input, nil
}

func (r *RuntimeInput) applyProvisioningParameters() error {
	updateString(&r.input.RuntimeInput.Name, &r.provisioningParameters.Name)

	updateInt(&r.input.ClusterConfig.GardenerConfig.MaxUnavailable, r.provisioningParameters.MaxUnavailable)
	updateInt(&r.input.ClusterConfig.GardenerConfig.MaxSurge, r.provisioningParameters.MaxSurge)
	updateInt(&r.input.ClusterConfig.GardenerConfig.AutoScalerMin, r.provisioningParameters.AutoScalerMin)
	updateInt(&r.input.ClusterConfig.GardenerConfig.AutoScalerMax, r.provisioningParameters.AutoScalerMax)
	updateInt(&r.input.ClusterConfig.GardenerConfig.VolumeSizeGb, r.provisioningParameters.VolumeSizeGb)
	updateString(&r.input.ClusterConfig.GardenerConfig.Region, r.provisioningParameters.Region)
	updateString(&r.input.ClusterConfig.GardenerConfig.MachineType, r.provisioningParameters.MachineType)
	updateString(&r.input.ClusterConfig.GardenerConfig.TargetSecret, r.provisioningParameters.TargetSecret)

	r.hyperscalerInputProvider.ApplyParameters(r.input.ClusterConfig, r.provisioningParameters)

	return nil
}

func (r *RuntimeInput) disableNotSelectedComponents() error {
	toDisable := r.optionalComponentsService.ComputeComponentsToDisable(r.provisioningParameters.OptionalComponentsToInstall)

	filterOut, err := r.optionalComponentsService.ExecuteDisablers(r.input.KymaConfig.Components, toDisable...)
	if err != nil {
		return errors.Wrapf(err, "while disabling components %v", toDisable)
	}

	r.input.KymaConfig.Components = filterOut

	return nil
}

func (r *RuntimeInput) applyOverrides() error {
	for i := range r.input.KymaConfig.Components {
		if entry, found := r.overrides[r.input.KymaConfig.Components[i].Component]; found {
			r.input.KymaConfig.Components[i].Configuration = append(r.input.KymaConfig.Components[i].Configuration, entry...)
		}
	}

	return nil
}

func (r *RuntimeInput) applyGlobalOverrides() error {
	r.input.KymaConfig.Configuration = r.globalOverrides
	return nil
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
