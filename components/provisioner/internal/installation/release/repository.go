package release

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
)

func NewReleaseRepository(connection *dbr.Connection, uuidGenerator uuid.UUIDGenerator) *artifactsRepository {
	return &artifactsRepository{
		connection: connection,

		uuidGenerator: uuidGenerator,
	}
}

type artifactsRepository struct {
	connection *dbr.Connection

	uuidGenerator uuid.UUIDGenerator
}

func (r artifactsRepository) GetRelease(id string) (Release, dberrors.Error) {
	return r.getRelease(dbr.Eq("id", id))
}

func (r artifactsRepository) GetByVersion(version string) (Release, dberrors.Error) {
	return r.getRelease(dbr.Eq("version", version))
}

func (r artifactsRepository) getRelease(selector dbr.Builder) (Release, dberrors.Error) {
	session := r.connection.NewSession(nil)

	var artifacts Release

	err := session.
		Select("id", "version", "tiller_yaml", "installer_yaml").
		From("kyma_release").
		Where(selector).
		LoadOne(&artifacts)

	if err != nil {
		if err == dbr.ErrNotFound {
			return Release{}, dberrors.NotFound("Kyma release artifacts for version %s not found", id)
		}
		return Release{}, dberrors.Internal("Failed to get Kyma release artifacts for version %s: %s", id, err.Error())
	}

	return artifacts, nil
}

func (r artifactsRepository) SaveRelease(artifacts Release) (Release, dberrors.Error) {
	artifacts.Id = r.uuidGenerator.New()

	session := r.connection.NewSession(nil)

	_, err := session.InsertInto("kyma_release").
		Columns("id", "version", "tiller_yaml", "installer_yaml").
		Record(artifacts).
		Exec()

	if err != nil {
		return Release{}, dberrors.Internal("Failed to save Kyma release artifacts for version %s: %s", artifacts.Version, err.Error())
	}

	return artifacts, nil
}
