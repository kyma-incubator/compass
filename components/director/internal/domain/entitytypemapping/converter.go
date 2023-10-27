package entitytypemapping

import (
	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

type converter struct {
}

// NewConverter returns new converter instance
func NewConverter() *converter {
	return &converter{}
}

// ToEntity converts to Entity
func (c *converter) ToEntity(entityModel *model.EntityTypeMapping) *Entity {
	if entityModel == nil {
		return nil
	}

	output := &Entity{
		BaseEntity: &repo.BaseEntity{
			ID:        entityModel.ID,
			Ready:     entityModel.Ready,
			CreatedAt: entityModel.CreatedAt,
			UpdatedAt: entityModel.UpdatedAt,
			DeletedAt: entityModel.DeletedAt,
			Error:     repo.NewNullableString(entityModel.Error),
		},
		APIDefinitionID:   repo.NewNullableString(entityModel.APIDefinitionID),
		EventDefinitionID: repo.NewNullableString(entityModel.EventDefinitionID),
		APIModelSelectors: repo.NewNullableStringFromJSONRawMessage(entityModel.APIModelSelectors),
		EntityTypeTargets: repo.NewNullableStringFromJSONRawMessage(entityModel.EntityTypeTargets),
	}

	return output
}

// FromEntity converts from model.Entity*
func (c *converter) FromEntity(entity *Entity) *model.EntityTypeMapping {
	if entity == nil {
		return nil
	}

	output := &model.EntityTypeMapping{
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
		APIDefinitionID:   repo.StringPtrFromNullableString(entity.APIDefinitionID),
		EventDefinitionID: repo.StringPtrFromNullableString(entity.EventDefinitionID),
		APIModelSelectors: repo.JSONRawMessageFromNullableString(entity.APIModelSelectors),
		EntityTypeTargets: repo.JSONRawMessageFromNullableString(entity.EntityTypeTargets),
	}
	return output
}
