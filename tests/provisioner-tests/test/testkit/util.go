package testkit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	gqlschema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	yaml "gopkg.in/yaml.v3"
)

const (
	installationCRURL = "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl"

	Azure = "Azure"
	GCP   = "GCP"
	AWS   = "AWS"
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

func RunParallelToMainFunction(timeout time.Duration, mainFunction func() error, parallelFunctions ...func() error) error {
	mainOut := make(chan error, 1)
	go func() {
		mainOut <- mainFunction()
	}()

	errOut := make(chan error, len(parallelFunctions))
	for _, fun := range parallelFunctions {
		go func(function func() error) {
			errOut <- function()
		}(fun)
	}

	funcErrors := make([]error, 0, len(parallelFunctions))

	for {
		select {
		case err := <-errOut:
			funcErrors = append(funcErrors, err)
		case err := <-mainOut:
			if err != nil {
				return errors.Errorf("Main function failed: %s", err.Error())
			}

			if len(funcErrors) < len(parallelFunctions) {
				return errors.Errorf("Not all parallel functions finished. Functions finished %d. Errors: %v", len(funcErrors), processErrors(funcErrors))
			}

			return processErrors(funcErrors)
		case <-time.After(timeout):
			return errors.Errorf("Timeout waiting for for parallel processes to finish. Functions finished %d. Errors: %v", len(funcErrors), processErrors(funcErrors))
		}
	}
}

func GetAndParseInstallerCR() ([]*gqlschema.ComponentConfigurationInput, error) {
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
			MachineType:  "Standard_D2_v3",
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

	logrus.Infof("Getting and parsing Kyma modules from Installation CR at: %s", installationCRURL)
	componentConfigInput, err := GetAndParseInstallerCR()
	if err != nil {
		return gqlschema.ProvisionRuntimeInput{}, fmt.Errorf("Failed to create component config input: %s", err.Error())
	}

	runtimeName := fmt.Sprintf("%s%s", "runtime", uuid.New().String()[:4])

	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name: runtimeName,
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				KubernetesVersion:      "1.15.4",
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

func processErrors(errorsArray []error) error {
	errorMsg := ""

	for i, err := range errorsArray {
		if err != nil {
			errorMsg = fmt.Sprintf("%s -- Error %d not nil: %s.", errorMsg, i, err.Error())
		}
	}

	if errorMsg != "" {
		return errors.Errorf("Errors: %s", errorMsg)
	}

	return nil
}

func toLowerCase(provider string) string {
	return strings.ToLower(provider)
}
