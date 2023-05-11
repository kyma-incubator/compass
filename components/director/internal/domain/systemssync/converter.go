package systemssync

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct{}

// NewConverter returns a new Converter used for conversion between repository and service representation of system sync model
func NewConverter() *converter {
	return &converter{}
}

// ToEntity converts the service model to repository entity
func (c *converter) ToEntity(in *model.SystemSynchronizationTimestamp) *Entity {
	if in == nil {
		return nil
	}

	return &Entity{
		ID:                in.ID,
		TenantID:          in.TenantID,
		ProductID:         in.ProductID,
		LastSyncTimestamp: in.LastSyncTimestamp,
	}
}

// FromEntity converts the repository entity to service model
func (c *converter) FromEntity(entity *Entity) *model.SystemSynchronizationTimestamp {
	if entity == nil {
		return nil
	}

	return &model.SystemSynchronizationTimestamp{
		ID:                entity.ID,
		TenantID:          entity.TenantID,
		ProductID:         entity.ProductID,
		LastSyncTimestamp: entity.LastSyncTimestamp,
	}
}
