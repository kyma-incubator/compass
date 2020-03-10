package testkit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	gqlschema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v3"
)

const (
	Azure = "Azure"
	GCP   = "GCP"
)

func WaitForFunction(interval, timeout time.Duration, isDone func() bool) error {
	done := time.After(timeout)

	for {
		if isDone() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}

func GetAndParseInstallerCR(installationCRURL string) ([]*gqlschema.ComponentConfigurationInput, error) {
	resp, err := http.Get(installationCRURL)
	if err != nil {
		return nil, fmt.Errorf("Error fetching installation CR: %s", err.Error())
	}
	crContent, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading body of installation CR GET response: %s", err.Error())
	}

	installationCR := v1alpha1.Installation{}
	err = yaml.NewDecoder(bytes.NewBuffer(crContent)).Decode(&installationCR)
	if err != nil {
		return nil, fmt.Errorf("Error decoding installer CR: %s", err.Error())
	}
	var components = make([]*gqlschema.ComponentConfigurationInput, 0, len(installationCR.Spec.Components))
	for _, component := range installationCR.Spec.Components {
		in := &gqlschema.ComponentConfigurationInput{
			Component: component.Name,
			Namespace: component.Namespace,
		}
		components = append(components, in)
	}
	return components, nil
}

func CreateGardenerProvisioningInput(config *TestConfig, provider string) (gqlschema.ProvisionRuntimeInput, error) {
	gardenerInputs := map[string]gqlschema.GardenerConfigInput{
		GCP: {
			MachineType:  "n1-standard-4",
			DiskType:     "pd-standard",
			Region:       "europe-west4",
			TargetSecret: config.Gardener.GCPSecret,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zone: "europe-west4-a",
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
				},
			},
		},
	}

	installationCRURL := createInstallationCRURL(config.Kyma.Version)
	logrus.Infof("Getting and parsing Kyma modules from Installation CR at: %s", installationCRURL)
	componentConfigInput, err := GetAndParseInstallerCR(installationCRURL)
	if err != nil {
		return gqlschema.ProvisionRuntimeInput{}, fmt.Errorf("Failed to create component config input: %s", err.Error())
	}

	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name: "",
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				KubernetesVersion:      "1.15.10",
				NodeCount:              3,
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
		KymaConfig: &gqlschema.KymaConfigInput{
			Version:    config.Kyma.Version,
			Components: componentConfigInput,
		},
	}, nil
}

func createInstallationCRURL(kymaVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kyma-project/kyma/%s/installation/resources/installer-cr-cluster-runtime.yaml.tpl", kymaVersion)
}

func toLowerCase(provider string) string {
	return strings.ToLower(provider)
}
