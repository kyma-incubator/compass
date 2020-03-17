package release

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/pkg/errors"
)

const (
	onDemandTillerFileFormat    = "https://storage.googleapis.com/kyma-development-artifacts/%s/tiller.yaml"
	onDemandInstallerFileFormat = "https://storage.googleapis.com/kyma-development-artifacts/%s/kyma-installer-cluster.yaml"
)

type httpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

// OnDemandWrapper wraps release.Repository with minimal functionality necessary for downloading the Kyma release from on-demand versions
type OnDemandWrapper struct {
	httpGetter httpGetter
	repository Repository
}

// NewOnDemandWrapper returns new instance of OnDemandWrapper
func NewOnDemandWrapper(httpGetter httpGetter, releaseRepo Repository) *OnDemandWrapper {
	return &OnDemandWrapper{
		httpGetter: httpGetter,
		repository: releaseRepo,
	}
}

// GetReleaseByVersion returns the release. If release cannot be found in db then
// it tries to download the tiller and installer yaml files from on-demand version (lazy init).
func (o *OnDemandWrapper) GetReleaseByVersion(version string) (model.Release, dberrors.Error) {
	release, err := o.repository.GetReleaseByVersion(version)

	switch {
	case err == nil: // release found in db
		return release, nil

	case err.Code() == dberrors.CodeNotFound: // release not found, if is on-demand version, try to download
		if !o.isOnDemandVersion(version) {
			return model.Release{}, err
		}

		rel, err := o.downloadRelease(version)
		if err != nil {
			return model.Release{}, err
		}

		rel, err = o.saveRelease(rel)
		if err != nil {
			return model.Release{}, err
		}
		return rel, nil

	default:
		return model.Release{}, dberrors.Internal("Failed to get Kyma release for version %s: %s", version, err.Error())
	}
}

// Detection rules:
//   For pull requests: PR-<number>
//   For changes to the master branch: master-<commit_sha>
//   For the latest changes in the master branch: master
//
// source: https://github.com/kyma-project/test-infra/blob/master/docs/prow/prow-architecture.md#generate-development-artifacts
func (o *OnDemandWrapper) isOnDemandVersion(version string) bool {
	isOnDemandVersion := strings.HasPrefix(version, "PR-") ||
		strings.HasPrefix(version, "master-") ||
		strings.EqualFold(version, "master")
	return isOnDemandVersion
}

func (o *OnDemandWrapper) get(url string) (string, error) {
	resp, err := o.httpGetter.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "while executing get request on url: %q", url)
	}
	defer util.Close(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return "", errors.Errorf("received unexpected http status %d", resp.StatusCode)
	}

	reqBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "while reading body")
	}

	return string(reqBody), nil
}

func (o *OnDemandWrapper) downloadRelease(version string) (model.Release, dberrors.Error) {
	tillerYAML, err := o.get(fmt.Sprintf(onDemandTillerFileFormat, version))
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to download tiller YAML release for version %s: %s", version, err)
	}

	installerYAML, err := o.get(fmt.Sprintf(onDemandInstallerFileFormat, version))
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to download installer YAML release for version %s: %s", version, err)
	}

	rel := model.Release{
		Version:       version,
		TillerYAML:    tillerYAML,
		InstallerYAML: installerYAML,
	}

	return rel, nil
}

func (o *OnDemandWrapper) saveRelease(rel model.Release) (model.Release, dberrors.Error) {
	rel, err := o.repository.SaveRelease(rel)
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to save Kyma release artifacts for version %s: %s", rel.Version, err.Error())
	}
	return rel, nil
}
