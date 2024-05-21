package operators

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// IsNotAssignedToAnyFormationOfTypeOperator represents the IsNotAssignedToAnyFormationOfType operator
	IsNotAssignedToAnyFormationOfTypeOperator = "IsNotAssignedToAnyFormationOfType"
)

// NewIsNotAssignedToAnyFormationOfTypeInput is input constructor for IsNotAssignedToAnyFormationOfType operator. It returns empty OperatorInput.
func NewIsNotAssignedToAnyFormationOfTypeInput() OperatorInput {
	return &formationconstraint.IsNotAssignedToAnyFormationOfTypeInput{}
}

// IsNotAssignedToAnyFormationOfType is a constraint operator. It checks if the resource from the OperatorInput is already part of formation of the type that the operator is associated with
func (e *ConstraintEngine) IsNotAssignedToAnyFormationOfType(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %s", IsNotAssignedToAnyFormationOfTypeOperator)

	i, ok := input.(*formationconstraint.IsNotAssignedToAnyFormationOfTypeInput)
	if !ok {
		return false, errors.New("Incompatible input")
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", IsNotAssignedToAnyFormationOfTypeOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

	var assignedFormations []string
	var err error
	switch i.ResourceType {
	case model.TenantResourceType:
		tenantInternalID, err := e.tenantSvc.GetInternalTenant(ctx, i.ResourceID)
		if err != nil {
			return false, err
		}

		assignments, err := e.asaSvc.ListForTargetTenant(ctx, tenantInternalID)
		if err != nil {
			return false, err
		}

		assignedFormations = make([]string, 0, len(assignments))
		for _, a := range assignments {
			assignedFormations = append(assignedFormations, a.ScenarioName)
		}
	case model.ApplicationResourceType:
		assignedFormations, err = e.listFormationsForObject(ctx, i.ResourceID)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	isAllowedToParticipateInFormationsOfType, err := e.isAllowedToParticipateInFormationsOfType(ctx, assignedFormations, i.Tenant, i.FormationTemplateID, i.ResourceSubtype, i.ExceptSystemTypes)
	if err != nil {
		return false, err
	}

	return isAllowedToParticipateInFormationsOfType, nil
}

func (e *ConstraintEngine) isAllowedToParticipateInFormationsOfType(ctx context.Context, assignedFormationNames []string, tenant, formationTemplateID, resourceSubtype string, exceptSystemTypes []string) (bool, error) {
	if len(assignedFormationNames) == 0 {
		return true, nil
	}

	if len(exceptSystemTypes) > 0 {
		for _, exceptType := range exceptSystemTypes {
			if resourceSubtype == exceptType {
				return true, nil
			}
		}
	}

	assignedFormations, err := e.formationRepo.ListByFormationNames(ctx, assignedFormationNames, tenant)
	if err != nil {
		return false, err
	}

	for _, formation := range assignedFormations {
		if formation.FormationTemplateID == formationTemplateID {
			return false, nil
		}
	}

	return true, nil
}

// listFormationsForObject returns all Formations that `objectID` is part of
func (e *ConstraintEngine) listFormationsForObject(ctx context.Context, objectID string) ([]string, error) {
	assignments, err := e.formationAssignmentService.ListAllForObjectGlobal(ctx, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing formations assignments for participant with ID %s", objectID)
	}

	if len(assignments) == 0 {
		return nil, nil
	}

	uniqueFormationIDsMap := make(map[string]struct{}, len(assignments))
	for _, assignment := range assignments {
		uniqueFormationIDsMap[assignment.FormationID] = struct{}{}
	}

	uniqueFormationIDs := make([]string, 0, len(uniqueFormationIDsMap))
	for formationID := range uniqueFormationIDsMap {
		uniqueFormationIDs = append(uniqueFormationIDs, formationID)
	}

	formations, err := e.formationRepo.ListByIDsGlobal(ctx, uniqueFormationIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formations by IDs for object with ID %s", objectID)
	}

	formationNames := make([]string, 0, len(formations))
	for _, formation := range formations {
		formationNames = append(formationNames, formation.Name)
	}

	return formationNames, nil
}
