package model

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/common"
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

// Validate missing godoc
func (a *AspectInput) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Name, validation.Required, validation.Length(common.MinTitleLength, common.MaxTitleLength), validation.NewStringRule(common.NoNewLines, "title should not contain line breaks")),
		validation.Field(&a.Description, validation.Length(common.MinDescriptionLength, common.MaxDescriptionLength)),
		validation.Field(&a.Mandatory, validation.By(func(value interface{}) error {
			return common.ValidateFieldMandatory(value, common.AspectMsg)
		})),
		validation.Field(&a.ApiResources, validation.By(func(value interface{}) error {
			return common.ValidateAspectApiResources(value)
		})),
		validation.Field(&a.EventResources, validation.By(func(value interface{}) error {
			return common.ValidateAspectEventResources(value)
		})))
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
