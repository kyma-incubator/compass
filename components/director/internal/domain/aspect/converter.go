package aspect

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type converter struct {
}

func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an Aspect to the service-layer representation model.Aspect.
func (c *converter) FromEntity(entity *Entity) *model.Aspect {
	return &model.Aspect{
		IntegrationDependencyID:  entity.IntegrationDependencyID,
		Name:                     entity.Name,
		Description:              repo.StringPtrFromNullableString(entity.Description),
		Mandatory:                entity.Mandatory,
		SupportMultipleProviders: repo.BoolPtrFromNullableBool(entity.SupportMultipleProviders),
		ApiResources:             repo.JSONRawMessageFromNullableString(entity.ApiResources),
		EventResources:           repo.JSONRawMessageFromNullableString(entity.EventResources),
		BaseEntity: &model.BaseEntity{
			ID:        entity.ID,
			Ready:     entity.Ready,
			CreatedAt: entity.CreatedAt,
			UpdatedAt: entity.UpdatedAt,
			DeletedAt: entity.DeletedAt,
			Error:     repo.StringPtrFromNullableString(entity.Error),
		},
	}
}

// ToEntity converts the provided service-layer representation of an Aspect to the repository-layer one.
func (c *converter) ToEntity(aspectModel *model.Aspect) *Entity {
	return &Entity{
		IntegrationDependencyID:  aspectModel.IntegrationDependencyID,
		Name:                     aspectModel.Name,
		Description:              repo.NewNullableString(aspectModel.Description),
		Mandatory:                aspectModel.Mandatory,
		SupportMultipleProviders: repo.NewNullableBool(aspectModel.SupportMultipleProviders),
		ApiResources:             repo.NewNullableStringFromJSONRawMessage(aspectModel.ApiResources),
		EventResources:           repo.NewNullableStringFromJSONRawMessage(aspectModel.EventResources),
		BaseEntity: &repo.BaseEntity{
			ID:        aspectModel.ID,
			Ready:     aspectModel.Ready,
			CreatedAt: aspectModel.CreatedAt,
			UpdatedAt: aspectModel.UpdatedAt,
			DeletedAt: aspectModel.DeletedAt,
			Error:     repo.NewNullableString(aspectModel.Error),
		},
	}
}
