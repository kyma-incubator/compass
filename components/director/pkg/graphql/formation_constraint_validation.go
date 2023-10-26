package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

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

	return nil
}