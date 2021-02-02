package webhook

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=WebhookRepository -output=automock -outpkg=automock -case=underscore
type WebhookRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.Webhook, error)
	ListByApplicationID(ctx context.Context, tenant, applicationID string) ([]*model.Webhook, error)
	Create(ctx context.Context, item *model.Webhook) error
	Update(ctx context.Context, item *model.Webhook) error
	Delete(ctx context.Context, tenant, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	repo   WebhookRepository
	uidSvc UIDService
}

func NewService(repo WebhookRepository, uidSvc UIDService) *service {
	return &service{
		repo:   repo,
		uidSvc: uidSvc,
	}
}

func (s *service) Get(ctx context.Context, id string) (*model.Webhook, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	webhook, err := s.repo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Webhook with ID %s", id)
	}

	return webhook, nil
}

func (s *service) List(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.repo.ListByApplicationID(ctx, tnt, applicationID)
}

func (s *service) Create(ctx context.Context, applicationID string, in model.WebhookInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}
	id := s.uidSvc.Generate()
	webhook := in.ToWebhook(id, tnt, applicationID)

	if err = s.repo.Create(ctx, webhook); err != nil {
		return "", errors.Wrapf(err, "while creating Webhook with type %s and id %s for Application with id %s", id, webhook.Type, applicationID)
	}
	log.C(ctx).Infof("Successfully created Webhook with type %s and id %s for Application with id %s", id, webhook.Type, applicationID)

	return webhook.ID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.WebhookInput) error {
	webhook, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	webhook = in.ToWebhook(id, webhook.TenantID, *webhook.ApplicationID)

	err = s.repo.Update(ctx, webhook)
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

	return s.repo.Delete(ctx, webhook.TenantID, webhook.ID)
}
