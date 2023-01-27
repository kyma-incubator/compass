package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
)

// IsNotAssignedToAnyFormationOfType contains the name of the IsNotAssignedToAnyFormationOfType operator
const IsNotAssignedToAnyFormationOfType string = "IsNotAssignedToAnyFormationOfType"

// OperatorInput represent the input needed by the operators
type OperatorInput interface{}

// FormationConstraintInputByOperator represents a mapping between operator names and OperatorInputs
var FormationConstraintInputByOperator = map[string]OperatorInput{
	IsNotAssignedToAnyFormationOfType: &formationconstraint.IsNotAssignedToAnyFormationOfTypeInput{},
}

// JoinPointDetailsByLocation represents a mapping between JoinPointLocation and JoinPointDetails
var JoinPointDetailsByLocation = map[formationconstraint.JoinPointLocation]formationconstraint.JoinPointDetails{
	formationconstraint.PreAssign:                 &formationconstraint.AssignFormationOperationDetails{},
	formationconstraint.PostAssign:                &formationconstraint.AssignFormationOperationDetails{},
	formationconstraint.PreUnassign:               &formationconstraint.UnassignFormationOperationDetails{},
	formationconstraint.PostUnassign:              &formationconstraint.UnassignFormationOperationDetails{},
	formationconstraint.PreCreate:                 &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PostCreate:                &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PreDelete:                 &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PostDelete:                &formationconstraint.CRUDFormationOperationDetails{},
	formationconstraint.PreGenerateNotifications:  &formationconstraint.GenerateNotificationOperationDetails{},
	formationconstraint.PostGenerateNotifications: &formationconstraint.GenerateNotificationOperationDetails{},
}

// Validate validates FormationConstraintInput
func (i FormationConstraintInput) Validate() error {
	if err := validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required),
		validation.Field(&i.ConstraintType, validation.Required, validation.In(ConstraintTypePre, ConstraintTypePost)),
		validation.Field(&i.TargetOperation, validation.Required, validation.In(TargetOperationAssignFormation, TargetOperationUnassignFormation, TargetOperationCreateFormation, TargetOperationDeleteFormation, TargetOperationGenerateNotification)),
		validation.Field(&i.Operator, validation.Required),
		validation.Field(&i.ResourceType, validation.Required, validation.In(ResourceTypeApplication, ResourceTypeRuntime, ResourceTypeFormation, ResourceTypeTenant, ResourceTypeRuntimeContext)),
		validation.Field(&i.ResourceSubtype, validation.Required),
		validation.Field(&i.InputTemplate, validation.Required),
		validation.Field(&i.ConstraintScope, validation.Required, validation.In(ConstraintScopeFormationType, ConstraintScopeGlobal)),
	); err != nil {
		return err
	}

	input := FormationConstraintInputByOperator[i.Operator]
	if err := formationconstraint.ParseInputTemplate(i.InputTemplate, JoinPointDetailsByLocation[formationconstraint.JoinPointLocation{ConstraintType: model.FormationConstraintType(i.ConstraintType), OperationName: model.TargetOperation(i.TargetOperation)}], input); err != nil {
		return apperrors.NewInvalidDataError("failed to parse input template: %s", err)
	}

	return nil
}
