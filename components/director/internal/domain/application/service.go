package application

import (
	"context"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ApplicationRepository -output=automock -outpkg=automock -case=underscore
type ApplicationRepository interface {
	GetByID(id string) (*model.Application, error)
	List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	Create(item *model.Application) error
	Update(item *model.Application) error
	Delete(item *model.Application) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type DocumentRepository interface {
	ListByApplicationID(applicationID string) ([]*model.Document, error)
	CreateMany(item []*model.Document) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=WebhookRepository -output=automock -outpkg=automock -case=underscore
type WebhookRepository interface {
	ListByApplicationID(applicationID string) ([]*model.ApplicationWebhook, error)
	CreateMany(item []*model.ApplicationWebhook) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=APIRepository -output=automock -outpkg=automock -case=underscore
type APIRepository interface {
	ListByApplicationID(applicationID string) ([]*model.APIDefinition, error)
	CreateMany(item []*model.APIDefinition) error
	DeleteAllByApplicationID(id string) error
}

//go:generate mockery -name=DocumentRepository -output=automock -outpkg=automock -case=underscore
type EventAPIRepository interface {
	ListByApplicationID(applicationID string) ([]*model.EventAPIDefinition, error)
	CreateMany(item *model.EventAPIDefinition) error
	DeleteAllByApplicationID(id string) error
}

type service struct {
	app ApplicationRepository
	document DocumentRepository
	webhook WebhookRepository
	api APIRepository
	eventAPI EventAPIRepository
}

func NewService(app ApplicationRepository, document DocumentRepository, webhook WebhookRepository, api APIRepository, eventAPI EventAPIRepository) *service {
	return &service{app: app, document: document, webhook: webhook, api: api, eventAPI: eventAPI}
}

func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	return s.app.List(filter, pageSize, cursor)
}

func (s *service) Get(ctx context.Context, id string) (*model.Application, error) {
	app, err := s.app.GetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	return app, nil
}

func (s *service) Create(ctx context.Context, in model.ApplicationInput) (string, error) {
	id := uuid.New().String()
	applicationTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	app := &model.Application{
		ID: id,
		Name:        in.Name,
		Description: in.Description,
		Tenant: applicationTenant,
		Labels:      in.Labels,
		Annotations: in.Annotations,
		HealthCheckURL: in.HealthCheckURL,
	}

	err = s.app.Create(app)
	if err != nil {
		return "", err
	}

	for _, api := range in.Apis {
		err = s.api.CreateMany(api)
		if err != nil {
			return "", errors.Wrapf(err, "while creating API")
		}
	}

	return id, nil
}

func (s *service) Update(ctx context.Context, id string, in model.ApplicationInput) error {
	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Application")
	}

	app.Name = in.Name
	app.Description = in.Description
	app.Labels = in.Labels
	app.Annotations = in.Annotations
	app.HealthCheckURL = in.HealthCheckURL

	err = s.app.Update(app)
	if err != nil {
		return errors.Wrapf(err, "while updating Application")
	}




	return nil
}

func (s *service) upsertApplication(id string, in model.ApplicationInput) error {

}

func (s *service) Delete(ctx context.Context, id string) error {
	app, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Application with ID %s", id)
	}

	err = s.api.DeleteAllByApplicationID(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting APIs for application %s", id)
	}

	err = s.eventAPI.DeleteAllByApplicationID(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting EventAPIs for application %s", id)
	}

	err = s.document.DeleteAllByApplicationID(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Documents for application %s", id)
	}

	err = s.webhook.DeleteAllByApplicationID(id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Webhooks for application %s", id)
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
