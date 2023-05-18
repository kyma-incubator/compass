package formationconstraint

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// IsNotAssignedToAnyFormationOfTypeOperator represents the IsNotAssignedToAnyFormationOfType operator
	IsNotAssignedToAnyFormationOfTypeOperator = "IsNotAssignedToAnyFormationOfType"
	// DoesNotContainResourceOfSubtypeOperator represents the DoesNotContainResourceOfSubtype operator
	DoesNotContainResourceOfSubtypeOperator = "DoesNotContainResourceOfSubtype"
	// DoNotGenerateFormationAssignmentNotificationOperator represents the DoNotGenerateFormationAssignmentNotification operator
	DoNotGenerateFormationAssignmentNotificationOperator = "DoNotGenerateFormationAssignmentNotification"
)

// OperatorName represents the constraint operator name
type OperatorName string

// OperatorInput represents the input needed by the constraint operator
type OperatorInput interface{}

// OperatorFunc provides an interface for functions implementing constraint operators
type OperatorFunc func(ctx context.Context, input OperatorInput) (bool, error)

// OperatorInputConstructor returns empty OperatorInput for a certain constraint operator
type OperatorInputConstructor func() OperatorInput

// NewIsNotAssignedToAnyFormationOfTypeInput is input constructor for IsNotAssignedToAnyFormationOfType operator. It returns empty OperatorInput.
func NewIsNotAssignedToAnyFormationOfTypeInput() OperatorInput {
	return &formationconstraint.IsNotAssignedToAnyFormationOfTypeInput{}
}

// NewDoesNotContainResourceOfSubtypeInput is input constructor for DoesNotContainResourceOfSubtypeOperator operator. It returns empty OperatorInput
func NewDoesNotContainResourceOfSubtypeInput() OperatorInput {
	return &formationconstraint.DoesNotContainResourceOfSubtypeInput{}
}

// NewDoNotGenerateFormationAssignmentNotificationInput is input constructor for DoNotGenerateFormationAssignmentNotificationOperator operator. It returns empty OperatorInput
func NewDoNotGenerateFormationAssignmentNotificationInput() OperatorInput {
	return &formationconstraint.DoNotGenerateFormationAssignmentNotificationInput{}
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
		scenariosLabel, err := e.labelRepo.GetByKey(ctx, i.Tenant, model.ApplicationLabelableObject, i.ResourceID, model.ScenariosKey)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return true, nil
			}
			return false, err
		}
		assignedFormations, err = label.ValueToStringsSlice(scenariosLabel.Value)
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

// DoesNotContainResourceOfSubtype is a constraint operator. It checks if the formation contains resource with the same subtype as the resource subtype from the OperatorInput
func (e *ConstraintEngine) DoesNotContainResourceOfSubtype(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DoesNotContainResourceOfSubtypeOperator)

	i, ok := input.(*formationconstraint.DoesNotContainResourceOfSubtypeInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator %q", DoesNotContainResourceOfSubtypeOperator)
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", DoesNotContainResourceOfSubtypeOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

	switch i.ResourceType {
	case model.ApplicationResourceType:
		applications, err := e.applicationRepository.ListByScenariosNoPaging(ctx, i.Tenant, []string{i.FormationName})
		if err != nil {
			return false, errors.Wrapf(err, "while listing applications in scenario %q", i.FormationName)
		}

		for _, application := range applications {
			appSubtype, err := e.getObjectSubtype(ctx, i.Tenant, model.ApplicationResourceType, application.ID)
			if err != nil {
				return false, errors.Wrapf(err, "while getting subtype of resource with type: %q and id: %q", model.ApplicationResourceType, application.ID)
			}

			if i.ResourceSubtype == appSubtype {
				return false, nil
			}
		}
	default:
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	return true, nil
}

// DoNotGenerateFormationAssignmentNotification is a constraint operator. It skips the generation of formation assignment notifications
func (e *ConstraintEngine) DoNotGenerateFormationAssignmentNotification(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DoNotGenerateFormationAssignmentNotificationOperator)

	i, ok := input.(*formationconstraint.DoNotGenerateFormationAssignmentNotificationInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator %q", DoNotGenerateFormationAssignmentNotificationOperator)
	}

	log.C(ctx).Infof("Enforcing %q constraint on resource of type: %q, subtype: %q and ID: %q", DoNotGenerateFormationAssignmentNotificationOperator, i.ResourceType, i.ResourceSubtype, i.ResourceID)

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

func (e *ConstraintEngine) getObjectSubtype(ctx context.Context, tenant string, objectType model.ResourceType, objectID string) (string, error) {
	switch objectType {
	case model.ApplicationResourceType:
		appTypeLbl, err := e.labelService.GetByKey(ctx, tenant, model.ApplicationLabelableObject, objectID, e.applicationTypeLabelKey)
		if err != nil {
			return "", errors.Wrapf(err, "while getting label with key %q of application with ID %q in tenant %q", e.applicationTypeLabelKey, objectID, tenant)
		}
		return appTypeLbl.Value.(string), nil
	case model.RuntimeResourceType:
		rtTypeLabel, err := e.labelService.GetByKey(ctx, tenant, model.RuntimeLabelableObject, objectID, e.runtimeTypeLabelKey)
		if err != nil {
			return "", errors.Wrapf(err, "while getting label with key %q of runtime with ID %q in tenant %q", e.runtimeTypeLabelKey, objectID, tenant)
		}
		return rtTypeLabel.Value.(string), nil
	case model.RuntimeContextResourceType:
		rtCtx, err := e.runtimeContextRepo.GetByID(ctx, tenant, objectID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting runtime context with ID %q in tenant %q", objectID, tenant)
		}
		rtTypeLabel, err := e.labelService.GetByKey(ctx, tenant, model.RuntimeLabelableObject, rtCtx.RuntimeID, e.runtimeTypeLabelKey)
		if err != nil {
			return "", errors.Wrapf(err, "while getting label with key %q of runtime with ID %q in tenant %q", e.runtimeTypeLabelKey, rtCtx.RuntimeID, tenant)
		}
		return rtTypeLabel.Value.(string), nil
	default:
		return "", errors.Errorf("Unsupported resource type %q", objectID)
	}
}
