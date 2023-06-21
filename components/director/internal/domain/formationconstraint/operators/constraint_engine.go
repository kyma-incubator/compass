package operators

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
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

//go:generate mockery --exported --name=destinationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationService interface {
	CreateDesignTimeDestinations(ctx context.Context, destinationDetails Destination, formationAssignment *model.FormationAssignment) (statusCode int, err error)
	CreateBasicCredentialDestinations(ctx context.Context, destinationDetails Destination, basicAuthenticationCredentials BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) (statusCode int, err error)
	CreateSAMLAssertionDestination(ctx context.Context, destinationDetails Destination, samlAssertionAuthCredentials *SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) (defaultStatusCode int, err error)
	CreateCertificateInDestinationService(ctx context.Context, destinationDetails Destination, formationAssignment *model.FormationAssignment) (defaultCertData destinationcreator.CertificateResponse, defaultStatusCode int, err error)
	DeleteDestinationFromDestinationService(ctx context.Context, destinationName, destinationSubaccount string, formationAssignment *model.FormationAssignment) error
	DeleteDestinations(ctx context.Context, formationAssignment *model.FormationAssignment) error
	DeleteCertificateFromDestinationService(ctx context.Context, certificateName, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) error
	EnrichAssignmentConfigWithCertificateData(assignmentConfig json.RawMessage, certData destinationcreator.CertificateResponse, destinationIndex int) (json.RawMessage, error)
}

//go:generate mockery --exported --name=formationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationRepository interface {
	ListByFormationNames(ctx context.Context, formationNames []string, tenantID string) ([]*model.Formation, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

//go:generate mockery --exported --name=runtimeContextRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepo interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=formationTemplateRepo --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationTemplateRepo interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
}

// FormationAssignmentRepository represents the Formation Assignment repository layer
//go:generate mockery --name=formationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentRepository interface {
	Update(ctx context.Context, model *model.FormationAssignment) error
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
	transact                  persistence.Transactioner
	constraintSvc             formationConstraintSvc
	tenantSvc                 tenantService
	asaSvc                    automaticScenarioAssignmentService
	destinationSvc            destinationService
	formationRepo             formationRepository
	labelRepo                 labelRepository
	labelService              labelService
	applicationRepository     applicationRepository
	runtimeContextRepo        runtimeContextRepo
	formationTemplateRepo     formationTemplateRepo
	formationAssignmentRepo   formationAssignmentRepository
	operators                 map[OperatorName]OperatorFunc
	operatorInputConstructors map[OperatorName]OperatorInputConstructor
	runtimeTypeLabelKey       string
	applicationTypeLabelKey   string
}

// NewConstraintEngine returns new ConstraintEngine
func NewConstraintEngine(transact persistence.Transactioner, constraintSvc formationConstraintSvc, tenantSvc tenantService, asaSvc automaticScenarioAssignmentService, destinationSvc destinationService, formationRepo formationRepository, labelRepo labelRepository, labelService labelService, applicationRepository applicationRepository, runtimeContextRepo runtimeContextRepo, formationTemplateRepo formationTemplateRepo, formationAssignmentRepo formationAssignmentRepository, runtimeTypeLabelKey string, applicationTypeLabelKey string) *ConstraintEngine {
	c := &ConstraintEngine{
		transact:                transact,
		constraintSvc:           constraintSvc,
		tenantSvc:               tenantSvc,
		asaSvc:                  asaSvc,
		destinationSvc:          destinationSvc,
		formationRepo:           formationRepo,
		labelRepo:               labelRepo,
		labelService:            labelService,
		applicationRepository:   applicationRepository,
		runtimeContextRepo:      runtimeContextRepo,
		formationTemplateRepo:   formationTemplateRepo,
		formationAssignmentRepo: formationAssignmentRepo,
		operatorInputConstructors: map[OperatorName]OperatorInputConstructor{
			IsNotAssignedToAnyFormationOfTypeOperator:            NewIsNotAssignedToAnyFormationOfTypeInput,
			DoesNotContainResourceOfSubtypeOperator:              NewDoesNotContainResourceOfSubtypeInput,
			DoNotGenerateFormationAssignmentNotificationOperator: NewDoNotGenerateFormationAssignmentNotificationInput,
			DestinationCreatorOperator:                           NewDestinationCreatorInput,
		},
		runtimeTypeLabelKey:     runtimeTypeLabelKey,
		applicationTypeLabelKey: applicationTypeLabelKey,
	}
	c.operators = map[OperatorName]OperatorFunc{
		IsNotAssignedToAnyFormationOfTypeOperator:            c.IsNotAssignedToAnyFormationOfType,
		DoesNotContainResourceOfSubtypeOperator:              c.DoesNotContainResourceOfSubtype,
		DoNotGenerateFormationAssignmentNotificationOperator: c.DoNotGenerateFormationAssignmentNotification,
		DestinationCreatorOperator:                           c.DestinationCreator,
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
