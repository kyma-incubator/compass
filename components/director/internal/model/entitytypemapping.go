package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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
	Type          string  `json:"type"`
	EntitySetName *string `json:"entitySetName,omitempty"`
	JsonPointer   *string `json:"jsonPointer,omitempty"`
}

// EntityTypeTarget represents the Entity Type Target entity
type EntityTypeTarget struct {
	OrdId         *string `json:"ordId,omitempty"`
	CorrelationId *string `json:"correlationId,omitempty"`
}
