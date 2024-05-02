package formation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"

	"github.com/hashicorp/go-multierror"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

//go:generate mockery --exported --name=labelDefRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefRepository interface {
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
	UpdateWithVersion(ctx context.Context, def model.LabelDefinition) error
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	Delete(context.Context, string, model.LabelableObject, string, string) error
	ListForObjectIDs(ctx context.Context, tenant string, objectType model.LabelableObject, objectIDs []string) (map[string]map[string]interface{}, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	GetByFiltersAndIDUsingUnion(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListAllWithUnionSetCombination(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Runtime, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	OwnerExistsByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (bool, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByRuntimeID(ctx context.Context, tenant, runtimeID string) (*model.RuntimeContext, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.RuntimeContext, error)
	ListByScenariosAndRuntimeIDs(ctx context.Context, tenant string, scenarios []string, runtimeIDs []string) ([]*model.RuntimeContext, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.RuntimeContext, error)
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
	ExistsByRuntimeID(ctx context.Context, tenant, rtmID string) (bool, error)
}

// FormationRepository represents the Formations repository layer
//
//go:generate mockery --name=FormationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationRepository interface {
	Get(ctx context.Context, id, tenantID string) (*model.Formation, error)
	GetByName(ctx context.Context, name, tenantID string) (*model.Formation, error)
	GetGlobalByID(ctx context.Context, id string) (*model.Formation, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.FormationPage, error)
	ListByIDsGlobal(ctx context.Context, formationIDs []string) ([]*model.Formation, error)
	Create(ctx context.Context, item *model.Formation) error
	DeleteByName(ctx context.Context, tenantID, name string) error
	Update(ctx context.Context, model *model.Formation) error
}

// FormationTemplateRepository represents the FormationTemplate repository layer
//
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	GetByNameAndTenant(ctx context.Context, templateName, tenantID string) (*model.FormationTemplate, error)
}

// NotificationsService represents the notification service for generating and sending notifications
//
//go:generate mockery --name=NotificationsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type NotificationsService interface {
	GenerateFormationAssignmentNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.FormationAssignmentNotificationRequestTargetMapping, error)
	GenerateFormationNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, formation *model.Formation, formationTemplateName, formationTemplateID string, formationOperation model.FormationOperation) ([]*webhookclient.FormationNotificationRequest, error)
	SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookExtRequest) (*webhookdir.Response, error)
	PrepareDetailsForNotificationStatusReturned(ctx context.Context, formation *model.Formation, operation model.FormationOperation) (*formationconstraint.NotificationStatusReturnedOperationDetails, error)
}

//go:generate mockery --exported --name=statusService --output=automock --outpkg=automock --case=underscore --disable-version-string
type statusService interface {
	UpdateWithConstraints(ctx context.Context, formation *model.Formation, operation model.FormationOperation) error
	SetFormationToErrorStateWithConstraints(ctx context.Context, formation *model.Formation, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationState, operation model.FormationOperation) error
}

// FormationAssignmentNotificationsService represents the notification service for generating and sending notifications
//
//go:generate mockery --name=FormationAssignmentNotificationsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentNotificationsService interface {
	GenerateFormationAssignmentNotification(ctx context.Context, formationAssignment *model.FormationAssignment, operation model.FormationOperation) (*webhookclient.FormationAssignmentNotificationRequest, error)
}

//go:generate mockery --exported --name=labelDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefService interface {
	CreateWithFormations(ctx context.Context, tnt string, formations []string) error
	ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error
	ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error
	GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error)
}

//go:generate mockery --exported --name=labelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelService interface {
	CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error
	GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error)
}

//go:generate mockery --exported --name=uuidService --output=automock --outpkg=automock --case=underscore --disable-version-string
type uuidService interface {
	Generate() string
}

//go:generate mockery --exported --name=automaticFormationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentService interface {
	GetForScenarioName(ctx context.Context, scenarioName string) (*model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=automaticFormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentRepository interface {
	Create(ctx context.Context, model *model.AutomaticScenarioAssignment) error
	DeleteForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) error
	DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error
	ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=tenantService --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantService interface {
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
	GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
}

//go:generate mockery --exported --name=constraintEngine --output=automock --outpkg=automock --case=underscore --disable-version-string
type constraintEngine interface {
	EnforceConstraints(ctx context.Context, location formationconstraint.JoinPointLocation, details formationconstraint.JoinPointDetails, formationTemplateID string) error
}

//go:generate mockery --exported --name=asaEngine --output=automock --outpkg=automock --case=underscore --disable-version-string
type asaEngine interface {
	EnsureScenarioAssigned(ctx context.Context, in *model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error
	RemoveAssignedScenario(ctx context.Context, in *model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error
	GetMatchingFuncByFormationObjectType(objType graphql.FormationObjectType) (MatchingFunc, error)
	GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error)
	IsFormationComingFromASA(ctx context.Context, objectID, formation string, objectType graphql.FormationObjectType) (bool, error)
}

//go:generate mockery --exported --name=assignmentOperationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type assignmentOperationService interface {
	Create(ctx context.Context, in *model.AssignmentOperationInput) (string, error)
	Finish(ctx context.Context, assignmentID, formationID string, operationType model.AssignmentOperationType) error
	Update(ctx context.Context, assignmentID, formationID string, operationType model.AssignmentOperationType, newTrigger model.OperationTrigger) error
	ListByFormationAssignmentIDs(ctx context.Context, formationAssignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error)
	DeleteByIDs(ctx context.Context, ids []string) error
}

type service struct {
	applicationRepository                  applicationRepository
	labelDefRepository                     labelDefRepository
	labelRepository                        labelRepository
	formationRepository                    FormationRepository
	formationTemplateRepository            FormationTemplateRepository
	labelService                           labelService
	labelDefService                        labelDefService
	asaService                             automaticFormationAssignmentService
	uuidService                            uuidService
	tenantSvc                              tenantService
	repo                                   automaticFormationAssignmentRepository
	runtimeRepo                            runtimeRepository
	runtimeContextRepo                     runtimeContextRepository
	formationAssignmentService             formationAssignmentService
	assignmentOperationService             assignmentOperationService
	formationAssignmentNotificationService FormationAssignmentNotificationsService
	notificationsService                   NotificationsService
	constraintEngine                       constraintEngine
	webhookRepository                      webhookRepository
	transact                               persistence.Transactioner
	asaEngine                              asaEngine
	statusService                          statusService
	runtimeTypeLabelKey                    string
	applicationTypeLabelKey                string
}

// NewService creates formation service
func NewService(
	transact persistence.Transactioner,
	applicationRepository applicationRepository,
	labelDefRepository labelDefRepository,
	labelRepository labelRepository,
	formationRepository FormationRepository,
	formationTemplateRepository FormationTemplateRepository,
	labelService labelService,
	uuidService uuidService,
	labelDefService labelDefService,
	asaRepo automaticFormationAssignmentRepository,
	asaService automaticFormationAssignmentService,
	tenantSvc tenantService, runtimeRepo runtimeRepository,
	runtimeContextRepo runtimeContextRepository,
	formationAssignmentService formationAssignmentService,
	assignmentOperationService assignmentOperationService,
	formationAssignmentNotificationService FormationAssignmentNotificationsService,
	notificationsService NotificationsService,
	constraintEngine constraintEngine,
	webhookRepository webhookRepository,
	statusService statusService,
	runtimeTypeLabelKey, applicationTypeLabelKey string) *service {
	return &service{
		transact:                               transact,
		applicationRepository:                  applicationRepository,
		labelDefRepository:                     labelDefRepository,
		labelRepository:                        labelRepository,
		formationRepository:                    formationRepository,
		formationTemplateRepository:            formationTemplateRepository,
		labelService:                           labelService,
		labelDefService:                        labelDefService,
		asaService:                             asaService,
		uuidService:                            uuidService,
		tenantSvc:                              tenantSvc,
		repo:                                   asaRepo,
		runtimeRepo:                            runtimeRepo,
		runtimeContextRepo:                     runtimeContextRepo,
		formationAssignmentNotificationService: formationAssignmentNotificationService,
		formationAssignmentService:             formationAssignmentService,
		assignmentOperationService:             assignmentOperationService,
		notificationsService:                   notificationsService,
		constraintEngine:                       constraintEngine,
		runtimeTypeLabelKey:                    runtimeTypeLabelKey,
		applicationTypeLabelKey:                applicationTypeLabelKey,
		asaEngine:                              NewASAEngine(asaRepo, runtimeRepo, runtimeContextRepo, formationRepository, formationTemplateRepository, runtimeTypeLabelKey, applicationTypeLabelKey),
		webhookRepository:                      webhookRepository,
		statusService:                          statusService,
	}
}

// Used for testing
//
//go:generate mockery --exported --name=processFunc --output=automock --outpkg=automock --case=underscore --disable-version-string
type processFunc interface { //nolint
	ProcessScenarioFunc(context.Context, string, string, graphql.FormationObjectType, model.Formation) (*model.Formation, error)
}

// ProcessScenarioFunc provides the signature for functions that process scenarios
type ProcessScenarioFunc func(context.Context, string, string, graphql.FormationObjectType, model.Formation) (*model.Formation, error)

// List returns paginated Formations based on pageSize and cursor
func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationPage, error) {
	formationTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.formationRepository.List(ctx, formationTenant, pageSize, cursor)
}

// ListFormationsForObject returns all Formations that `objectID` is part of
func (s *service) ListFormationsForObject(ctx context.Context, objectID string) ([]*model.Formation, error) {
	assignments, err := s.formationAssignmentService.ListAllForObjectGlobal(ctx, objectID)
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

	return s.formationRepository.ListByIDsGlobal(ctx, uniqueFormationIDs)
}

// Get returns the Formation by its id
func (s *service) Get(ctx context.Context, id string) (*model.Formation, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	formation, err := s.formationRepository.Get(ctx, id, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formation with ID %q", id)
	}

	return formation, nil
}

// GetFormationByName returns the Formation by its name
func (s *service) GetFormationByName(ctx context.Context, formationName, tnt string) (*model.Formation, error) {
	f, err := s.formationRepository.GetByName(ctx, formationName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by name: %q: %v", formationName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by name: %q", formationName)
	}

	return f, nil
}

// GetGlobalByID retrieves formation by `id` globally
func (s *service) GetGlobalByID(ctx context.Context, id string) (*model.Formation, error) {
	f, err := s.formationRepository.GetGlobalByID(ctx, id)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by ID: %q globally", id)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by ID: %q globally", id)
	}

	return f, nil
}

func (s *service) Update(ctx context.Context, model *model.Formation) error {
	if err := s.formationRepository.Update(ctx, model); err != nil {
		log.C(ctx).Errorf("An error occurred while updating formation with ID: %q", model.ID)
		return errors.Wrapf(err, "An error occurred while updating formation with ID: %q", model.ID)
	}
	return nil
}

// GetFormationsForObject returns slice of formations for entity with ID objID and type objType
func (s *service) GetFormationsForObject(ctx context.Context, tnt string, objType model.LabelableObject, objID string) ([]string, error) {
	labelInput := &model.LabelInput{
		Key:        model.ScenariosKey,
		ObjectID:   objID,
		ObjectType: objType,
	}
	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching scenario label for %q with id %q", objType, objID)
	}

	return label.ValueToStringsSlice(existingLabel.Value)
}

// CreateFormation is responsible for a couple of things:
//   - Enforce any "pre" and "post" operation formation constraints
//   - Adds the provided formation to the scenario label definitions of the given tenant, if the scenario label definition does not exist it will be created
//   - Creates a new Formation entity based on the provided template name or the default one is used if it's not provided
//   - Generate and send notification(s) if the template from which the formation is created has a webhook attached. And maintain a state based on the executed formation notification(s) - either synchronous or asynchronous
func (s *service) CreateFormation(ctx context.Context, tnt string, formation model.Formation, templateName string) (*model.Formation, error) {
	fTmpl, err := s.formationTemplateRepository.GetByNameAndTenant(ctx, templateName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by name: %q: %v", templateName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by name: %q", templateName)
	}

	formationName := formation.Name
	formationTemplateID := fTmpl.ID
	formationTemplateName := fTmpl.Name

	CRUDJoinPointDetails := &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       templateName,
		FormationTemplateID: formationTemplateID,
		FormationName:       formationName,
		TenantID:            tnt,
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreCreate, CRUDJoinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.CreateFormationOperation, model.PreOperation)
	}

	if err := s.modifyFormations(ctx, tnt, formationName, addFormation); err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, err
		}
		if err = s.labelDefService.CreateWithFormations(ctx, tnt, []string{formationName}); err != nil {
			return nil, err
		}
	}

	formationTemplateWebhooks, err := s.webhookRepository.ListByReferenceObjectIDGlobal(ctx, formationTemplateID, model.FormationTemplateWebhookReference)
	if err != nil {
		return nil, errors.Wrapf(err, "when listing formation lifecycle webhooks for formation template with ID: %q", formationTemplateID)
	}

	formationState := determineFormationState(ctx, formationTemplateID, formationTemplateName, formationTemplateWebhooks, formation.State)

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	newFormation, err := s.createFormation(ctx, tnt, formationTemplateID, formationName, formationState)
	if err != nil {
		return nil, err
	}

	if newFormation.State == model.DraftFormationState {
		log.C(ctx).Infof("The formation is created with %s state. No Lifecycle notification will be executed until the formation is finalized", newFormation.State)
	} else {
		formationReqs, err := s.notificationsService.GenerateFormationNotifications(ctx, formationTemplateWebhooks, tnt, newFormation, formationTemplateName, formationTemplateID, model.CreateFormation)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for formation with ID: %q and name: %q", newFormation.ID, newFormation.Name)
		}

		for _, formationReq := range formationReqs {
			if err := s.processFormationNotifications(ctx, newFormation, formationReq, model.CreateErrorFormationState); err != nil {
				processErr := errors.Wrapf(err, "while processing notifications for formation with ID: %q and name: %q", newFormation.ID, newFormation.Name)
				log.C(ctx).Error(processErr)
				return nil, processErr
			}
		}
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostCreate, CRUDJoinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.CreateFormationOperation, model.PostOperation)
	}

	return newFormation, nil
}

// DeleteFormation is responsible for a couple of things:
//   - Enforce any "pre" and "post" operation formation constraints
//   - Generate and send notification(s) if the template from which the formation is created has a webhook attached. And maintain a state based on the executed formation notification(s) - either synchronous or asynchronous
//   - Removes the provided formation from the scenario label definitions of the given tenant and deletes the formation entity from the DB
func (s *service) DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	ft, err := s.getFormationWithTemplate(ctx, formation.Name, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while deleting formation")
	}

	formationID := ft.formation.ID
	formationName := ft.formation.Name
	formationTemplateID := ft.formationTemplate.ID
	formationTemplateName := ft.formationTemplate.Name

	joinPointDetails := &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       formationTemplateName,
		FormationTemplateID: formationTemplateID,
		FormationName:       formationName,
		TenantID:            tnt,
	}

	assignmentsForFormation, err := s.formationAssignmentService.GetAssignmentsForFormation(ctx, tnt, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignments for formation with ID %q", formationID)
	}

	if len(assignmentsForFormation) > 0 {
		return nil, errors.Errorf("cannot delete formation with ID %q, because it is not empty", formationID)
	}

	asa, err := s.asaService.GetForScenarioName(ctx, formationName)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return nil, errors.Wrapf(err, "while getting automatic scenario assignment for formation with name %q", formationName)
	}
	if asa != nil {
		return nil, errors.Errorf("cannot delete formation with ID %q, because there is still a subaccount part of it", formationID)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreDelete, joinPointDetails, formationTemplateID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.DeleteFormationOperation, model.PreOperation)
	}

	hasWebhook := false
	if formation.State == model.DraftFormationState {
		log.C(ctx).Infof("Formation is in %q state. Skipping notifications...", model.DraftFormationState)
	} else {
		formationTemplateWebhooks, err := s.webhookRepository.ListByReferenceObjectIDGlobal(ctx, formationTemplateID, model.FormationTemplateWebhookReference)
		if err != nil {
			return nil, errors.Wrapf(err, "when listing formation lifecycle webhooks for formation template with ID: %q", formationTemplateID)
		}

		if len(formationTemplateWebhooks) > 0 {
			hasWebhook = true
		}

		formationReqs, err := s.notificationsService.GenerateFormationNotifications(ctx, formationTemplateWebhooks, tnt, ft.formation, formationTemplateName, formationTemplateID, model.DeleteFormation)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for formation with ID: %q and name: %q", formationID, formationName)
		}

		for _, formationReq := range formationReqs {
			if err := s.processFormationNotifications(ctx, ft.formation, formationReq, model.DeleteErrorFormationState); err != nil {
				processErr := errors.Wrapf(err, "while processing notifications for formation with ID: %q and name: %q", formationID, formationName)
				log.C(ctx).Error(processErr)
				return nil, processErr
			}
		}
	}

	if !hasWebhook || ft.formation.State == model.ReadyFormationState || ft.formation.State == model.DraftFormationState {
		if err := s.DeleteFormationEntityAndScenarios(ctx, tnt, formationName); err != nil {
			return nil, errors.Wrapf(err, "An error occurred while deleting formation entity with name: %q and its scenarios label", formationName)
		}

		if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostDelete, joinPointDetails, formationTemplateID); err != nil {
			return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.DeleteFormationOperation, model.PostOperation)
		}
	}

	return ft.formation, nil
}

// DeleteFormationEntityAndScenarios removes the formation name from scenarios label definitions and deletes the formation entity from the DB
func (s *service) DeleteFormationEntityAndScenarios(ctx context.Context, tnt, formationName string) error {
	if err := s.modifyFormations(ctx, tnt, formationName, deleteFormation); err != nil {
		return err
	}

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	if err := s.formationRepository.DeleteByName(ctx, tnt, formationName); err != nil {
		log.C(ctx).Errorf("An error occurred while deleting formation with name: %q", formationName)
		return errors.Wrapf(err, "An error occurred while deleting formation with name: %q", formationName)
	}

	return nil
}

// AssignFormation assigns object based on graphql.FormationObjectType.
//
// When objectType graphql.FormationObjectType is graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime and
// graphql.FormationObjectTypeRuntimeContext it adds the provided formation to the scenario label of the entity if such exists,
// otherwise new scenario label is created for the entity with the provided formation.
//
// FormationAssignments for the object that is being assigned and the already assigned objects are generated and stored.
// For each object X already part of the formation formationAssignment with source=X and target=objectID and formationAssignment
// with source=objectID and target=X are generated.
//
// Additionally, notifications are sent to the interested participants for that formation change.
//   - If objectType is graphql.FormationObjectTypeApplication:
//   - A notification about the assigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//   - A notification about the assigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//   - If the assigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
//   - If the assigned application has a configuration change webhook, a notification about each runtime/runtimeContext in the formation is sent to this application.
//   - If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext:
//   - If the assigned runtime/runtimeContext has configuration change webhook, a notification about each application in the formation is sent to this runtime.
//   - A notification about the assigned runtime/runtimeContext is sent to all the applications that are in the formation and have configuration change webhook.
//
// If an error occurs during the formationAssignment processing the failed formationAssignment's value is updated with the error and the processing proceeds. The error should not
// be returned but only logged. If the error is returned the assign operation will be rolled back and all the created resources(labels, formationAssignments etc.) will be rolled
// back. On the other hand the participants in the formation will have been notified for the assignment and there is no mechanism for informing them that the assignment was not executed successfully.
//
// After the assigning there may be formationAssignments in CREATE_ERROR state. They can be fixed by assigning the object to the same formation again. This will result in retrying only
// the formationAssignments that are in state different from READY.
//
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// create automatic scenario assignment with the caller and target tenant which then will assign the right Runtime / RuntimeContexts based on the formation template's runtimeType.
func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (f *model.Formation, err error) {
	log.C(ctx).Infof("Assigning object with ID %q of type %q to formation %q", objectID, objectType, formation.Name)

	ft, err := s.getFormationWithTemplate(ctx, formation.Name, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while assigning formation with name %q", formation.Name)
	}

	if !isObjectTypeSupported(ft.formationTemplate, objectType) {
		return nil, errors.Errorf("Formation %q of type %q does not support resources of type %q", ft.formation.Name, ft.formationTemplate.Name, objectType)
	}

	joinPointDetails, err := s.prepareDetailsForAssign(ctx, tnt, objectID, objectType, ft.formation, ft.formationTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing joinpoint details for target operation %q and constraint type %q", model.AssignFormationOperation, model.PreOperation)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreAssign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.AssignFormationOperation, model.PreOperation)
	}

	formationFromDB := ft.formation
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		// If we assign it to the label definitions when it is in deleting state we risk leaving incorrect data
		// in the LabelDefinition and formation assignments and failing to delete the formation later on
		if formationFromDB.State == model.DeletingFormationState || formationFromDB.State == model.DeleteErrorFormationState {
			return nil, fmt.Errorf("cannot assign to formation with ID %q as it is in %q state", formationFromDB.ID, formationFromDB.State)
		}

		err = s.assign(ctx, tnt, objectID, objectType, formationFromDB, ft.formationTemplate)
		if err != nil {
			return nil, err
		}

		// The defer statement after the formation assignment persistence depends on the value of the variable err.
		// If 'err' is used for the name of the returned error, a new variable that shadows the 'err' variable from the outer scope
		// is created. As the defer statement is declared in the scope of the case fragment of the switch it will be bound to the 'err' variable in the same scope
		// which is the new one. Then the deffer will not execute its logic in case of error in the outer scope.
		assignmentInputs, terr := s.formationAssignmentService.GenerateAssignments(ctx, tnt, objectID, objectType, formationFromDB)
		if terr != nil {
			return nil, terr
		}

		// We need to persist the FAs before we proceed to notification processing as for scenarios where there are both
		// participants with ASYNC notifications and SYNC notifications it is possible that a FA status update request is
		// received before the FA is persisted in the database.
		// Example: Participant 1(P1) has ASYNC webhook(W1) and Participant2(P2) has SYNC webhook(W2). Assign both P1 and P2 to a formation.
		// Formation assignments are generated.
		// Execute W1, then execute W2. W2 takes, for example, 10 seconds. Before the W2 processing finishes, an FA status update request for W1 is received.
		// When fetching the corresponding FA from the DB we get an object not found error, as the status update is performed in new transaction and the transaction in which the FA were generated is still running.
		var assignments []*model.FormationAssignment
		if terr = s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			assignments, terr = s.formationAssignmentService.PersistAssignments(ctxWithTransact, tnt, assignmentInputs)
			if terr != nil {
				return terr
			}

			return nil
		}); terr != nil {
			return nil, terr
		}

		// todo::: delete/update below comments
		// create operations in a transaction similar to the FAs above (using terr as well)
		// we want them in a separate transaction similar to the FAs case - if someone reports on the status API and we try to get the operations we have to be sure that the operation will be persisted so that we don't get not found error
		if terr = s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			for _, a := range assignments {
				if _, terr = s.assignmentOperationService.Create(ctxWithTransact, &model.AssignmentOperationInput{
					Type:                  model.Assign,
					FormationAssignmentID: a.ID,
					FormationID:           a.FormationID,
					TriggeredBy:           model.AssignObject,
				}); terr != nil {
					return errors.Wrapf(terr, "while creating %s Operation for assignment with ID: %s", model.Assign, a.ID)
				}
			}

			return nil
		}); terr != nil {
			return nil, terr
		}

		// If the assigning of the object fails, the transaction opened in the resolver will be rolled back.
		// The FA records and the labels for the object will not be reverted as they were persisted as part of another
		// transaction. The leftover resources should be deleted separately.
		defer func() {
			if err == nil {
				return
			}

			log.C(ctx).Infof("Failed to assign object with ID %q of type %q to formation %q. Deleting Created Formation Assignment records...", objectID, objectType, formation.Name)
			if terr = s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
				if deleteErr := s.formationAssignmentService.DeleteAssignmentsForObjectID(ctxWithTransact, formationFromDB.ID, objectID); deleteErr != nil {
					log.C(ctx).WithError(deleteErr).Errorf("Failed to delete assignments fo object with ID %q of type %q to formation %q", objectID, objectType, formation.Name)
					return deleteErr
				}

				return nil
			}); terr != nil {
				log.C(ctx).Error(terr)
			}
		}()

		// When it is in initial state, the notification generation will be handled by the async API via resynchronizing the formation later
		// If we are in create error or draft state, the formation is not ready, and we should not send notifications
		if formationFromDB.State == model.InitialFormationState || formationFromDB.State == model.CreateErrorFormationState || formationFromDB.State == model.DraftFormationState {
			log.C(ctx).Infof("Formation with id %q is not in %q state. Waiting for state to be updated...", formationFromDB.ID, model.ReadyFormationState)
			return ft.formation, nil
		}

		requests, terr := s.notificationsService.GenerateFormationAssignmentNotifications(ctx, tnt, objectID, formationFromDB, model.AssignFormation, objectType)
		err = terr
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s assignment", objectType)
		}

		if err = s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			if err = s.formationAssignmentService.ProcessFormationAssignments(ctxWithTransact, assignments, requests, s.formationAssignmentService.ProcessFormationAssignmentPair, model.AssignFormation); err != nil {
				log.C(ctxWithTransact).Errorf("Error occurred while processing formationAssignments %s", err.Error())
				return err
			}

			return nil
		}); err != nil {
			return nil, err
		}

	case graphql.FormationObjectTypeTenant:
		targetTenantID, err := s.tenantSvc.GetInternalTenant(ctx, objectID)
		if err != nil {
			return nil, err
		}

		if _, err = s.CreateAutomaticScenarioAssignment(ctx, newAutomaticScenarioAssignmentModel(formationFromDB.Name, tnt, targetTenantID)); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostAssign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.AssignFormationOperation, model.PostOperation)
	}

	return formationFromDB, nil
}

func (s *service) prepareDetailsForAssign(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, formationTemplate *model.FormationTemplate) (*formationconstraint.AssignFormationOperationDetails, error) {
	resourceSubtype, err := s.getObjectSubtype(ctx, tnt, objectID, objectType)
	if err != nil {
		return nil, err
	}

	joinPointDetails := &formationconstraint.AssignFormationOperationDetails{
		ResourceType:        model.ResourceType(objectType),
		ResourceSubtype:     resourceSubtype,
		ResourceID:          objectID,
		FormationType:       formationTemplate.Name,
		FormationTemplateID: formationTemplate.ID,
		FormationID:         formation.ID,
		FormationName:       formation.Name,
		TenantID:            tnt,
	}
	return joinPointDetails, nil
}

func (s *service) prepareDetailsForUnassign(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, formationTemplate *model.FormationTemplate) (*formationconstraint.UnassignFormationOperationDetails, error) {
	resourceSubtype, err := s.getObjectSubtype(ctx, tnt, objectID, objectType)
	if err != nil {
		return nil, err
	}

	joinPointDetails := &formationconstraint.UnassignFormationOperationDetails{
		ResourceType:        model.ResourceType(objectType),
		ResourceSubtype:     resourceSubtype,
		ResourceID:          objectID,
		FormationType:       formationTemplate.Name,
		FormationTemplateID: formationTemplate.ID,
		FormationID:         formation.ID,
		TenantID:            tnt,
	}
	return joinPointDetails, nil
}

func (s *service) getObjectSubtype(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType) (string, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		applicationTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.applicationTypeLabelKey,
			ObjectID:   objectID,
			ObjectType: model.ApplicationLabelableObject,
		})
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return "", nil
			}
			return "", errors.Wrapf(err, "while getting label %q for application with ID %q", s.applicationTypeLabelKey, objectID)
		}

		applicationType, ok := applicationTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing application type for application %q", objectID)
		}
		return applicationType, nil

	case graphql.FormationObjectTypeRuntime:
		runtimeTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.runtimeTypeLabelKey,
			ObjectID:   objectID,
			ObjectType: model.RuntimeLabelableObject,
		})
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return "", nil
			}
			return "", errors.Wrapf(err, "while getting label %q for runtime with ID %q", s.runtimeTypeLabelKey, objectID)
		}

		runtimeType, ok := runtimeTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing runtime type for runtime %q", objectID)
		}
		return runtimeType, nil

	case graphql.FormationObjectTypeRuntimeContext:
		rtmCtx, err := s.runtimeContextRepo.GetByID(ctx, tnt, objectID)
		if err != nil {
			return "", errors.Wrapf(err, "while fetching runtime context with ID %q", objectID)
		}

		runtimeTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
			Key:        s.runtimeTypeLabelKey,
			ObjectID:   rtmCtx.RuntimeID,
			ObjectType: model.RuntimeLabelableObject,
		})
		if err != nil {
			return "", errors.Wrapf(err, "while getting label %q for runtime with ID %q", s.runtimeTypeLabelKey, objectID)
		}

		runtimeType, ok := runtimeTypeLabel.Value.(string)
		if !ok {
			return "", errors.Errorf("Missing runtime type for runtime %q", rtmCtx.RuntimeID)
		}
		return runtimeType, nil

	case graphql.FormationObjectTypeTenant:
		t, err := s.tenantSvc.GetTenantByExternalID(ctx, objectID)
		if err != nil {
			return "", errors.Wrapf(err, "while getting tenant by external ID")
		}

		return string(t.Type), nil

	default:
		return "", fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) assign(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation, formationTemplate *model.FormationTemplate) error {
	if err := s.checkFormationTemplateTypes(ctx, tnt, objectID, objectType, formationTemplate); err != nil {
		return err
	}

	if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation.Name, objectTypeToLabelableObject(objectType), addFormation); err != nil {
		if apperrors.IsNotFoundError(err) {
			labelInput := newLabelInput(formation.Name, objectID, objectTypeToLabelableObject(objectType))
			if err = s.labelService.CreateLabel(ctx, tnt, s.uuidService.Generate(), labelInput); err != nil {
				return err
			}
			return nil
		}
		return err
	}

	return nil
}

// UnassignFromScenarioLabel unassigns object from scenario label
func (s *service) UnassignFromScenarioLabel(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation) error {
	if objectType == graphql.FormationObjectTypeApplication ||
		objectType == graphql.FormationObjectTypeRuntime ||
		objectType == graphql.FormationObjectTypeRuntimeContext {
		return s.modifyAssignedFormations(ctx, tnt, objectID, formation.Name, objectTypeToLabelableObject(objectType), deleteFormation)
	}
	return nil
}

func (s *service) checkFormationTemplateTypes(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formationTemplate *model.FormationTemplate) error {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		app, err := s.applicationRepository.GetByID(ctx, tnt, objectID)
		if err != nil {
			return errors.Wrapf(err, "while getting application with ID: %q", objectID)
		}
		if err := s.isValidApplicationType(ctx, tnt, objectID, formationTemplate); err != nil {
			return errors.Wrapf(err, "while validating application type for application %q", objectID)
		}
		if err := s.isValidApplication(app); err != nil {
			return errors.Wrapf(err, "while validating application with ID: %q", objectID)
		}
	case graphql.FormationObjectTypeRuntime:
		if _, err := s.runtimeRepo.GetByID(ctx, tnt, objectID); err != nil {
			return errors.Wrapf(err, "while getting runtime with ID: %q", objectID)
		}
		if err := s.isValidRuntimeType(ctx, tnt, objectID, formationTemplate); err != nil {
			return errors.Wrapf(err, "while validating runtime type")
		}
	case graphql.FormationObjectTypeRuntimeContext:
		runtimeCtx, err := s.runtimeContextRepo.GetByID(ctx, tnt, objectID)
		if err != nil {
			return errors.Wrapf(err, "while getting runtime context")
		}
		if err = s.isValidRuntimeType(ctx, tnt, runtimeCtx.RuntimeID, formationTemplate); err != nil {
			return errors.Wrapf(err, "while validating runtime type of runtime")
		}
	}
	return nil
}

// UnassignFormation unassigns object base on graphql.FormationObjectType.
//
// For objectType graphql.FormationObjectTypeApplication it removes the provided formation from the
// scenario label of the application.
//
// For objectTypes graphql.FormationObjectTypeRuntime and graphql.FormationObjectTypeRuntimeContext
// it removes the formation from the scenario label of the runtime/runtime context if the provided
// formation is NOT assigned from ASA and does nothing if it is assigned from ASA.
//
//	 Additionally, notifications are sent to the interested participants for that formation change.
//			- If objectType is graphql.FormationObjectTypeApplication:
//					- A notification about the unassigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//	 			- A notification about the unassigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//					- If the unassigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
//					- If the unassigned application has a configuration change webhook, a notification about each runtime/runtimeContext in the formation is sent to this application.
//			- If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext:
//					- If the unassigned runtime/runtimeContext has configuration change webhook, a notification about each application in the formation is sent to this runtime.
//	  			- A notification about the unassigned runtime/runtimeContext is sent to all the applications that are in the formation and have configuration change webhook.
//
// For the formationAssignments that have their source or target field set to objectID:
//   - If the formationAssignment does not have notification associated with it
//   - the formation assignment is deleted
//   - If the formationAssignment is associated with a notification
//   - If the response from the notification is success
//   - the formationAssignment is deleted
//   - If the response from the notification is different from success
//   - the formation assignment is updated with an error
//
// After the processing of the formationAssignments the state is persisted regardless of whether there were any errors.
// If an error has occurred during the formationAssignment processing the unassign operation is rolled back(the updated
// with the error formationAssignments are already persisted in the database).
//
// For objectType graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant which then will unassign the right Runtime / RuntimeContexts based on the formation template's runtimeType.
func (s *service) UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (f *model.Formation, err error) {
	log.C(ctx).Infof("Unassigning object with ID: %q of type: %q from formation %q", objectID, objectType, formation.Name)

	if !isObjectTypeAllowed(objectType) {
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}

	formationName := formation.Name
	ft, err := s.getFormationWithTemplate(ctx, formationName, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while unassigning formation with name %q", formationName)
	}

	formationFromDB := ft.formation

	if isFormationComingFromASA, err := s.asaEngine.IsFormationComingFromASA(ctx, objectID, formation.Name, objectType); err != nil {
		return nil, err
	} else if isFormationComingFromASA {
		return formationFromDB, nil
	}

	joinPointDetails, err := s.prepareDetailsForUnassign(ctx, tnt, objectID, objectType, ft.formation, ft.formationTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing joinpoint details for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PreOperation)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreUnassign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PreOperation)
	}

	if objectType == graphql.FormationObjectTypeTenant {
		asa, err := s.asaService.GetForScenarioName(ctx, formationName)
		if err != nil {
			return nil, err
		}
		if err = s.DeleteAutomaticScenarioAssignment(ctx, asa); err != nil {
			return nil, err
		}

		if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostUnassign, joinPointDetails, ft.formationTemplate.ID); err != nil {
			return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PostOperation)
		}

		return formationFromDB, nil
	}

	initialAssignmentsData, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(ctx, formationFromDB.ID, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing formation assignments for object with type %q and ID %q", objectType, objectID)
	}

	// We can reach this only if we are in INITIAL or DRAFT state and there are assigned objects to the formation
	// there are no notifications sent for them, and we have created formation assignments for them.
	// If we by any chance reach it from ERROR state, the formation should be empty, with no formation assignments in it, and the deletion shouldn't do anything.
	if formationFromDB.State != model.ReadyFormationState {
		log.C(ctx).Infof("Formation with id %q is not in %q state. Waiting for response on status API before sending notifications...", formationFromDB.ID, model.ReadyFormationState)
		err = s.formationAssignmentService.DeleteAssignmentsForObjectID(ctx, formationFromDB.ID, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while deleting formationAssignments for object with type %q and ID %q", objectType, objectID)
		}
		err = s.UnassignFromScenarioLabel(ctx, tnt, objectID, objectType, formationFromDB)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while unassigning from formation")
		}
		return ft.formation, nil
	}

	// We need the formation assignments data before updating them to DELETING state (and resetting their configuration) in the transaction below,
	// so in case of any failures that can happen before the processing of these formation assignments (e.g. generating notifications for them),
	// we could revert the changes made in the transaction below
	initialAssignmentsClones := make([]*model.FormationAssignment, 0, len(initialAssignmentsData))
	initialParticipants := make(map[string]bool, len(initialAssignmentsData)*2)
	// If by any chance we are coming from the status API or are being called to clean up the scenario label,
	// we still want to check for the object that is unassigned
	initialParticipants[objectID] = true
	for _, ia := range initialAssignmentsData {
		initialAssignmentsClones = append(initialAssignmentsClones, ia.Clone())
		initialParticipants[ia.Source] = true
		initialParticipants[ia.Target] = true
	}

	// Flag that is used to determine whether to revert the changes made in the transaction below or not.
	// If any errors occur after we committed the transaction below but before the formation assignments processing, we need to revert them.
	// If errors occur after formation assignments processing, we should not revert them
	shouldRevertAssignmentsUpdateFromTransaction := true

	// In case of Unassign, we want to persist the formation assignments in DELETING state and resetting their configuration and error in isolated transaction
	// similar to the Assign operation and to cover the case when we have both type of webhook - sync and async.
	// So if the async notification was sent, and we're processing the sync notification, meanwhile the async participant sends
	// FA status update request, he will have the latest state of the formation assignment even when we didn't finish the sync notification processing.
	if err := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
		for _, ia := range initialAssignmentsData {
			if ia.SetStateToDeleting() {
				log.C(ctx).Infof("Update and persist in the DB '%s' state of formation assignment with ID: '%s'", ia.State, ia.ID)
				formationassignment.ResetAssignmentConfigAndError(ia)
				if err := s.formationAssignmentService.Update(ctxWithTransact, ia.ID, ia); err != nil {
					return errors.Wrapf(err, "while updating formation assignment with ID: '%s' to '%s' state", ia.ID, ia.State)
				}
			} else {
				log.C(ctx).Infof("State of formation assignment with ID %q is already in '%s', proceeding without updating it", ia.ID, ia.State)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	// create `Unassign` Operation here in separate transaction similar to the `Assign` case
	var operationAssignmentIDs []string
	if opErr := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
		for _, ia := range initialAssignmentsData {
			opID, err := s.assignmentOperationService.Create(ctxWithTransact, &model.AssignmentOperationInput{
				Type:                  model.Unassign,
				FormationAssignmentID: ia.ID,
				FormationID:           ia.FormationID,
				TriggeredBy:           model.UnassignObject,
			})
			if err != nil {
				return errors.Wrapf(err, "while creating %s Operation for assignment with ID: %s", model.Unassign, ia.ID)
			}
			operationAssignmentIDs = append(operationAssignmentIDs, opID)
		}

		return nil
	}); opErr != nil {
		return nil, opErr
	}

	defer func() {
		if err == nil {
			return
		}

		if !shouldRevertAssignmentsUpdateFromTransaction {
			return
		}

		log.C(ctx).Infof("Reverting formation assignment changes that updated them to DELETING state and reset their configuration in the first transaction...")
		if terr := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			for _, FAClone := range initialAssignmentsClones {
				if updateErr := s.formationAssignmentService.Update(ctxWithTransact, FAClone.ID, FAClone); updateErr != nil {
					log.C(ctx).WithError(updateErr).Errorf("while updating formation assignment with ID: %s", FAClone.ID)
					return updateErr
				}
			}

			return nil
		}); terr != nil {
			log.C(ctx).WithError(terr).Error("An error occurred while reverting formation assignments with their original data")
		}

		log.C(ctx).Infof("Deleting Operations related to assignments that were updated to DELETING state...")
		if terr := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			if deleteErr := s.assignmentOperationService.DeleteByIDs(ctxWithTransact, operationAssignmentIDs); deleteErr != nil {
				return deleteErr
			}
			return nil
		}); terr != nil {
			log.C(ctx).WithError(terr).Errorf("while deleting Operations for formation assignments with IDs: %s", operationAssignmentIDs)
		}
	}()

	// It is important that all operations regarding formation assignments should be in the inner transaction
	tx, terr := s.transact.Begin()
	err = terr
	if err != nil {
		return nil, err
	}
	transactionCtx := persistence.SaveToContext(ctx, tx)
	defer s.transact.RollbackUnlessCommitted(transactionCtx, tx)

	requests, nerr := s.notificationsService.GenerateFormationAssignmentNotifications(transactionCtx, tnt, objectID, formationFromDB, model.UnassignFormation, objectType)
	err = nerr
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			if commitErr := tx.Commit(); commitErr != nil {
				err = errors.Wrapf(
					errors.Wrapf(commitErr, "while committing transaction"),
					errors.Wrapf(nerr, "while generating formation assignment notifications").Error(),
				)
				return nil, err
			}

			return formationFromDB, nil
		}
		return nil, errors.Wrapf(err, "while generating notifications for %s unassignment", objectType)
	}

	if nerr = s.formationAssignmentService.ProcessFormationAssignments(transactionCtx, initialAssignmentsData, requests, s.formationAssignmentService.CleanupFormationAssignment, model.UnassignFormation); nerr != nil {
		err = nerr
		if commitErr := tx.Commit(); commitErr != nil {
			err = errors.Wrapf(
				errors.Wrapf(commitErr, "while committing transaction"),
				errors.Wrapf(nerr, "while processing formation assignments").Error(),
			)
			return nil, err
		}

		shouldRevertAssignmentsUpdateFromTransaction = false
		return nil, err
	}

	nerr = tx.Commit()
	err = nerr
	if err != nil {
		return nil, errors.Wrapf(err, "while committing transaction")
	}
	// The formation assignments processing executed in the current transaction has finished,
	// and if an error occurred in any of the operations executed afterward should not revert the initially updated formation assignments in the first transaction.
	shouldRevertAssignmentsUpdateFromTransaction = false

	// It is important that we have committed the previous transaction before formation assignments are listed
	// They could be deleted by either it or another operation altogether (e.g. async API status report)
	scenarioTx, terr := s.transact.Begin()
	err = terr
	if terr != nil {
		return nil, terr
	}
	scenarioTransactionCtx := persistence.SaveToContext(ctx, scenarioTx)
	defer s.transact.RollbackUnlessCommitted(scenarioTransactionCtx, scenarioTx)

	for participantID := range initialParticipants {
		pendingAsyncAssignments, nerr := s.formationAssignmentService.ListFormationAssignmentsForObjectID(scenarioTransactionCtx, formationFromDB.ID, participantID)
		err = nerr
		if err != nil {
			return nil, errors.Wrapf(err, "while listing formationAssignments for object with type %q and ID %q", objectType, participantID)
		}

		if len(pendingAsyncAssignments) == 0 {
			log.C(ctx).Infof("There are no formation assignments left for formation with ID: %q. Unassigning the object with type %q and ID %q from formation %q", formationFromDB.ID, objectType, objectID, formationFromDB.ID)
			err = s.UnassignFromScenarioLabel(scenarioTransactionCtx, tnt, participantID, objectType, formationFromDB)
			if err != nil && !apperrors.IsNotFoundError(err) {
				return nil, errors.Wrapf(err, "while unassigning from formation")
			}
		}
	}

	err = scenarioTx.Commit()
	if err != nil {
		return nil, errors.Wrapf(err, "while committing transaction")
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostUnassign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "while enforcing constraints for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PostOperation)
	}

	return formationFromDB, nil
}

// FinalizeDraftFormation changes the formation state do initial and start processing the formation and formation assignment notifications
func (s *service) FinalizeDraftFormation(ctx context.Context, formationID string) (*model.Formation, error) {
	log.C(ctx).Infof("Finalizing formation with ID: %q", formationID)
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	formation, err := s.formationRepository.Get(ctx, formationID, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation with ID %q for tenant %q", formationID, tenantID)
	}

	if formation.State != model.DraftFormationState {
		return nil, errors.Errorf("The formation with ID %s is not in %s state", formationID, model.DraftFormationState)
	}

	fTmpl, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "An error occurred while getting formation template with ID: %q", formation.FormationTemplateID)
	}

	formationTemplateID := fTmpl.ID

	formationTemplateWebhooks, err := s.webhookRepository.ListByReferenceObjectIDGlobal(ctx, formationTemplateID, model.FormationTemplateWebhookReference)
	if err != nil {
		return nil, errors.Wrapf(err, "when listing formation lifecycle webhooks for formation template with ID: %q", formationTemplateID)
	}

	newState := model.ReadyFormationState
	if len(formationTemplateWebhooks) > 0 {
		newState = model.InitialFormationState
	}

	log.C(ctx).Infof("Setting formation with ID %s to %s state and starting resynchronization", formationID, newState)
	formation.State = newState

	formationStateTx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	formationStateTransactionCtx := persistence.SaveToContext(ctx, formationStateTx)

	defer s.transact.RollbackUnlessCommitted(formationStateTransactionCtx, formationStateTx)

	if err = s.Update(formationStateTransactionCtx, formation); err != nil {
		return nil, err
	}

	err = formationStateTx.Commit()
	if err != nil {
		return nil, err
	}

	return s.resynchronizeFormation(ctx, formation, tenantID, false)
}

// ResynchronizeFormationNotifications sends all notifications that are in error or initial state
func (s *service) ResynchronizeFormationNotifications(ctx context.Context, formationID string, shouldReset bool) (*model.Formation, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	formation, err := s.formationRepository.Get(ctx, formationID, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation with ID %q for tenant %q", formationID, tenantID)
	}
	log.C(ctx).Infof("Resynchronizing formation with ID: %s and name: %s", formationID, formation.Name)

	return s.resynchronizeFormation(ctx, formation, tenantID, shouldReset)
}

func (s *service) resynchronizeFormation(ctx context.Context, formation *model.Formation, tenantID string, shouldReset bool) (*model.Formation, error) {
	if formation.State == model.DraftFormationState {
		return nil, errors.Errorf("Formations in state %s can not be resynchronized", model.DraftFormationState)
	}

	if formation.State != model.ReadyFormationState {
		previousState := formation.State
		resynchronizedFormation, isDeleted, err := s.resynchronizeFormationNotifications(ctx, tenantID, formation, previousState)
		if err != nil {
			return nil, errors.Wrapf(err, "while resynchronizing formation notifications for formation with ID %q", formation.ID)
		}
		if isDeleted {
			return resynchronizedFormation, nil
		}
		if resynchronizedFormation.State != model.ReadyFormationState {
			return resynchronizedFormation, nil
		}
	}

	return s.resynchronizeFormationAssignmentNotifications(ctx, tenantID, formation, shouldReset)
}

func (s *service) resynchronizeFormationAssignmentNotifications(ctx context.Context, tenantID string, formation *model.Formation, shouldReset bool) (*model.Formation, error) {
	resyncableFormationAssignments := make([]*model.FormationAssignment, 0)
	failedDeleteErrorFormationAssignments := make([]*model.FormationAssignment, 0)
	assignmentMappingNoNotificationPairs := make([]*formationassignment.AssignmentMappingPairWithOperation, 0)
	assignmentMappingSyncPairs := make([]*formationassignment.AssignmentMappingPairWithOperation, 0)
	assignmentMappingAsyncPairs := make([]*formationassignment.AssignmentMappingPairWithOperation, 0)

	assignmentOperationTriggeredBy := model.ResyncAssignment
	formationID := formation.ID
	if err := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
		if shouldReset {
			formationTemplate, err := s.formationTemplateRepository.Get(ctxWithTransact, formation.FormationTemplateID)
			if err != nil {
				return errors.Wrapf(err, "while getting formation template with ID %q", formation.FormationTemplateID)
			}
			if !formationTemplate.SupportsReset {
				return apperrors.NewInvalidOperationError(fmt.Sprintf("formation template %q does not support resetting", formationTemplate.Name))
			}
			assignmentsForFormation, err := s.formationAssignmentService.GetAssignmentsForFormation(ctxWithTransact, tenantID, formationID)
			if err != nil {
				return errors.Wrapf(err, "while getting formation assignments for formation with ID %q", formationID)
			}
			for _, assignment := range assignmentsForFormation {
				assignment.State = string(model.InitialAssignmentState)
				formationassignment.ResetAssignmentConfigAndError(assignment) // reset the assignments
				err = s.formationAssignmentService.Update(ctxWithTransact, assignment.ID, assignment)
				if err != nil {
					return err
				}
				assignmentOperationTriggeredBy = model.ResetAssignment
				if _, err = s.assignmentOperationService.Create(ctxWithTransact, &model.AssignmentOperationInput{
					Type:                  model.Assign,
					FormationAssignmentID: assignment.ID,
					FormationID:           assignment.FormationID,
					TriggeredBy:           assignmentOperationTriggeredBy,
				}); err != nil {
					return err
				}
			}
		}

		resyncableFAs, err := s.formationAssignmentService.GetAssignmentsForFormationWithStates(ctxWithTransact, tenantID, formationID, model.ResynchronizableFormationAssignmentStates)
		if err != nil {
			return errors.Wrap(err, "while getting formation assignments with synchronizing and error states")
		}

		for _, rfa := range resyncableFAs {
			if rfa.State == string(model.DeleteErrorAssignmentState) {
				failedDeleteErrorFormationAssignments = append(failedDeleteErrorFormationAssignments, rfa)
			}
		}
		resyncableFormationAssignments = resyncableFAs

		for _, fa := range resyncableFormationAssignments {
			operation := fa.GetOperation()
			var notificationForReverseFA *webhookclient.FormationAssignmentNotificationRequest
			notificationForFA, err := s.formationAssignmentNotificationService.GenerateFormationAssignmentNotification(ctxWithTransact, fa, operation)
			if err != nil {
				return err
			}

			reverseFA, err := s.formationAssignmentService.GetReverseBySourceAndTarget(ctxWithTransact, fa.FormationID, fa.Source, fa.Target)
			if err != nil && !apperrors.IsNotFoundError(err) {
				return err
			}
			if reverseFA != nil {
				notificationForReverseFA, err = s.formationAssignmentNotificationService.GenerateFormationAssignmentNotification(ctxWithTransact, reverseFA, operation)
				if err != nil && !apperrors.IsNotFoundError(err) {
					return err
				}
			}

			var reverseReqMapping *formationassignment.FormationAssignmentRequestMapping
			if reverseFA != nil || notificationForReverseFA != nil {
				reverseReqMapping = &formationassignment.FormationAssignmentRequestMapping{
					Request:             notificationForReverseFA,
					FormationAssignment: reverseFA,
				}
			}

			faClone := fa.Clone()
			if notificationForFA != nil && operation == model.UnassignFormation {
				faClone.SetStateToDeleting()
				formationassignment.ResetAssignmentConfigAndError(faClone)
				if err := s.formationAssignmentService.Update(ctxWithTransact, faClone.ID, faClone); err != nil {
					return errors.Wrapf(err, "while updating formation assignment with ID: '%s' to '%s' state", faClone.ID, faClone.State)
				}
				if err := s.assignmentOperationService.Update(ctxWithTransact, faClone.ID, faClone.FormationID, model.Unassign, assignmentOperationTriggeredBy); err != nil {
					return errors.Wrapf(err, "while updating %s Operation for assignment with ID: %s triggered by resync", model.Unassign, faClone.ID)
				}
			}
			if operation == model.AssignFormation {
				faClone.State = string(model.InitialAssignmentState)
				// Cleanup the error if present as new notification will be sent. The previous configuration should be left intact.
				faClone.Error = nil
				if err := s.formationAssignmentService.Update(ctxWithTransact, faClone.ID, faClone); err != nil {
					return errors.Wrapf(err, "while updating formation assignment with ID: '%s' to '%s' state", faClone.ID, faClone.State)
				}
				if err := s.assignmentOperationService.Update(ctxWithTransact, faClone.ID, faClone.FormationID, model.Assign, assignmentOperationTriggeredBy); err != nil {
					return errors.Wrapf(err, "while updating %s Operation for assignment with ID: %s triggered by resync", model.Assign, faClone.ID)
				}
			}

			assignmentPair := formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             notificationForFA,
						FormationAssignment: fa,
					},
					ReverseAssignmentReqMapping: reverseReqMapping,
				},
				Operation: operation,
			}

			// We separate the assignment pairs in 3 groups
			// 1. With no requests for the assignment
			// 2. With synchronous webhook requests for the assignments
			// 3. With asynchronous webhook requests for the assignments
			// We do this, so that we can order the processing of the formation assignments
			// This makes the notification count deterministic (we don't send asynchronous notifications before synchronous ones),
			// and we assure that the notification receivers always receive the reverse as READY,
			// if it has no request associated, rather than being sometimes INITIAL, sometimes READY.
			if notificationForFA == nil {
				assignmentMappingNoNotificationPairs = append(assignmentMappingNoNotificationPairs, &assignmentPair)
			} else if notificationForFA != nil &&
				notificationForFA.Webhook != nil &&
				notificationForFA.Webhook.Mode != nil &&
				*notificationForFA.Webhook.Mode == graphql.WebhookModeAsyncCallback {
				assignmentMappingAsyncPairs = append(assignmentMappingAsyncPairs, &assignmentPair)
			} else {
				assignmentMappingSyncPairs = append(assignmentMappingSyncPairs, &assignmentPair)
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	alreadyProcessedFAs := make(map[string]bool, 0)
	assignmentMappingPairs := append(assignmentMappingNoNotificationPairs, append(assignmentMappingSyncPairs, assignmentMappingAsyncPairs...)...)
	var errs *multierror.Error
	if err := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
		for _, assignmentPair := range assignmentMappingPairs {
			if alreadyProcessedFAs[assignmentPair.AssignmentReqMapping.FormationAssignment.ID] {
				continue
			}
			switch assignmentPair.Operation {
			case model.AssignFormation:
				isReverseProcessed, err := s.formationAssignmentService.ProcessFormationAssignmentPair(ctxWithTransact, assignmentPair)
				if err != nil {
					errs = multierror.Append(errs, err)
				}
				// It is probably impossible for the reverse assignment to be nil if the reverse is processed, but it's better to safeguard against it anyway
				if isReverseProcessed && assignmentPair.ReverseAssignmentReqMapping != nil && assignmentPair.ReverseAssignmentReqMapping.FormationAssignment != nil {
					alreadyProcessedFAs[assignmentPair.ReverseAssignmentReqMapping.FormationAssignment.ID] = true
				}
			case model.UnassignFormation:
				if _, err := s.formationAssignmentService.CleanupFormationAssignment(ctxWithTransact, assignmentPair); err != nil {
					errs = multierror.Append(errs, err)
				}
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if len(failedDeleteErrorFormationAssignments) > 0 {
		objectIDToTypeMap := make(map[string]graphql.FormationObjectType, len(failedDeleteErrorFormationAssignments)*2)
		for _, assignment := range failedDeleteErrorFormationAssignments {
			objectIDToTypeMap[assignment.Source] = formationAssignmentTypeToFormationObjectType(assignment.SourceType)
			objectIDToTypeMap[assignment.Target] = formationAssignmentTypeToFormationObjectType(assignment.TargetType)
		}

		if err := s.executeInTransaction(ctx, func(ctxWithTransact context.Context) error {
			for objectID, objectType := range objectIDToTypeMap {
				leftAssignmentsInFormation, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(ctxWithTransact, formation.ID, objectID)
				if err != nil {
					return errors.Wrapf(err, "while listing formationAssignments for object with type %q and ID %q", objectType, objectID)
				}

				if len(leftAssignmentsInFormation) == 0 {
					log.C(ctx).Infof("There are no formation assignments left for formation with ID: %q. Unassigning the object with type %q and ID %q from formation %q", formation.ID, objectType, objectID, formation.ID)
					err = s.UnassignFromScenarioLabel(ctxWithTransact, tenantID, objectID, objectType, formation)
					if err != nil && !apperrors.IsNotFoundError(err) {
						return errors.Wrapf(err, "while unassigning the object with type %q and ID %q", objectType, objectID)
					}
				}
			}

			return nil
		}); err != nil {
			return nil, err
		}
	}

	return formation, errs.ErrorOrNil()
}

func (s *service) resynchronizeFormationNotifications(ctx context.Context, tenantID string, formation *model.Formation, previousState model.FormationState) (*model.Formation, bool, error) {
	formationResyncTx, err := s.transact.Begin()
	if err != nil {
		return nil, false, err
	}
	formationResyncTransactionCtx := persistence.SaveToContext(ctx, formationResyncTx)

	defer s.transact.RollbackUnlessCommitted(formationResyncTransactionCtx, formationResyncTx)

	fTmpl, err := s.formationTemplateRepository.Get(formationResyncTransactionCtx, formation.FormationTemplateID)
	if err != nil {
		return nil, false, errors.Wrapf(err, "An error occurred while getting formation template with ID: %q", formation.FormationTemplateID)
	}
	formationTemplateID := fTmpl.ID
	formationTemplateName := fTmpl.Name
	formationID := formation.ID

	formationTemplateWebhooks, err := s.webhookRepository.ListByReferenceObjectIDGlobal(formationResyncTransactionCtx, formationTemplateID, model.FormationTemplateWebhookReference)
	if err != nil {
		return nil, false, errors.Wrapf(err, "when listing formation lifecycle webhooks for formation template with ID: %q", formationTemplateID)
	}
	operation := determineFormationOperationFromState(formation.State)
	errorState := determineFormationErrorStateFromOperation(operation)

	formationReqs, err := s.notificationsService.GenerateFormationNotifications(formationResyncTransactionCtx, formationTemplateWebhooks, tenantID, formation, formationTemplateName, formationTemplateID, operation)
	if err != nil {
		return nil, false, errors.Wrapf(err, "while generating notifications for formation with ID: %q and name: %q", formationID, formation.Name)
	}

	for _, formationReq := range formationReqs {
		if err = s.processFormationNotifications(formationResyncTransactionCtx, formation, formationReq, errorState); err != nil {
			processErr := errors.Wrapf(err, "while processing notifications for formation with ID: %q and name: %q", formationID, formation.Name)
			log.C(ctx).Error(processErr)
			return nil, false, processErr
		}
		if errorState == model.DeleteErrorFormationState && formation.State == model.ReadyFormationState && formationReq.Webhook != nil && formationReq.Webhook.Mode != nil && *formationReq.Webhook.Mode == graphql.WebhookModeSync {
			if err = s.DeleteFormationEntityAndScenarios(formationResyncTransactionCtx, tenantID, formation.Name); err != nil {
				return nil, false, errors.Wrapf(err, "while deleting formation with name %s", formation.Name)
			}
		}
	}

	if previousState == model.DeleteErrorFormationState && formation.State == model.ReadyFormationState {
		err = formationResyncTx.Commit()
		if err != nil {
			return nil, false, err
		}
		return formation, true, nil
	}

	formation, err = s.formationRepository.Get(formationResyncTransactionCtx, formationID, tenantID)
	if err != nil {
		return nil, false, errors.Wrapf(err, "while getting formation with ID %q for tenant %q", formationID, tenantID)
	}

	err = formationResyncTx.Commit()
	if err != nil {
		return nil, false, err
	}

	return formation, false, nil
}

// CreateAutomaticScenarioAssignment creates a new AutomaticScenarioAssignment for a given ScenarioName, Tenant and TargetTenantID
// It also ensures that all runtimes(or/and runtime contexts) with given scenarios are assigned for the TargetTenantID
func (s *service) CreateAutomaticScenarioAssignment(ctx context.Context, in *model.AutomaticScenarioAssignment) (*model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	in.Tenant = tenantID
	if err := s.validateThatScenarioExists(ctx, in); err != nil {
		return nil, err
	}

	if err = s.repo.Create(ctx, in); err != nil {
		if apperrors.IsNotUniqueError(err) {
			return nil, apperrors.NewInvalidOperationError("a given scenario already has an assignment")
		}

		return nil, errors.Wrap(err, "while persisting Assignment")
	}

	if err = s.asaEngine.EnsureScenarioAssigned(ctx, in, s.AssignFormation); err != nil {
		return nil, errors.Wrap(err, "while assigning scenario to runtimes matching selector")
	}

	return in, nil
}

// DeleteAutomaticScenarioAssignment deletes the assignment for a given scenario in a scope of a tenant
// It also removes corresponding assigned scenarios for the ASA
func (s *service) DeleteAutomaticScenarioAssignment(ctx context.Context, in *model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	if err = s.repo.DeleteForScenarioName(ctx, tenantID, in.ScenarioName); err != nil {
		return errors.Wrap(err, "while deleting the Assignment")
	}

	if err = s.asaEngine.RemoveAssignedScenario(ctx, in, s.UnassignFormation); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	return nil
}

// RemoveAssignedScenarios removes all the scenarios that are coming from any of the provided ASAs
func (s *service) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		if err := s.asaEngine.RemoveAssignedScenario(ctx, asa, s.UnassignFormation); err != nil {
			return errors.Wrapf(err, "while deleting automatic scenario assigment: %s", asa.ScenarioName)
		}
	}
	return nil
}

// DeleteManyASAForSameTargetTenant deletes a list of ASAs for the same targetTenant
// It also removes corresponding scenario assignments coming from the ASAs
func (s *service) DeleteManyASAForSameTargetTenant(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	targetTenant, err := s.ensureSameTargetTenant(in)
	if err != nil {
		return errors.Wrap(err, "while ensuring input is valid")
	}

	if err = s.repo.DeleteForTargetTenant(ctx, tenantID, targetTenant); err != nil {
		return errors.Wrap(err, "while deleting the Assignments")
	}

	if err = s.RemoveAssignedScenarios(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	return nil
}

func (s *service) GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error) {
	return s.asaEngine.GetScenariosFromMatchingASAs(ctx, objectID, objType)
}

// MergeScenariosFromInputLabelsAndAssignments merges all the scenarios that are part of the resource labels (already added + to be added with the current operation)
// with all the scenarios that should be assigned based on ASAs.
func (s *service) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}, runtimeID string) ([]interface{}, error) {
	scenariosFromAssignments, err := s.asaEngine.GetScenariosFromMatchingASAs(ctx, runtimeID, graphql.FormationObjectTypeRuntime)
	scenariosSet := make(map[string]struct{}, len(scenariosFromAssignments))

	if err != nil {
		return nil, errors.Wrapf(err, "while getting scenarios for selector labels")
	}

	for _, scenario := range scenariosFromAssignments {
		scenariosSet[scenario] = struct{}{}
	}

	scenariosFromInput, isScenarioLabelInInput := inputLabels[model.ScenariosKey]

	if isScenarioLabelInInput {
		scenarioLabels, err := label.ValueToStringsSlice(scenariosFromInput)
		if err != nil {
			return nil, errors.Wrap(err, "while converting scenarios label to a string slice")
		}

		for _, scenario := range scenarioLabels {
			scenariosSet[scenario] = struct{}{}
		}
	}

	scenarios := make([]interface{}, 0, len(scenariosSet))
	for k := range scenariosSet {
		scenarios = append(scenarios, k)
	}
	return scenarios, nil
}

func (s *service) SetFormationToErrorState(ctx context.Context, formation *model.Formation, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationState) error {
	log.C(ctx).Infof("Setting formation with ID: %q to state: %q", formation.ID, state)
	formation.State = state

	formationError := formationassignment.AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}

	marshaledErr, err := json.Marshal(formationError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for formation with ID: %q", formation.ID)
	}
	formation.Error = marshaledErr

	if err := s.formationRepository.Update(ctx, formation); err != nil {
		if state == model.DeleteErrorFormationState && (apperrors.IsNotFoundError(err) || apperrors.IsUnauthorizedError(err)) { // the not found error is disguised behind the unauthorized error in case of update
			return nil
		}
		return err
	}
	return nil
}

// MatchingFunc provides signature for functions used for matching asa against runtimeID
type MatchingFunc func(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error)

func (s *service) modifyFormations(ctx context.Context, tnt, formationName string, modificationFunc modificationFunc) error {
	def, err := s.labelDefRepository.GetByKey(ctx, tnt, model.ScenariosKey)
	if err != nil {
		return errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	formationNames, err := labeldef.ParseFormationsFromSchema(def.Schema)
	if err != nil {
		return err
	}

	formationNames = modificationFunc(formationNames, formationName)

	schema, err := labeldef.NewSchemaForFormations(formationNames)
	if err != nil {
		return errors.Wrap(err, "while parsing scenarios")
	}

	if err = s.labelDefService.ValidateExistingLabelsAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return err
	}
	if err = s.labelDefService.ValidateAutomaticScenarioAssignmentAgainstSchema(ctx, schema, tnt, model.ScenariosKey); err != nil {
		return errors.Wrap(err, "while validating Scenario Assignments against a new schema")
	}

	return s.labelDefRepository.UpdateWithVersion(ctx, model.LabelDefinition{
		ID:      def.ID,
		Tenant:  tnt,
		Key:     model.ScenariosKey,
		Schema:  &schema,
		Version: def.Version,
	})
}

func (s *service) modifyAssignedFormations(ctx context.Context, tnt, objectID, formationName string, objectType model.LabelableObject, modificationFunc modificationFunc) error {
	log.C(ctx).Infof("Modifying formation with name: %q for object with type: %q and ID: %q", formationName, objectType, objectID)

	labelInput := newLabelInput(formationName, objectID, objectType)
	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return err
	}

	existingFormations, err := label.ValueToStringsSlice(existingLabel.Value)
	if err != nil {
		return err
	}

	formations := modificationFunc(existingFormations, formationName)

	// can not set scenario label to empty value, violates the scenario label definition
	if len(formations) == 0 {
		log.C(ctx).Infof("After the modifications, the %q label is empty. Deleting empty label...", model.ScenariosKey)
		if err = s.labelRepository.Delete(ctx, tnt, objectType, objectID, model.ScenariosKey); err != nil {
			if apperrors.IsUnauthorizedError(err) {
				return apperrors.NewNotFoundError(resource.Label, existingLabel.ID)
			}
			return err
		}
		return nil
	}

	labelInput.Value = formations
	labelInput.Version = existingLabel.Version
	log.C(ctx).Infof("Updating formations list to %q", formations)
	if err = s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput); err != nil {
		if apperrors.IsUnauthorizedError(err) || apperrors.IsNewInvalidOperationError(err) {
			return apperrors.NewNotFoundError(resource.Label, existingLabel.ID)
		}
		return err
	}
	return nil
}

type modificationFunc func(formationNames []string, formationName string) []string

func addFormation(formationNames []string, formationName string) []string {
	for _, f := range formationNames {
		if f == formationName {
			return formationNames
		}
	}

	return append(formationNames, formationName)
}

func deleteFormation(formations []string, formation string) []string {
	filteredFormations := make([]string, 0, len(formations))
	for _, f := range formations {
		if f != formation {
			filteredFormations = append(filteredFormations, f)
		}
	}

	return filteredFormations
}

func newLabelInput(formation, objectID string, objectType model.LabelableObject) *model.LabelInput {
	return &model.LabelInput{
		Key:        model.ScenariosKey,
		Value:      []string{formation},
		ObjectID:   objectID,
		ObjectType: objectType,
		Version:    0,
	}
}

func newAutomaticScenarioAssignmentModel(formation, callerTenant, targetTenant string) *model.AutomaticScenarioAssignment {
	return &model.AutomaticScenarioAssignment{
		ScenarioName:   formation,
		Tenant:         callerTenant,
		TargetTenantID: targetTenant,
	}
}

func objectTypeToLabelableObject(objectType graphql.FormationObjectType) (labelableObj model.LabelableObject) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		labelableObj = model.ApplicationLabelableObject
	case graphql.FormationObjectTypeRuntime:
		labelableObj = model.RuntimeLabelableObject
	case graphql.FormationObjectTypeTenant:
		labelableObj = model.TenantLabelableObject
	case graphql.FormationObjectTypeRuntimeContext:
		labelableObj = model.RuntimeContextLabelableObject
	}
	return labelableObj
}

func formationAssignmentTypeToFormationObjectType(objectType model.FormationAssignmentType) (formationObjectType graphql.FormationObjectType) {
	switch objectType {
	case model.FormationAssignmentTypeApplication:
		formationObjectType = graphql.FormationObjectTypeApplication
	case model.FormationAssignmentTypeRuntime:
		formationObjectType = graphql.FormationObjectTypeRuntime
	case model.FormationAssignmentTypeRuntimeContext:
		formationObjectType = graphql.FormationObjectTypeRuntimeContext
	}
	return formationObjectType
}

func (s *service) ensureSameTargetTenant(in []*model.AutomaticScenarioAssignment) (string, error) {
	if len(in) == 0 || in[0] == nil {
		return "", apperrors.NewInternalError("expected at least one item in Assignments slice")
	}

	targetTenant := in[0].TargetTenantID

	for _, item := range in {
		if item != nil && item.TargetTenantID != targetTenant {
			return "", apperrors.NewInternalError("all input items have to have the same target tenant")
		}
	}

	return targetTenant, nil
}

func (s *service) validateThatScenarioExists(ctx context.Context, in *model.AutomaticScenarioAssignment) error {
	availableScenarios, err := s.getAvailableScenarios(ctx, in.Tenant)
	if err != nil {
		return err
	}

	for _, av := range availableScenarios {
		if av == in.ScenarioName {
			return nil
		}
	}

	return apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, in.ScenarioName)
}

func (s *service) getAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	out, err := s.labelDefService.GetAvailableScenarios(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting available scenarios")
	}
	return out, nil
}

func (s *service) createFormation(ctx context.Context, tenant, templateID, formationName string, state model.FormationState) (*model.Formation, error) {
	formation := &model.Formation{
		ID:                  s.uuidService.Generate(),
		TenantID:            tenant,
		FormationTemplateID: templateID,
		Name:                formationName,
		State:               state,
	}

	log.C(ctx).Debugf("Creating formation with name: %q and template ID: %q...", formationName, templateID)
	if err := s.formationRepository.Create(ctx, formation); err != nil {
		log.C(ctx).Errorf("An error occurred while creating formation with name: %q and template ID: %q", formationName, templateID)
		return nil, errors.Wrapf(err, "An error occurred while creating formation with name: %q and template ID: %q", formationName, templateID)
	}

	return formation, nil
}

type formationWithTemplate struct {
	formation         *model.Formation
	formationTemplate *model.FormationTemplate
}

func (s *service) getFormationWithTemplate(ctx context.Context, formationName, tnt string) (*formationWithTemplate, error) {
	formation, err := s.formationRepository.GetByName(ctx, formationName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by name: %q: %v", formationName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by name: %q", formationName)
	}

	template, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by ID: %q: %v", formation.FormationTemplateID, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by ID: %q", formation.FormationTemplateID)
	}

	return &formationWithTemplate{formation: formation, formationTemplate: template}, nil
}

func (s *service) isValidRuntimeType(ctx context.Context, tnt string, runtimeID string, formationTemplate *model.FormationTemplate) error {
	runtimeTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
		Key:        s.runtimeTypeLabelKey,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	})
	if err != nil {
		return errors.Wrapf(err, "while getting label %q for runtime with ID %q", s.runtimeTypeLabelKey, runtimeID)
	}

	runtimeType, ok := runtimeTypeLabel.Value.(string)
	if !ok {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("missing runtimeType for formation template %q, allowing only %q", formationTemplate.Name, formationTemplate.RuntimeTypes))
	}
	isAllowed := false
	for _, allowedType := range formationTemplate.RuntimeTypes {
		if allowedType == runtimeType {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeType, formationTemplate.Name, formationTemplate.RuntimeTypes))
	}
	return nil
}

func (s *service) isValidApplication(application *model.Application) error {
	if application.DeletedAt != nil {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("application with ID %q is currently being deleted", application.ID))
	}
	if !application.Ready {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("application with ID %q is not ready", application.ID))
	}
	return nil
}

func (s *service) isValidApplicationType(ctx context.Context, tnt string, applicationID string, formationTemplate *model.FormationTemplate) error {
	applicationTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
		Key:        s.applicationTypeLabelKey,
		ObjectID:   applicationID,
		ObjectType: model.ApplicationLabelableObject,
	})
	if err != nil {
		return errors.Wrapf(err, "while getting label %q for application with ID %q", s.applicationTypeLabelKey, applicationID)
	}

	applicationType, ok := applicationTypeLabel.Value.(string)
	if !ok {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("missing %s label for formation template %q, allowing only %q", s.applicationTypeLabelKey, formationTemplate.Name, formationTemplate.ApplicationTypes))
	}
	isAllowed := false
	for _, allowedType := range formationTemplate.ApplicationTypes {
		if allowedType == applicationType {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("unsupported applicationType %q for formation template %q, allowing only %q", applicationType, formationTemplate.Name, formationTemplate.ApplicationTypes))
	}
	return nil
}

func setToSlice(set map[string]bool) []string {
	result := make([]string, 0, len(set))
	for key := range set {
		result = append(result, key)
	}
	return result
}

func (s *service) processFormationNotifications(ctx context.Context, formation *model.Formation, formationReq *webhookclient.FormationNotificationRequest, errorState model.FormationState) error {
	response, err := s.notificationsService.SendNotification(ctx, formationReq)
	if err != nil {
		updateError := s.SetFormationToErrorState(ctx, formation, err.Error(), formationassignment.TechnicalError, errorState)
		if updateError != nil {
			return errors.Wrapf(updateError, "while updating error state: %s", errors.Wrapf(err, "while sending notification for formation with ID: %q", formation.ID).Error())
		}
		notificationErr := errors.Wrapf(err, "while sending notification for formation with ID: %q and name: %q", formation.ID, formation.Name)
		log.C(ctx).Error(notificationErr)
		return notificationErr
	}

	if response.Error != nil && *response.Error != "" {
		if err = s.statusService.SetFormationToErrorStateWithConstraints(ctx, formation, *response.Error, formationassignment.ClientError, errorState, determineFormationOperationFromState(errorState)); err != nil {
			return errors.Wrapf(err, "while updating error state for formation with ID: %q and name: %q", formation.ID, formation.Name)
		}

		log.C(ctx).Errorf("Received error from formation webhook response: %v", *response.Error)
		// This is the client error case, and we should not return an error,
		// because otherwise the transaction will be rolled back
		return nil
	}

	if formationReq.Webhook != nil && formationReq.Webhook.Mode != nil && *formationReq.Webhook.Mode == graphql.WebhookModeAsyncCallback {
		log.C(ctx).Infof("The webhook with ID: %q in the notification is in %q mode. Waiting for the receiver to report the status on the status API...", formationReq.Webhook.ID, graphql.WebhookModeAsyncCallback)
		if errorState == model.CreateErrorFormationState {
			formation.State = model.InitialFormationState
			formation.Error = nil
		}
		if errorState == model.DeleteErrorFormationState {
			formation.State = model.DeletingFormationState
			formation.Error = nil
		}
		log.C(ctx).Infof("Updating formation with ID: %q and name: %q to: %q state and waiting for the receiver to report the status on the status API...", formation.ID, formation.Name, formation.State)
		if err = s.formationRepository.Update(ctx, formation); err != nil {
			if errorState == model.DeleteErrorFormationState && (apperrors.IsNotFoundError(err) || apperrors.IsUnauthorizedError(err)) { // the not found error is disguised behind the unauthorized error in case of update
				return nil
			}

			return errors.Wrapf(err, "while updating formation with ID: %q", formation.ID)
		}
		return nil
	}

	if *response.ActualStatusCode == *response.SuccessStatusCode {
		formation.State = model.ReadyFormationState
		formation.Error = nil
		log.C(ctx).Infof("Updating formation with ID: %q and name: %q to: %q state", formation.ID, formation.Name, model.ReadyFormationState)
		if err := s.statusService.UpdateWithConstraints(ctx, formation, determineFormationOperationFromState(errorState)); err != nil {
			return errors.Wrapf(err, "while updating formation with ID: %q and name: %q to state: %s", formation.ID, formation.Name, model.ReadyFormationState)
		}
	}

	return nil
}

func determineFormationState(ctx context.Context, formationTemplateID, formationTemplateName string, formationTemplateWebhooks []*model.Webhook, externallyProvidedFormationState model.FormationState) model.FormationState {
	if len(formationTemplateWebhooks) == 0 {
		if len(externallyProvidedFormationState) > 0 {
			log.C(ctx).Infof("Formation template with ID: %q and name: %q does not have any webhooks. The formation will be created with %s state as it was provided externally", formationTemplateID, formationTemplateName, externallyProvidedFormationState)
			return externallyProvidedFormationState
		}
		log.C(ctx).Infof("Formation template with ID: %q and name: %q does not have any webhooks. The formation will be created with %s state", formationTemplateID, formationTemplateName, model.ReadyFormationState)
		return model.ReadyFormationState
	}

	if externallyProvidedFormationState == model.DraftFormationState {
		return externallyProvidedFormationState
	}

	return model.InitialFormationState
}

func determineFormationOperationFromState(state model.FormationState) model.FormationOperation {
	switch state {
	case model.InitialFormationState, model.CreateErrorFormationState:
		return model.CreateFormation
	case model.DeletingFormationState, model.DeleteErrorFormationState:
		return model.DeleteFormation
	default:
		return ""
	}
}

func determineFormationErrorStateFromOperation(operation model.FormationOperation) model.FormationState {
	switch operation {
	case model.CreateFormation:
		return model.CreateErrorFormationState
	case model.DeleteFormation:
		return model.DeleteErrorFormationState
	default:
		return ""
	}
}

func isObjectTypeAllowed(objectType graphql.FormationObjectType) bool {
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext, graphql.FormationObjectTypeTenant:
		return true
	default:
		return false
	}
}

func isObjectTypeSupported(formationTemplate *model.FormationTemplate, objectType graphql.FormationObjectType) bool {
	if formationTemplate.RuntimeArtifactKind == nil && formationTemplate.RuntimeTypeDisplayName == nil && len(formationTemplate.RuntimeTypes) == 0 {
		switch objectType {
		case graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext, graphql.FormationObjectTypeTenant:
			return false
		default:
			return true
		}
	}

	return true
}

// executeInTransaction wraps a given function into an isolated DB transaction
func (s *service) executeInTransaction(ctx context.Context, dbCalls func(ctxWithTransact context.Context) error) error {
	tx, err := s.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to begin DB transaction")
		return err
	}
	ctx = persistence.SaveToContext(ctx, tx)
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	if err := dbCalls(ctx); err != nil {
		log.C(ctx).WithError(err).Error("Failed to execute database calls")
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit database transaction")
		return err
	}
	return nil
}
