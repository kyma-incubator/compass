package graphql

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/formation_constraint_input"
)

// IsNotAssignedToAnyFormationOfType contains the name of the IsNotAssignedToAnyFormationOfType operator
const IsNotAssignedToAnyFormationOfType string = "IsNotAssignedToAnyFormationOfType"

// OperatorInput represent the input needed by the operators
type OperatorInput interface{}

// FormationConstraintInputByOperator represents a mapping between operator names and OperatorInputs
var FormationConstraintInputByOperator = map[string]OperatorInput{
	IsNotAssignedToAnyFormationOfType: &formation_constraint_input.IsNotAssignedToAnyFormationOfTypeInput{},
}

type templateSource struct {
	FormationTemplateID string
	ResourceType        ResourceType
	ResourceSubtype     string
	ResourceID          string
	TenantID            string
}

var validationSource = templateSource{
	FormationTemplateID: "",
	ResourceType:        "",
	ResourceSubtype:     "",
	ResourceID:          "",
	TenantID:            "",
}

// Validate validates FormationConstraintInput
func (i FormationConstraintInput) Validate() error {
	if err := validation.ValidateStruct(&i,
		validation.Field(&i.Name, validation.Required),
		validation.Field(&i.ConstraintType, validation.Required, validation.In(ConstraintTypePre, ConstraintTypePost)),
		validation.Field(&i.TargetOperation, validation.Required, validation.In(TargetOperationAssignFormation, TargetOperationUnassignedFormation, TargetOperationCreateFormation, TargetOperationDeleteFormation, TargetOperationGenerateNotification)),
		validation.Field(&i.Operator, validation.Required),
		validation.Field(&i.ResourceType, validation.Required, validation.In(ResourceTypeApplication, ResourceTypeRuntime, ResourceTypeFormation, ResourceTypeTenant, ResourceTypeRuntimeContext)),
		validation.Field(&i.ResourceSubtype, validation.Required),
		validation.Field(&i.OperatorScope, validation.Required, validation.In(OperatorScopeGlobal, OperatorScopeTenant)),
		validation.Field(&i.InputTemplate, validation.Required),
		validation.Field(&i.ConstraintScope, validation.Required, validation.In(ConstraintScopeFormationType, ConstraintScopeGlobal)),
		validation.Field(&i.FormationTemplateID, validation.Required),
	); err != nil {
		return err
	}

	input := FormationConstraintInputByOperator[i.Operator]
	if err := formation_constraint_input.ParseInputTemplate(i.InputTemplate, validationSource, input); err != nil {
		return apperrors.NewInvalidDataError("failed to parse input template: %s", err)
	}

	return nil
}
