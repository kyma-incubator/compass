package model

import (
	"encoding/json"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/common"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Aspect represent structure for Aspect
type Aspect struct {
	IntegrationDependencyID  string
	Title                    string
	Description              *string
	Mandatory                *bool
	SupportMultipleProviders *bool
	ApiResources             json.RawMessage
	EventResources           json.RawMessage
	*BaseEntity
}

// GetType returns Type aspect
func (*Aspect) GetType() resource.Type {
	return resource.Aspect
}

// AspectInput is an input for creating a new Aspect
type AspectInput struct {
	Title                    string          `json:"title"`
	Description              *string         `json:"description"`
	Mandatory                *bool           `json:"mandatory"`
	SupportMultipleProviders *bool           `json:"supportMultipleProviders"`
	ApiResources             json.RawMessage `json:"apiResources"`
	EventResources           json.RawMessage `json:"eventResources"`
}

// Validate validates Aspect fields
func (a *AspectInput) Validate() error {
	return validation.ValidateStruct(a,
		validation.Field(&a.Title, validation.Required, validation.Length(common.MinTitleLength, common.MaxTitleLength), validation.NewStringRule(common.NoNewLines, "title should not contain line breaks")),
		validation.Field(&a.Description, validation.NilOrNotEmpty, validation.Length(common.MinDescriptionLength, common.MaxDescriptionLength)),
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

// ToAspect converts AspectInput to Aspect
func (a *AspectInput) ToAspect(id string, integrationDependencyId string) *Aspect {
	if a == nil {
		return nil
	}

	aspect := &Aspect{
		IntegrationDependencyID:  integrationDependencyId,
		Title:                    a.Title,
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
