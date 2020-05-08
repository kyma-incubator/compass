package testkit

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

func CreateGardenerProvisioningInput(config *TestConfig, version, provider string) (gqlschema.ProvisionRuntimeInput, error) {
	gardenerInputs := map[string]gqlschema.GardenerConfigInput{
		GCP: {
			MachineType:  "n1-standard-4",
			DiskType:     "pd-standard",
			Region:       "europe-west4",
			TargetSecret: config.Gardener.GCPSecret,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zones: []string{"europe-west4-a", "europe-west4-b", "europe-west4-c"},
				},
			},
		},
		Azure: {
			MachineType:  "Standard_D4_v3",
			DiskType:     "Standard_LRS",
			Region:       "westeurope",
			TargetSecret: config.Gardener.AzureSecret,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: "10.250.0.0/19",
					Zones:    []string{"westeurope-1", "westeurope-2", "westeurope-3"},
				},
			},
		},
	}

	kymaConfigInput, err := CreateKymaConfigInput(version)
	if err != nil {
		return gqlschema.ProvisionRuntimeInput{}, fmt.Errorf("failed to create kyma config input: %s", err.Error())
	}

	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name: "",
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				KubernetesVersion:      "1.15.10",
				DiskType:               gardenerInputs[provider].DiskType,
				VolumeSizeGb:           35,
				MachineType:            gardenerInputs[provider].MachineType,
				Region:                 gardenerInputs[provider].Region,
				Provider:               toLowerCase(provider),
				TargetSecret:           gardenerInputs[provider].TargetSecret,
				WorkerCidr:             "10.250.0.0/19",
				AutoScalerMin:          2,
				AutoScalerMax:          4,
				MaxSurge:               4,
				MaxUnavailable:         1,
				ProviderSpecificConfig: gardenerInputs[provider].ProviderSpecificConfig,
			},
		},
		KymaConfig: kymaConfigInput,
	}, nil
}

func CreateKymaConfigInput(version string) (*gqlschema.KymaConfigInput, error) {
	installationCRURL := createInstallationCRURL(version)
	logrus.Infof("Getting and parsing Kyma modules from Installation CR at: %s", installationCRURL)
	componentConfigInput, err := GetAndParseInstallerCR(installationCRURL)
	if err != nil {
		return &gqlschema.KymaConfigInput{}, fmt.Errorf("failed to create component config input: %s", err.Error())
	}

	return &gqlschema.KymaConfigInput{Version: version, Components: componentConfigInput}, nil
}
