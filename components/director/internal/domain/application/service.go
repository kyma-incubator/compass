package application

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ApplicationRepository -output=automock -outpkg=automock -case=underscore
type ApplicationRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	ListByScenarios(ctx context.Context, tenantID uuid.UUID, scenarios []string, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	Create(ctx context.Context, item *model.Application) error
	Update(ctx context.Context, item *model.Application) error
	Delete(ctx context.Context, item *model.Application) error
}

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error
	DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	ListAllByApplicationID(applicationID string) ([]*model.Document, error)
	CreateMany(items []*model.Document) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=WebhookRepository -output=automock -outpkg=automock -case=underscore
type WebhookRepository interface {
	ListByApplicationID(applicationID string) ([]*model.Webhook, error)
	CreateMany(items []*model.Webhook) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	CreateMany(items []*model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=EventAPIRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	ListByApplicationID(applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error)
	CreateMany(items []*model.EventAPIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=RuntimeRepository -output=automock -outpkg=automock -case=underscore
type RuntimeRepository interface {
	Exists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery -name=LabelUpsertService -output=automock -outpkg=automock -case=underscore
type LabelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
	UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error
}

//go:generate mockery -name=ScenariosService -output=automock -outpkg=automock -case=underscore
type ScenariosService interface {
	EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	appRepo      ApplicationRepository
	apiRepo      APIRepository
	eventAPIRepo EventAPIRepository
	documentRepo DocumentRepository
	webhookRepo  WebhookRepository
	labelRepo    LabelRepository
	runtimeRepo  RuntimeRepository

	labelUpsertService LabelUpsertService
	scenariosService   ScenariosService
	uidService         UIDService
}

func NewService(app ApplicationRepository, webhook WebhookRepository, api APIRepository, eventAPI EventAPIRepository, documentRepo DocumentRepository, runtimeRepo RuntimeRepository, labelRepo LabelRepository, labelUpsertService LabelUpsertService, scenariosService ScenariosService, uidService UIDService) *service {
	return &service{
		appRepo:            app,
		webhookRepo:        webhook,
		apiRepo:            api,
		eventAPIRepo:       eventAPI,
		documentRepo:       documentRepo,
		runtimeRepo:        runtimeRepo,
		labelRepo:          labelRepo,
		labelUpsertService: labelUpsertService,
		scenariosService:   scenariosService,
		uidService:         uidService}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.appRepo.List(ctx, appTenant, filter, pageSize, cursor)
}

func (s *service) ListByRuntimeID(ctx context.Context, runtimeID uuid.UUID, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	tenantID, err := tenant.LoadFromContext(ctx)

	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	tenantUUID, err := uuid.Parse(tenantID)
	if err != nil {
		return nil, errors.New("tenantID is not UUID")
	}

	exist, err := s.runtimeRepo.Exists(ctx, tenantID, runtimeID.String())
	if err != nil {
		return nil, errors.Wrap(err, "while checking if runtime exits")
	}

	if !exist {
		return nil, errors.New("runtime does not exist")
	}

	label, err := s.labelRepo.GetByKey(ctx, tenantID, model.RuntimeLabelableObject, runtimeID.String(), model.ScenariosKey)
	if err != nil {
		return nil, errors.Wrap(err, "while getting scenarios for runtime")
	}

	scenarios, err := getScenariosValues(label.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while converting scenarios labels")
	}
	if len(scenarios) == 0 {
		return &model.ApplicationPage{
			Data:       []*model.Application{},
			TotalCount: 0,
			PageInfo: &pagination.Page{
				StartCursor: "",
				EndCursor:   "",
				HasNextPage: false,
			},
		}, nil
	}

	return s.appRepo.ListByScenarios(ctx, tenantUUID, scenarios, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.appRepo.GetByID(ctx, appTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	return app, nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exist, err := s.appRepo.Exists(ctx, appTenant, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	return exist, nil
}

func (s *service) Create(ctx context.Context, in model.ApplicationInput) (string, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrap(err, "while loading tenant from context")
	}

	err = in.Validate()
	if err != nil {
		return "", errors.Wrap(err, "while validating Application input")
	}

	id := s.uidService.Generate()
	app := in.ToApplication(id, appTenant)

	// TODO: Checking if Label Definition exists could be moved after application creation, once application repository is ported to sql
	err = s.scenariosService.EnsureScenariosLabelDefinitionExists(ctx, appTenant)
	if err != nil {
		return "", err
	}

	if _, ok := in.Labels[model.ScenariosKey]; !ok {
		if in.Labels == nil {
			in.Labels = map[string]interface{}{}
		}
		in.Labels[model.ScenariosKey] = model.ScenariosDefaultValue
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, appTenant, model.ApplicationLabelableObject, id, in.Labels)
	if err != nil {
		return id, errors.Wrapf(err, "while creating multiple labels for Application")
	}

	err = s.appRepo.Create(ctx, app)
	if err != nil {
		return "", err
	}

	err = s.createRelatedResources(in, app.ID)
	if err != nil {
		return "", errors.Wrap(err, "while creating related Application resources")
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.ApplicationInput) error {
	err := in.Validate()
	if err != nil {
		return errors.Wrap(err, "while validating Application input")
	}

	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}
	app = in.ToApplication(app.ID, app.Tenant)

	err = s.appRepo.Update(ctx, app)
	if err != nil {
		return errors.Wrap(err, "while updating Application")
	}

	err = s.deleteRelatedResources(id)
	if err != nil {
		return errors.Wrap(err, "while deleting related Application resources")
	}

	err = s.labelRepo.DeleteAll(ctx, appTenant, model.ApplicationLabelableObject, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting all labels for Application")
	}

	err = s.createRelatedResources(in, app.ID)
	if err != nil {
		return errors.Wrap(err, "while creating related Application resources")
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, appTenant, model.ApplicationLabelableObject, id, in.Labels)
	if err != nil {
		return errors.Wrapf(err, "while creating multiple labels for Application")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	err = s.deleteRelatedResources(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting related Application resources")
	}

	err = s.appRepo.Delete(ctx, app)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application")
	}

	// TODO: Set cascade delete when implementing DB repository for Application domain
	err = s.labelRepo.DeleteAll(ctx, appTenant, model.ApplicationLabelableObject, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting all labels for Runtime")
	}

	return nil
}

func (s *service) SetLabel(ctx context.Context, labelInput *model.LabelInput) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, labelInput.ObjectID)
	if err != nil {
		return errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return fmt.Errorf("Application with ID %s doesn't exist", labelInput.ObjectID)
	}

	err = s.labelUpsertService.UpsertLabel(ctx, appTenant, labelInput)
	if err != nil {
		return errors.Wrapf(err, "while creating label for Application")
	}

	return nil
}

func (s *service) GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return nil, fmt.Errorf("Application with ID %s doesn't exist", applicationID)
	}

	label, err := s.labelRepo.GetByKey(ctx, appTenant, model.ApplicationLabelableObject, applicationID, key)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Application")
	}

	return label, nil
}

//TODO: In future consider using `map[string]*model.Label`
func (s *service) ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Application existence")
	}

	if !appExists {
		return nil, fmt.Errorf("Application with ID %s doesn't exist", applicationID)
	}

	labels, err := s.labelRepo.ListForObject(ctx, appTenant, model.ApplicationLabelableObject, applicationID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting label for Application")
	}

	return labels, nil
}

func (s *service) DeleteLabel(ctx context.Context, applicationID string, key string) error {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if key == model.ScenariosKey {
		return fmt.Errorf("%s label can not be deleted from application", model.ScenariosKey)
	}

	appExists, err := s.appRepo.Exists(ctx, appTenant, applicationID)
	if err != nil {
		return errors.Wrap(err, "while checking Application existence")
	}
	if !appExists {
		return fmt.Errorf("Application with ID %s doesn't exist", applicationID)
	}

	err = s.labelRepo.Delete(ctx, appTenant, model.ApplicationLabelableObject, applicationID, key)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application label")
	}

	return nil
}

func (s *service) createRelatedResources(in model.ApplicationInput, applicationID string) error {
	var err error

	var webhooks []*model.Webhook
	for _, item := range in.Webhooks {
		webhooks = append(webhooks, item.ToWebhook(s.uidService.Generate(), applicationID))
	}
	err = s.webhookRepo.CreateMany(webhooks)
	if err != nil {
		return errors.Wrapf(err, "while creating Webhooks for application")
	}

	var apis []*model.APIDefinition
	for _, item := range in.Apis {
		apis = append(apis, item.ToAPIDefinition(s.uidService.Generate(), applicationID))
	}

	err = s.apiRepo.CreateMany(apis)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	var eventAPIs []*model.EventAPIDefinition
	for _, item := range in.EventAPIs {
		eventAPIs = append(eventAPIs, item.ToEventAPIDefinition(s.uidService.Generate(), applicationID))
	}
	err = s.eventAPIRepo.CreateMany(eventAPIs)
	if err != nil {
		return errors.Wrapf(err, "while creating EventAPIs for application")
	}

	var documents []*model.Document
	for _, item := range in.Documents {
		documents = append(documents, item.ToDocument(s.uidService.Generate(), applicationID))
	}
	err = s.documentRepo.CreateMany(documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) deleteRelatedResources(applicationID string) error {
	var err error

	err = s.webhookRepo.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting Webhooks for application %s", applicationID)
	}

	err = s.apiRepo.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIs for application %s", applicationID)
	}

	err = s.eventAPIRepo.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIs for application %s", applicationID)
	}

	err = s.documentRepo.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting Documents for application %s", applicationID)
	}

	return nil
}

func getScenariosValues(labels interface{}) ([]string, error) {
	tmpScenarios, ok := labels.([]interface{})
	if !ok {
		return nil, errors.New("Cannot convert scenario labels to array of string")
	}

	var scenarios []string
	for _, label := range tmpScenarios {
		scenario, ok := label.(string)
		if !ok {
			return nil, errors.New("Cannot convert scenario label to string")
		}
		scenarios = append(scenarios, scenario)
	}

	return scenarios, nil
}
