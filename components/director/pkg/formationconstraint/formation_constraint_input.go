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

// ContainsScenarioGroupsInput input for ContainsScenarioGroups operator
type ContainsScenarioGroupsInput struct {
	ResourceType           model.ResourceType `json:"resource_type"`
	ResourceSubtype        string             `json:"resource_subtype"`
	ResourceID             string             `json:"resource_id"`
	Tenant                 string             `json:"tenant"`
	RequiredScenarioGroups []string           `json:"requiredScenarioGroups"`
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
	Operation                             model.FormationOperation `json:"operation"`
	ResourceType                          model.ResourceType       `json:"resource_type"`
	ResourceSubtype                       string                   `json:"resource_subtype"`
	NotificationStatusReportMemoryAddress uintptr                  `json:"notification_status_report_memory_address"`
	FAMemoryAddress                       uintptr                  `json:"formation_assignment_memory_address"`         // contains the memory address of the join point details' formation assignment in form of an integer
	ReverseFAMemoryAddress                uintptr                  `json:"reverse_formation_assignment_memory_address"` // contains the memory address of the join point details' reverse formation assignment in form of an integer
	Location                              JoinPointLocation        `json:"join_point_location"`
	SkipSubaccountValidation              bool                     `json:"skip_subaccount_validation"`
	UseCertSvcKeystoreForSAML             bool                     `json:"use_cert_svc_keystore_for_saml"`
}

// ConfigMutatorInput input for ConfigMutator operator
type ConfigMutatorInput struct {
	State                                 *string                  `json:"state"`
	Tenant                                string                   `json:"tenant"`
	OnlyForSourceSubtypes                 []string                 `json:"only_for_source_subtypes"`
	ModifiedConfiguration                 *string                  `json:"modified_configuration"`
	Operation                             model.FormationOperation `json:"operation"`
	ResourceType                          model.ResourceType       `json:"resource_type"`
	ResourceSubtype                       string                   `json:"resource_subtype"`
	NotificationStatusReportMemoryAddress uintptr                  `json:"notification_status_report_memory_address"`
	SourceResourceType                    model.ResourceType       `json:"source_resource_type"`
	SourceResourceID                      string                   `json:"source_resource_id"`
	Location                              JoinPointLocation        `json:"join_point_location"`
}

// RedirectNotificationInput is an input for RedirectNotification operator
type RedirectNotificationInput struct {
	ShouldRedirect       bool                     `json:"should_redirect"`
	URLTemplate          string                   `json:"url_template"`
	URL                  string                   `json:"url"`
	WebhookMemoryAddress uintptr                  `json:"webhook_memory_address"` // contains the memory address of the join point details' webhook in form of an integer
	Operation            model.FormationOperation `json:"operation"`
	ResourceType         model.ResourceType       `json:"resource_type"`
	ResourceSubtype      string                   `json:"resource_subtype"`
	Location             JoinPointLocation        `json:"join_point_location"`
}

// AsynchronousFlowControlOperatorInput is an input for AsynchronousFlowControlOperator operator
type AsynchronousFlowControlOperatorInput struct {
	RedirectNotificationInput
	FailOnSyncParticipants                bool    `json:"fail_on_sync_participants"`
	FailOnNonBTPParticipants              bool    `json:"fail_on_non_btp_participants"`
	NotificationStatusReportMemoryAddress uintptr `json:"notification_status_report_memory_address"`
	FAMemoryAddress                       uintptr `json:"formation_assignment_memory_address"`         // contains the memory address of the join point details' formation assignment in form of an integer
	ReverseFAMemoryAddress                uintptr `json:"reverse_formation_assignment_memory_address"` // contains the memory address of the join point details' reverse formation assignment in form of an integer
}

// JSONSchemaValidatorOperatorInput is an input for JSONSchemaValidatorOperator operator
type JSONSchemaValidatorOperatorInput struct {
	ResourceType          model.ResourceType `json:"resource_type"`
	ResourceSubtype       string             `json:"resource_subtype"`
	ResourceID            string             `json:"resource_id"`
	SourceResourceType    model.ResourceType `json:"source_resource_type"`
	SourceResourceID      string             `json:"source_resource_id"`
	TenantID              string             `json:"tenant_id"`
	FormationTemplateName string             `json:"formation_template_name"`
	FAMemoryAddress       uintptr            `json:"formation_assignment_memory_address"` // contains the memory address of the join point details' formation assignment in form of an integer
	ExceptSubtypes        []string           `json:"except_subtypes"`
	ExceptFormationTypes  []string           `json:"except_formation_types"`
	OnlyForSourceSubtypes []string           `json:"only_for_source_subtypes"`
	JSONSchema            string             `json:"json_schema"`
}

// WithExceptFormationTypes sets the ExceptFormationTypes field of the JSONSchemaValidatorOperatorInput
func (i *JSONSchemaValidatorOperatorInput) WithExceptFormationTypes(formationTypes []string) *JSONSchemaValidatorOperatorInput {
	i.ExceptFormationTypes = formationTypes
	return i
}

// WithExceptSubtypes sets the ExceptSubtypes field of the JSONSchemaValidatorOperatorInput
func (i *JSONSchemaValidatorOperatorInput) WithExceptSubtypes(subtypes []string) *JSONSchemaValidatorOperatorInput {
	i.ExceptSubtypes = subtypes
	return i
}

// WithOnlyForSourceSubtypes sets the OnlyForSourceSubtypes field of the JSONSchemaValidatorOperatorInput
func (i *JSONSchemaValidatorOperatorInput) WithOnlyForSourceSubtypes(sourceSubtypes []string) *JSONSchemaValidatorOperatorInput {
	i.OnlyForSourceSubtypes = sourceSubtypes
	return i
}
