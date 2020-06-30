package api

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

var providerCredentials = &gqlschema.CredentialsInput{SecretName: "secret_1"}

type provisingTestCase struct {
	runtimeID    string
	description  string
	config       *gqlschema.ClusterConfigInput
	runtimeInput *gqlschema.RuntimeInput
}

type upgradeShootTestCase struct {
	runtimeID   string
	description string
	config      *gqlschema.UpgradeShootInput
}

func newTestProvisioningConfigs() []provisingTestCase {
	clusterConfigForGardenerWithGCP := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "version",
			Provider:          "GCP",
			TargetSecret:      "secret",
			Seed:              util.StringPtr("gcp-eu1"),
			Region:            "europe-west1",
			MachineType:       "n1-standard-1",
			DiskType:          "pd-ssd",
			VolumeSizeGb:      40,
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zones: []string{"fix-gcp-zone1", "fix-gcp-zone-2"},
				},
			},
		},
	}

	clusterConfigForGardenerWithAzure := func(zones []string) *gqlschema.ClusterConfigInput {
		return &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				KubernetesVersion: "version",
				Provider:          "Azure",
				TargetSecret:      "secret",
				Seed:              util.StringPtr("az-eu1"),
				Region:            "westeurope",
				MachineType:       "Standard_D8_v3",
				DiskType:          "Standard_LRS",
				VolumeSizeGb:      40,
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					AzureConfig: &gqlschema.AzureProviderConfigInput{
						VnetCidr: "cidr",
						Zones:    zones,
					},
				},
			},
		}
	}

	clusterConfigForGardenerWithAWS := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			Provider:       "AWS",
			TargetSecret:   "secret",
			Seed:           nil,
			Region:         "eu-central-1",
			MachineType:    "t3-xlarge",
			DiskType:       "gp2",
			VolumeSizeGb:   40,
			WorkerCidr:     "cidr",
			AutoScalerMin:  1,
			AutoScalerMax:  5,
			MaxSurge:       1,
			MaxUnavailable: 2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AwsConfig: &gqlschema.AWSProviderConfigInput{
					Zone:         "zone",
					InternalCidr: "cidr",
					VpcCidr:      "cidr",
					PublicCidr:   "cidr",
				},
			},
		},
	}

	zones := []string{"fix-az-zone-1", "fix-az-zone-2"}

	testConfigs := []provisingTestCase{
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bbb", config: clusterConfigForGardenerWithGCP, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with GCP configuration 1",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime1",
				Description: new(string),
			}},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb4", config: clusterConfigForGardenerWithAzure(zones), description: "Should provision and deprovision a runtime with happy flow using correct Gardener with Azure configuration when zones passed",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime2",
				Description: new(string),
			}},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb1", config: clusterConfigForGardenerWithAzure(nil), description: "Should provision and deprovision a runtime with happy flow using correct Gardener with Azure configuration when no zones passed",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime3",
				Description: new(string),
			}},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb5", config: clusterConfigForGardenerWithAWS, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with AWS configuration",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime4",
				Description: new(string),
			}},
	}
	return testConfigs
}

func newTestShootUpgradeConfigs() []upgradeShootTestCase {

	newMachineType := "new-machine"
	newDiskType := "vinyl"
	newVolumeSizeGb := 50
	newCidr := "cidr2"

	upgradeShootInputForGCP := &gqlschema.UpgradeShootInput{
		GardenerConfig: &gqlschema.GardenerUpgradeInput{
			MachineType:    &newMachineType,
			DiskType:       &newDiskType,
			VolumeSizeGb:   &newVolumeSizeGb,
			WorkerCidr:     &newCidr,
			AutoScalerMin:  util.IntPtr(2),
			AutoScalerMax:  util.IntPtr(6),
			MaxSurge:       util.IntPtr(2),
			MaxUnavailable: util.IntPtr(1),
		},
	}

	upgradeShootInputForAzure := &gqlschema.UpgradeShootInput{
		GardenerConfig: &gqlschema.GardenerUpgradeInput{
			MachineType:    &newMachineType,
			DiskType:       &newDiskType,
			VolumeSizeGb:   &newVolumeSizeGb,
			WorkerCidr:     &newCidr,
			AutoScalerMin:  util.IntPtr(2),
			AutoScalerMax:  util.IntPtr(6),
			MaxSurge:       util.IntPtr(2),
			MaxUnavailable: util.IntPtr(1),
			ProviderSpecificConfig: &gqlschema.ProviderSpecificUpgradeInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: newCidr,
				},
			},
		},
	}

	upgradeShootInputForAWS := &gqlschema.UpgradeShootInput{
		GardenerConfig: &gqlschema.GardenerUpgradeInput{
			MachineType:    &newMachineType,
			DiskType:       &newDiskType,
			VolumeSizeGb:   &newVolumeSizeGb,
			WorkerCidr:     &newCidr,
			AutoScalerMin:  util.IntPtr(2),
			AutoScalerMax:  util.IntPtr(6),
			MaxSurge:       util.IntPtr(2),
			MaxUnavailable: util.IntPtr(1),
		},
	}

	testConfigs := []upgradeShootTestCase{
		{
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bbb",
			config:      upgradeShootInputForGCP,
			description: "Should upgrade a runtime with happy flow using correct Gardener with GCP configuration",
		},
		{
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb4",
			config:      upgradeShootInputForAzure,
			description: "Should upgrade a runtime with happy flow using correct Gardener with Azure configuration",
		},
		{
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb1",
			config:      upgradeShootInputForAWS,
			description: "Should provision and deprovision a runtime with happy flow using correct Gardener with Azure configuration when no zones passed",
		},
	}
	return testConfigs
}
