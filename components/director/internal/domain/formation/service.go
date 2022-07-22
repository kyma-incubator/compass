package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=labelDefRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
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
	ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Runtime, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	ListAll(ctx context.Context, tenant string) ([]*model.RuntimeContext, error)
	Exists(ctx context.Context, tenant, id string) (bool, error)
	ListByScenariosAndRuntimeIDs(ctx context.Context, tenant string, scenarios []string, runtimeIDs []string) ([]*model.RuntimeContext, error)
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

// FormationRepository represents the Formations repository layer
//go:generate mockery --name=FormationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationRepository interface {
	Get(ctx context.Context, id, tenantID string) (*model.Formation, error)
	GetByName(ctx context.Context, name, tenantID string) (*model.Formation, error)
	List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.FormationPage, error)
	Create(ctx context.Context, item *model.Formation) error
	DeleteByName(ctx context.Context, tenantID, name string) error
}

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	GetByName(ctx context.Context, templateName string) (*model.FormationTemplate, error)
}

//go:generate mockery --exported --name=labelDefService --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelDefService interface {
	CreateWithFormations(ctx context.Context, tnt string, formations []string) error
	ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant, key string) error
	ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID, key string) error
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenantID string) error
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
	CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error
	GetInternalTenant(ctx context.Context, externalTenant string) (string, error)
}

//go:generate mockery --exported --name=webhookClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookClient interface {
	Do(ctx context.Context, request *webhook_client.Request) (*webhookdir.Response, error)
}

//go:generate mockery --exported --name=webhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookRepository interface {
	ListByReferenceObjectTypeAndWebhookType(ctx context.Context, tenant string, whType model.WebhookType, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error)
	GetByIDAndWebhookType(ctx context.Context, tenant, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error)
}

//go:generate mockery --exported --name=webhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
}

type service struct {
	labelDefRepository            labelDefRepository
	labelRepository               labelRepository
	formationRepository           FormationRepository
	formationTemplateRepository   FormationTemplateRepository
	labelService                  labelService
	labelDefService               labelDefService
	asaService                    automaticFormationAssignmentService
	uuidService                   uuidService
	tenantSvc                     tenantService
	repo                          automaticFormationAssignmentRepository
	runtimeRepo                   runtimeRepository
	runtimeContextRepo            runtimeContextRepository
	webhookRepository             webhookRepository
	webhookClient                 webhookClient
	applicationRepository         applicationRepository
	applicationTemplateRepository applicationTemplateRepository
	webhookConverter              webhookConverter
}

// NewService creates formation service
func NewService(labelDefRepository labelDefRepository, labelRepository labelRepository, formationRepository FormationRepository, formationTemplateRepository FormationTemplateRepository, labelService labelService, uuidService uuidService, labelDefService labelDefService, asaRepo automaticFormationAssignmentRepository, asaService automaticFormationAssignmentService, tenantSvc tenantService, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, webhookRepository webhookRepository, webhookClient webhookClient, applicationRepository applicationRepository, applicationTemplateRepository applicationTemplateRepository, webhookConverter webhookConverter) *service {
	return &service{
		labelDefRepository:            labelDefRepository,
		labelRepository:               labelRepository,
		formationRepository:           formationRepository,
		formationTemplateRepository:   formationTemplateRepository,
		labelService:                  labelService,
		labelDefService:               labelDefService,
		asaService:                    asaService,
		uuidService:                   uuidService,
		tenantSvc:                     tenantSvc,
		repo:                          asaRepo,
		runtimeRepo:                   runtimeRepo,
		runtimeContextRepo:            runtimeContextRepo,
		webhookRepository:             webhookRepository,
		webhookClient:                 webhookClient,
		applicationRepository:         applicationRepository,
		applicationTemplateRepository: applicationTemplateRepository,
		webhookConverter:              webhookConverter,
	}
}

type processScenarioFunc func(context.Context, string, string, graphql.FormationObjectType, model.Formation) (*model.Formation, error)

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
	return s.createFormation(ctx, tnt, templateName, formationName)
}

// DeleteFormation removes the provided formation from the scenario label definitions of the given tenant.
// Also, removes the Formation entity from the DB
func (s *service) DeleteFormation(ctx context.Context, tnt string, formation model.Formation) (*model.Formation, error) {
	formationName := formation.Name
	if err := s.modifyFormations(ctx, tnt, formationName, deleteFormation); err != nil {
		return nil, err
	}

	f, err := s.getFormationByName(ctx, formation.Name, tnt)
	if err != nil {
		return nil, err
	}

	// TODO:: Currently we need to support both mechanisms of formation creation/deletion(through label definitions and Formations entity) for backwards compatibility
	if err = s.formationRepository.DeleteByName(ctx, tnt, formationName); err != nil {
		log.C(ctx).Errorf("An error occurred while deleting formation with name: %q", formationName)
		return nil, errors.Wrapf(err, "An error occurred while deleting formation with name: %q", formationName)
	}

	return f, nil
}

// AssignFormation assigns object based on graphql.FormationObjectType.
// For objectTypes graphql.FormationObjectType is graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime and
// graphql.FormationObjectTypeRuntimeContext it adds the provided formation to the scenario label of the entity if such exists,
// otherwise new scenario label is created for the entity with the provided formation.
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// create automatic scenario assignment with the caller and target tenant.
func (s *service) AssignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		formationFromDB, err := s.assign(ctx, tnt, objectID, objectType, formation)
		if err != nil {
			return nil, err
		}
		requests, err := s.generateNotifications(ctx, tnt, objectID, formationFromDB, model.AssignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s assignment", objectType)
		}
		err = s.sendNotifications(ctx, requests)
		if err != nil {
			return nil, err
		}
		return formationFromDB, nil
	case graphql.FormationObjectTypeTenant:
		tenantID, err := s.tenantSvc.GetInternalTenant(ctx, objectID)
		if err != nil {
			return nil, err
		}

		if _, err = s.CreateAutomaticScenarioAssignment(ctx, newAutomaticScenarioAssignmentModel(formation.Name, tnt, tenantID)); err != nil {
			return nil, err
		}
		return s.getFormationByName(ctx, formation.Name, tnt)
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) createWebhookRequest(ctx context.Context, webhook *model.Webhook, input *webhookdir.FormationConfigurationChangeInput) (*webhook_client.Request, error) {
	gqlWebhook, err := s.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhook_client.Request{
		Webhook:       *gqlWebhook,
		Object:        input,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

func (s *service) sendNotifications(ctx context.Context, notifications []*webhook_client.Request) error {
	for _, notification := range notifications {
		if _, err := s.webhookClient.Do(ctx, notification); err != nil {
			return errors.Wrapf(err, "while executing webhook with ID %s for Runtime with ID %s", notification.Webhook.ID, *notification.Webhook.RuntimeID)
		}
	}
	return nil
}

func (s *service) assign(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), addFormation); err != nil {
		if apperrors.IsNotFoundError(err) {
			labelInput := newLabelInput(formation.Name, objectID, objectTypeToLabelableObject(objectType))
			if err = s.labelService.CreateLabel(ctx, tnt, s.uuidService.Generate(), labelInput); err != nil {
				return nil, err
			}

			return s.getFormationByName(ctx, formation.Name, tnt)
		}
		return nil, err
	}

	return s.getFormationByName(ctx, formation.Name, tnt)
}

// UnassignFormation unassigns object base on graphql.FormationObjectType.
// For objectType graphql.FormationObjectTypeApplication it removes the provided formation from the
// scenario label of the application.
// For objectTypes graphql.FormationObjectTypeRuntime and graphql.FormationObjectTypeRuntimeContext
// it removes the formation from the scenario label of the runtime/runtime context if the provided
// formation is NOT assigned from ASA and does nothing if it is assigned from ASA.
// For objectType graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant.
func (s *service) UnassignFormation(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation); err != nil {
			return nil, err
		}

		formationFromDB, err := s.getFormationByName(ctx, formation.Name, tnt)
		if err != nil {
			return nil, err
		}
		requests, err := s.generateNotificationsForApplicationAssignment(ctx, tnt, objectID, formationFromDB, model.UnassignFormation)
		if err != nil {
			return nil, errors.Wrap(err, "while generating notifications for application unassignment")
		}
		err = s.sendNotifications(ctx, requests)
		if err != nil {
			return nil, err
		}
		return formationFromDB, nil

	case graphql.FormationObjectTypeRuntime, graphql.FormationObjectTypeRuntimeContext:
		if isFormationComingFromASA, err := s.isFormationComingFromASA(ctx, objectID, formation.Name, objectType); err != nil {
			return nil, err
		} else if isFormationComingFromASA {
			return &formation, nil
		}

		if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation); err != nil {
			return nil, err
		}

		formationFromDB, err := s.getFormationByName(ctx, formation.Name, tnt)
		if err != nil {
			return nil, err
		}
		requests, err := s.generateNotifications(ctx, tnt, objectID, formationFromDB, model.UnassignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s unassignment", objectType)
		}
		err = s.sendNotifications(ctx, requests)
		if err != nil {
			return nil, err
		}
		return formationFromDB, nil
	case graphql.FormationObjectTypeTenant:
		asa, err := s.asaService.GetForScenarioName(ctx, formation.Name)
		if err != nil {
			return nil, err
		}
		if err = s.DeleteAutomaticScenarioAssignment(ctx, asa); err != nil {
			return nil, err
		}

		return s.getFormationByName(ctx, formation.Name, tnt)
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) generateNotifications(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhook_client.Request, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		return s.generateNotificationsForApplicationAssignment(ctx, tenant, appID, formation, operation)
	case graphql.FormationObjectTypeRuntime:
		return s.generateNotificationsForRuntimeAssignment(ctx, tenant, appID, formation, operation)
	case graphql.FormationObjectTypeRuntimeContext:
		return s.generateNotificationsForRuntimeContextAssignment(ctx, tenant, appID, formation, operation)
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}
func (s *service) generateNotificationsForApplicationAssignment(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation) ([]*webhook_client.Request, error) {
	application, err := s.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application with id %s", appID)
	}
	applicationLabels, err := s.getLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, err
	}
	applicationWithLabels := &webhookdir.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
	}
	if err != nil {
		return nil, err
	}

	var appTemplateWithLabels *webhookdir.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := s.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting application template with id %s", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := s.getLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, err
		}
		appTemplateWithLabels = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: appTemplate,
			Labels:              applicationTemplateLabels,
		}
	}

	webhooks, err := s.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeConfigurationChanged, model.RuntimeWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing configuration changed webhooks for runtimes")
	}

	listeningRuntimeIDs := make([]string, 0, len(webhooks))
	for _, wh := range webhooks {
		listeningRuntimeIDs = append(listeningRuntimeIDs, wh.ObjectID)
	}

	if len(listeningRuntimeIDs) == 0 {
		return nil, nil
	}

	listeningRuntimes, err := s.runtimeRepo.ListByIDs(ctx, tenant, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtimes")
	}

	listeningRuntimesLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeLabelableObject, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing lables")
	}

	listeningRuntimesMapping := make(map[string]*webhookdir.RuntimeWithLabels, len(listeningRuntimes))
	for i, rt := range listeningRuntimes {
		listeningRuntimesMapping[rt.ID] = &webhookdir.RuntimeWithLabels{
			Runtime: listeningRuntimes[i],
			Labels:  listeningRuntimesLabels[rt.ID],
		}
	}

	listeningRuntimesInScenario, err := s.runtimeRepo.ListByScenariosAndIDs(ctx, tenant, []string{formation.Name}, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtimes")
	}

	runtimeContextsInScenarioForListeningRuntimes, err := s.runtimeContextRepo.ListByScenariosAndRuntimeIDs(ctx, tenant, []string{formation.Name}, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtime contexts")
	}

	runtimeContextsInScenarioForListeningRuntimesIDs := make([]string, 0, len(runtimeContextsInScenarioForListeningRuntimes))
	for _, rtCtx := range runtimeContextsInScenarioForListeningRuntimes {
		runtimeContextsInScenarioForListeningRuntimesIDs = append(runtimeContextsInScenarioForListeningRuntimesIDs, rtCtx.ID)
	}

	runtimeContextsLables, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeContextLabelableObject, runtimeContextsInScenarioForListeningRuntimesIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing lables")
	}

	runtimeIDsToBeNotified := make(map[string]bool, len(listeningRuntimesInScenario)+len(runtimeContextsInScenarioForListeningRuntimes))
	runtimeContextsInScenarioForListeningRuntimesMapping := make(map[string]*webhookdir.RuntimeContextWithLabels, len(runtimeContextsInScenarioForListeningRuntimes))
	for _, rt := range listeningRuntimesInScenario {
		runtimeIDsToBeNotified[rt.ID] = true
	}
	for i, rtCtx := range runtimeContextsInScenarioForListeningRuntimes {
		runtimeIDsToBeNotified[rtCtx.RuntimeID] = true
		runtimeContextsInScenarioForListeningRuntimesMapping[rtCtx.RuntimeID] = &webhookdir.RuntimeContextWithLabels{
			RuntimeContext: runtimeContextsInScenarioForListeningRuntimes[i],
			Labels:         runtimeContextsLables[rtCtx.ID],
		}
	}

	webhooksToCall := make(map[string]*model.Webhook, len(runtimeIDsToBeNotified))
	for i := range webhooks {
		if runtimeIDsToBeNotified[webhooks[i].ObjectID] {
			webhooksToCall[webhooks[i].ObjectID] = webhooks[i]
		}
	}

	requests := make([]*webhook_client.Request, 0, len(runtimeIDsToBeNotified))
	for rtID := range runtimeIDsToBeNotified {
		rtCtx := runtimeContextsInScenarioForListeningRuntimesMapping[rtID]
		runtime := listeningRuntimesMapping[rtID]
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:           operation,
			FormationID:         formation.ID,
			ApplicationTemplate: appTemplateWithLabels,
			Application:         applicationWithLabels,
			Runtime:             runtime,
			RuntimeContext:      rtCtx,
		}
		req, err := s.createWebhookRequest(ctx, webhooksToCall[runtime.ID], input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *service) generateNotificationsForRuntimeContextAssignment(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation) ([]*webhook_client.Request, error) {
	runtimeCtx, err := s.runtimeContextRepo.GetByID(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context with id %s", runtimeCtxID)
	}
	runtimeCtxLabels, err := s.getLabelsForObject(ctx, tenant, runtimeCtxID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, err
	}
	runtimeCtxWithLabels := &webhookdir.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context labels with id %s", runtimeCtxID)
	}

	requests, err := s.generateNotificationsForRuntimeAssignment(ctx, tenant, runtimeCtxWithLabels.RuntimeID, formation, operation)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		request.Object.(*webhookdir.FormationConfigurationChangeInput).RuntimeContext = runtimeCtxWithLabels
	}
	return requests, nil
}

func (s *service) generateNotificationsForRuntimeAssignment(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation) ([]*webhook_client.Request, error) {
	runtime, err := s.runtimeRepo.GetByID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime with id %s", runtimeID)
	}
	runtimeLabels, err := s.getLabelsForObject(ctx, tenant, runtimeID, model.RuntimeLabelableObject)
	if err != nil {
		return nil, err
	}
	runtimeWithLabels := &webhookdir.RuntimeWithLabels{
		Runtime: runtime,
		Labels:  runtimeLabels,
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime labels for id %s", runtimeID)
	}

	webhook, err := s.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "while listing configuration changes for applications")
	}

	applicationsToBeNotifiedFor, err := s.applicationRepository.ListByScenariosNoPaging(ctx, tenant, []string{formation.Name})
	if err != nil {
		return nil, errors.Wrap(err, "while listing scenario labels for applications")
	}
	if len(applicationsToBeNotifiedFor) == 0 {
		return nil, nil
	}
	applicationsToBeNotifiedForIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	applicationsTemplateIDs := make([]string, 0, len(applicationsToBeNotifiedFor))
	for _, app := range applicationsToBeNotifiedFor {
		applicationsToBeNotifiedForIDs = append(applicationsToBeNotifiedForIDs, app.ID)
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	applicationsToBeNotifiedForLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.ApplicationLabelableObject, applicationsToBeNotifiedForIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for applications")
	}
	applicationTemplates, err := s.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing application templates")
	}
	applicationTemplatesLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.AppTemplateLabelableObject, applicationsTemplateIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for application templates")
	}
	applicationTemplatesMapping := make(map[string]*webhookdir.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
		}
	}

	requests := make([]*webhook_client.Request, 0, len(applicationsToBeNotifiedFor))
	for i, app := range applicationsToBeNotifiedFor {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:           operation,
			FormationID:         formation.ID,
			ApplicationTemplate: appTemplate,
			Application: &webhookdir.ApplicationWithLabels{
				Application: applicationsToBeNotifiedFor[i],
				Labels:      applicationsToBeNotifiedForLabels[app.ID],
			},
			Runtime:        runtimeWithLabels,
			RuntimeContext: nil,
		}
		req, err := s.createWebhookRequest(ctx, webhook, input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *service) getLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]interface{}, error) {
	labels, err := s.labelRepository.ListForObject(ctx, tenant, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing labels for %s with id %s", objectType, objectID)
	}
	labelsMap := make(map[string]interface{}, len(labels))
	for _, l := range labels {
		labelsMap[l.Key] = l.Value
	}
	return labelsMap, nil
}

// CreateAutomaticScenarioAssignment creates a new AutomaticScenarioAssignment for a given ScenarioName, Tenant and TargetTenantID
// It also ensures that all runtimes with given scenarios are assigned for the TargetTenantID
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

	if err = s.EnsureScenarioAssigned(ctx, in); err != nil {
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

	if err = s.RemoveAssignedScenario(ctx, in); err != nil {
		return errors.Wrap(err, "while unassigning scenario from runtimes")
	}

	return nil
}

// EnsureScenarioAssigned ensures that the scenario is assigned to all the runtimes and runtimeContexts that are in the ASAs target_tenant_id
func (s *service) EnsureScenarioAssigned(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	return s.processScenario(ctx, in, s.AssignFormation, model.AssignFormation)
}

// RemoveAssignedScenario removes all the scenarios that are coming from the provided ASA
func (s *service) RemoveAssignedScenario(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	return s.processScenario(ctx, in, s.UnassignFormation, model.UnassignFormation)
}

func (s *service) processScenario(ctx context.Context, in model.AutomaticScenarioAssignment, processScenarioFunc processScenarioFunc, processingType model.FormationOperation) error {
	runtimes, err := s.runtimeRepo.ListOwnedRuntimes(ctx, in.TargetTenantID, nil)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range runtimes {
		if _, err = processScenarioFunc(ctx, in.Tenant, r.ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime with id %s from formation %s coming from ASA", processingType, r.ID, in.ScenarioName)
		}
	}

	runtimeContexts, err := s.runtimeContextRepo.ListAll(ctx, in.TargetTenantID)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtime contexts in target tenant: %s", in.TargetTenantID)
	}

	for _, rc := range runtimeContexts {
		if _, err = processScenarioFunc(ctx, in.Tenant, rc.ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime context with id %s from formation %s coming from ASA", processingType, rc.ID, in.ScenarioName)
		}
	}

	return nil
}

// RemoveAssignedScenarios removes all the scenarios that are coming from any of the provided ASAs
func (s *service) RemoveAssignedScenarios(ctx context.Context, in []*model.AutomaticScenarioAssignment) error {
	for _, asa := range in {
		if err := s.RemoveAssignedScenario(ctx, *asa); err != nil {
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

// MergeScenariosFromInputLabelsAndAssignments merges all the scenarios that are part of the resource labels (already added + to be added with the current operation)
// with all the scenarios that should be assigned based on ASAs.
func (s *service) MergeScenariosFromInputLabelsAndAssignments(ctx context.Context, inputLabels map[string]interface{}, runtimeID string) ([]interface{}, error) {
	scenariosFromAssignments, err := s.GetScenariosFromMatchingASAs(ctx, runtimeID, graphql.FormationObjectTypeRuntime)
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

// GetScenariosFromMatchingASAs gets all the scenarios that should be added to the runtime based on the matching Automatic Scenario Assignments
// In order to do that, the ASAs should be searched in the caller tenant as this is the tenant that modifies the runtime and this is the tenant that the ASA
// produced labels should be added to.
func (s *service) GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	match, err := s.getMatchingFuncByFormationObjectType(objType)
	if err != nil {
		return nil, err
	}

	scenarioAssignments, err := s.repo.ListAll(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listinng Automatic Scenario Assignments in tenant: %s", tenantID)
	}

	matchingASAs := make([]*model.AutomaticScenarioAssignment, 0, len(scenarioAssignments))

	for _, scenarioAssignment := range scenarioAssignments {
		matches, err := match(ctx, scenarioAssignment, objectID)
		if err != nil {
			return nil, errors.Wrapf(err, "while checkig if asa matches runtime with ID %s", objectID)
		}
		if matches {
			matchingASAs = append(matchingASAs, scenarioAssignment)
		}
	}

	scenarios := make([]string, 0)
	for _, sa := range matchingASAs {
		scenarios = append(scenarios, sa.ScenarioName)
	}
	return scenarios, nil
}

type matchingFunc func(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error)

func (s *service) getMatchingFuncByFormationObjectType(objType graphql.FormationObjectType) (matchingFunc, error) {
	switch objType {
	case graphql.FormationObjectTypeRuntime:
		return s.isASAMatchingRuntime, nil
	case graphql.FormationObjectTypeRuntimeContext:
		return s.isASAMatchingRuntimeContext, nil
	}
	return nil, errors.Errorf("unexpected formation object type %q", objType)
}

func (s *service) isASAMatchingRuntime(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeID string) (bool, error) {
	return s.runtimeRepo.Exists(ctx, asa.TargetTenantID, runtimeID)
}

func (s *service) isASAMatchingRuntimeContext(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeContextID string) (bool, error) {
	return s.runtimeContextRepo.Exists(ctx, asa.TargetTenantID, runtimeContextID)
}

func (s *service) isFormationComingFromASA(ctx context.Context, objectID, formation string, objectType graphql.FormationObjectType) (bool, error) {
	formationsFromASA, err := s.GetScenariosFromMatchingASAs(ctx, objectID, objectType)
	if err != nil {
		return false, errors.Wrapf(err, "while getting formations from ASAs for %s with id: %q", objectType, objectID)
	}

	for _, formationFromASA := range formationsFromASA {
		if formation == formationFromASA {
			return true, nil
		}
	}

	return false, nil
}

func (s *service) modifyFormations(ctx context.Context, tnt, formationName string, modificationFunc modificationFunc) error {
	def, err := s.labelDefRepository.GetByKey(ctx, tnt, model.ScenariosKey)
	if err != nil {
		return errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	formations, err := labeldef.ParseFormationsFromSchema(def.Schema)
	if err != nil {
		return err
	}

	formations = modificationFunc(formations, formationName)

	schema, err := labeldef.NewSchemaForFormations(formations)
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

func (s *service) modifyAssignedFormations(ctx context.Context, tnt, objectID string, formation model.Formation, objectType model.LabelableObject, modificationFunc modificationFunc) error {
	labelInput := newLabelInput(formation.Name, objectID, objectType)

	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return err
	}

	existingFormations, err := label.ValueToStringsSlice(existingLabel.Value)
	if err != nil {
		return err
	}

	formations := modificationFunc(existingFormations, formation.Name)
	// can not set scenario label to empty value, violates the scenario label definition
	if len(formations) == 0 {
		return s.labelRepository.Delete(ctx, tnt, objectType, objectID, model.ScenariosKey)
	}

	labelInput.Value = formations
	labelInput.Version = existingLabel.Version
	return s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput)
}

type modificationFunc func([]string, string) []string

func addFormation(formations []string, formation string) []string {
	for _, f := range formations {
		if f == formation {
			return formations
		}
	}

	return append(formations, formation)
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
	if err := s.labelDefService.EnsureScenariosLabelDefinitionExists(ctx, tenantID); err != nil {
		return nil, errors.Wrap(err, "while ensuring that `scenarios` label definition exist")
	}

	out, err := s.labelDefService.GetAvailableScenarios(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting available scenarios")
	}
	return out, nil
}

func (s *service) createFormation(ctx context.Context, tenant, templateName, formationName string) (*model.Formation, error) {
	fTmpl, err := s.formationTemplateRepository.GetByName(ctx, templateName)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation template by name: %q: %v", templateName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation template by name: %q", templateName)
	}

	formation := &model.Formation{
		ID:                  s.uuidService.Generate(),
		TenantID:            tenant,
		FormationTemplateID: fTmpl.ID,
		Name:                formationName,
	}
	log.C(ctx).Debugf("Creating formation with name: %q and template ID: %q...", formationName, fTmpl.ID)
	if err = s.formationRepository.Create(ctx, formation); err != nil {
		log.C(ctx).Errorf("An error occurred while creating formation with name: %q and template ID: %q", formationName, fTmpl.ID)
		return nil, errors.Wrapf(err, "An error occurred while creating formation with name: %q and template ID: %q", formationName, fTmpl.ID)
	}

	return formation, nil
}

func (s *service) getFormationByName(ctx context.Context, formationName, tnt string) (*model.Formation, error) {
	// TODO:: Workaround for the DEFAULT scenario, because it is not in the 'formations' table, and getting it will fail.
	// Soon this label will be removed and then we can get rid of this check.
	if formationName == model.DefaultScenario {
		return &model.Formation{Name: model.DefaultScenario}, nil
	}

	f, err := s.formationRepository.GetByName(ctx, formationName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by name: %q: %v", formationName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by name: %q", formationName)
	}

	return f, nil
}
