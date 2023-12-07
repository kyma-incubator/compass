package operators

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/templatehelper"
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
	CreateDesignTimeDestinations(ctx context.Context, destinationsDetails []Destination, formationAssignment *model.FormationAssignment, skipSubaccountValidation bool) error
	CreateBasicCredentialDestinations(ctx context.Context, destinationsDetails []Destination, basicAuthenticationCredentials BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error
	CreateSAMLAssertionDestination(ctx context.Context, destinationsDetails []Destination, samlAssertionAuthCredentials *SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error
	CreateClientCertificateAuthenticationDestination(ctx context.Context, destinationsDetails []Destination, clientCertAuthCredentials *ClientCertAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error
	CreateOAuth2ClientCredentialsDestinations(ctx context.Context, destinationsDetails []Destination, oauth2ClientCredsCredentials *OAuth2ClientCredentialsAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error
	DeleteDestinations(ctx context.Context, formationAssignment *model.FormationAssignment, skipSubaccountValidation bool) error
}

//go:generate mockery --exported --name=destinationCreatorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationCreatorService interface {
	CreateCertificate(ctx context.Context, destinationsDetails []Destination, destinationAuthType destinationcreatorpkg.AuthType, formationAssignment *model.FormationAssignment, depth uint8, skipSubaccountValidation, useSelfSignedCert bool) (*CertificateData, error)
	EnrichAssignmentConfigWithCertificateData(assignmentConfig json.RawMessage, destinationTypePath string, certData *CertificateData) (json.RawMessage, error)
	EnrichAssignmentConfigWithSAMLCertificateData(assignmentConfig json.RawMessage, destinationTypePath string, certData *CertificateData) (json.RawMessage, error)
}

//go:generate mockery --exported --name=systemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type systemAuthService interface {
	ListForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string) ([]pkgmodel.SystemAuth, error)
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
//
//go:generate mockery --name=formationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentRepository interface {
	Update(ctx context.Context, model *model.FormationAssignment) error
}

// FormationAssignmentService represents the formation assignment notification service for generating notifications
//
//go:generate mockery --name=formationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentService interface {
	CleanupFormationAssignment(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPairWithOperation) (bool, error)
}

// FormationAssignmentNotificationService represents the formation assignment notification service for generating notifications
//
//go:generate mockery --name=formationAssignmentNotificationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentNotificationService interface {
	GenerateFormationAssignmentPair(ctx context.Context, fa, reverseFA *model.FormationAssignment, operation model.FormationOperation) (*formationassignment.AssignmentMappingPairWithOperation, error)
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
	transact                           persistence.Transactioner
	constraintSvc                      formationConstraintSvc
	tenantSvc                          tenantService
	asaSvc                             automaticScenarioAssignmentService
	destinationSvc                     destinationService
	destinationCreatorSvc              destinationCreatorService
	systemAuthSvc                      systemAuthService
	formationRepo                      formationRepository
	labelRepo                          labelRepository
	labelService                       labelService
	applicationRepository              applicationRepository
	runtimeContextRepo                 runtimeContextRepo
	formationTemplateRepo              formationTemplateRepo
	formationAssignmentRepo            formationAssignmentRepository
	formationAssignmentService         formationAssignmentService
	formationAssignmentNotificationSvc formationAssignmentNotificationService
	operators                          map[OperatorName]OperatorFunc
	operatorInputConstructors          map[OperatorName]OperatorInputConstructor
	runtimeTypeLabelKey                string
	applicationTypeLabelKey            string
}

// NewConstraintEngine returns new ConstraintEngine
func NewConstraintEngine(transact persistence.Transactioner, constraintSvc formationConstraintSvc, tenantSvc tenantService, asaSvc automaticScenarioAssignmentService, destinationSvc destinationService, destinationCreatorSvc destinationCreatorService, systemAuthSvc systemAuthService, formationRepo formationRepository, labelRepo labelRepository, labelService labelService, applicationRepository applicationRepository, runtimeContextRepo runtimeContextRepo, formationTemplateRepo formationTemplateRepo, formationAssignmentRepo formationAssignmentRepository, formationAssignmentService formationAssignmentService, formationAssignmentNotificationSvc formationAssignmentNotificationService, runtimeTypeLabelKey string, applicationTypeLabelKey string) *ConstraintEngine {
	ce := &ConstraintEngine{
		transact:                           transact,
		constraintSvc:                      constraintSvc,
		tenantSvc:                          tenantSvc,
		asaSvc:                             asaSvc,
		destinationSvc:                     destinationSvc,
		destinationCreatorSvc:              destinationCreatorSvc,
		systemAuthSvc:                      systemAuthSvc,
		formationRepo:                      formationRepo,
		labelRepo:                          labelRepo,
		labelService:                       labelService,
		applicationRepository:              applicationRepository,
		runtimeContextRepo:                 runtimeContextRepo,
		formationTemplateRepo:              formationTemplateRepo,
		formationAssignmentRepo:            formationAssignmentRepo,
		formationAssignmentService:         formationAssignmentService,
		formationAssignmentNotificationSvc: formationAssignmentNotificationSvc,
		operatorInputConstructors: map[OperatorName]OperatorInputConstructor{
			IsNotAssignedToAnyFormationOfTypeOperator:                    NewIsNotAssignedToAnyFormationOfTypeInput,
			DoesNotContainResourceOfSubtypeOperator:                      NewDoesNotContainResourceOfSubtypeInput,
			ContainsScenarioGroupsOperator:                               NewContainsScenarioGroupsInput,
			DoNotGenerateFormationAssignmentNotificationOperator:         NewDoNotGenerateFormationAssignmentNotificationInput,
			DoNotGenerateFormationAssignmentNotificationForLoopsOperator: NewDoNotGenerateFormationAssignmentNotificationForLoopsInput,
			DestinationCreatorOperator:                                   NewDestinationCreatorInput,
			ConfigMutatorOperator:                                        NewConfigMutatorInput,
			RedirectNotificationOperator:                                 NewRedirectNotificationInput,
			AsynchronousFlowControlOperator:                              AsynchronousFlowControlOperatorInput,
		},
		runtimeTypeLabelKey:     runtimeTypeLabelKey,
		applicationTypeLabelKey: applicationTypeLabelKey,
	}
	ce.operators = map[OperatorName]OperatorFunc{
		IsNotAssignedToAnyFormationOfTypeOperator:                    ce.IsNotAssignedToAnyFormationOfType,
		DoesNotContainResourceOfSubtypeOperator:                      ce.DoesNotContainResourceOfSubtype,
		ContainsScenarioGroupsOperator:                               ce.ContainsScenarioGroups,
		DoNotGenerateFormationAssignmentNotificationOperator:         ce.DoNotGenerateFormationAssignmentNotification,
		DoNotGenerateFormationAssignmentNotificationForLoopsOperator: ce.DoNotGenerateFormationAssignmentNotificationForLoops,
		DestinationCreatorOperator:                                   ce.DestinationCreator,
		ConfigMutatorOperator:                                        ce.MutateConfig,
		RedirectNotificationOperator:                                 ce.RedirectNotification,
		AsynchronousFlowControlOperator:                              ce.AsynchronousFlowControlOperator,
	}
	return ce
}

// SetFormationAssignmentService sets the formation assignment service of the constraint engine
func (e *ConstraintEngine) SetFormationAssignmentService(formationAssignmentService formationAssignmentService) {
	e.formationAssignmentService = formationAssignmentService
}

// SetFormationAssignmentNotificationService sets the formation assignment notification service of the constraint engine
func (e *ConstraintEngine) SetFormationAssignmentNotificationService(formationAssignmentNotificationSvc formationAssignmentNotificationService) {
	e.formationAssignmentNotificationSvc = formationAssignmentNotificationSvc
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

	sort.Slice(constraints, func(i, j int) bool {
		return constraints[i].Priority > constraints[j].Priority ||
			(constraints[i].Priority == constraints[j].Priority && constraints[i].CreatedAt.Before(*constraints[j].CreatedAt))
	})

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
		if err := templatehelper.ParseTemplate(&mc.InputTemplate, details, operatorInput); err != nil {
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
