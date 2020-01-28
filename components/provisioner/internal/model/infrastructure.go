package model

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/aws"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/azure"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/gcp"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
)

func NewGCPInfrastructure(workerCIDR string) *gcp.InfrastructureConfig {
	return &gcp.InfrastructureConfig{
		Networks: gcp.NetworkConfig{
			Worker: workerCIDR,
		},
	}
}

func NewAzureInfrastructure(workerCIDR string, azConfig AzureGardenerConfig) *azure.InfrastructureConfig {
	return &azure.InfrastructureConfig{
		Networks: azure.NetworkConfig{
			Workers: workerCIDR,
			VNet: azure.VNet{
				CIDR: &azConfig.input.VnetCidr,
			},
		},
	}
}

func NewAWSInfrastructure(workerCIDR string, awsConfig AWSGardenerConfig) *aws.InfrastructureConfig {
	return &aws.InfrastructureConfig{
		Networks: aws.Networks{
			Zones: []aws.Zone{
				{
					Name:     awsConfig.input.Zone,
					Internal: awsConfig.input.InternalCidr,
					Public:   awsConfig.input.PublicCidr,
					Workers:  workerCIDR,
				},
			},
			VPC: aws.VPC{
				CIDR: util.StringPtr(awsConfig.input.VpcCidr),
			},
		},
	}
}
