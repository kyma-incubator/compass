package formationconstraint

import "github.com/kyma-incubator/compass/components/director/internal/model"

// IsNotAssignedToAnyFormationOfTypeInput input for IsNotAssignedToAnyFormationOfType operator
type IsNotAssignedToAnyFormationOfTypeInput struct {
	FormationTemplateID string             `json:"formation_template_id"`
	ResourceType        model.ResourceType `json:"resource_type"`
	ResourceSubtype     string             `json:"resource_subtype"`
	ResourceID          string             `json:"resource_id"`
	Tenant              string             `json:"tenant"`
	ExceptSystemTypes   []string           `json:"exceptSystemTypes"`
}

// DoesNotContainResourceOfSubtypeInput input for DoesNotContainResourceOfSubtype operator
type DoesNotContainResourceOfSubtypeInput struct {
	FormationName   string             `json:"formation_name"`
	ResourceType    model.ResourceType `json:"resource_type"`
	ResourceSubtype string             `json:"resource_subtype"`
	ResourceID      string             `json:"resource_id"`
	Tenant          string             `json:"tenant"`
}
