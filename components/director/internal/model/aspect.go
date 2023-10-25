package model

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Aspect missing godoc
type Aspect struct {
	IntegrationDependencyID  string
	Name                     string
	Description              *string
	Mandatory                bool
	SupportMultipleProviders *bool
	ApiResources             json.RawMessage
	EventResources           json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*Aspect) GetType() resource.Type {
	return resource.Aspect
}

// AspectInput missing godoc
type AspectInput struct {
	Name                     string
	Description              *string
	Mandatory                bool
	SupportMultipleProviders *bool
	ApiResources             json.RawMessage
	EventResources           json.RawMessage
}

// ToAspect missing godoc
func (a *AspectInput) ToAspect(id string, integrationDependencyId string) *Aspect {
	if a == nil {
		return nil
	}

	aspect := &Aspect{
		IntegrationDependencyID:  integrationDependencyId,
		Name:                     a.Name,
		Description:              a.Description,
		Mandatory:                a.Mandatory,
		SupportMultipleProviders: a.SupportMultipleProviders,
		ApiResources:             a.ApiResources,
		EventResources:           a.EventResources,
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
	}

	return aspect
}
