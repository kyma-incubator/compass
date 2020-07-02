package release

import (
	"fmt"
	"strings"

	"github.com/kyma-project/control-plane/components/provisioner/internal/model"
	"github.com/kyma-project/control-plane/components/provisioner/internal/persistence/dberrors"
)

const (
	onDemandTillerFileFormat    = "https://storage.googleapis.com/kyma-development-artifacts/%s/tiller.yaml"
	onDemandInstallerFileFormat = "https://storage.googleapis.com/kyma-development-artifacts/%s/kyma-installer-cluster.yaml"
)

type TextFileDownloader interface {
	Download(url string) (string, error)
	DownloadOrEmpty(url string) (string, error)
}

// OnDemandWrapper wraps release.Repository with minimal functionality necessary for downloading the Kyma release from on-demand versions
type OnDemandWrapper struct {
	downloader TextFileDownloader
	repository Repository
}

// NewOnDemandWrapper returns new instance of OnDemandWrapper
func NewOnDemandWrapper(downloader TextFileDownloader, releaseRepo Repository) *OnDemandWrapper {
	return &OnDemandWrapper{
		downloader: downloader,
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

func (o *OnDemandWrapper) downloadRelease(version string) (model.Release, dberrors.Error) {
	tillerYAML, err := o.downloader.DownloadOrEmpty(fmt.Sprintf(onDemandTillerFileFormat, version))
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to download tiller YAML release for version %s: %s", version, err)
	}

	installerYAML, err := o.downloader.Download(fmt.Sprintf(onDemandInstallerFileFormat, version))
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
