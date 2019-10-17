package persistence

import (
	"github.com/gofrs/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
)

//go:generate mockery -name=UUIDGenerator
type UUIDGenerator interface {
	New() (string, dberrors.Error)
}

type uuidGenerator struct {
}

func NewUUIDGenerator() UUIDGenerator {
	return uuidGenerator{}
}

func (u uuidGenerator) New() (string, dberrors.Error) {
	id, err := uuid.NewV4()
	if err != nil {
		return "", dberrors.Internal("Failed to generate UUID: %s", err)
	}

	return id.String(), nil
}
