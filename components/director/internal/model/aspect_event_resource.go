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
