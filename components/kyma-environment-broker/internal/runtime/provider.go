package runtime

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	kebError "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/error"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/iosafety"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	releaseInstallerURLFormat  = "https://github.com/kyma-project/kyma/releases/download/%s/kyma-installer-cluster.yaml"
	onDemandInstallerURLFormat = "https://storage.googleapis.com/kyma-development-artifacts/%s/kyma-installer-cluster.yaml"
)

// ComponentsListProvider provides the whole components list for creating a Kyma Runtime
type ComponentsListProvider struct {
	managedRuntimeComponentsYAMLPath string
	httpClient                       HTTPDoer
	components                       map[string][]v1alpha1.KymaComponent
}

type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewComponentsListProvider returns new instance of the ComponentsListProvider
func NewComponentsListProvider(managedRuntimeComponentsYAMLPath string) *ComponentsListProvider {
	return &ComponentsListProvider{
		httpClient:                       http.DefaultClient,
		managedRuntimeComponentsYAMLPath: managedRuntimeComponentsYAMLPath,
		components:                       make(map[string][]v1alpha1.KymaComponent, 0),
	}
}

// AllComponents returns all components for Kyma Runtime. It fetches always the
// Kyma open-source components from the given url and management components from
// the file system and merge them together.
func (r *ComponentsListProvider) AllComponents(kymaVersion string) ([]v1alpha1.KymaComponent, error) {
	if cmps, ok := r.components[kymaVersion]; ok {
		return cmps, nil
	}

	// Read Kyma installer yaml (url)
	openSourceKymaComponents, err := r.getOpenSourceKymaComponents(kymaVersion)
	if err != nil {
		return nil, errors.Wrap(err, "while getting open source kyma components")
	}

	// Read mounted config (path)
	managedRuntimeComponents, err := r.getManagedRuntimeComponents()
	if err != nil {
		return nil, errors.Wrap(err, "while getting managed runtime components list")
	}

	// Return merged list, managed components added at the end
	merged := append(openSourceKymaComponents, managedRuntimeComponents...)

	r.components[kymaVersion] = merged
	return merged, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (r *ComponentsListProvider) getOpenSourceKymaComponents(version string) (comp []v1alpha1.KymaComponent, err error) {
	installerYamlURL := r.getInstallerYamlURL(version)

	req, err := http.NewRequest(http.MethodGet, installerYamlURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "while creating http request")
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, kebError.AsTemporaryError(err, "while making request for Kyma components list")
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

func (r *ComponentsListProvider) getManagedRuntimeComponents() ([]v1alpha1.KymaComponent, error) {
	yamlFile, err := ioutil.ReadFile(r.managedRuntimeComponentsYAMLPath)
	if err != nil {
		return nil, errors.Wrap(err, "while reading YAML file with managed components list")
	}

	var managedList struct {
		Components []v1alpha1.KymaComponent `json:"components"`
	}
	err = yaml.Unmarshal(yamlFile, &managedList)
	logrus.Infof("%+v", managedList)
	for _, c := range managedList.Components {
		if c.Source != nil {
			logrus.Infof(c.Source.URL)
		}
	}
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshaling YAML file with managed components list")
	}
	return managedList.Components, nil
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

func (r *ComponentsListProvider) checkStatusCode(resp *http.Response) error {
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// limited buff to ready only ~4kb, so big response will not blowup our component
	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		body = []byte(fmt.Sprintf("cannot read body, got error: %s", err))
	}
	msg := fmt.Sprintf("while checking response status code for Kyma components list: "+
		"got unexpected status code, want %d, got %d, url: %s, body: %s",
		http.StatusOK, resp.StatusCode, resp.Request.URL.String(), body)

	switch {
	case resp.StatusCode == http.StatusRequestTimeout:
		return kebError.NewTemporaryError(msg)
	case resp.StatusCode >= http.StatusInternalServerError:
		return kebError.NewTemporaryError(msg)
	default:
		return errors.New(msg)
	}
}

func (r *ComponentsListProvider) getInstallerYamlURL(kymaVersion string) string {
	if r.isOnDemandRelease(kymaVersion) {
		return fmt.Sprintf(onDemandInstallerURLFormat, kymaVersion)
	}
	return fmt.Sprintf(releaseInstallerURLFormat, kymaVersion)
}

// isOnDemandRelease returns true if the version is recognized as on-demand.
//
// Detection rules:
//   For pull requests: PR-<number>
//   For changes to the master branch: master-<commit_sha>
//
// source: https://github.com/kyma-project/test-infra/blob/master/docs/prow/prow-architecture.md#generate-development-artifacts
func (r *ComponentsListProvider) isOnDemandRelease(version string) bool {
	isOnDemandVersion := strings.HasPrefix(version, "PR-") ||
		strings.HasPrefix(version, "master-")
	return isOnDemandVersion
}
