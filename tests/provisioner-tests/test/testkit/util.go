package testkit

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	gqlschema "github.com/kyma-project/control-plane/components/provisioner/pkg/gqlschema"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
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

func createInstallationCRURL(kymaVersion string) string {
	return fmt.Sprintf("https://raw.githubusercontent.com/kyma-project/kyma/%s/installation/resources/installer-cr-cluster-runtime.yaml.tpl", kymaVersion)
}

func toLowerCase(provider string) string {
	return strings.ToLower(provider)
}
