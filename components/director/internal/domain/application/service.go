package application

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ApplicationRepository -output=automock -outpkg=automock -case=underscore
type ApplicationRepository interface {
	GetByID(tenant, id string) (*model.Application, error)
	List(tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	Create(item *model.Application) error
	Update(item *model.Application) error
	Delete(item *model.Application) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	ListAllByApplicationID(applicationID string) ([]*model.Document, error)
	CreateMany(items []*model.Document) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=WebhookRepository -output=automock -outpkg=automock -case=underscore
type WebhookRepository interface {
	ListByApplicationID(applicationID string) ([]*model.ApplicationWebhook, error)
	CreateMany(items []*model.ApplicationWebhook) error
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

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	app        ApplicationRepository
	api        APIRepository
	eventAPI   EventAPIRepository
	document   DocumentRepository
	webhook    WebhookRepository
	uidService UIDService
}

func NewService(app ApplicationRepository, webhook WebhookRepository, api APIRepository, eventAPI EventAPIRepository, document DocumentRepository, uidService UIDService) *service {
	return &service{app: app, webhook: webhook, api: api, eventAPI: eventAPI, document: document, uidService: uidService}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.app.List(appTenant, filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Application, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	app, err := s.app.GetByID(appTenant, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	return app, nil
}

func (s *service) Create(ctx context.Context, in model.ApplicationInput) (string, error) {
	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	id := s.uidService.Generate()
	app := in.ToApplication(id, appTenant)

	err = s.app.Create(app)
	if err != nil {
		return "", err
	}

	err = s.createRelatedResources(in, app.ID)
	if err != nil {
		return "", errors.Wrapf(err, "while creating related Application resources")
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.ApplicationInput) error {
	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	app = in.ToApplication(app.ID, app.Tenant)

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application")
	}

	err = s.deleteRelatedResources(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting related Application resources")
	}

	err = s.createRelatedResources(in, app.ID)
	if err != nil {
		return errors.Wrapf(err, "while creating related Application resources")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	err = s.deleteRelatedResources(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting related Application resources")
	}

	return s.app.Delete(app)
}

func (s *service) AddLabel(ctx context.Context, applicationID string, key string, values []string) error {
	app, err := s.Get(ctx, applicationID)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	app.AddLabel(key, values)

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application")
	}

	return nil
}

func (s *service) DeleteLabel(ctx context.Context, applicationID string, key string, values []string) error {
	app, err := s.Get(ctx, applicationID)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	err = app.DeleteLabel(key, values)
	if err != nil {
		return errors.Wrapf(err, "while deleting label with key %s", key)
	}

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application")
	}

	return nil
}

func (s *service) AddAnnotation(ctx context.Context, applicationID string, key string, value interface{}) error {
	app, err := s.Get(ctx, applicationID)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	err = app.AddAnnotation(key, value)
	if err != nil {
		return errors.Wrapf(err, "while adding new annotation %s", key)
	}

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application")
	}

	return nil
}

func (s *service) DeleteAnnotation(ctx context.Context, applicationID string, key string) error {
	app, err := s.Get(ctx, applicationID)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	err = app.DeleteAnnotation(key)
	if err != nil {
		return errors.Wrapf(err, "while deleting annotation with key %s", key)
	}

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application with ID %s", applicationID)
	}

	return nil
}

func (s *service) createRelatedResources(in model.ApplicationInput, applicationID string) error {
	var err error

	var webhooks []*model.ApplicationWebhook
	for _, item := range in.Webhooks {
		webhooks = append(webhooks, item.ToWebhook(s.uidService.Generate(), applicationID))
	}
	err = s.webhook.CreateMany(webhooks)
	if err != nil {
		return errors.Wrapf(err, "while creating Webhooks for application")
	}

	var apis []*model.APIDefinition
	for _, item := range in.Apis {
		apis = append(apis, item.ToAPIDefinition(s.uidService.Generate(), applicationID))
	}

	err = s.api.CreateMany(apis)
	if err != nil {
		return errors.Wrapf(err, "while creating APIs for application")
	}

	var eventAPIs []*model.EventAPIDefinition
	for _, item := range in.EventAPIs {
		eventAPIs = append(eventAPIs, item.ToEventAPIDefinition(s.uidService.Generate(), applicationID))
	}
	err = s.eventAPI.CreateMany(eventAPIs)
	if err != nil {
		return errors.Wrapf(err, "while creating EventAPIs for application")
	}

	var documents []*model.Document
	for _, item := range in.Documents {
		documents = append(documents, item.ToDocument(s.uidService.Generate(), applicationID))
	}
	err = s.document.CreateMany(documents)
	if err != nil {
		return errors.Wrapf(err, "while creating Documents for application")
	}

	return nil
}

func (s *service) deleteRelatedResources(applicationID string) error {
	var err error

	err = s.webhook.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting Webhooks for application %s", applicationID)
	}

	err = s.api.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIs for application %s", applicationID)
	}

	err = s.eventAPI.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIs for application %s", applicationID)
	}

	err = s.document.DeleteAllByApplicationID(applicationID)
	if err != nil {
		return errors.Wrapf(err, "while deleting Documents for application %s", applicationID)
	}

	return nil
}
