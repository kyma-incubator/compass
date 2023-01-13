package formationconstraint

import (
	"context"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/formation_constraint_input"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

// formationConstraintSvc represents the Formation Constraint service layer
//go:generate mockery --exported --name=FormationConstraintService --output=automock --outpkg=automock --case=underscore --disable-version-string
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

//go:generate mockery --exported --name=formationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.Formation, error)
}

// LabelRepository missing godoc
//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

type MatchingDetails struct {
	resourceType    model.ResourceType
	resourceSubtype string
}

type JoinPointDetails interface {
	GetMatchingDetails() MatchingDetails
}

type CRUDFormationOperationDetails struct {
	FormationType       string
	FormationTemplateID string
	FormationName       string
	CheckScope          model.OperatorScopeType
	TenantID            string
}

func (d *CRUDFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    model.FormationResourceType,
		resourceSubtype: d.FormationType,
	}
}

type AssignFormationOperationDetails struct {
	ResourceType        model.ResourceType
	ResourceSubtype     string
	ResourceID          string
	FormationType       string
	FormationTemplateID string
	FormationID         string
	TenantID            string
}

func (d *AssignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    d.ResourceType,
		resourceSubtype: d.ResourceSubtype,
	}
}

type UnassignFormationOperationDetails struct {
	ResourceType        model.ResourceType
	ResourceSubtype     string
	ResourceID          string
	FormationType       string
	FormationTemplateID string
	FormationID         string
	TenantID            string
}

func (d *UnassignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    d.ResourceType,
		resourceSubtype: d.ResourceSubtype,
	}
}

//todo add comments
type GenerateNotificationOperationDetails struct {
	Operation           model.FormationOperation
	FormationID         string
	ApplicationTemplate *webhook.ApplicationTemplateWithLabels
	Application         *webhook.ApplicationWithLabels
	Runtime             *webhook.RuntimeWithLabels
	RuntimeContext      *webhook.RuntimeContextWithLabels
	Assignment          *webhook.FormationAssignment
	ReverseAssignment   *webhook.FormationAssignment

	SourceApplicationTemplate *webhook.ApplicationTemplateWithLabels
	// SourceApplication is the application that the notification is about
	SourceApplication         *webhook.ApplicationWithLabels
	TargetApplicationTemplate *webhook.ApplicationTemplateWithLabels
	// TargetApplication is the application that the notification is for (the one with the webhook / the one receiving the notification)
	TargetApplication *webhook.ApplicationWithLabels

	ResourceType    model.ResourceType
	ResourceSubtype string
	ResourceID      string
}

func (d *GenerateNotificationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		resourceType:    d.ResourceType,
		resourceSubtype: d.ResourceSubtype,
	}
}

type OperatorName string

type OperatorInput interface{}

type OperatorFunc func(ctx context.Context, input OperatorInput) (bool, error)

type OperatorInputConstructor func() OperatorInput

func NewIsNotAssignedToAnyFormationOfTypeInput() OperatorInput {
	return &formation_constraint_input.IsNotAssignedToAnyFormationOfTypeInput{}
}

func (e *ConstraintEngine) IsNotAssignedToAnyFormationOfType(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: IsNotAssignedToAnyFormationOfType")
	spew.Dump(input)
	i, ok := input.(*formation_constraint_input.IsNotAssignedToAnyFormationOfTypeInput)
	if !ok {
		return false, errors.New("Incompatible input")
	}
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
		fmt.Println("SHOULD BE HERE")
		return false, nil
	}

	return true, nil

}

func (e *ConstraintEngine) participatesInFormationsOfType(ctx context.Context, assignedFormations []string, formationTemplateID string) (bool, error) {
	formations, err := e.formationRepo.ListByFormationTemplateID(ctx, formationTemplateID)
	spew.Dump(formations)
	if err != nil {
		return false, err
	}

	for _, assignedFormationNAme := range assignedFormations {
		for _, f := range formations {
			if f.Name == assignedFormationNAme {
				return true, nil
			}
		}
	}

	return false, nil
}

type JoinPointLocation struct {
	OperationName  model.TargetOperation
	ConstraintType model.FormationConstraintType
}

type ConstraintEngine struct {
	constraintSvc             formationConstraintSvc
	tenantSvc                 tenantService
	asaSvc                    automaticFormationAssignmentService
	formationRepo             formationRepository
	labelRepo                 labelRepository
	operators                 map[OperatorName]OperatorFunc
	operatorInputConstructors map[OperatorName]OperatorInputConstructor
}

func NewConstraintEngine(constraintSvc formationConstraintSvc, tenantSvc tenantService, asaSvc automaticFormationAssignmentService, formationRepo formationRepository, labelRepo labelRepository) *ConstraintEngine {
	c := &ConstraintEngine{
		constraintSvc:             constraintSvc,
		tenantSvc:                 tenantSvc,
		asaSvc:                    asaSvc,
		formationRepo:             formationRepo,
		labelRepo:                 labelRepo,
		operatorInputConstructors: map[OperatorName]OperatorInputConstructor{"participatesInFormationsOfType": NewIsNotAssignedToAnyFormationOfTypeInput},
	}
	c.operators = map[OperatorName]OperatorFunc{"participatesInFormationsOfType": c.IsNotAssignedToAnyFormationOfType}
	return c
}

func (e *ConstraintEngine) EnforceConstraints(ctx context.Context, location JoinPointLocation, details JoinPointDetails, formationTemplateID string) error {
	matchigDetails := details.GetMatchingDetails()
	log.C(ctx).Infof("Enforcing constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchigDetails.resourceType, matchigDetails.resourceSubtype)

	constraints, err := e.constraintSvc.ListMatchingConstraints(ctx, formationTemplateID, location, matchigDetails)
	if err != nil {
		return errors.Wrapf(err, "While listing matching constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchigDetails.resourceType, matchigDetails.resourceSubtype)
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
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInputConstructor, ok := e.operatorInputConstructors[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator input constructor for operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInput := operatorInputConstructor()
		spew.Dump(mc.InputTemplate, " ", details, " ", operatorInput)
		if err := formation_constraint_input.ParseInputTemplate(mc.InputTemplate, details, operatorInput); err != nil {
			log.C(ctx).Errorf("An error occured while parsing input template for formation constraint %q: %s", mc.Name, err.Error())
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Failed to parse operator input template for operator %q", mc.Operator),
			})
			continue
		}

		operatorResult, err := operator(ctx, operatorInput)
		if err != nil {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("An error occured while executing operator %q for formation constraint %q: %v", mc.Operator, mc.Name, err),
			})
		}

		if !operatorResult {
			errs = multierror.Append(errs, FormationConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q is not satisfied", mc.Operator),
			})
		}

	}

	return errs.ErrorOrNil()
}
