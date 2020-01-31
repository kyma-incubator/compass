package model

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/aws"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/azure"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model/infrastructure/gcp"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	infrastructureConfigKind = "InfrastructureConfig"
	controlPlaneConfigKind   = "ControlPlaneConfig"

	gcpAPIVersion   = "gcp.provider.extensions.gardener.cloud/v1alpha1"
	azureAPIVersion = "azure.provider.extensions.gardener.cloud/v1alpha1"
	awsAPIVersion   = "aws.provider.extensions.gardener.cloud/v1alpha1"
)

func NewGCPInfrastructure(workerCIDR string) *gcp.InfrastructureConfig {
	return &gcp.InfrastructureConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       infrastructureConfigKind,
			APIVersion: gcpAPIVersion,
		},
		Networks: gcp.NetworkConfig{
			Worker: workerCIDR,
			Workers: util.StringPtr(workerCIDR),
		},
	}
}

func NewGCPControlPlane(zone string) *gcp.ControlPlaneConfig {
	return &gcp.ControlPlaneConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       controlPlaneConfigKind,
			APIVersion: gcpAPIVersion,
		},
		Zone: zone,
	}
}

func NewAzureInfrastructure(workerCIDR string, azConfig AzureGardenerConfig) *azure.InfrastructureConfig {
	return &azure.InfrastructureConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       infrastructureConfigKind,
			APIVersion: azureAPIVersion,
		},
		Networks: azure.NetworkConfig{
			Workers: workerCIDR,
			VNet: azure.VNet{
				CIDR: &azConfig.input.VnetCidr,
			},
		},
	}
}

func NewAzureControlPlane() *azure.ControlPlaneConfig {
	return &azure.ControlPlaneConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       controlPlaneConfigKind,
			APIVersion: azureAPIVersion,
		},
	}
}

func NewAWSInfrastructure(workerCIDR string, awsConfig AWSGardenerConfig) *aws.InfrastructureConfig {
	return &aws.InfrastructureConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       infrastructureConfigKind,
			APIVersion: awsAPIVersion,
		},
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

func NewAWSControlPlane() *aws.ControlPlaneConfig {
	return &aws.ControlPlaneConfig{
		TypeMeta: v1.TypeMeta{
			Kind:       controlPlaneConfigKind,
			APIVersion: awsAPIVersion,
		},
	}
}
