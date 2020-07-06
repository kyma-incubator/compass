package api

import (
	"github.com/kyma-project/control-plane/components/provisioner/internal/util"
	"github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
)

type testCase struct {
	name              string
	description       string
	runtimeID         string
	provisioningInput provisioningInput
	upgradeShootInput gqlschema.UpgradeShootInput
}

type provisioningInput struct {
	config       gqlschema.ClusterConfigInput
	runtimeInput gqlschema.RuntimeInput
}

var providerCredentials = &gqlschema.CredentialsInput{SecretName: "secret_1"}

func newTestProvisioningConfigs() []testCase {
	return []testCase{
		{name: "GCP on Gardener",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct GCP configuration for Gardener",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bbb",
			provisioningInput: provisioningInput{
				config: gcpGardenerClusterConfigInput(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 1",
					Description: new(string),
				}},
			upgradeShootInput: NewUpgradeShootInput(),
		},
		{name: "Azure on Gardener (with zones)",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct Azure configuration for Gardener, when zones passed",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb4",
			provisioningInput: provisioningInput{
				config: azureGardenerClusterConfigInput("fix-az-zone-1", "fix-az-zone-2"),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 2",
					Description: new(string),
				}},
			upgradeShootInput: NewAzureUpgradeShootInput(),
		},
		{name: "Azure on Gardener (without zones)",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct Azure configuration for Gardener, when zones are empty",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb1",
			provisioningInput: provisioningInput{
				config: azureGardenerClusterConfigInput(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 3",
					Description: new(string),
				}},
			upgradeShootInput: NewAzureUpgradeShootInput(),
		},
		{name: "AWS on Gardener",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct AWS configuration for Gardener",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb5",
			provisioningInput: provisioningInput{
				config: awsGardenerClusterConfigInput(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime4",
					Description: new(string),
				}},
			upgradeShootInput: NewUpgradeShootInput(),
		},
	}
}

func gcpGardenerClusterConfigInput() gqlschema.ClusterConfigInput {
	return gqlschema.ClusterConfigInput{
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
}

func azureGardenerClusterConfigInput(zones ...string) gqlschema.ClusterConfigInput {
	return gqlschema.ClusterConfigInput{
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

func awsGardenerClusterConfigInput() gqlschema.ClusterConfigInput {
	return gqlschema.ClusterConfigInput{
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
}

func NewUpgradeShootInput() gqlschema.UpgradeShootInput {
	newMachineType := "new-machine"
	newDiskType := "papyrus"
	newVolumeSizeGb := 50
	newCidr := "cidr2"

	return gqlschema.UpgradeShootInput{
		GardenerConfig: &gqlschema.GardenerUpgradeInput{
			MachineType:            &newMachineType,
			DiskType:               &newDiskType,
			VolumeSizeGb:           &newVolumeSizeGb,
			WorkerCidr:             &newCidr,
			AutoScalerMin:          util.IntPtr(2),
			AutoScalerMax:          util.IntPtr(6),
			MaxSurge:               util.IntPtr(2),
			MaxUnavailable:         util.IntPtr(1),
			ProviderSpecificConfig: nil,
		},
	}
}

func NewAzureUpgradeShootInput() gqlschema.UpgradeShootInput {
	input := NewUpgradeShootInput()
	input.GardenerConfig.ProviderSpecificConfig = &gqlschema.ProviderSpecificInput{
		AzureConfig: &gqlschema.AzureProviderConfigInput{
			VnetCidr: "cidr2",
		},
	}
	return input
}
