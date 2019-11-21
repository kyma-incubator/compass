package artifacts

import (
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"
)

func NewArtifactsRepository(connection *dbr.Connection, uuidGenerator uuid.UUIDGenerator) *artifactsRepository {
	return &artifactsRepository{
		connection: connection,

		uuidGenerator: uuidGenerator,
	}
}

type artifactsRepository struct {
	connection *dbr.Connection

	uuidGenerator uuid.UUIDGenerator
}

func (r artifactsRepository) GetArtifacts(version string) (ReleaseArtifacts, dberrors.Error) {
	session := r.connection.NewSession(nil)

	var artifacts ReleaseArtifacts

	err := session.
		Select("id", "version", "tiller_yaml", "installer_yaml").
		From("kyma_artifacts").
		Where(dbr.Eq("version", version)).
		LoadOne(&artifacts)

	if err != nil {
		if err == dbr.ErrNotFound {
			return ReleaseArtifacts{}, dberrors.NotFound("Kyma release artifacts for version %s not found", version)
		}
		return ReleaseArtifacts{}, dberrors.Internal("Failed to get Kyma release artifacts for version %s: %s", version, err.Error())
	}

	return artifacts, nil
}

func (r artifactsRepository) SaveArtifacts(artifacts ReleaseArtifacts) (ReleaseArtifacts, dberrors.Error) {
	artifacts.Id = r.uuidGenerator.New()

	session := r.connection.NewSession(nil)

	_, err := session.InsertInto("kyma_artifacts").
		Columns("id", "version", "tiller_yaml", "installer_yaml").
		Record(artifacts).
		Exec()

	if err != nil {
		return ReleaseArtifacts{}, dberrors.Internal("Failed to save Kyma release artifacts for version %s: %s", artifacts.Version, err.Error())
	}

	return artifacts, nil
}
