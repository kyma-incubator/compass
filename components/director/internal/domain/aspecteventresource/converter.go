package aspecteventresource

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
func (c *converter) FromEntity(entity *Entity) *model.AspectEventResource {
	if entity == nil {
		return nil
	}

	return &model.AspectEventResource{
		ApplicationID:                repo.StringPtrFromNullableString(entity.ApplicationID),
		ApplicationTemplateVersionID: repo.StringPtrFromNullableString(entity.ApplicationTemplateVersionID),
		AspectID:                     entity.AspectID,
		OrdID:                        entity.OrdID,
		MinVersion:                   repo.StringPtrFromNullableString(entity.MinVersion),
		Subset:                       repo.JSONRawMessageFromNullableString(entity.Subset),
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
func (c *converter) ToEntity(aspectEventResourceModel *model.AspectEventResource) *Entity {
	if aspectEventResourceModel == nil {
		return nil
	}

	return &Entity{
		ApplicationID:                repo.NewNullableString(aspectEventResourceModel.ApplicationID),
		ApplicationTemplateVersionID: repo.NewNullableString(aspectEventResourceModel.ApplicationTemplateVersionID),
		AspectID:                     aspectEventResourceModel.AspectID,
		OrdID:                        aspectEventResourceModel.OrdID,
		MinVersion:                   repo.NewNullableString(aspectEventResourceModel.MinVersion),
		Subset:                       repo.NewNullableStringFromJSONRawMessage(aspectEventResourceModel.Subset),
		BaseEntity: &repo.BaseEntity{
			ID:        aspectEventResourceModel.ID,
			Ready:     aspectEventResourceModel.Ready,
			CreatedAt: aspectEventResourceModel.CreatedAt,
			UpdatedAt: aspectEventResourceModel.UpdatedAt,
			DeletedAt: aspectEventResourceModel.DeletedAt,
			Error:     repo.NewNullableString(aspectEventResourceModel.Error),
		},
	}
}
