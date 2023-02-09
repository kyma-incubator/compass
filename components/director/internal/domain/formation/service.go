package formation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"

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
//go:generate mockery --name=FormationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationRepository interface {
	Get(ctx context.Context, id, tenantID string) (*model.Formation, error)
	GetByName(ctx context.Context, name, tenantID string) (*model.Formation, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.FormationPage, error)
	Create(ctx context.Context, item *model.Formation) error
	DeleteByName(ctx context.Context, tenantID, name string) error
	Update(ctx context.Context, model *model.Formation) error
}

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	GetByNameAndTenant(ctx context.Context, templateName, tenantID string) (*model.FormationTemplate, error)
}

// NotificationsService represents the notification service for generating and sending notifications
//go:generate mockery --name=NotificationsService --output=automock --outpkg=automock --case=underscore --disable-version-string
type NotificationsService interface {
	GenerateFormationAssignmentNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.FormationAssignmentNotificationRequest, error)
	GenerateFormationNotifications(ctx context.Context, tenantID string, formation *model.Formation, formationTemplateID string, operation model.FormationOperation) ([]*webhookclient.FormationNotificationRequest, error)
	SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookRequest) (*webhookdir.Response, error)
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
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
}

//go:generate mockery --exported --name=automaticFormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type automaticFormationAssignmentRepository interface {
	Create(ctx context.Context, model model.AutomaticScenarioAssignment) error
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
	EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error
	RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc ProcessScenarioFunc) error
	GetMatchingFuncByFormationObjectType(objType graphql.FormationObjectType) (MatchingFunc, error)
	GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error)
	IsFormationComingFromASA(ctx context.Context, objectID, formation string, objectType graphql.FormationObjectType) (bool, error)
}

type service struct {
	labelDefRepository          labelDefRepository
	labelRepository             labelRepository
	formationRepository         FormationRepository
	formationTemplateRepository FormationTemplateRepository
	labelService                labelService
	labelDefService             labelDefService
	asaService                  automaticFormationAssignmentService
	uuidService                 uuidService
	tenantSvc                   tenantService
	repo                        automaticFormationAssignmentRepository
	runtimeRepo                 runtimeRepository
	runtimeContextRepo          runtimeContextRepository
	formationAssignmentService  formationAssignmentService
	notificationsService        NotificationsService
	constraintEngine            constraintEngine
	transact                    persistence.Transactioner
	asaEngine                   asaEngine
	runtimeTypeLabelKey         string
	applicationTypeLabelKey     string
}

// NewService creates formation service
func NewService(transact persistence.Transactioner, labelDefRepository labelDefRepository, labelRepository labelRepository, formationRepository FormationRepository, formationTemplateRepository FormationTemplateRepository, labelService labelService, uuidService uuidService, labelDefService labelDefService, asaRepo automaticFormationAssignmentRepository, asaService automaticFormationAssignmentService, tenantSvc tenantService, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, formationAssignmentService formationAssignmentService, notificationsService NotificationsService, constraintEngine constraintEngine, runtimeTypeLabelKey, applicationTypeLabelKey string) *service {
	return &service{
		transact:                    transact,
		labelDefRepository:          labelDefRepository,
		labelRepository:             labelRepository,
		formationRepository:         formationRepository,
		formationTemplateRepository: formationTemplateRepository,
		labelService:                labelService,
		labelDefService:             labelDefService,
		asaService:                  asaService,
		uuidService:                 uuidService,
		tenantSvc:                   tenantSvc,
		repo:                        asaRepo,
		runtimeRepo:                 runtimeRepo,
		runtimeContextRepo:          runtimeContextRepo,
		formationAssignmentService:  formationAssignmentService,
		notificationsService:        notificationsService,
		constraintEngine:            constraintEngine,
		runtimeTypeLabelKey:         runtimeTypeLabelKey,
		applicationTypeLabelKey:     applicationTypeLabelKey,
		asaEngine:                   NewASAEngine(asaRepo, runtimeRepo, runtimeContextRepo, formationRepository, formationTemplateRepository, runtimeTypeLabelKey, applicationTypeLabelKey),
	}
}

//go:generate mockery --exported --name=processFunc --output=automock --outpkg=automock --case=underscore --disable-version-string
// Used for testing
//nolint
type processFunc interface {
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

func (s *service) GetFormationByName(ctx context.Context, formationName, tnt string) (*model.Formation, error) {
	f, err := s.formationRepository.GetByName(ctx, formationName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by name: %q: %v", formationName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by name: %q", formationName)
	}

	return f, nil
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

// CreateFormation adds the provided formation to the scenario label definitions of the given tenant.
// If the scenario label definition does not exist it will be created
// Also, a new Formation entity is created based on the provided template name or the default one is used if it's not provided
func (s *service) CreateFormation(ctx context.Context, tnt string, formation model.Formation, templateName string) (*model.Formation, error) {
	fTmpl, err := s.formationTemplateRepository.GetByNameAndTenant(ctx, templateName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by name: %q: %v", templateName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by name: %q", templateName)
	}

	joinPointDetails := &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       templateName,
		FormationTemplateID: fTmpl.ID,
		FormationName:       formation.Name,
		TenantID:            tnt,
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreCreate, joinPointDetails, fTmpl.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.CreateFormationOperation, model.PreOperation)
	}

	formationName := formation.Name
	if err := s.modifyFormations(ctx, tnt, formationName, addFormation); err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, err
		}
		if err = s.labelDefService.CreateWithFormations(ctx, tnt, []string{formationName}); err != nil {
			return nil, err
		}
	}

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	newFormation, err := s.createFormation(ctx, tnt, fTmpl.ID, formationName)
	if err != nil {
		return nil, err
	}

	formationReqs, err := s.notificationsService.GenerateFormationNotifications(ctx, tnt, newFormation, fTmpl.ID, model.CreateFormation)
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

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostCreate, joinPointDetails, fTmpl.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.CreateFormationOperation, model.PostOperation)
	}

	return newFormation, nil
}

// DeleteFormation removes the provided formation from the scenario label definitions of the given tenant.
// Also, removes the Formation entity from the DB
func (s *service) DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	ft, err := s.getFormationWithTemplate(ctx, formation.Name, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "While deleting formation")
	}

	joinPointDetails := &formationconstraint.CRUDFormationOperationDetails{
		FormationType:       ft.formationTemplate.Name,
		FormationTemplateID: ft.formationTemplate.ID,
		FormationName:       ft.formation.Name,
		TenantID:            tnt,
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreDelete, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.DeleteFormationOperation, model.PreOperation)
	}

	formationID := ft.formation.ID
	formationName := ft.formation.Name
	log.C(ctx).Infof("Updating formation with ID: %q and name: %q to state: %s", formationID, formationName, model.DeletingFormationState)
	if err := s.formationRepository.Update(ctx, ft.formation); err != nil {
		return nil, errors.Wrapf(err, "while updating formation with ID: %q and name: %q to state: %s", formationID, formationName, model.DeletingFormationState)
	}

	formationReqs, err := s.notificationsService.GenerateFormationNotifications(ctx, tnt, ft.formation, ft.formationTemplate.ID, model.DeleteFormation)
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

	if err := s.modifyFormations(ctx, tnt, formationName, deleteFormation); err != nil {
		return nil, err
	}

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	if err = s.formationRepository.DeleteByName(ctx, tnt, formationName); err != nil {
		log.C(ctx).Errorf("An error occurred while deleting formation with name: %q", formationName)
		return nil, errors.Wrapf(err, "An error occurred while deleting formation with name: %q", formationName)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostDelete, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.DeleteFormationOperation, model.PostOperation)
	}

	return ft.formation, nil
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
// 		- If objectType is graphql.FormationObjectTypeApplication:
//				- A notification about the assigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//  			- A notification about the assigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//				- If the assigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
//				- If the assigned application has a configuration change webhook, a notification about each runtime/runtimeContext in the formation is sent to this application.
// 		- If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext:
//				- If the assigned runtime/runtimeContext has configuration change webhook, a notification about each application in the formation is sent to this runtime.
//				- A notification about the assigned runtime/runtimeContext is sent to all the applications that are in the formation and have configuration change webhook.
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
func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	log.C(ctx).Infof("Assigning object with ID %q of type %q to formation %q", objectID, objectType, formation.Name)

	ft, err := s.getFormationWithTemplate(ctx, formation.Name, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "While assigning formation with name %q", formation.Name)
	}

	joinPointDetails, err := s.prepareDetailsForAssign(ctx, tnt, objectID, objectType, ft.formation, ft.formationTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "While preparing joinpoint details for target operation %q and constraint type %q", model.AssignFormationOperation, model.PreOperation)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreAssign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.AssignFormationOperation, model.PreOperation)
	}

	formationFromDB := ft.formation
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		err := s.assign(ctx, tnt, objectID, objectType, formationFromDB, ft.formationTemplate)
		if err != nil {
			return nil, err
		}

		assignments, err := s.formationAssignmentService.GenerateAssignments(ctx, tnt, objectID, objectType, formationFromDB)
		if err != nil {
			return nil, err
		}

		rtmContextIDsMapping, err := s.getRuntimeContextIDToRuntimeIDMapping(ctx, tnt, assignments)
		if err != nil {
			return nil, err
		}

		requests, err := s.notificationsService.GenerateFormationAssignmentNotifications(ctx, tnt, objectID, formationFromDB, model.AssignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s assignment", objectType)
		}

		if err = s.formationAssignmentService.ProcessFormationAssignments(ctx, assignments, rtmContextIDsMapping, requests, s.formationAssignmentService.ProcessFormationAssignmentPair); err != nil {
			log.C(ctx).Errorf("Error occurred while processing formationAssignments %s", err.Error())
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
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.AssignFormationOperation, model.PostOperation)
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

func (s *service) checkFormationTemplateTypes(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formationTemplate *model.FormationTemplate) error {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		if err := s.isValidApplicationType(ctx, tnt, objectID, formationTemplate); err != nil {
			return errors.Wrapf(err, "while validating application type for application %q", objectID)
		}
	case graphql.FormationObjectTypeRuntime:
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
//  Additionally, notifications are sent to the interested participants for that formation change.
// 		- If objectType is graphql.FormationObjectTypeApplication:
//				- A notification about the unassigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//  			- A notification about the unassigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//				- If the unassigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
//				- If the unassigned application has a configuration change webhook, a notification about each runtime/runtimeContext in the formation is sent to this application.
// 		- If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext:
//				- If the unassigned runtime/runtimeContext has configuration change webhook, a notification about each application in the formation is sent to this runtime.
//   			- A notification about the unassigned runtime/runtimeContext is sent to all the applications that are in the formation and have configuration change webhook.
//
// For the formationAssignments that have their source or target field set to objectID:
// 		- If the formationAssignment does not have notification associated with it
//				- the formation assignment is deleted
//		- If the formationAssignment is associated with a notification
//				- If the response from the notification is success
//						- the formationAssignment is deleted
// 				- If the response from the notification is different from success
//						- the formation assignment is updated with an error
//
// After the processing of the formationAssignments the state is persisted regardless of whether there were any errors.
// If an error has occurred during the formationAssignment processing the unassign operation is rolled back(the updated
// with the error formationAssignments are already persisted in the database).
//
// For objectType graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant which then will unassign the right Runtime / RuntimeContexts based on the formation template's runtimeType.
func (s *service) UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	log.C(ctx).Infof("Unassigning object with ID %q of type %q to formation %q", objectID, objectType, formation.Name)

	formationName := formation.Name
	ft, err := s.getFormationWithTemplate(ctx, formationName, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "While unassigning formation with name %q", formationName)
	}

	joinPointDetails, err := s.prepareDetailsForUnassign(ctx, tnt, objectID, objectType, ft.formation, ft.formationTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "While preparing joinpoint details for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PreOperation)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PreUnassign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PreOperation)
	}

	formationFromDB := ft.formation
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		if err := s.modifyAssignedFormations(ctx, tnt, objectID, formationName, objectTypeToLabelableObject(objectType), deleteFormation); err != nil {
			return nil, err
		}

		requests, err := s.notificationsService.GenerateFormationAssignmentNotifications(ctx, tnt, objectID, formationFromDB, model.UnassignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s unassignment", objectType)
		}

		formationAssignmentsForObject, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(ctx, formationFromDB.ID, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "While listing formationAssignments for object with type %q and ID %q", objectType, objectID)
		}

		rtmContextIDsMapping, err := s.getRuntimeContextIDToRuntimeIDMapping(ctx, tnt, formationAssignmentsForObject)
		if err != nil {
			return nil, err
		}

		tx, err := s.transact.Begin()
		if err != nil {
			return nil, err
		}
		transactionCtx := persistence.SaveToContext(ctx, tx)
		defer s.transact.RollbackUnlessCommitted(transactionCtx, tx)

		if err = s.formationAssignmentService.ProcessFormationAssignments(transactionCtx, formationAssignmentsForObject, rtmContextIDsMapping, requests, s.formationAssignmentService.CleanupFormationAssignment); err != nil {
			commitErr := tx.Commit()
			if commitErr != nil {
				return nil, errors.Wrapf(err, "while committing transaction with error")
			}
			return nil, err
		}

		// It is important to do the list in the inner transaction
		pendingAsyncAssignments, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(transactionCtx, formationFromDB.ID, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "While listing formationAssignments for object with type %q and ID %q", objectType, objectID)
		}

		err = tx.Commit()
		if err != nil {
			return nil, errors.Wrapf(err, "while committing transaction")
		}

		if len(pendingAsyncAssignments) > 0 {
			log.C(ctx).Infof("There is an async delete notification in progress. Re-assigning the object with type %q and ID %q to formation %q until status is reported by the notification receiver", objectType, objectID, formationName)
			err := s.assign(ctx, tnt, objectID, objectType, formationFromDB, ft.formationTemplate) // It is important to do the re-assign in the outer transaction.
			if err != nil {
				return nil, errors.Wrapf(err, "While re-assigning the object with type %q and ID %q that is being unassigned asynchronously", objectType, objectID)
			}
		}

	case graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		if isFormationComingFromASA, err := s.asaEngine.IsFormationComingFromASA(ctx, objectID, formationName, objectType); err != nil {
			return nil, err
		} else if isFormationComingFromASA {
			break
		}

		if err = s.modifyAssignedFormations(ctx, tnt, objectID, formationName, objectTypeToLabelableObject(objectType), deleteFormation); err != nil {
			if apperrors.IsNotFoundError(err) {
				return formationFromDB, nil
			}
			return nil, err
		}

		requests, err := s.notificationsService.GenerateFormationAssignmentNotifications(ctx, tnt, objectID, formationFromDB, model.UnassignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s unassignment", objectType)
		}

		formationAssignmentsForObject, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(ctx, formationFromDB.ID, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "While listing formationAssignments for object with type %q and ID %q", objectType, objectID)
		}

		rtmContextIDsMapping, err := s.getRuntimeContextIDToRuntimeIDMapping(ctx, tnt, formationAssignmentsForObject)
		if err != nil {
			return nil, err
		}

		tx, err := s.transact.Begin()
		if err != nil {
			return nil, err
		}

		transactionCtx := persistence.SaveToContext(ctx, tx)
		defer s.transact.RollbackUnlessCommitted(transactionCtx, tx)

		if err = s.formationAssignmentService.ProcessFormationAssignments(transactionCtx, formationAssignmentsForObject, rtmContextIDsMapping, requests, s.formationAssignmentService.CleanupFormationAssignment); err != nil {
			commitErr := tx.Commit()
			if commitErr != nil {
				return nil, errors.Wrapf(err, "while committing transaction with error")
			}
			return nil, err
		}

		// It is important to do the list in the inner transaction
		pendingAsyncAssignments, err := s.formationAssignmentService.ListFormationAssignmentsForObjectID(transactionCtx, formationFromDB.ID, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "While listing formationAssignments for object with type %q and ID %q", objectType, objectID)
		}

		err = tx.Commit()
		if err != nil {
			return nil, errors.Wrapf(err, "while committing transaction")
		}

		if len(pendingAsyncAssignments) > 0 {
			log.C(ctx).Infof("There is an async delete notification in progress. Re-assigning the object with type %q and ID %q to formation %q until status is reported by the notification receiver", objectType, objectID, formation.Name)
			err := s.assign(ctx, tnt, objectID, objectType, formationFromDB, ft.formationTemplate) // It is importnat to do the re-assign in the outer transaction.
			if err != nil {
				return nil, errors.Wrapf(err, "While re-assigning the object with type %q and ID %q that is being unassigned asynchronously", objectType, objectID)
			}
		}

	case graphql.FormationObjectTypeTenant:
		asa, err := s.asaService.GetForScenarioName(ctx, formationName)
		if err != nil {
			return nil, err
		}
		if err = s.DeleteAutomaticScenarioAssignment(ctx, asa); err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}

	if err = s.constraintEngine.EnforceConstraints(ctx, formationconstraint.PostUnassign, joinPointDetails, ft.formationTemplate.ID); err != nil {
		return nil, errors.Wrapf(err, "While enforcing constraints for target operation %q and constraint type %q", model.UnassignFormationOperation, model.PostOperation)
	}

	return formationFromDB, nil
}

func (s *service) getRuntimeContextIDToRuntimeIDMapping(ctx context.Context, tnt string, formationAssignmentsForObject []*model.FormationAssignment) (map[string]string, error) {
	rtmContextIDs := make([]string, 0)
	for _, assignment := range formationAssignmentsForObject {
		if assignment.TargetType == model.FormationAssignmentTypeRuntimeContext {
			rtmContextIDs = append(rtmContextIDs, assignment.Target)
		}
	}
	rtmContexts, err := s.runtimeContextRepo.ListByIDs(ctx, tnt, rtmContextIDs)
	if err != nil {
		return nil, err
	}
	rtmContextIDsToRuntimeMap := make(map[string]string, len(rtmContexts))
	for _, rtmContext := range rtmContexts {
		rtmContextIDsToRuntimeMap[rtmContext.ID] = rtmContext.RuntimeID
	}
	return rtmContextIDsToRuntimeMap, nil
}

// CreateAutomaticScenarioAssignment creates a new AutomaticScenarioAssignment for a given ScenarioName, Tenant and TargetTenantID
// It also ensures that all runtimes(or/and runtime contexts) with given scenarios are assigned for the TargetTenantID
func (s *service) CreateAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	in.Tenant = tenantID
	if err := s.validateThatScenarioExists(ctx, in); err != nil {
		return model.AutomaticScenarioAssignment{}, err
	}

	if err = s.repo.Create(ctx, in); err != nil {
		if apperrors.IsNotUniqueError(err) {
			return model.AutomaticScenarioAssignment{}, apperrors.NewInvalidOperationError("a given scenario already has an assignment")
		}

		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while persisting Assignment")
	}

	if err = s.asaEngine.EnsureScenarioAssigned(ctx, in, s.AssignFormation); err != nil {
		return model.AutomaticScenarioAssignment{}, errors.Wrap(err, "while assigning scenario to runtimes matching selector")
	}

	return in, nil
}

// DeleteAutomaticScenarioAssignment deletes the assignment for a given scenario in a scope of a tenant
// It also removes corresponding assigned scenarios for the ASA
func (s *service) DeleteAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) error {
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
		if err := s.asaEngine.RemoveAssignedScenario(ctx, *asa, s.UnassignFormation); err != nil {
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
		log.C(ctx).Infof("The object is not part of any formations. Deleting empty label")
		return s.labelRepository.Delete(ctx, tnt, objectType, objectID, model.ScenariosKey)
	}

	labelInput.Value = formations
	labelInput.Version = existingLabel.Version
	log.C(ctx).Infof("Updating formations list to %q", formations)
	return s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput)
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

func newAutomaticScenarioAssignmentModel(formation, callerTenant, targetTenant string) model.AutomaticScenarioAssignment {
	return model.AutomaticScenarioAssignment{
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

func (s *service) validateThatScenarioExists(ctx context.Context, in model.AutomaticScenarioAssignment) error {
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

func (s *service) createFormation(ctx context.Context, tenant, templateID, formationName string) (*model.Formation, error) {
	formation := &model.Formation{
		ID:                  s.uuidService.Generate(),
		TenantID:            tenant,
		FormationTemplateID: templateID,
		Name:                formationName,
		State:               model.InitialFormationState,
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

func (s *service) processFormationNotifications(ctx context.Context, formation *model.Formation, formationReq *webhookclient.FormationNotificationRequest, state model.FormationState) error {
	response, err := s.notificationsService.SendNotification(ctx, formationReq)
	if err != nil {
		updateError := s.setFormationToErrorState(ctx, formation, err.Error(), formationassignment.TechnicalError, state)
		if updateError != nil {
			return errors.Wrapf(updateError, "while updating error state: %s", errors.Wrapf(err, "while sending notification for formation with ID: %q", formation.ID).Error())
		}
		notificationErr := errors.Wrapf(err, "while sending notification for formation with ID: %q and name: %q", formation.ID, formation.Name)
		log.C(ctx).Error(notificationErr)
		return notificationErr
	}

	if response.Error != nil && *response.Error != "" {
		err = s.setFormationToErrorState(ctx, formation, *response.Error, formationassignment.ClientError, state)
		if err != nil {
			return errors.Wrapf(err, "while updating error state for formation with ID: %q and name: %q", formation.ID, formation.Name)
		}

		err = errors.Errorf("Received error from formation webhook response: %v", *response.Error)
		log.C(ctx).Error(err)
		return err
	}

	if *response.ActualStatusCode == *response.SuccessStatusCode {
		formation.State = model.ReadyFormationState
		log.C(ctx).Infof("Updating formation with ID: %q and name: %q to state: %s", formation.ID, formation.Name, model.ReadyFormationState)
		if err := s.formationRepository.Update(ctx, formation); err != nil {
			return errors.Wrapf(err, "while updating formation with ID: %q and name: %q to state: %s", formation.ID, formation.Name, model.ReadyFormationState)
		}
	}

	return nil
}

func (s *service) setFormationToErrorState(ctx context.Context, formation *model.Formation, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationState) error {
	formation.State = state

	formationError := formationassignment.AssignmentErrorWrapper{Error: formationassignment.AssignmentError{
		Message:   errorMessage,
		ErrorCode: errorCode,
	}}

	marshaledErr, err := json.Marshal(formationError)
	if err != nil {
		return errors.Wrapf(err, "While preparing error message for formation with ID: %q", formation.ID)
	}
	formation.Error = marshaledErr

	if err := s.formationRepository.Update(ctx, formation); err != nil {
		return err
	}
	log.C(ctx).Infof("Formation with ID: %s set to state: %s", formation.ID, formation.State)
	return nil
}
