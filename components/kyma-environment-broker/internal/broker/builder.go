package broker

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

func NewInputBuilderForPlan(planID string) (*provisioningParamsBuilder, bool) {
	var builder *provisioningParamsBuilder
	switch planID {
	case azurePlanID:
		builder = newProvisioningParamsBuilder(&azureInputProvider{})
	// insert cases for other providers like AWS or GCP
	default:
		return nil, false
	}
	return builder, true
}

type inputProvider interface {
	Defaults() *gqlschema.ClusterConfigInput
	ApplyParameters(input *gqlschema.ClusterConfigInput, params *provisioningParameters)
}

type provisioningParamsBuilder struct {
	provider inputProvider
	input    *gqlschema.ProvisionRuntimeInput
}

func newProvisioningParamsBuilder(ip inputProvider) *provisioningParamsBuilder {
	builder := &provisioningParamsBuilder{
		input: &gqlschema.ProvisionRuntimeInput{
			ClusterConfig: ip.Defaults(),
			KymaConfig:    &gqlschema.KymaConfigInput{Version: "1.6", Modules: gqlschema.AllKymaModule},
		},
		provider: ip,
	}
	return builder
}

type provisioningParameters struct {
	Name           string  `json:"name"`
	NodeCount      *int    `json:"nodeCount"`
	VolumeSizeGb   *int    `json:"volumeSizeGb"`
	MachineType    *string `json:"machineType"`
	Region         *string `json:"region"`
	Zone           *string `json:"zone"`
	AutoScalerMin  *int    `json:"autoScalerMin"`
	AutoScalerMax  *int    `json:"autoScalerMax"`
	MaxSurge       *int    `json:"maxSurge"`
	MaxUnavailable *int    `json:"maxUnavailable"`
}

func (b *provisioningParameters) ToModel() *internal.ProvisioningParameters {
	return &internal.ProvisioningParameters{
		Name:           b.Name,
		NodeCount:      b.NodeCount,
		VolumeSizeGb:   b.VolumeSizeGb,
		MachineType:    b.MachineType,
		Region:         b.Region,
		Zone:           b.Zone,
		AutoScalerMin:  b.AutoScalerMin,
		AutoScalerMax:  b.AutoScalerMax,
		MaxSurge:       b.MaxSurge,
		MaxUnavailable: b.MaxUnavailable,
	}
}

func (b *provisioningParamsBuilder) ApplyParameters(params *provisioningParameters) {
	b.input.ClusterConfig.GardenerConfig.Name = params.Name
	updateInt(&b.input.ClusterConfig.GardenerConfig.NodeCount, params.NodeCount)
	updateInt(&b.input.ClusterConfig.GardenerConfig.MaxUnavailable, params.MaxUnavailable)
	updateInt(&b.input.ClusterConfig.GardenerConfig.MaxSurge, params.MaxSurge)
	updateInt(&b.input.ClusterConfig.GardenerConfig.AutoScalerMin, params.AutoScalerMin)
	updateInt(&b.input.ClusterConfig.GardenerConfig.AutoScalerMax, params.AutoScalerMax)
	updateString(&b.input.ClusterConfig.GardenerConfig.Region, params.Region)
	updateString(&b.input.ClusterConfig.GardenerConfig.MachineType, params.MachineType)
	updateInt(&b.input.ClusterConfig.GardenerConfig.VolumeSizeGb, params.VolumeSizeGb)

	b.provider.ApplyParameters(b.input.ClusterConfig, params)
}

func (b *provisioningParamsBuilder) ClusterConfigInput() *gqlschema.ProvisionRuntimeInput {
	return b.input
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
