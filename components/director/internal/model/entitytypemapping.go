package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
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
