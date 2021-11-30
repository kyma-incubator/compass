package tombstone

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

// NewConverter missing godoc
func NewConverter() *converter {
	return &converter{}
}

// ToEntity missing godoc
func (c *converter) ToEntity(in *model.Tombstone) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		ID:            in.ID,
		OrdID:         in.OrdID,
		ApplicationID: in.ApplicationID,
		RemovalDate:   in.RemovalDate,
	}

	return output
}

// FromEntity missing godoc
func (c *converter) FromEntity(entity *Entity) (*model.Tombstone, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Tombstone entity is nil")
	}

	output := &model.Tombstone{
		ID:            entity.ID,
		OrdID:         entity.OrdID,
		ApplicationID: entity.ApplicationID,
		RemovalDate:   entity.RemovalDate,
	}

	return output, nil
}
