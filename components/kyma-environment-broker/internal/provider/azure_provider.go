package provider

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type AzureInput struct{}

func (p *AzureInput) Defaults() *gqlschema.ClusterConfigInput {
	return &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "1.15.11",
			DiskType:          "Standard_LRS",
			VolumeSizeGb:      50,
			MachineType:       "Standard_D8_v3",
			Region:            "westeurope",
			Provider:          "azure",
			WorkerCidr:        "10.250.0.0/19",
			AutoScalerMin:     3,
			AutoScalerMax:     10,
			MaxSurge:          4,
			MaxUnavailable:    1,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: "10.250.0.0/19",
				},
			},
		},
	}
}

func (p *AzureInput) ApplyParameters(input *gqlschema.ClusterConfigInput, params internal.ProvisioningParametersDTO) {
}
