package model

import (
	"encoding/json"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// APIModelSelectorType is the type of the API APIModelSelector.
type APIModelSelectorType string

const (
	// APIModelSelectorType for odata selector tyor.
	APIModelSelectorTypeODATA APIModelSelectorType = "odata"
	// APIModelSelectorTypeJSONPointer for json pointer selector tyor.
	APIModelSelectorTypeJSONPointer APIModelSelectorType = "json-pointer"
)

// EntityTypeMapping missing godoc
type EntityTypeMapping struct {
	APIDefinitionID   *string
	EventDefinitionID *string
	APIModelSelectors json.RawMessage
	EntityTypeTargets json.RawMessage
	*BaseEntity
}

// GetType missing godoc
func (*EntityTypeMapping) GetType() resource.Type {
	return resource.EntityTypeMapping
}

// EntityTypePage missing godoc
type EntityTypeMappingPage struct {
	Data       []*EntityTypeMapping
	PageInfo   *pagination.Page
	TotalCount int
}

// IsPageable missing godoc
func (EntityTypeMappingPage) IsPageable() {}

// EntityTypeMappingInput missing godoc
type EntityTypeMappingInput struct {
	APIModelSelectors json.RawMessage `json:"apiModelSelectors,omitempty"`
	EntityTypeTargets json.RawMessage `json:"entityTypeTargets,omitempty"`
}

// ToEntityTypeMapping missing godoc
func (i *EntityTypeMappingInput) ToEntityTypeMapping(id string, resourceType resource.Type, resourceID string) *EntityTypeMapping {
	if i == nil {
		return nil
	}

	entityTypeMapping := &EntityTypeMapping{
		BaseEntity: &BaseEntity{
			ID:    id,
			Ready: true,
		},
		APIModelSelectors: i.APIModelSelectors,
		EntityTypeTargets: i.EntityTypeTargets,
	}

	switch resourceType {
	case resource.API:
		entityTypeMapping.APIDefinitionID = &resourceID
	case resource.EventDefinition:
		entityTypeMapping.EventDefinitionID = &resourceID
	}

	return entityTypeMapping
}

// SetFromUpdateInput missing godoc
func (entityTypeMapping *EntityTypeMapping) SetFromUpdateInput(update EntityTypeMappingInput) {
	entityTypeMapping.APIModelSelectors = update.APIModelSelectors
	entityTypeMapping.EntityTypeTargets = update.EntityTypeTargets
}

// APIModelSelector represents the API Model Selector entity
type APIModelSelector struct {
	Type          APIModelSelectorType `json:"type"`
	EntitySetName *string              `json:"entitySetName,omitempty"`
	JsonPointer   *string              `json:"jsonPointer,omitempty"`
}

// Validate missing godoc
func (apiModelSelector *APIModelSelector) Validate() error {
	return validation.ValidateStruct(apiModelSelector,
		validation.Field(&apiModelSelector.Type, validation.Required, validation.In(APIModelSelectorTypeODATA, APIModelSelectorTypeJSONPointer)),
		validation.Field(&apiModelSelector.EntitySetName,
			validation.When(apiModelSelector.Type == APIModelSelectorTypeODATA, validation.Required),
			validation.When(apiModelSelector.Type == APIModelSelectorTypeJSONPointer, validation.Nil),
		),
		validation.Field(&apiModelSelector.JsonPointer,
			validation.When(apiModelSelector.Type == APIModelSelectorTypeJSONPointer, validation.Required),
			validation.When(apiModelSelector.Type == APIModelSelectorTypeODATA, validation.Nil),
		),
	)
}

// EntityTypeTarget represents the Entity Type Target entity
type EntityTypeTarget struct {
	OrdId         *string `json:"ordId,omitempty"`
	CorrelationId *string `json:"correlationId,omitempty"`
}

// Validate missing godoc
func (entityTypeTarget *EntityTypeTarget) Validate() error {
	return validation.ValidateStruct(entityTypeTarget,
		validation.Field(&entityTypeTarget.OrdId,
			validation.When(entityTypeTarget.CorrelationId != nil, validation.Nil),
		),
		validation.Field(&entityTypeTarget.CorrelationId,
			validation.When(entityTypeTarget.OrdId != nil, validation.Nil),
		),
	)
}
