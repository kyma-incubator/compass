package operators

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=formationConstraintSvc --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationConstraintSvc interface {
	ListMatchingConstraints(ctx context.Context, formationTemplateID string, location formationconstraintpkg.JoinPointLocation, details formationconstraintpkg.MatchingDetails) ([]*model.FormationConstraint, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --exported --name=automaticScenarioAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticScenarioAssignmentService interface {
	ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=formationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	ListByFormationNames(ctx context.Context, formationNames []string, tenantID string) ([]*model.Formation, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeCtxRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeCtxRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

// OperatorInput represents the input needed by the constraint operator
type OperatorInput interface{}

// OperatorName represents the constraint operator name
type OperatorName string

// OperatorFunc provides an interface for functions implementing constraint operators
type OperatorFunc func(ctx context.Context, input OperatorInput) (bool, error)

// OperatorInputConstructor returns empty OperatorInput for a certain constraint operator
type OperatorInputConstructor func() OperatorInput

// ConstraintEngine determines which constraints are applicable to the reached join point and enforces them
type ConstraintEngine struct {
	constraintSvc             formationConstraintSvc
	tenantSvc                 tenantService
	asaSvc                    automaticScenarioAssignmentService
	formationRepo             formationRepository
	labelRepo                 labelRepository
	labelService              labelService
	applicationRepository     applicationRepository
	runtimeRepository         runtimeRepository
	runtimeCtxRepository      runtimeCtxRepository
	mtlsHTTPClient            *http.Client
	destinationCfg            DestinationConfig
	operators                 map[OperatorName]OperatorFunc
	operatorInputConstructors map[OperatorName]OperatorInputConstructor
}

// NewConstraintEngine returns new ConstraintEngine
func NewConstraintEngine(constraintSvc formationConstraintSvc, tenantSvc tenantService, asaSvc automaticScenarioAssignmentService, formationRepo formationRepository, labelRepo labelRepository, labelService labelService, applicationRepository applicationRepository, runtimeRepository runtimeRepository, runtimeCtxRepository runtimeCtxRepository, mtlsHTTPClient *http.Client, destinationCfg DestinationConfig) *ConstraintEngine {
	c := &ConstraintEngine{
		constraintSvc:         constraintSvc,
		tenantSvc:             tenantSvc,
		asaSvc:                asaSvc,
		formationRepo:         formationRepo,
		labelRepo:             labelRepo,
		labelService:          labelService,
		applicationRepository: applicationRepository,
		runtimeRepository:     runtimeRepository,
		runtimeCtxRepository:  runtimeCtxRepository,
		mtlsHTTPClient:        mtlsHTTPClient,
		destinationCfg:        destinationCfg,
		operatorInputConstructors: map[OperatorName]OperatorInputConstructor{
			IsNotAssignedToAnyFormationOfTypeOperator: NewIsNotAssignedToAnyFormationOfTypeInput,
			DoesNotContainResourceOfSubtypeOperator:   NewDoesNotContainResourceOfSubtypeInput,
			DestinationCreatorOperator:                NewDestinationCreatorInput,
		},
	}
	c.operators = map[OperatorName]OperatorFunc{
		IsNotAssignedToAnyFormationOfTypeOperator: c.IsNotAssignedToAnyFormationOfType,
		DoesNotContainResourceOfSubtypeOperator:   c.DoesNotContainResourceOfSubtype,
		DestinationCreatorOperator:                c.DestinationCreator,
	}
	return c
}

// EnforceConstraints finds all the applicable constraints based on JoinPointLocation and JoinPointDetails. Checks for each constraint if it is satisfied.
// If any constraint is not satisfied this information is stored and the engine proceeds with enforcing the next constraint if such exists. In the end if
// any constraint was not satisfied an error is returned.
func (e *ConstraintEngine) EnforceConstraints(ctx context.Context, location formationconstraintpkg.JoinPointLocation, details formationconstraintpkg.JoinPointDetails, formationTemplateID string) error {
	matchingDetails := details.GetMatchingDetails()
	log.C(ctx).Infof("Enforcing constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchingDetails.ResourceType, matchingDetails.ResourceSubtype)

	constraints, err := e.constraintSvc.ListMatchingConstraints(ctx, formationTemplateID, location, matchingDetails)
	if err != nil {
		return errors.Wrapf(err, "While listing matching constraints for target operation %q, constraint type %q, resource type %q and resource subtype %q", location.OperationName, location.ConstraintType, matchingDetails.ResourceType, matchingDetails.ResourceSubtype)
	}

	matchedConstraintsNames := make([]string, 0, len(constraints))
	for _, c := range constraints {
		matchedConstraintsNames = append(matchedConstraintsNames, c.Name)
	}

	log.C(ctx).Infof("Matched constraints: %v", matchedConstraintsNames)

	var errs *multierror.Error
	for _, mc := range constraints {
		operatorFunc, ok := e.operators[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, formationconstraint.ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInputConstructor, ok := e.operatorInputConstructors[OperatorName(mc.Operator)]
		if !ok {
			errs = multierror.Append(errs, formationconstraint.ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator input constructor for operator %q not found", mc.Operator),
			})
			continue
		}

		operatorInput := operatorInputConstructor()
		if err := formationconstraintpkg.ParseInputTemplate(mc.InputTemplate, details, operatorInput); err != nil {
			log.C(ctx).Errorf("An error occurred while parsing input template for formation constraint %q: %s", mc.Name, err.Error())
			errs = multierror.Append(errs, formationconstraint.ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Failed to parse operator input template for operator %q", mc.Operator),
			})
			continue
		}

		operatorResult, err := operatorFunc(ctx, operatorInput)
		if err != nil {
			errs = multierror.Append(errs, formationconstraint.ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("An error occurred while executing operator %q for formation constraint %q: %v", mc.Operator, mc.Name, err),
			})
		}

		if !operatorResult {
			errs = multierror.Append(errs, formationconstraint.ConstraintError{
				ConstraintName: mc.Name,
				Reason:         fmt.Sprintf("Operator %q is not satisfied", mc.Operator),
			})
		}
	}

	return errs.ErrorOrNil()
}
