package testkit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	gqlschema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"
)

const InstallationCRURL = "https://raw.githubusercontent.com/kyma-project/kyma/master/installation/resources/installer-cr-cluster-with-compass.yaml.tpl"

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
	resp, err := http.Get(InstallationCRURL)
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
