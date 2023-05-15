package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

// IsNotAssignedToAnyFormationOfType contains the name of the IsNotAssignedToAnyFormationOfType operator
const IsNotAssignedToAnyFormationOfType string = "IsNotAssignedToAnyFormationOfType"

// DoesNotContainResourceOfSubtype contains the name of the DoesNotContainResourceOfSubtype operator
const DoesNotContainResourceOfSubtype = "DoesNotContainResourceOfSubtype"

// DoNotGenerateFormationAssignmentNotificationOperator represents the DoNotGenerateFormationAssignmentNotification operator
const DoNotGenerateFormationAssignmentNotificationOperator = "DoNotGenerateFormationAssignmentNotification"

// OperatorInput represent the input needed by the operators
type OperatorInput interface{}

// FormationConstraintInputByOperator represents a mapping between operator names and OperatorInputs
var FormationConstraintInputByOperator = map[string]OperatorInput{
	IsNotAssignedToAnyFormationOfType:                    &formationconstraint.IsNotAssignedToAnyFormationOfTypeInput{},
	DoesNotContainResourceOfSubtype:                      &formationconstraint.DoesNotContainResourceOfSubtypeInput{},
	DoNotGenerateFormationAssignmentNotificationOperator: &formationconstraint.DoNotGenerateFormationAssignmentNotificationInput{},
}

// JoinPointDetailsByLocation represents a mapping between JoinPointLocation and JoinPointDetails
var JoinPointDetailsByLocation = map[formationconstraint.JoinPointLocation]formationconstraint.JoinPointDetails{
	formationconstraint.PreAssign:                                    &formationconstraint.AssignFormationOperationDetails{},
	formationconstraint.PostAssign:                                   &formationconstraint.AssignFormationOperationDetails{},
	formationconstraint.PreUnassign:                                  &formationconstraint.UnassignFormationOperationDetails{},
	formationconstraint.PostUnassign:                                 &formationconstraint.UnassignFormationOperationDetails{},
	formationconstraint.PreCreate:                                    &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PostCreate:                                   &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PreDelete:                                    &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PostDelete:                                   &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PreGenerateFormationAssignmentNotifications:  emptyGenerateFormationAssignmentNotificationOperationDetails(),
	formationconstraint.PostGenerateFormationAssignmentNotifications: emptyGenerateFormationAssignmentNotificationOperationDetails(),
	formationconstraint.PreGenerateFormationNotifications:            emptyGenerateFormationNotificationOperationDetails(),
	formationconstraint.PostGenerateFormationNotifications:           emptyGenerateFormationNotificationOperationDetails(),
	formationconstraint.PreSendNotification:                          &formationconstraint.SendNotificationOperationDetails{},
	formationconstraint.PostSendNotification:                         &formationconstraint.SendNotificationOperationDetails{},
	formationconstraint.PreNotificationStatusReturned:                &formationconstraint.NotificationStatusReturnedOperationDetails{},
	formationconstraint.PostNotificationStatusReturned:               &formationconstraint.NotificationStatusReturnedOperationDetails{},
}

// Validate validates FormationConstraintInput
func (i FormationConstraintInput) Validate() error {
	if err := validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required),
		validation.Field(&i.ConstraintType, validation.Required, validation.In(ConstraintTypePre, ConstraintTypePost, ConstraintTypeUI)),
		validation.Field(&i.TargetOperation, validation.Required, validation.In(TargetOperationAssignFormation, TargetOperationUnassignFormation, TargetOperationCreateFormation, TargetOperationDeleteFormation, TargetOperationGenerateFormationAssignmentNotification, TargetOperationGenerateFormationNotification, TargetOperationLoadFormations, TargetOperationSelectSystemsForFormation, TargetOperationSendNotification, TargetOperationNotificationStatusReturned)),
		validation.Field(&i.Operator, validation.Required),
		validation.Field(&i.ResourceType, validation.Required, validation.In(ResourceTypeApplication, ResourceTypeRuntime, ResourceTypeFormation, ResourceTypeTenant, ResourceTypeRuntimeContext)),
		validation.Field(&i.ResourceSubtype, validation.Required),
		validation.Field(&i.InputTemplate, validation.Required),
		validation.Field(&i.ConstraintScope, validation.Required, validation.In(ConstraintScopeFormationType, ConstraintScopeGlobal)),
	); err != nil {
		return err
	}

	if i.ConstraintType != ConstraintTypeUI {
		input := FormationConstraintInputByOperator[i.Operator]
		if err := formationconstraint.ParseInputTemplate(i.InputTemplate, JoinPointDetailsByLocation[formationconstraint.JoinPointLocation{ConstraintType: model.FormationConstraintType(i.ConstraintType), OperationName: model.TargetOperation(i.TargetOperation)}], input); err != nil {
			return apperrors.NewInvalidDataError("failed to parse input template: %s", err)
		}
	}

	return nil
}

func emptyGenerateFormationAssignmentNotificationOperationDetails() *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails {
	return &formationconstraint.GenerateFormationAssignmentNotificationOperationDetails{
		CustomerTenantContext: &webhook.CustomerTenantContext{},
		ApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
		},
		Application: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
		},
		Runtime: &webhook.RuntimeWithLabels{
			Runtime: &model.Runtime{},
			Labels:  map[string]string{},
		},
		RuntimeContext: &webhook.RuntimeContextWithLabels{
			RuntimeContext: &model.RuntimeContext{},
			Labels:         map[string]string{},
		},
		Assignment: &webhook.FormationAssignment{
			Value: "\"\"",
		},
		ReverseAssignment: &webhook.FormationAssignment{
			Value: "\"\"",
		},
		SourceApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
		},
		SourceApplication: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
		},
		TargetApplicationTemplate: &webhook.ApplicationTemplateWithLabels{
			ApplicationTemplate: &model.ApplicationTemplate{},
			Labels:              map[string]string{},
		},
		TargetApplication: &webhook.ApplicationWithLabels{
			Application: &model.Application{
				BaseEntity: &model.BaseEntity{},
			},
			Labels: map[string]string{},
		},
	}
}

func emptyGenerateFormationNotificationOperationDetails() *formationconstraint.GenerateFormationNotificationOperationDetails {
	return &formationconstraint.GenerateFormationNotificationOperationDetails{
		CustomerTenantContext: &webhook.CustomerTenantContext{},
	}
}
