package broker

import "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

type azureInputProvider struct {

}

var _ inputProvider = &azureInputProvider{}

func (p *azureInputProvider) Defaults() *gqlschema.ClusterConfigInput {
	return &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "1.15.4",
			DiskType:          "Standard_LRS",
			VolumeSizeGb:      50,
			NodeCount:         3,
			MachineType:       "Standard_D2_v3",
			Region:            "westeurope",
			Provider:          "azure",
			WorkerCidr:        "10.250.0.0/19",
			AutoScalerMin:     2,
			AutoScalerMax:     4,
			MaxSurge:          4,
			MaxUnavailable:    1,
			Seed:              "az-eu1",
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: "10.250.0.0/19",
				},
			},
		},
	}
}

func (p *azureInputProvider) ApplyParameters(input *gqlschema.ClusterConfigInput, params *ProvisioningParameters) {
}
