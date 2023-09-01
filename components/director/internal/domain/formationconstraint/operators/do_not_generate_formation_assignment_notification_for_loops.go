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

	if len(i.ExceptFormationTypes) > 0 {
		formationTemplate, err := e.formationTemplateRepo.Get(ctx, i.FormationTemplateID)
		if err != nil {
			return false, errors.Wrapf(err, "while getting formation template with ID %q", i.FormationTemplateID)
		}
		for _, exceptFormationType := range i.ExceptFormationTypes {
			if formationTemplate.Name == exceptFormationType {
				return true, nil
			}
		}
	}

	if len(i.ExceptSubtypes) == 0 {
		log.C(ctx).Infof("Skipping notifications to target resource of type: %q, subtype: %q and ID: %q for source resource of type: %q and ID: %q", i.ResourceType, i.ResourceSubtype, i.ResourceID, i.SourceResourceType, i.SourceResourceID)
		return false, nil
	}

	sourceSubType, err := e.getObjectSubtype(ctx, i.Tenant, i.SourceResourceType, i.SourceResourceID)
	if err != nil {
		return false, errors.Wrapf(err, "while getting subtype of resource with type: %q and id: %q", i.SourceResourceType, i.SourceResourceID)
	}

	for _, exceptSubtype := range i.ExceptSubtypes {
		if sourceSubType == exceptSubtype {
			return true, nil
		}
	}

	log.C(ctx).Infof("Skipping notifications to target resource of type: %q, subtype: %q and ID: %q for source resource of type: %q, subtype: %q, and ID: %q", i.ResourceType, i.ResourceSubtype, i.ResourceID, i.SourceResourceType, sourceSubType, i.SourceResourceID)
	return false, nil
}
