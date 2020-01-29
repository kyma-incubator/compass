package runtime

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/iosafety"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const (
	releaseInstallerURLFormat  = "https://github.com/kyma-project/kyma/releases/download/%s/kyma-installer-cluster.yaml"
	onDemandInstallerURLFormat = "https://storage.googleapis.com/kyma-development-artifacts/%s/kyma-installer-cluster.yaml"
)

// ComponentsListProvider provides the whole components list for creating a Kyma Runtime
type ComponentsListProvider struct {
	kymaVersion                      string
	managedRuntimeComponentsYAMLPath string
	httpClient                       HTTPDoer
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewComponentsListProvider returns new instance of the ComponentsListProvider
func NewComponentsListProvider(client HTTPDoer, kymaVersion string, managedRuntimeComponentsYAMLPath string) *ComponentsListProvider {
	return &ComponentsListProvider{
		httpClient:                       client,
		kymaVersion:                      kymaVersion,
		managedRuntimeComponentsYAMLPath: managedRuntimeComponentsYAMLPath,
	}
}

// AllComponents returns all components for Kyma Runtime. It fetches always the
// Kyma open-source components from the given url and management components from
// the file system and merge them together.
func (r ComponentsListProvider) AllComponents() ([]v1alpha1.KymaComponent, error) {
	// Read Kyma installer yaml (url)
	openSourceKymaComponents, err := r.getOpenSourceKymaComponents()
	if err != nil {
		return nil, errors.Wrap(err, "while getting open source Kyma components list")
	}

	// Read mounted config (path)
	managedRuntimeComponents, err := r.getManagedRuntimeComponents()
	if err != nil {
		return nil, errors.Wrap(err, "while getting managed runtime components list")
	}

	// Return merged list, managed components added at the end
	merged := append(openSourceKymaComponents, managedRuntimeComponents...)

	return merged, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (r ComponentsListProvider) getOpenSourceKymaComponents() (comp []v1alpha1.KymaComponent, err error) {
	installerYamlURL := r.getInstallerYamlURL()

	req, err := http.NewRequest(http.MethodGet, installerYamlURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while creating http request")
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if drainErr := iosafety.DrainReader(resp.Body); drainErr != nil {
			err = multierror.Append(err, errors.Wrap(drainErr, "while trying to drain body reader"))
		}

		if closeErr := resp.Body.Close(); closeErr != nil {
			err = multierror.Append(err, errors.Wrap(closeErr, "while trying to close body reader"))
		}
	}()

	if err := r.checkStatusCode(resp); err != nil {
		return nil, err
	}

	dec := yaml.NewDecoder(resp.Body)

	var t Installation
	for dec.Decode(&t) == nil {
		if t.Kind == "Installation" {
			return t.Spec.Components, nil
		}
	}
	return nil, errors.New("installer cr not found")

}

func (r ComponentsListProvider) getManagedRuntimeComponents() ([]v1alpha1.KymaComponent, error) {
	yamlFile, err := ioutil.ReadFile(r.managedRuntimeComponentsYAMLPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading YAML file with managed components mangedList")
	}

	var mangedList struct {
		Components []v1alpha1.KymaComponent `json:"components"`
	}
	err = yaml.Unmarshal(yamlFile, &mangedList)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling YAML file with managed components mangedList")
	}
	return mangedList.Components, nil
}

// Installation represents the installer CR.
// It is copied because using directly the installer CR
// with such fields:
//
// 	metav1.TypeMeta   `json:",inline"`
//	metav1.ObjectMeta `json:"metadata,omitempty"`
//
// is not working with "gopkg.in/yaml.v2" stream decoder.
// On the other hand "sigs.k8s.io/yaml" does not support
// stream decoding.
type Installation struct {
	Kind string                    `json:"kind"`
	Spec v1alpha1.InstallationSpec `json:"spec"`
}

func (r ComponentsListProvider) checkStatusCode(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		// limited buff to ready only ~4kb, so big response will not blowup our component
		body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			body = []byte(fmt.Sprintf("cannot read body, got error: %s", err))
		}
		return errors.Errorf("got unexpected status code, want %d, got %d, url: %s, body: %s",
			http.StatusOK, resp.StatusCode, resp.Request.URL.String(), body)
	}
	return nil
}

func (r ComponentsListProvider) getInstallerYamlURL() string {
	if r.isOnDemandRelease(r.kymaVersion) {
		return fmt.Sprintf(onDemandInstallerURLFormat, r.kymaVersion)
	}
	return fmt.Sprintf(releaseInstallerURLFormat, r.kymaVersion)

}

// isOnDemandRelease returns true if the version is recognized as on-demand.
//
// Detection rules:
//   For pull requests: PR-<number>
//   For changes to the master branch: master-<commit_sha>
//   For the latest changes in the master branch: master
//
// source: https://github.com/kyma-project/test-infra/blob/master/docs/prow/prow-architecture.md#generate-development-artifacts
func (r ComponentsListProvider) isOnDemandRelease(version string) bool {
	isOnDemandVersion := strings.HasPrefix(version, "PR-") ||
		strings.HasPrefix(version, "master-") ||
		strings.EqualFold(version, "master")
	return isOnDemandVersion
}
