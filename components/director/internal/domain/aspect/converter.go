package aspect

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

type converter struct {
}

// NewConverter returns a new Converter that can later be used to make the conversions between the service and repository layer representations of a Compass Aspect.
func NewConverter() *converter {
	return &converter{}
}

// FromEntity converts the provided Entity repo-layer representation of an Aspect to the service-layer representation model.Aspect.
func (c *converter) FromEntity(entity *Entity) *model.Aspect {
	return &model.Aspect{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		IntegrationDependencyID:      entity.IntegrationDependencyID,
		Title:                        entity.Title,
		Description:                  repo.StringPtrFromNullableString(entity.Description),
		Mandatory:                    repo.BoolPtrFromNullableBool(entity.Mandatory),
		SupportMultipleProviders:     repo.BoolPtrFromNullableBool(entity.SupportMultipleProviders),
		APIResources:                 repo.JSONRawMessageFromNullableString(entity.APIResources),
		EventResources:               repo.JSONRawMessageFromNullableString(entity.EventResources),
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
		ApplicationID:                repo.NewNullableString(aspectModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(aspectModel.ApplicationTemplateVersionID),
		IntegrationDependencyID:      aspectModel.IntegrationDependencyID,
		Title:                        aspectModel.Title,
		Description:                  repo.NewNullableString(aspectModel.Description),
		Mandatory:                    repo.NewNullableBool(aspectModel.Mandatory),
		SupportMultipleProviders:     repo.NewNullableBool(aspectModel.SupportMultipleProviders),
		APIResources:                 repo.NewNullableStringFromJSONRawMessage(aspectModel.APIResources),
		EventResources:               repo.NewNullableStringFromJSONRawMessage(aspectModel.EventResources),
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
