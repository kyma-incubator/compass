package formationconstraint

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
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
	Operation                              model.FormationOperation `json:"operation"`
	ResourceType                           model.ResourceType       `json:"resource_type"`
	ResourceSubtype                        string                   `json:"resource_subtype"`
	JoinPointDetailsFAMemoryAddress        uintptr                  `json:"details_formation_assignment_memory_address"`         // contains the memory address of the join point details' formation assignment in form of an integer
	JoinPointDetailsReverseFAMemoryAddress uintptr                  `json:"details_reverse_formation_assignment_memory_address"` // contains the memory address of the join point details' reverse formation assignment in form of an integer
	Location                               JoinPointLocation        `json:"join_point_location"`
	SkipSubaccountValidation               bool                     `json:"skip_subaccount_validation"`
}

// ConfigMutatorInput input for ConfigMutator operator
type ConfigMutatorInput struct {
	State                                  string                   `json:"state"`
	Configuration                          string                   `json:"configuration"`
	Operation                              model.FormationOperation `json:"operation"`
	ResourceType                           model.ResourceType       `json:"resource_type"`
	ResourceSubtype                        string                   `json:"resource_subtype"`
	JoinPointDetailsFAMemoryAddress        uintptr                  `json:"details_formation_assignment_memory_address"`         // contains the memory address of the join point details' formation assignment in form of an integer
	JoinPointDetailsReverseFAMemoryAddress uintptr                  `json:"details_reverse_formation_assignment_memory_address"` // contains the memory address of the join point details' reverse formation assignment in form of an integer
	Location                               JoinPointLocation        `json:"join_point_location"`
}
