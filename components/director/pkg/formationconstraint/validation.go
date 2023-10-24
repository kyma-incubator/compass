package formationconstraint

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

const (
	// IsNotAssignedToAnyFormationOfType contains the name of the IsNotAssignedToAnyFormationOfType operator
	IsNotAssignedToAnyFormationOfType string = "IsNotAssignedToAnyFormationOfType"
	// DoesNotContainResourceOfSubtype contains the name of the DoesNotContainResourceOfSubtype operator
	DoesNotContainResourceOfSubtype = "DoesNotContainResourceOfSubtype"
	// DoNotGenerateFormationAssignmentNotificationOperator represents the DoNotGenerateFormationAssignmentNotification operator
	DoNotGenerateFormationAssignmentNotificationOperator = "DoNotGenerateFormationAssignmentNotification"
	// DoNotGenerateFormationAssignmentNotificationForLoopsOperator represents the DoNotGenerateFormationAssignmentNotificationForLoops operator
	DoNotGenerateFormationAssignmentNotificationForLoopsOperator = "DoNotGenerateFormationAssignmentNotificationForLoops"
	// DestinationCreator contains the name of the DestinationCreator operator
	DestinationCreator = "DestinationCreator"
	// ConfigMutatorOperator contains the name of the ConfigMutator
	ConfigMutatorOperator = "ConfigMutator"
	// RedirectNotificationOperator contains the name of the RedirectNotificationOperator
	RedirectNotificationOperator = "RedirectNotification"
)

// OperatorInput represent the input needed by the operators
type OperatorInput interface{}

// FormationConstraintInputByOperator represents a mapping between operator names and OperatorInputs
var FormationConstraintInputByOperator = map[string]OperatorInput{
	IsNotAssignedToAnyFormationOfType:                            &IsNotAssignedToAnyFormationOfTypeInput{},
	DoesNotContainResourceOfSubtype:                              &DoesNotContainResourceOfSubtypeInput{},
	DoNotGenerateFormationAssignmentNotificationOperator:         &DoNotGenerateFormationAssignmentNotificationInput{},
	DoNotGenerateFormationAssignmentNotificationForLoopsOperator: &DoNotGenerateFormationAssignmentNotificationInput{},
	DestinationCreator:                                           &DestinationCreatorInput{},
	ConfigMutatorOperator:                                        &ConfigMutatorInput{},
	RedirectNotificationOperator:                                 &RedirectNotificationInput{},
}

// JoinPointDetailsByLocation represents a mapping between JoinPointLocation and JoinPointDetails
var JoinPointDetailsByLocation = map[JoinPointLocation]JoinPointDetails{
	PreAssign:    &AssignFormationOperationDetails{},
	PostAssign:   &AssignFormationOperationDetails{},
	PreUnassign:  &UnassignFormationOperationDetails{},
	PostUnassign: &UnassignFormationOperationDetails{},
	PreCreate:    &CRUDFormationOperationDetails{},
	PostCreate:   &CRUDFormationOperationDetails{},
	PreDelete:    &CRUDFormationOperationDetails{},
	PostDelete:   &CRUDFormationOperationDetails{},
	PreGenerateFormationAssignmentNotifications:  emptyGenerateFormationAssignmentNotificationOperationDetails(),
	PostGenerateFormationAssignmentNotifications: emptyGenerateFormationAssignmentNotificationOperationDetails(),
	PreGenerateFormationNotifications:            emptyGenerateFormationNotificationOperationDetails(),
	PostGenerateFormationNotifications:           emptyGenerateFormationNotificationOperationDetails(),
	PreSendNotification:                          emptySendNotificationOperationDetails(),
	PostSendNotification:                         emptySendNotificationOperationDetails(),
	PreNotificationStatusReturned:                emptyNotificationStatusReturnedOperationDetails(),
	PostNotificationStatusReturned:               emptyNotificationStatusReturnedOperationDetails(),
}

// FormationConstraintInputWrapper is a wrapper structure of the graphql.FormationConstraintInput,
// so we could extract the template validation in a separate package due to import cycle
type FormationConstraintInputWrapper struct {
	*graphql.FormationConstraintInput
}

// NewFormationConstraintInputWrapper creates a new FormationConstraintInputWrapper object
func NewFormationConstraintInputWrapper(input *graphql.FormationConstraintInput) *FormationConstraintInputWrapper {
	return &FormationConstraintInputWrapper{input}
}

// Validate validates the input template as part of the graphql.FormationConstraintInput
func (i FormationConstraintInputWrapper) Validate() error {
	if i.ConstraintType != graphql.ConstraintTypeUI {
		input := FormationConstraintInputByOperator[i.Operator]
		if err := ParseInputTemplate(i.InputTemplate, JoinPointDetailsByLocation[JoinPointLocation{ConstraintType: model.FormationConstraintType(i.ConstraintType), OperationName: model.TargetOperation(i.TargetOperation)}], input); err != nil {
			return apperrors.NewInvalidDataError("failed to parse input template: %s", err)
		}
	}

	return nil
}

func emptyGenerateFormationAssignmentNotificationOperationDetails() *GenerateFormationAssignmentNotificationOperationDetails {
	return &GenerateFormationAssignmentNotificationOperationDetails{
		CustomerTenantContext: &webhook.CustomerTenantContext{},
		Formation: &model.Formation{
			Error: json.RawMessage{},
		},
		ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		Application: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		Runtime: &webhook.RuntimeWithLabels{
			Runtime: &model.Runtime{},
			Labels:  map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		RuntimeContext: &webhook.RuntimeContextWithLabels{
			RuntimeContext: &model.RuntimeContext{},
			Labels:         map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		Assignment:        &webhook.FormationAssignment{},
		ReverseAssignment: &webhook.FormationAssignment{},
		SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		SourceApplication: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
		TargetApplication: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
			Tenant: &webhook.TenantWithLabels{
				BusinessTenantMapping: &model.BusinessTenantMapping{},
				Labels:                map[string]string{},
			},
		},
	}
}

func emptyGenerateFormationNotificationOperationDetails() *GenerateFormationNotificationOperationDetails {
	return &GenerateFormationNotificationOperationDetails{
		CustomerTenantContext: &webhook.CustomerTenantContext{},
	}
}

func emptySendNotificationOperationDetails() *SendNotificationOperationDetails {
	return &SendNotificationOperationDetails{
		Location: JoinPointLocation{},
		Webhook: &graphql.Webhook{
			CreatedAt: &graphql.Timestamp{},
		},
		TemplateInput: nil,
		FormationAssignment: &model.FormationAssignment{
			Value: json.RawMessage("\"\""),
			Error: json.RawMessage("\"\""),
		},
		ReverseFormationAssignment: &model.FormationAssignment{
			Value: json.RawMessage("\"\""),
			Error: json.RawMessage("\"\""),
		},
		Formation: &model.Formation{
			Error: json.RawMessage("\"\""),
		},
	}
}

func emptyNotificationStatusReturnedOperationDetails() *NotificationStatusReturnedOperationDetails {
	return &NotificationStatusReturnedOperationDetails{
		Location: JoinPointLocation{},
		FormationAssignment: &model.FormationAssignment{
			Value: json.RawMessage("\"\""),
			Error: json.RawMessage("\"\""),
		},
		FormationAssignmentTemplateInput: &webhook.ApplicationTenantMappingInput{
			Operation: model.AssignFormation,
			Formation: &model.Formation{},
			SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixLabels(),
			},
			SourceApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(),
				Labels:      fixLabels(),
			},
			TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
				ApplicationTemplate: fixApplicationTemplateModel(),
				Labels:              fixLabels(),
			},
			TargetApplication: &webhook.ApplicationWithLabels{
				Application: fixApplicationModel(),
				Labels:      fixLabels(),
			},
			Assignment:        nil,
			ReverseAssignment: nil,
		},
		ReverseFormationAssignment: &model.FormationAssignment{
			Value: json.RawMessage("\"\""),
			Error: json.RawMessage("\"\""),
		},
		Formation: &model.Formation{
			Error: json.RawMessage("\"\""),
		},
		FormationTemplate: &model.FormationTemplate{},
	}
}

func fixApplicationModel() *model.Application {
	return &model.Application{
		ProviderName:          str.Ptr(""),
		ApplicationTemplateID: str.Ptr(""),
		Description:           str.Ptr(""),
		Status: &model.ApplicationStatus{
			Condition: model.ApplicationStatusConditionInitial,
			Timestamp: time.Time{},
		},
		HealthCheckURL:      str.Ptr(""),
		BaseURL:             str.Ptr(""),
		OrdLabels:           json.RawMessage("[]"),
		CorrelationIDs:      json.RawMessage("[]"),
		SystemStatus:        str.Ptr(""),
		DocumentationLabels: json.RawMessage("[]"),
		BaseEntity: &model.BaseEntity{
			Ready:     true,
			Error:     nil,
			CreatedAt: &time.Time{},
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
		},
	}
}

func fixLabels() map[string]string {
	return map[string]string{"foo": "bar"}
}

func fixApplicationTemplateModel() *model.ApplicationTemplate {
	return &model.ApplicationTemplate{
		Description:          str.Ptr(""),
		ApplicationInputJSON: `{}`,
	}
}
