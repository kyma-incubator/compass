package webhook

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore
type WebhookRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Webhook, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Webhook, error)
	ListByApplicationID(ctx context.Context, tenant, applicationID string) ([]*model.Webhook, error)
	ListByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	Create(ctx context.Context, item *model.Webhook) error
	Update(ctx context.Context, item *model.Webhook) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery --name=ApplicationRepository --output=automock --outpkg=automock --case=underscore
type ApplicationRepository interface {
	GetGlobalByID(ctx context.Context, id string) (*model.Application, error)
}

//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type OwningResource string

type service struct {
	webhookRepo WebhookRepository
	appRepo     ApplicationRepository
	uidSvc      UIDService
}

func NewService(repo WebhookRepository, appRepo ApplicationRepository, uidSvc UIDService) *service {
	return &service{
		webhookRepo: repo,
		uidSvc:      uidSvc,
		appRepo:     appRepo,
	}
}

func (s *service) Get(ctx context.Context, id string) (webhook *model.Webhook, err error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil || tnt == "" {
		log.C(ctx).Infof("tenant was not loaded while getting Webhook id %s", id)
		webhook, err = s.webhookRepo.GetByIDGlobal(ctx, id)
	} else {
		webhook, err = s.webhookRepo.GetByID(ctx, tnt, id)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Webhook with ID %s", id)
	}
	return
}

func (s *service) ListForApplication(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.webhookRepo.ListByApplicationID(ctx, tnt, applicationID)
}

func (s *service) ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
	return s.webhookRepo.ListByApplicationTemplateID(ctx, applicationTemplateID)
}

func (s *service) ListAllApplicationWebhooks(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	application, err := s.appRepo.GetGlobalByID(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Applicaiton with ID %s", applicationID)
	}

	return s.retrieveWebhooks(ctx, application)
}

func (s *service) Create(ctx context.Context, owningResourceID string, in model.WebhookInput, converterFunc model.WebhookConverterFunc) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if apperrors.IsTenantRequired(err) {
		log.C(ctx).Debugf("Creating Webhook with type %s without tenant", in.Type)
	} else if err != nil {
		return "", err
	}
	id := s.uidSvc.Generate()
	webhook := converterFunc(&in, id, &tnt, owningResourceID)

	if err = s.webhookRepo.Create(ctx, webhook); err != nil {
		return "", errors.Wrapf(err, "while creating Webhook with type %s and id %s for Application with id %s", id, webhook.Type, owningResourceID)
	}
	log.C(ctx).Infof("Successfully created Webhook with type %s and id %s for Application with id %s", id, webhook.Type, owningResourceID)

	return webhook.ID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.WebhookInput) error {
	webhook, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	if webhook.ApplicationID != nil {
		webhook = in.ToApplicationWebhook(id, webhook.TenantID, *webhook.ApplicationID)
	} else if webhook.ApplicationTemplateID != nil {
		webhook = in.ToApplicationTemplateWebhook(id, webhook.TenantID, *webhook.ApplicationTemplateID)
	} else {
		return errors.New("while updating Webhook: webhook doesn't have neither of application_id and application_template_id")
	}

	err = s.webhookRepo.Update(ctx, webhook)
	if err != nil {
		return errors.Wrapf(err, "while updating Webhook")
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	webhook, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	return s.webhookRepo.Delete(ctx, webhook.ID)
}

func (s *service) retrieveWebhooks(ctx context.Context, application *model.Application) ([]*model.Webhook, error) {
	appWebhooks, err := s.ListForApplication(ctx, application.ID)
	if err != nil {
		return nil, err
	}

	var appTemplateWebhooks []*model.Webhook
	if application.ApplicationTemplateID != nil {
		appTemplateWebhooks, err = s.ListForApplicationTemplate(ctx, *application.ApplicationTemplateID)
		if err != nil {
			return nil, err
		}
	}

	webhooksMap := make(map[model.WebhookType]*model.Webhook)
	for i, webhook := range appTemplateWebhooks {
		webhooksMap[webhook.Type] = appTemplateWebhooks[i]
	}
	//Override values derived from template
	for i, webhook := range appWebhooks {
		webhooksMap[webhook.Type] = appWebhooks[i]
	}

	webhooks := make([]*model.Webhook, 0)
	for key := range webhooksMap {
		webhooks = append(webhooks, webhooksMap[key])
	}

	return webhooks, nil
}
