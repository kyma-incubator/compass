package broker

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type awsInputProvider struct {
}

var _ inputProvider = &awsInputProvider{}

func (p *awsInputProvider) Defaults() *gqlschema.ClusterConfigInput {
	return &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "1.15.4",
			DiskType:          "gp2",
			VolumeSizeGb:      50,
			NodeCount:         3,
			MachineType:       "m4.2xlarge",
			Region:            "eu-west-1",
			Provider:          "aws",
			Seed:              "aws-eu1",
			WorkerCidr:        "10.250.0.0/19",
			AutoScalerMin:     2,
			AutoScalerMax:     4,
			MaxSurge:          4,
			MaxUnavailable:    1,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AwsConfig: &gqlschema.AWSProviderConfigInput{
					Zone:         "eu-west-1b",
					InternalCidr: "10.250.112.0/22",
					PublicCidr:   "10.250.96.0/22",
					VpcCidr:      "10.250.0.0/16",
				},
			},
		},
	}
}

func (p *awsInputProvider) ApplyParameters(input *gqlschema.ClusterConfigInput, params *internal.ProvisioningParametersDTO) {
	updateString(&input.GardenerConfig.ProviderSpecificConfig.AwsConfig.Zone, params.Zone)
}
