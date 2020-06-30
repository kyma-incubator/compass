package api

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
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
				config: gcpClusterConfig(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 1",
					Description: new(string),
				}},
			upgradeShootInput: newUpgradeShootInput(),
		},
		{name: "Azure on Gardener (with zones)",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct Azure configuration for Gardener, when zones passed",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb4",
			provisioningInput: provisioningInput{
				config: azureClusterConfigInput("fix-az-zone-1", "fix-az-zone-2"),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 2",
					Description: new(string),
				}},
			upgradeShootInput: newAzureUpgradeShootInput(),
		},
		{name: "Azure on Gardener (without zones)",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct Azure configuration for Gardener, when zones are empty",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb1",
			provisioningInput: provisioningInput{
				config: azureClusterConfigInput(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime 3",
					Description: new(string),
				}},
			upgradeShootInput: newAzureUpgradeShootInput(),
		},
		{name: "AWS on Gardener",
			description: "Should provision, deprovision a runtime and upgrade shoot on happy path, using correct AWS configuration for Gardener",
			runtimeID:   "1100bb59-9c40-4ebb-b846-7477c4dc5bb5",
			provisioningInput: provisioningInput{
				config: awsClusterConfigInput(),
				runtimeInput: gqlschema.RuntimeInput{
					Name:        "test runtime4",
					Description: new(string),
				}},
			upgradeShootInput: newUpgradeShootInput(),
		},
	}
}

func gcpClusterConfig() gqlschema.ClusterConfigInput {
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

func azureClusterConfigInput(zones ...string) gqlschema.ClusterConfigInput {
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

func awsClusterConfigInput() gqlschema.ClusterConfigInput {
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

func newUpgradeShootInput() gqlschema.UpgradeShootInput {
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

func newAzureUpgradeShootInput() gqlschema.UpgradeShootInput {
	input := newUpgradeShootInput()
	input.GardenerConfig.ProviderSpecificConfig = &gqlschema.ProviderSpecificInput{
		AzureConfig: &gqlschema.AzureProviderConfigInput{
			VnetCidr: "cidr2",
		},
	}
	return input
}

func expectedShootConfig(initConfig, upgradeConfig gqlschema.GardenerConfig) gqlschema.GardenerConfig {
	var expectedProviderConfig gqlschema.ProviderSpecificConfig

	switch upgradedConfig := upgradeConfig.ProviderSpecificConfig.(type) {
	case gqlschema.AWSProviderConfig, gqlschema.GCPProviderConfig:
		expectedProviderConfig = upgradedConfig
	case gqlschema.AzureProviderConfig:
		azureConfig := initConfig.ProviderSpecificConfig.(gqlschema.AzureProviderConfig)
		expectedProviderConfig = gqlschema.AzureProviderConfig{
			Zones:    azureConfig.Zones,
			VnetCidr: upgradedConfig.VnetCidr,
		}
	}

	// if upgraded.ProviderSpecificConfig != nil {
	// 	expectedProviderConfig = config.ProviderSpecificConfig
	// }
	// if upgraded.ProviderSpecificConfig.AzureConfig != nil {
	// 	initConfig := config.ProviderSpecificConfig.(gqlschema.AzureProviderConfig)

	// 	expectedProviderConfig = gqlschema.AzureProviderConfig{
	// 		Zones:    initConfig.Zones,
	// 		VnetCidr: &upgraded.ProviderSpecificConfig.AzureConfig.VnetCidr,
	// 	}
	// }
	// if upgraded.ProviderSpecificConfig.GcpConfig != nil {
	// 	expectedProviderConfig = config.ProviderSpecificConfig
	// }

	return gqlschema.GardenerConfig{
		Name:         initConfig.Name,
		Provider:     initConfig.Provider,
		TargetSecret: initConfig.TargetSecret,
		Seed:         initConfig.Seed,
		Region:       initConfig.Region,

		KubernetesVersion: upgradeConfig.KubernetesVersion,
		MachineType:       upgradeConfig.MachineType,
		DiskType:          upgradeConfig.DiskType,
		VolumeSizeGb:      upgradeConfig.VolumeSizeGb,
		AutoScalerMin:     upgradeConfig.AutoScalerMin,
		AutoScalerMax:     upgradeConfig.AutoScalerMax,
		MaxSurge:          upgradeConfig.MaxSurge,
		MaxUnavailable:    upgradeConfig.MaxUnavailable,
		WorkerCidr:        upgradeConfig.WorkerCidr,

		ProviderSpecificConfig: expectedProviderConfig,
	}
}
