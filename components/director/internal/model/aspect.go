package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Aspect represent structure for Aspect
type Aspect struct {
	ApplicationID                *string
	ApplicationTemplateVersionID *string
	IntegrationDependencyID      string
	Title                        string
	Description                  *string
	Mandatory                    *bool
	SupportMultipleProviders     *bool
	APIResources                 json.RawMessage
	*BaseEntity
}

// GetType returns Type aspect
func (*Aspect) GetType() resource.Type {
	return resource.Aspect
}

// AspectInput is an input for creating a new Aspect
type AspectInput struct {
	Title                    string                      `json:"title"`
	Description              *string                     `json:"description"`
	Mandatory                *bool                       `json:"mandatory"`
	SupportMultipleProviders *bool                       `json:"supportMultipleProviders"`
	APIResources             json.RawMessage             `json:"apiResources"`
	EventResources           []*AspectEventResourceInput `json:"eventResources"`
}

// ToAspect converts AspectInput to Aspect
func (a *AspectInput) ToAspect(id string, resourceType resource.Type, resourceID string, integrationDependencyID string) *Aspect {
	if a == nil {
		return nil
	}

	aspect := &Aspect{
		IntegrationDependencyID:  integrationDependencyID,
		Title:                    a.Title,
		Description:              a.Description,
		Mandatory:                a.Mandatory,
		SupportMultipleProviders: a.SupportMultipleProviders,
		APIResources:             a.APIResources,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	if resourceType.IsTenantIgnorable() {
		aspect.ApplicationTemplateVersionID = &resourceID
	} else if resourceType == resource.Application {
		aspect.ApplicationID = &resourceID
	}

	return aspect
}
