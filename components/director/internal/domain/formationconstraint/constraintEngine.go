package formationconstraint

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/formation_constraint_input"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=formationConstraintSvc --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationConstraintSvc interface {
	ListMatchingConstraints(ctx context.Context, formationTemplateID string, location JoinPointLocation, details MatchingDetails) ([]*model.FormationConstraint, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --exported --name=automaticFormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentService interface {
	ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=formationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.Formation, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

// OperatorName represents the constraint operator name
type OperatorName string

// OperatorInput represents the input needed by the constraint operator
type OperatorInput interface{}

// OperatorFunc provides an interface for functions implementing formation operators
type OperatorFunc func(ctx context.Context, input OperatorInput) (bool, error)

// OperatorInputConstructor returns empty OperatorInput
type OperatorInputConstructor func() OperatorInput

// NewIsNotAssignedToAnyFormationOfTypeInput returns empty OperatorInput for operator IsNotAssignedToAnyFormationOfType
func NewIsNotAssignedToAnyFormationOfTypeInput() OperatorInput {
	return &formation_constraint_input.IsNotAssignedToAnyFormationOfTypeInput{}
}

// JoinPointLocation contains information to distinguish join points
type JoinPointLocation struct {
	OperationName  model.TargetOperation
	ConstraintType model.FormationConstraintType
}

// ConstraintEngine determines which constraints are applicable to the reached join point and enforces them
type ConstraintEngine struct {
	constraintSvc             formationConstraintSvc
	tenantSvc                 tenantService
	asaSvc                    automaticFormationAssignmentService
	formationRepo             formationRepository
	labelRepo                 labelRepository
	operators                 map[OperatorName]OperatorFunc
	operatorInputConstructors map[OperatorName]OperatorInputConstructor
}

// NewConstraintEngine returns new ConstraintEngine
func NewConstraintEngine(constraintSvc formationConstraintSvc, tenantSvc tenantService, asaSvc automaticFormationAssignmentService, formationRepo formationRepository, labelRepo labelRepository) *ConstraintEngine {
	c := &ConstraintEngine{
		constraintSvc:             constraintSvc,
		tenantSvc:                 tenantSvc,
		asaSvc:                    asaSvc,
		formationRepo:             formationRepo,
		labelRepo:                 labelRepo,
		operatorInputConstructors: map[OperatorName]OperatorInputConstructor{"IsNotAssignedToAnyFormationOfType": NewIsNotAssignedToAnyFormationOfTypeInput},
	}
	c.operators = map[OperatorName]OperatorFunc{"IsNotAssignedToAnyFormationOfType": c.IsNotAssignedToAnyFormationOfType}
	return c
}

// EnforceConstraints finds all the applicable constraints based on JoinPointLocation and JoinPointDetails. Checks for each constraint if it is satisfied.
// If any constraint is not satisfied this information is stored and the engine proceeds with enforcing the next constraint if such exists. In the end if
// any constraint was not satisfied an error is returned.
func (e *ConstraintEngine) EnforceConstraints(ctx context.Context, location JoinPointLocation, details JoinPointDetails, formationTemplateID string) error {
	matchingDetails := details.GetMatchingDetails()
	log.C(ctx).Infof("Enforcing constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchingDetails.resourceType, matchingDetails.resourceSubtype)

	constraints, err := e.constraintSvc.ListMatchingConstraints(ctx, formationTemplateID, location, matchingDetails)
	if err != nil {
		return errors.Wrapf(err, "While listing matching constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchingDetails.resourceType, matchingDetails.resourceSubtype)
	}

	matchedConstraintsNames := make([]string, 0, len(constraints))
	for _, c := range constraints {
		matchedConstraintsNames = append(matchedConstraintsNames, c.Name)
	}

	log.C(ctx).Infof("Matched constraints: %v", matchedConstraintsNames)

	var errs *multierror.Error
	for _, mc := range constraints {
		operator, ok := e.operators[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInputConstructor, ok := e.operatorInputConstructors[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator input constructor for operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInput := operatorInputConstructor()
		if err := formation_constraint_input.ParseInputTemplate(mc.InputTemplate, details, operatorInput); err != nil {
			log.C(ctx).Errorf("An error occured while parsing input template for formation constraint %q: %s", mc.Name, err.Error())
			errs = multierror.Append(errs, ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Failed to parse operator input template for operator %q", mc.Operator),
			})
			continue
		}

		operatorResult, err := operator(ctx, operatorInput)
		if err != nil {
			errs = multierror.Append(errs, ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("An error occured while executing operator %q for formation constraint %q: %v", mc.Operator, mc.Name, err),
			})
		}

		if !operatorResult {
			errs = multierror.Append(errs, ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q is not satisfied", mc.Operator),
			})
		}

	}

	return errs.ErrorOrNil()
}

// IsNotAssignedToAnyFormationOfType checks if the resource from the OperatorInput is already part of formation of the type that the operator is associated with
func (e *ConstraintEngine) IsNotAssignedToAnyFormationOfType(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: IsNotAssignedToAnyFormationOfType")

	i, ok := input.(*formation_constraint_input.IsNotAssignedToAnyFormationOfTypeInput)
	if !ok {
		return false, errors.New("Incompatible input")
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q", i.ResourceType, i.ResourceSubtype, i.ResourceID)

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
		return false, errors.Errorf("Unsupportedd resource type %q", i.ResourceType)
	}

	participatesInFormationsOfType, err := e.participatesInFormationsOfType(ctx, assignedFormations, i.FormationTemplateID)
	if err != nil {
		return false, err
	}

	if participatesInFormationsOfType {
		return false, nil
	}

	return true, nil

}

func (e *ConstraintEngine) participatesInFormationsOfType(ctx context.Context, assignedFormationNames []string, formationTemplateID string) (bool, error) {
	formationsFromTemplate, err := e.formationRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	if err != nil {
		return false, err
	}

	for _, assignedFormationName := range assignedFormationNames {
		for _, f := range formationsFromTemplate {
			if f.Name == assignedFormationName {
				return true, nil
			}
		}
	}

	return false, nil
}
