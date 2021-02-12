package tombstone

import (
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

func (c *converter) ToEntity(in *model.Tombstone) *Entity {
	if in == nil {
		return nil
	}

	output := &Entity{
		OrdID:         in.OrdID,
		TenantID:      in.TenantID,
		ApplicationID: in.ApplicationID,
		RemovalDate:   in.RemovalDate,
	}

	return output
}

func (c *converter) FromEntity(entity *Entity) (*model.Tombstone, error) {
	if entity == nil {
		return nil, apperrors.NewInternalError("the Tombstone entity is nil")
	}

	output := &model.Tombstone{
		OrdID:         entity.OrdID,
		TenantID:      entity.TenantID,
		ApplicationID: entity.ApplicationID,
		RemovalDate:   entity.RemovalDate,
	}

	return output, nil
}
