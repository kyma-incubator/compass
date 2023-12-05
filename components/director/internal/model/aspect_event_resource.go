package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// AspectEventResource represent structure for Aspect Event Resource
type AspectEventResource struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	AspectID                     string
	OrdID                        string
	MinVersion                   *string
	Subset                       json.RawMessage
	*BaseEntity
}

// GetType returns Type aspectEventResource
func (*AspectEventResource) GetType() resource.Type {
	return resource.AspectEventResource
}

// AspectEventResourceInput is an input for creating a new AspectEventResource
type AspectEventResourceInput struct {
	OrdID      string          `json:"ordId"`
	MinVersion *string         `json:"minVersion"`
	Subset     json.RawMessage `json:"subset"`
}

//// Validate validates Aspect fields
//func (a *AspectEventResourceInput) Validate() error {
//	return validation.ValidateStruct(a,
//		validation.Field(&a.Title, validation.Required, validation.Length(common.MinTitleLength, common.MaxTitleLength), validation.NewStringRule(common.NoNewLines, "title should not contain line breaks")),
//		validation.Field(&a.Description, validation.NilOrNotEmpty, validation.Length(common.MinDescriptionLength, common.MaxDescriptionLength)),
//		validation.Field(&a.Mandatory, validation.By(func(value interface{}) error {
//			return common.ValidateFieldMandatory(value, common.AspectMsg)
//		})),
//		validation.Field(&a.APIResources, validation.By(common.ValidateAspectAPIResources)),
//		validation.Field(&a.EventResources, validation.By(common.ValidateAspectEventResources)),
//	)
//}

// ToAspectEventResource converts AspectEventResourceInput to AspectEventResource
func (a *AspectEventResourceInput) ToAspectEventResource(id string, resourceType resource.Type, resourceID string, aspectID string) *AspectEventResource {
	if a == nil {
		return nil
	}

	aspectEventResource := &AspectEventResource{
		AspectID:   aspectID,
		OrdID:      a.OrdID,
		MinVersion: a.MinVersion,
		Subset:     a.Subset,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType.IsTenantIgnorable() {
		aspectEventResource.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		aspectEventResource.ApplicationID = &resourceID
	}

	return aspectEventResource
}
