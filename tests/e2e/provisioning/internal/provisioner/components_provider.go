package provisioner

import (
	"fmt"
	"github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
	"time"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

// This is code is copied from KEB
// This is temporary solution until the upgrade endpoint will be implemented in KEB itself

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

	var openSourceKymaComponents []v1alpha1.KymaComponent
	err := retry.Do(func() error {
		var err error
		openSourceKymaComponents, err = r.getOpenSourceKymaComponents(kymaVersion)
		if err != nil {
			log.Errorf("error getting open source Kyma components: %s", err.Error())
			return err
		}
		return nil
	}, retry.Delay(2*time.Second), retry.Attempts(10), retry.DelayType(retry.FixedDelay))
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
		return nil, errors.Wrap(err, "while making request for Kyma components list")
	}

	defer func() {
		drainAndClose(resp.Body)
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

	return errors.New(msg)
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

func drainAndClose(reader io.ReadCloser) {
	if reader == nil {
		return
	}

	_, drainError := io.Copy(ioutil.Discard, io.LimitReader(reader, 4096))
	if drainError != nil {
		log.Warnf("failed to drain ReadCloser: %s", drainError.Error())
	}
	err := reader.Close()
	if err != nil {
		log.Warnf("failed to close ReadCloser: %s", err.Error())
	}
}
