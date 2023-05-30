package formationconstraint

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

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

// DoNotGenerateFormationAssignmentNotificationInput input for DoNotGenerateFormationAssignmentNotification operator
type DoNotGenerateFormationAssignmentNotificationInput struct {
	ResourceType         model.ResourceType `json:"resource_type"`
	ResourceSubtype      string             `json:"resource_subtype"`
	ResourceID           string             `json:"resource_id"`
	SourceResourceType   model.ResourceType `json:"source_resource_type"`
	SourceResourceID     string             `json:"source_resource_id"`
	Tenant               string             `json:"tenant"`
	FormationTemplateID  string             `json:"formation_template_id"`
	ExceptSubtypes       []string           `json:"except_subtypes"`
	ExceptFormationTypes []string           `json:"except_formation_types"`
}

// DestinationCreatorInput input for DestinationCreator operator
type DestinationCreatorInput struct {
	Operation                  model.FormationOperation     `json:"operation"`
	ResourceType               model.ResourceType           `json:"resource_type"`
	ResourceSubtype            string                       `json:"resource_subtype"`
	FormationAssignment        *webhook.FormationAssignment `json:"assignment"`
	ReverseFormationAssignment *webhook.FormationAssignment `json:"reverseAssignment"`
	Location                   JoinPointLocation            `json:"joinPointLocation"`
	ShouldRewriteCredentials   bool                         `json:"shouldRewriteCredentials"`
}
