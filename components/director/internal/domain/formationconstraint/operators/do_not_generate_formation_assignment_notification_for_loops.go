package operators

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// DoNotGenerateFormationAssignmentNotificationForLoopsOperator represents the DoNotGenerateFormationAssignmentNotificationForLoops operator
	DoNotGenerateFormationAssignmentNotificationForLoopsOperator = "DoNotGenerateFormationAssignmentNotificationForLoops"
)

// NewDoNotGenerateFormationAssignmentNotificationForLoopsInput is input constructor for DoNotGenerateFormationAssignmentNotificationForLoopsOperator operator. It returns empty OperatorInput
func NewDoNotGenerateFormationAssignmentNotificationForLoopsInput() OperatorInput {
	return &formationconstraint.DoNotGenerateFormationAssignmentNotificationInput{}
}

// DoNotGenerateFormationAssignmentNotificationForLoops is a constraint operator. It skips the generation of formation assignment notifications for loops
func (e *ConstraintEngine) DoNotGenerateFormationAssignmentNotificationForLoops(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DoNotGenerateFormationAssignmentNotificationForLoopsOperator)

	i, ok := input.(*formationconstraint.DoNotGenerateFormationAssignmentNotificationInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator %q", DoNotGenerateFormationAssignmentNotificationForLoopsOperator)
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", DoNotGenerateFormationAssignmentNotificationForLoopsOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

	if i.ResourceID != i.SourceResourceID || i.ResourceType != i.SourceResourceType {
		return true, nil
	}

	return e.DoNotGenerateFormationAssignmentNotification(ctx, input)
}
