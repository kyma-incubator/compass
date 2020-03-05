package release

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/gocraft/dbr/v2"
	"github.com/pkg/errors"
)

const (
	onDemandTillerFileFormat    = "https://storage.googleapis.com/kyma-development-artifacts/%s/tiller.yaml"
	onDemandInstallerFileFormat = "https://storage.googleapis.com/kyma-development-artifacts/%s/kyma-installer-cluster.yaml"
)

type httpGetter interface {
	Get(url string) (resp *http.Response, err error)
}

// OnDemand implements minimal functionality necessary for download the Kyma release from on-demand versions
type OnDemand struct {
	httpGetter httpGetter
	generator  uuid.UUIDGenerator
	session    *dbr.Session
}

// NewOnDemand returns new instance of OnDemand
func NewOnDemand(httpGetter httpGetter, generator uuid.UUIDGenerator, connection *dbr.Connection) *OnDemand {
	return &OnDemand{
		httpGetter: httpGetter,
		generator:  generator,
		session:    connection.NewSession(nil),
	}
}

// GetReleaseByVersion returns the release. If release cannot be found in db then
// it tries to download the tiller and installer yaml files from on-demand version (lazy init).
func (o *OnDemand) GetReleaseByVersion(version string) (model.Release, dberrors.Error) {
	var release model.Release
	err := o.session.
		Select("id", "version", "tiller_yaml", "installer_yaml").
		From("kyma_release").
		Where(dbr.Eq("version", version)).
		LoadOne(&release)

	switch {
	case err == nil: // release found in db
		return release, nil

	case err == dbr.ErrNotFound: // release not found, try to download
		rel, err := o.downloadRelease(version)
		if err != nil {
			return model.Release{}, err
		}

		if err = o.saveRelease(rel); err != nil {
			return model.Release{}, err
		}
		return rel, nil

	default:
		return model.Release{}, dberrors.Internal("Failed to get Kyma release for version %s: %s", version, err.Error())
	}
}

// ReleaseExists returns true if the version is recognized as on-demand.
//
// Detection rules:
//   For pull requests: PR-<number>
//   For changes to the master branch: master-<commit_sha>
//   For the latest changes in the master branch: master
//
// source: https://github.com/kyma-project/test-infra/blob/master/docs/prow/prow-architecture.md#generate-development-artifacts
func (o *OnDemand) ReleaseExists(version string) (bool, dberrors.Error) {
	isOnDemandVersion := strings.HasPrefix(version, "PR-") ||
		strings.HasPrefix(version, "master-") ||
		strings.EqualFold(version, "master")
	return isOnDemandVersion, nil
}

func (o *OnDemand) get(url string) (string, error) {
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

func (o *OnDemand) downloadRelease(version string) (model.Release, dberrors.Error) {
	tillerYAML, err := o.get(fmt.Sprintf(onDemandTillerFileFormat, version))
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to download tiller YAML release for version %s: %s", version, err)
	}

	installerYAML, err := o.get(fmt.Sprintf(onDemandInstallerFileFormat, version))
	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to download installer YAML release for version %s: %s", version, err)
	}

	rel := model.Release{
		Id:            o.generator.New(),
		Version:       version,
		TillerYAML:    tillerYAML,
		InstallerYAML: installerYAML,
	}

	return rel, nil
}

func (o *OnDemand) saveRelease(rel model.Release) dberrors.Error {
	_, err := o.session.InsertInto("kyma_release").
		Columns("id", "version", "tiller_yaml", "installer_yaml").
		Record(rel).
		Exec()
	if err != nil {
		return dberrors.Internal("Failed to save Kyma release artifacts for version %s: %s", rel.Version, err.Error())
	}
	return nil
}
