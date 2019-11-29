package release

import (
	"github.com/gocraft/dbr/v2"
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
)

//go:generate mockery -name=Repository
type Repository interface {
	ReadRepository
	SaveRelease(artifacts model.Release) (model.Release, dberrors.Error)
}

type ReadRepository interface {
	GetReleaseByVersion(version string) (model.Release, dberrors.Error)
	ReleaseExists(version string) (bool, dberrors.Error)
}

func NewReleaseRepository(connection *dbr.Connection, generator uuid.UUIDGenerator) *releaseRepository {
	return &releaseRepository{
		connection: connection,
		generator:  generator,
	}
}

type releaseRepository struct {
	connection *dbr.Connection
	generator  uuid.UUIDGenerator
}

func (r releaseRepository) GetReleaseByVersion(version string) (model.Release, dberrors.Error) {
	session := r.connection.NewSession(nil)

	var release model.Release

	err := session.
		Select("id", "version", "tiller_yaml", "installer_yaml").
		From("kyma_release").
		Where(dbr.Eq("version", version)).
		LoadOne(&release)

	if err != nil {
		if err == dbr.ErrNotFound {
			return model.Release{}, dberrors.NotFound("Kyma release for version %s not found", version)
		}
		return model.Release{}, dberrors.Internal("Failed to get Kyma release for version %s: %s", version, err.Error())
	}

	return release, nil
}

func (r releaseRepository) ReleaseExists(version string) (bool, dberrors.Error) {
	_, err := r.GetReleaseByVersion(version)

	if err != nil {
		if err.Code() == dberrors.CodeNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r releaseRepository) SaveRelease(artifacts model.Release) (model.Release, dberrors.Error) {
	artifacts.Id = r.generator.New()
	session := r.connection.NewSession(nil)

	_, err := session.InsertInto("kyma_release").
		Columns("id", "version", "tiller_yaml", "installer_yaml").
		Record(artifacts).
		Exec()

	if err != nil {
		return model.Release{}, dberrors.Internal("Failed to save Kyma release artifacts for version %s: %s", artifacts.Version, err.Error())
	}

	return artifacts, nil
}
