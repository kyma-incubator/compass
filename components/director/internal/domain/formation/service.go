package formation

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	webhookdir "github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

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
	GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error)
	ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Runtime, error)
	ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error)
	ListByIDs(ctx context.Context, tenant string, ids []string) ([]*model.Runtime, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error)
	OwnerExistsByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (bool, error)
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error)
}

//go:generate mockery --exported --name=applicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplate, error)
}

//go:generate mockery --exported --name=runtimeContextRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeContextRepository interface {
	GetByRuntimeID(ctx context.Context, tenant, runtimeID string) (*model.RuntimeContext, error)
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
}

// FormationTemplateRepository represents the FormationTemplate repository layer
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	GetByName(ctx context.Context, templateName string) (*model.FormationTemplate, error)
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
}

//go:generate mockery --exported --name=webhookClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type webhookClient interface {
	Do(ctx context.Context, request *webhookclient.Request) (*webhookdir.Response, error)
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

//go:generate mockery --exported --name=formationAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type formationAssignmentService interface {
	Create(ctx context.Context, in *model.FormationAssignmentInput) (string, error)
	Get(ctx context.Context, id string) (*model.FormationAssignment, error)
	GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error)
	ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	Update(ctx context.Context, id string, in *model.FormationAssignmentInput) error
	Delete(ctx context.Context, id string) error
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
	formationAssignmentService    formationAssignmentService
	runtimeTypeLabelKey           string
	applicationTypeLabelKey       string
}

type FormationAssignmentRequestMapping struct {
	Request             *webhookclient.Request
	FormationAssignment *model.FormationAssignment
}

// NewService creates formation service
func NewService(labelDefRepository labelDefRepository, labelRepository labelRepository, formationRepository FormationRepository, formationTemplateRepository FormationTemplateRepository, labelService labelService, uuidService uuidService, labelDefService labelDefService, asaRepo automaticFormationAssignmentRepository, asaService automaticFormationAssignmentService, tenantSvc tenantService, runtimeRepo runtimeRepository, runtimeContextRepo runtimeContextRepository, webhookRepository webhookRepository, webhookClient webhookClient, applicationRepository applicationRepository, applicationTemplateRepository applicationTemplateRepository, webhookConverter webhookConverter, formationAssignmentService formationAssignmentService, runtimeTypeLabelKey, applicationTypeLabelKey string) *service {
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
		formationAssignmentService:    formationAssignmentService,
		runtimeTypeLabelKey:           runtimeTypeLabelKey,
		applicationTypeLabelKey:       applicationTypeLabelKey,
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
//
// When objectType graphql.FormationObjectType is graphql.FormationObjectTypeApplication, graphql.FormationObjectTypeRuntime and
// graphql.FormationObjectTypeRuntimeContext it adds the provided formation to the scenario label of the entity if such exists,
// otherwise new scenario label is created for the entity with the provided formation.
//
// Additionally, notifications are sent to the interested participants for that formation change.
// 		- If objectType is graphql.FormationObjectTypeApplication:
//				- A notification about the assigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//  			- A notification about the assigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//				- If the assigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
// 		- If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext, and the runtime has configuration change webhook,
//			a notification about each application in the formation is sent to this runtime.
//
// If the graphql.FormationObjectType is graphql.FormationObjectTypeTenant it will
// create automatic scenario assignment with the caller and target tenant which then will assign the right Runtime / RuntimeContexts based on the formation template's runtimeType.
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
		targetTenantID, err := s.tenantSvc.GetInternalTenant(ctx, objectID)
		if err != nil {
			return nil, err
		}

		if _, err = s.CreateAutomaticScenarioAssignment(ctx, newAutomaticScenarioAssignmentModel(formation.Name, tnt, targetTenantID)); err != nil {
			return nil, err
		}
		return s.getFormationByName(ctx, formation.Name, tnt)
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) isValidRuntimeType(ctx context.Context, tnt string, runtimeID string, formation *model.Formation) error {
	formationTemplate, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		return errors.Wrapf(err, "while getting formation template with ID %q", formation.FormationTemplateID)
	}
	runtimeTypeLabel, err := s.labelService.GetLabel(ctx, tnt, &model.LabelInput{
		Key:        s.runtimeTypeLabelKey,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	})
	if err != nil {
		return errors.Wrapf(err, "while getting label %q for runtime with ID %q", s.runtimeTypeLabelKey, runtimeID)
	}

	if runtimeType, ok := runtimeTypeLabel.Value.(string); !ok || runtimeType != formationTemplate.RuntimeType {
		return apperrors.NewInvalidOperationError(fmt.Sprintf("unsupported runtimeType %q for formation template %q, allowing only %q", runtimeType, formationTemplate.Name, formationTemplate.RuntimeType))
	}
	return nil
}

func (s *service) generateAssignments(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation) ([]*model.FormationAssignment, error) {
	applications, err := s.applicationRepository.ListByScenariosNoPaging(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}
	runtimes, err := s.runtimeRepo.ListByScenarios(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}
	runtimeContexts, err := s.runtimeContextRepo.ListByScenarios(ctx, tnt, []string{formation.Name})
	if err != nil {
		return nil, err
	}
	assignments := make([]*model.FormationAssignment, 0, (len(applications)+len(runtimes))*2)
	for _, app := range applications {
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      app.ID,
			TargetType:  string(graphql.FormationObjectTypeApplication),
			Value:       nil,
		})
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      app.ID,
			SourceType:  string(graphql.FormationObjectTypeApplication),
			Target:      objectID,
			TargetType:  string(objectType),
			Value:       nil,
		})
	}
	for _, runtime := range runtimes {
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      runtime.ID,
			TargetType:  string(graphql.FormationObjectTypeRuntime),
			Value:       nil,
		})
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      runtime.ID,
			SourceType:  string(graphql.FormationObjectTypeRuntime),
			Target:      objectID,
			TargetType:  string(objectType),
			Value:       nil,
		})
	}
	for _, runtimeCtx := range runtimeContexts {
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      runtimeCtx.ID,
			TargetType:  string(graphql.FormationObjectTypeRuntimeContext),
			Value:       nil,
		})
		assignments = append(assignments, &model.FormationAssignment{
			FormationID: formation.ID,
			TenantID:    tnt,
			Source:      runtimeCtx.ID,
			SourceType:  string(graphql.FormationObjectTypeRuntimeContext),
			Target:      objectID,
			TargetType:  string(objectType),
			Value:       nil,
		})
	}
	return assignments, nil
}

func (s *service) matchFormationAssignmentsWithRequests(assignments []*model.FormationAssignment, requests []*webhookclient.Request) []*FormationAssignmentRequestMapping {
	formationAssignmentMapping := make([]*FormationAssignmentRequestMapping, 0, len(assignments))
	for i, assignment := range assignments {
		mappingObject := &FormationAssignmentRequestMapping{
			Request:             nil,
			FormationAssignment: assignments[i],
		}
		for j, request := range requests {
			var objectID string
			if request.Webhook.RuntimeID != nil {
				objectID = *request.Webhook.RuntimeID
			}
			if request.Webhook.ApplicationID != nil {
				objectID = *request.Webhook.ApplicationID
			}
			if objectID != assignment.Target {
				continue
			}
			participants := request.Object.GetParticipants()
			for _, id := range participants {
				if assignment.Source == id {
					mappingObject.Request = requests[j]
					break
				}
			}
		}
		formationAssignmentMapping = append(formationAssignmentMapping, mappingObject)
	}
	return formationAssignmentMapping
}

func (s *service) createWebhookRequest(ctx context.Context, webhook *model.Webhook, input webhookdir.TemplateInput) (*webhookclient.Request, error) {
	gqlWebhook, err := s.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting webhook with ID %s", webhook.ID)
	}
	return &webhookclient.Request{
		Webhook:       *gqlWebhook,
		Object:        input,
		CorrelationID: correlation.CorrelationIDFromContext(ctx),
	}, nil
}

func (s *service) sendNotifications(ctx context.Context, notifications []*webhookclient.Request) error {
	log.C(ctx).Infof("Sending %d notifications", len(notifications))
	for i, notification := range notifications {
		log.C(ctx).Infof("Sending notification %d out of %d for webhook with ID %s", i+1, len(notifications), notification.Webhook.ID)
		if _, err := s.webhookClient.Do(ctx, notification); err != nil {
			return errors.Wrapf(err, "while executing webhook with ID %s", notification.Webhook.ID)
		}
		log.C(ctx).Infof("Successfully sent notification %d out of %d for webhook with %s", i+1, len(notifications), notification.Webhook.ID)
	}
	return nil
}

func (s *service) assign(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	formationFromDB, err := s.getFormationByName(ctx, formation.Name, tnt)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation %q", formation.Name)
	}
	if err = s.checkFormationTemplateTypes(ctx, tnt, objectID, objectType, formationFromDB); err != nil {
		return nil, err
	}

	if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), addFormation); err != nil {
		if apperrors.IsNotFoundError(err) {
			labelInput := newLabelInput(formation.Name, objectID, objectTypeToLabelableObject(objectType))
			if err = s.labelService.CreateLabel(ctx, tnt, s.uuidService.Generate(), labelInput); err != nil {
				return nil, err
			}
			return formationFromDB, nil
		}
		return nil, err
	}

	return formationFromDB, nil
}

func (s *service) checkFormationTemplateTypes(ctx context.Context, tnt, objectID string, objectType graphql.FormationObjectType, formation *model.Formation) error {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		if err := s.isValidApplicationType(ctx, tnt, objectID, formation); err != nil {
			return errors.Wrapf(err, "while validating application type for application %q", objectID)
		}
	case graphql.FormationObjectTypeRuntime:
		if err := s.isValidRuntimeType(ctx, tnt, objectID, formation); err != nil {
			return errors.Wrapf(err, "while validating runtime type")
		}
	case graphql.FormationObjectTypeRuntimeContext:
		runtimeCtx, err := s.runtimeContextRepo.GetByID(ctx, tnt, objectID)
		if err != nil {
			return errors.Wrapf(err, "while getting runtime context")
		}
		if err = s.isValidRuntimeType(ctx, tnt, runtimeCtx.RuntimeID, formation); err != nil {
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
// Additionally, a notification is sent to each runtime that needs to be notified (has a configuration change webhook) and is part of the formation either directly or via runtimeContext.
// 		- If objectType is graphql.FormationObjectTypeApplication:
//				- A notification about the unassigned application is sent to all the runtimes that are in the formation (either directly or via runtimeContext) and has configuration change webhook.
//  			- A notification about the unassigned application is sent to all the applications that are in the formation and has application tenant mapping webhook.
//				- If the unassigned application has an application tenant mapping webhook, a notification about each application in the formation is sent to this application.
// 		- If objectType is graphql.FormationObjectTypeRuntime or graphql.FormationObjectTypeRuntimeContext, and the runtime has configuration change webhook,
//			a notification for each application in the formation is sent to this runtime.
//
// For objectType graphql.FormationObjectTypeTenant it will
// delete the automatic scenario assignment with the caller and target tenant which then will unassign the right Runtime / RuntimeContexts based on the formation template's runtimeType.
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
		requests, err := s.generateNotifications(ctx, tnt, objectID, formationFromDB, model.UnassignFormation, objectType)
		if err != nil {
			return nil, errors.Wrapf(err, "while generating notifications for %s unassignment", objectType)
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

		formationFromDB, err := s.getFormationByName(ctx, formation.Name, tnt)
		if err != nil {
			return nil, err
		}

		if err := s.modifyAssignedFormations(ctx, tnt, objectID, formation, objectTypeToLabelableObject(objectType), deleteFormation); err != nil {
			if apperrors.IsNotFoundError(err) {
				return formationFromDB, nil
			}
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

func (s *service) generateNotifications(ctx context.Context, tenant, objectID string, formation *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.Request, error) {
	switch objectType {
	case graphql.FormationObjectTypeApplication:
		rtNotifications, err := s.generateRuntimeNotificationsForApplicationAssignment(ctx, tenant, objectID, formation, operation)
		if err != nil {
			return nil, err
		}
		appNotifications, err := s.generateApplicationNotificationsForApplicationAssignment(ctx, tenant, objectID, formation, operation)
		if err != nil {
			return nil, err
		}
		return append(rtNotifications, appNotifications...), nil
	case graphql.FormationObjectTypeRuntime:
		return s.generateRuntimeNotificationsForRuntimeAssignment(ctx, tenant, objectID, formation, operation)
	case graphql.FormationObjectTypeRuntimeContext:
		return s.generateRuntimeNotificationsForRuntimeContextAssignment(ctx, tenant, objectID, formation, operation)
	default:
		return nil, fmt.Errorf("unknown formation type %s", objectType)
	}
}

func (s *service) generateApplicationNotificationsForApplicationAssignment(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation) ([]*webhookclient.Request, error) {
	log.C(ctx).Infof("Generating %s app-to-app formation notifications for application %s", operation, appID)
	application, err := s.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application with id %s", appID)
	}
	applicationLabels, err := s.getLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels for application with id %s", appID)
	}
	applicationWithLabels := &webhookdir.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
	}

	var appTemplateWithLabels *webhookdir.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := s.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting application template with id %s", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := s.getLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting labels for application template with id %s", appTemplate.ID)
		}
		appTemplateWithLabels = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: appTemplate,
			Labels:              applicationTemplateLabels,
		}
	}

	webhooks, err := s.webhookRepository.ListByReferenceObjectTypeAndWebhookType(ctx, tenant, model.WebhookTypeApplicationTenantMapping, model.ApplicationWebhookReference)
	if err != nil {
		return nil, errors.Wrap(err, "when listing application tenant mapping webhooks for applications")
	}

	listeningAppIDs := make(map[string]bool, len(webhooks))
	for _, wh := range webhooks {
		listeningAppIDs[wh.ObjectID] = true
	}

	if len(listeningAppIDs) == 0 {
		log.C(ctx).Infof("There are no applications is listening for app-to-app formation notifications in tenant %s", tenant)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d applications listening for app-to-app formation notifications in tenant %s", len(listeningAppIDs), tenant)

	requests := make([]*webhookclient.Request, 0, len(listeningAppIDs))
	if listeningAppIDs[appID] {
		log.C(ctx).Infof("The application with ID %s that is being %s is also listening for app-to-app formation notifications. Will create notifications about all other apps in the formation...", appID, operation)
		var webhook *model.Webhook
		for i := range webhooks {
			if webhooks[i].ObjectID == appID {
				webhook = webhooks[i]
			}
		}

		applicationMappingsToBeNotifiedFor, applicationTemplatesMapping, err := s.prepareApplicationMappingsInFormation(ctx, tenant, formation, appID)
		if err != nil {
			return nil, err
		}

		appsInFormationCountExcludingAppCurrentlyAssigned := len(applicationMappingsToBeNotifiedFor)
		if operation == model.AssignFormation {
			appsInFormationCountExcludingAppCurrentlyAssigned -= 1
		}

		log.C(ctx).Infof("There are %d applications in formation %s. Notification will be sent about them to application with id %s that is being %s.", appsInFormationCountExcludingAppCurrentlyAssigned, formation.Name, appID, operation)

		for _, sourceApp := range applicationMappingsToBeNotifiedFor {
			if sourceApp.ID == appID {
				continue // Do not notify about itself
			}
			var appTemplate *webhookdir.ApplicationTemplateWithLabels
			if sourceApp.ApplicationTemplateID != nil {
				appTemplate = applicationTemplatesMapping[*sourceApp.ApplicationTemplateID]
			} else {
				log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for source application in the input for webhook %s", sourceApp.ID, webhook.ID)
			}
			if appTemplateWithLabels == nil {
				log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for target application in the input for webhook %s", appID, webhook.ID)
			}
			input := &webhookdir.ApplicationTenantMappingInput{
				Operation:                 operation,
				FormationID:               formation.ID,
				SourceApplicationTemplate: appTemplate,
				SourceApplication:         sourceApp,
				TargetApplicationTemplate: appTemplateWithLabels,
				TargetApplication:         applicationWithLabels,
			}
			req, err := s.createWebhookRequest(ctx, webhook, input)
			if err != nil {
				return nil, err
			}
			requests = append(requests, req)
		}

		delete(listeningAppIDs, appID)
	}

	listeningAppsInScenario, err := s.applicationRepository.ListByScenariosAndIDs(ctx, tenant, []string{formation.Name}, setToSlice(listeningAppIDs))
	if err != nil {
		return nil, errors.Wrapf(err, "while listing applications in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d out of %d applications listening for app-to-app formation notifications in tenant %s that are in scenario %s", len(listeningAppsInScenario), len(listeningAppIDs), tenant, formation.Name)

	appIDsToBeNotified := make(map[string]bool, len(listeningAppsInScenario))
	applicationsTemplateIDs := make([]string, 0, len(listeningAppsInScenario))
	for _, app := range listeningAppsInScenario {
		appIDsToBeNotified[app.ID] = true
		if app.ApplicationTemplateID != nil {
			applicationsTemplateIDs = append(applicationsTemplateIDs, *app.ApplicationTemplateID)
		}
	}

	listeningAppsLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.ApplicationLabelableObject, setToSlice(appIDsToBeNotified))
	if err != nil {
		return nil, errors.Wrap(err, "while listing application labels")
	}

	listeningAppsMapping := make(map[string]*webhookdir.ApplicationWithLabels, len(listeningAppsInScenario))
	for i, app := range listeningAppsInScenario {
		listeningAppsMapping[app.ID] = &webhookdir.ApplicationWithLabels{
			Application: listeningAppsInScenario[i],
			Labels:      listeningAppsLabels[app.ID],
		}
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

	webhooksToCall := make(map[string]*model.Webhook, len(appIDsToBeNotified))
	for i := range webhooks {
		if appIDsToBeNotified[webhooks[i].ObjectID] {
			webhooksToCall[webhooks[i].ObjectID] = webhooks[i]
		}
	}

	for _, targetApp := range listeningAppsMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if targetApp.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*targetApp.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for the target application in the input for webhook %s", targetApp.ID, webhooksToCall[targetApp.ID].ID)
		}
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template for source application in the input for webhook %s", appID, webhooksToCall[targetApp.ID].ID)
		}
		input := &webhookdir.ApplicationTenantMappingInput{
			Operation:                 operation,
			FormationID:               formation.ID,
			SourceApplicationTemplate: appTemplateWithLabels,
			SourceApplication:         applicationWithLabels,
			TargetApplicationTemplate: appTemplate,
			TargetApplication:         targetApp,
		}
		req, err := s.createWebhookRequest(ctx, webhooksToCall[targetApp.ID], input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	log.C(ctx).Infof("Total number of app-to-app notifications for application with ID %s that is being %s is %d", appID, operation, len(requests))

	return requests, nil
}

func (s *service) generateRuntimeNotificationsForApplicationAssignment(ctx context.Context, tenant string, appID string, formation *model.Formation, operation model.FormationOperation) ([]*webhookclient.Request, error) {
	log.C(ctx).Infof("Generating %s notifications for application %s", operation, appID)
	application, err := s.applicationRepository.GetByID(ctx, tenant, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application with id %s", appID)
	}
	applicationLabels, err := s.getLabelsForObject(ctx, tenant, appID, model.ApplicationLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels for application with id %s", appID)
	}
	applicationWithLabels := &webhookdir.ApplicationWithLabels{
		Application: application,
		Labels:      applicationLabels,
	}

	var appTemplateWithLabels *webhookdir.ApplicationTemplateWithLabels
	if application.ApplicationTemplateID != nil {
		appTemplate, err := s.applicationTemplateRepository.Get(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting application template with id %s", *application.ApplicationTemplateID)
		}
		applicationTemplateLabels, err := s.getLabelsForObject(ctx, tenant, appTemplate.ID, model.AppTemplateLabelableObject)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting labels for application template with id %s", appTemplate.ID)
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
		log.C(ctx).Infof("There are no runtimes is listening for formation notifications in tenant %s", tenant)
		return nil, nil
	}

	log.C(ctx).Infof("There are %d runtimes listening for formation notifications in tenant %s", len(listeningRuntimeIDs), tenant)

	listeningRuntimes, err := s.runtimeRepo.ListByIDs(ctx, tenant, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtimes")
	}

	listeningRuntimesLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeLabelableObject, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing runtime labels")
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
		return nil, errors.Wrapf(err, "while listing runtimes in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d out of %d runtimes listening for formation notifications in tenant %s that are in scenario %s", len(listeningRuntimesInScenario), len(listeningRuntimeIDs), tenant, formation.Name)

	runtimeContextsInScenarioForListeningRuntimes, err := s.runtimeContextRepo.ListByScenariosAndRuntimeIDs(ctx, tenant, []string{formation.Name}, listeningRuntimeIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing runtime contexts in scenario %s", formation.Name)
	}

	log.C(ctx).Infof("There are %d runtime contexts in tenant %s that are in scenario %s and are for any of the listening runtimes", len(runtimeContextsInScenarioForListeningRuntimes), tenant, formation.Name)

	runtimeContextsInScenarioForListeningRuntimesIDs := make([]string, 0, len(runtimeContextsInScenarioForListeningRuntimes))
	for _, rtCtx := range runtimeContextsInScenarioForListeningRuntimes {
		runtimeContextsInScenarioForListeningRuntimesIDs = append(runtimeContextsInScenarioForListeningRuntimesIDs, rtCtx.ID)
	}

	runtimeContextsLables, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.RuntimeContextLabelableObject, runtimeContextsInScenarioForListeningRuntimesIDs)
	if err != nil {
		return nil, errors.Wrap(err, "while listing labels for runtime contexts")
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

	requests := make([]*webhookclient.Request, 0, len(runtimeIDsToBeNotified))
	for rtID := range runtimeIDsToBeNotified {
		rtCtx := runtimeContextsInScenarioForListeningRuntimesMapping[rtID]
		if rtCtx == nil {
			log.C(ctx).Infof("There is no runtime context for runtime %s in scenario %s. Will proceed without runtime context in the input for webhook %s", rtID, formation.Name, webhooksToCall[rtID].ID)
		}
		runtime := listeningRuntimesMapping[rtID]
		if appTemplateWithLabels == nil {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", appID, webhooksToCall[rtID].ID)
		}
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

func (s *service) generateRuntimeNotificationsForRuntimeContextAssignment(ctx context.Context, tenant, runtimeCtxID string, formation *model.Formation, operation model.FormationOperation) ([]*webhookclient.Request, error) {
	log.C(ctx).Infof("Generating %s notifications for runtime context %s", operation, runtimeCtxID)
	runtimeCtx, err := s.runtimeContextRepo.GetByID(ctx, tenant, runtimeCtxID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context with id %s", runtimeCtxID)
	}
	runtimeCtxLabels, err := s.getLabelsForObject(ctx, tenant, runtimeCtxID, model.RuntimeContextLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime context labels with id %s", runtimeCtxID)
	}

	runtimeCtxWithLabels := &webhookdir.RuntimeContextWithLabels{
		RuntimeContext: runtimeCtx,
		Labels:         runtimeCtxLabels,
	}

	requests, err := s.generateRuntimeNotificationsForRuntimeAssignment(ctx, tenant, runtimeCtxWithLabels.RuntimeID, formation, operation)
	if err != nil {
		return nil, err
	}
	for _, request := range requests {
		request.Object.(*webhookdir.FormationConfigurationChangeInput).RuntimeContext = runtimeCtxWithLabels
	}
	return requests, nil
}

func (s *service) generateRuntimeNotificationsForRuntimeAssignment(ctx context.Context, tenant, runtimeID string, formation *model.Formation, operation model.FormationOperation) ([]*webhookclient.Request, error) {
	log.C(ctx).Infof("Generating %s notifications for runtime %s", operation, runtimeID)
	runtime, err := s.runtimeRepo.GetByID(ctx, tenant, runtimeID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime with id %s", runtimeID)
	}
	runtimeLabels, err := s.getLabelsForObject(ctx, tenant, runtimeID, model.RuntimeLabelableObject)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting runtime labels for id %s", runtimeID)
	}
	runtimeWithLabels := &webhookdir.RuntimeWithLabels{
		Runtime: runtime,
		Labels:  runtimeLabels,
	}

	webhook, err := s.webhookRepository.GetByIDAndWebhookType(ctx, tenant, runtimeID, model.RuntimeWebhookReference, model.WebhookTypeConfigurationChanged)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There is no configuration chaged webhook for runtime %s. There are no notifications to be generated.", runtimeID)
			return nil, nil
		}
		return nil, errors.Wrapf(err, "while listing configuration changed webhooks for runtime %s", runtimeID)
	}

	applicationMapping, applicationTemplatesMapping, err := s.prepareApplicationMappingsInFormation(ctx, tenant, formation, runtimeID)
	if err != nil {
		return nil, err
	}

	requests := make([]*webhookclient.Request, 0, len(applicationMapping))
	for _, app := range applicationMapping {
		var appTemplate *webhookdir.ApplicationTemplateWithLabels
		if app.ApplicationTemplateID != nil {
			appTemplate = applicationTemplatesMapping[*app.ApplicationTemplateID]
		} else {
			log.C(ctx).Infof("Application %s has no application template. Will proceed without application template in the input for webhook %s", app.ID, webhook.ID)
		}
		input := &webhookdir.FormationConfigurationChangeInput{
			Operation:           operation,
			FormationID:         formation.ID,
			ApplicationTemplate: appTemplate,
			Application:         app,
			Runtime:             runtimeWithLabels,
			RuntimeContext:      nil,
		}
		req, err := s.createWebhookRequest(ctx, webhook, input)
		if err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *service) prepareApplicationMappingsInFormation(ctx context.Context, tenant string, formation *model.Formation, targetID string) (map[string]*webhookdir.ApplicationWithLabels, map[string]*webhookdir.ApplicationTemplateWithLabels, error) {
	applicationsToBeNotifiedFor, err := s.applicationRepository.ListByScenariosNoPaging(ctx, tenant, []string{formation.Name})
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing scenario labels for applications")
	}
	if len(applicationsToBeNotifiedFor) == 0 {
		log.C(ctx).Infof("There are no applications in scenario %s. No notifications will be generated for %s", formation.Name, targetID)
		return nil, nil, nil
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
		return nil, nil, errors.Wrap(err, "while listing labels for applications")
	}
	applicationMapping := make(map[string]*webhookdir.ApplicationWithLabels, len(applicationsToBeNotifiedForIDs))
	for i, app := range applicationsToBeNotifiedFor {
		applicationMapping[app.ID] = &webhookdir.ApplicationWithLabels{
			Application: applicationsToBeNotifiedFor[i],
			Labels:      applicationsToBeNotifiedForLabels[app.ID],
		}
	}

	applicationTemplates, err := s.applicationTemplateRepository.ListByIDs(ctx, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing application templates")
	}
	applicationTemplatesLabels, err := s.labelRepository.ListForObjectIDs(ctx, tenant, model.AppTemplateLabelableObject, applicationsTemplateIDs)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while listing labels for application templates")
	}
	applicationTemplatesMapping := make(map[string]*webhookdir.ApplicationTemplateWithLabels, len(applicationTemplates))
	for i, appTemplate := range applicationTemplates {
		applicationTemplatesMapping[appTemplate.ID] = &webhookdir.ApplicationTemplateWithLabels{
			ApplicationTemplate: applicationTemplates[i],
			Labels:              applicationTemplatesLabels[appTemplate.ID],
		}
	}

	return applicationMapping, applicationTemplatesMapping, nil
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
	runtimeType, err := s.getFormationTemplateRuntimeType(ctx, in.ScenarioName, in.Tenant)
	if err != nil {
		return err
	}

	lblFilters := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtimeType", fmt.Sprintf("\"%s\"", runtimeType))}

	ownedRuntimes, err := s.runtimeRepo.ListOwnedRuntimes(ctx, in.TargetTenantID, lblFilters)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range ownedRuntimes {
		hasRuntimeContext, err := s.runtimeContextRepo.ExistsByRuntimeID(ctx, in.TargetTenantID, r.ID)
		if err != nil {
			return errors.Wrapf(err, "while getting runtime contexts for runtime with id %q", r.ID)
		}

		// If the runtime has runtime context that is so called "multi-tenant" runtime, and we should not assign the runtime to formation.
		// In such cases only the runtime context should be assigned to formation. That happens with the "for" cycle below.
		if hasRuntimeContext {
			continue
		}

		// If the runtime has not runtime context, then it's a "single tenant" runtime, and we have to assign it to formation.
		if _, err = processScenarioFunc(ctx, in.Tenant, r.ID, graphql.FormationObjectTypeRuntime, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime with id %s from formation %s coming from ASA", processingType, r.ID, in.ScenarioName)
		}
	}

	// The part below covers the "multi-tenant" runtime case that we skipped above and
	// gets all runtimes(with and without owner access) and assign every runtime context(if there is any) for each of the runtimes to formation.
	runtimes, err := s.runtimeRepo.ListAll(ctx, in.TargetTenantID, lblFilters)
	if err != nil {
		return errors.Wrapf(err, "while fetching runtimes in target tenant: %s", in.TargetTenantID)
	}

	for _, r := range runtimes {
		runtimeContext, err := s.runtimeContextRepo.GetByRuntimeID(ctx, in.TargetTenantID, r.ID)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				continue
			}
			return errors.Wrapf(err, "while getting runtime context for runtime with ID: %q", r.ID)
		}

		if _, err = processScenarioFunc(ctx, in.Tenant, runtimeContext.ID, graphql.FormationObjectTypeRuntimeContext, model.Formation{Name: in.ScenarioName}); err != nil {
			return errors.Wrapf(err, "while %s runtime context with id %s from formation %s coming from ASA", processingType, runtimeContext.ID, in.ScenarioName)
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
	log.C(ctx).Infof("Getting scenarios matching from ASA for object with ID: %q and type: %q", objectID, objType)
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	matchFunc, err := s.getMatchingFuncByFormationObjectType(objType)
	if err != nil {
		return nil, err
	}

	scenarioAssignments, err := s.repo.ListAll(ctx, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listinng Automatic Scenario Assignments in tenant: %s", tenantID)
	}
	log.C(ctx).Infof("Found %d ASA(s) in tenant with ID: %q", len(scenarioAssignments), tenantID)

	matchingASAs := make([]*model.AutomaticScenarioAssignment, 0, len(scenarioAssignments))
	for _, scenarioAssignment := range scenarioAssignments {
		matches, err := matchFunc(ctx, scenarioAssignment, objectID)
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
	log.C(ctx).Infof("Matched scenarios from ASA are: %v", scenarios)

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
	runtimeType, err := s.getFormationTemplateRuntimeType(ctx, asa.ScenarioName, asa.Tenant)
	if err != nil {
		return false, err
	}

	lblFilters := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery("runtimeType", fmt.Sprintf("\"%s\"", runtimeType))}

	runtimeExists, err := s.runtimeRepo.OwnerExistsByFiltersAndID(ctx, asa.TargetTenantID, runtimeID, lblFilters)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if runtime with id %q have owner=true", runtimeID)
	}

	if !runtimeExists {
		return false, nil
	}

	// If the runtime has runtime contexts then it's a "multi-tenant" runtime, and it should NOT be matched by the ASA and should NOT be added to formation.
	hasRuntimeContext, err := s.runtimeContextRepo.ExistsByRuntimeID(ctx, asa.TargetTenantID, runtimeID)
	if err != nil {
		return false, errors.Wrapf(err, "while cheking runtime context existence for runtime with ID: %q", runtimeID)
	}

	return !hasRuntimeContext, nil
}

func (s *service) isASAMatchingRuntimeContext(ctx context.Context, asa *model.AutomaticScenarioAssignment, runtimeContextID string) (bool, error) {
	runtimeType, err := s.getFormationTemplateRuntimeType(ctx, asa.ScenarioName, asa.Tenant)
	if err != nil {
		return false, err
	}

	runtimeTypeKey := "runtimeType"
	lblFilters := []*labelfilter.LabelFilter{labelfilter.NewForKeyWithQuery(runtimeTypeKey, fmt.Sprintf("\"%s\"", runtimeType))}

	rtmCtx, err := s.runtimeContextRepo.GetByID(ctx, asa.TargetTenantID, runtimeContextID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "while getting runtime contexts with ID: %q", runtimeContextID)
	}

	_, err = s.runtimeRepo.GetByFiltersAndID(ctx, asa.TargetTenantID, rtmCtx.RuntimeID, lblFilters)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "while getting runtime with ID: %q and label with key: %q and value: %q", rtmCtx.RuntimeID, runtimeTypeKey, runtimeType)
	}

	return true, nil
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

	formations, err = modificationFunc(formations, formationName)
	if err != nil {
		return err
	}

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
	log.C(ctx).Infof("Modifying formation with name: %q for object with type: %q and ID: %q", formation.Name, objectType, objectID)

	labelInput := newLabelInput(formation.Name, objectID, objectType)
	existingLabel, err := s.labelService.GetLabel(ctx, tnt, labelInput)
	if err != nil {
		return err
	}

	existingFormations, err := label.ValueToStringsSlice(existingLabel.Value)
	if err != nil {
		return err
	}

	formations, err := modificationFunc(existingFormations, formation.Name)
	if err != nil {
		return err
	}

	// can not set scenario label to empty value, violates the scenario label definition
	if len(formations) == 0 {
		return s.labelRepository.Delete(ctx, tnt, objectType, objectID, model.ScenariosKey)
	}

	labelInput.Value = formations
	labelInput.Version = existingLabel.Version
	return s.labelService.UpdateLabel(ctx, tnt, existingLabel.ID, labelInput)
}

type modificationFunc func([]string, string) ([]string, error)

func addFormation(formations []string, formation string) ([]string, error) {
	for _, f := range formations {
		if f == formation {
			return nil, apperrors.NewNotUniqueErrorWithMessage(resource.Formations, fmt.Sprintf("Formation %s already exists", formation))
		}
	}

	return append(formations, formation), nil
}

func deleteFormation(formations []string, formation string) ([]string, error) {
	filteredFormations := make([]string, 0, len(formations))
	for _, f := range formations {
		if f != formation {
			filteredFormations = append(filteredFormations, f)
		}
	}

	return filteredFormations, nil
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
	f, err := s.formationRepository.GetByName(ctx, formationName, tnt)
	if err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation by name: %q: %v", formationName, err)
		return nil, errors.Wrapf(err, "An error occurred while getting formation by name: %q", formationName)
	}

	return f, nil
}

func (s *service) getFormationTemplateRuntimeType(ctx context.Context, scenarioName, tenant string) (string, error) {
	log.C(ctx).Debugf("Getting formation with name: %q in tenant: %q", scenarioName, tenant)
	formation, err := s.formationRepository.GetByName(ctx, scenarioName, tenant)
	if err != nil {
		return "", errors.Wrapf(err, "while getting formation by name %q", scenarioName)
	}

	log.C(ctx).Debugf("Getting formation template with ID: %q", formation.FormationTemplateID)
	formationTemplate, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		return "", errors.Wrapf(err, "while getting formation template by id %q", formation.FormationTemplateID)
	}

	return formationTemplate.RuntimeType, nil
}

func (s *service) isValidApplicationType(ctx context.Context, tnt string, applicationID string, formation *model.Formation) error {
	formationTemplate, err := s.formationTemplateRepository.Get(ctx, formation.FormationTemplateID)
	if err != nil {
		return errors.Wrapf(err, "while getting formation template with ID %q", formation.FormationTemplateID)
	}
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
