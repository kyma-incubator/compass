package model

import (
	"encoding/json"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/common"

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

// Validate validates Aspect Event Resource fields
func (a *AspectEventResourceInput) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.OrdID, validation.Required, validation.Length(common.MinOrdIDLength, common.MaxOrdIDLength), validation.Match(regexp.MustCompile(common.AspectEventResourceRegex))),
		validation.Field(&a.MinVersion, validation.NilOrNotEmpty, validation.Match(regexp.MustCompile(common.AspectResourcesMinVersionRegex))),
		validation.Field(&a.Subset, validation.By(common.ValidateAspectEventResourceSubset)),
	)
}

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
