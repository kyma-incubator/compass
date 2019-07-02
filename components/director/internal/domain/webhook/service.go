package webhook

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery -name=WebhookRepository -output=automock -outpkg=automock -case=underscore
type WebhookRepository interface {
	GetByID(id string) (*model.ApplicationWebhook, error)
	ListByApplicationID(applicationID string) ([]*model.ApplicationWebhook, error)
	Create(item *model.ApplicationWebhook) error
	Update(item *model.ApplicationWebhook) error
	Delete(item *model.ApplicationWebhook) error
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

func (s *service) Get(ctx context.Context, id string) (*model.ApplicationWebhook, error) {
	webhook, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Webhook with ID %s", id)
	}

	return webhook, nil
}

func (s *service) List(ctx context.Context, applicationID string) ([]*model.ApplicationWebhook, error) {
	return s.repo.ListByApplicationID(applicationID)
}

func (s *service) Create(ctx context.Context, applicationID string, in model.ApplicationWebhookInput) (string, error) {
	id := s.uidSvc.Generate()
	webhook := in.ToWebhook(id, applicationID)

	err := s.repo.Create(webhook)
	if err != nil {
		return "", errors.Wrap(err, "while creating Webhook")
	}

	return webhook.ID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.ApplicationWebhookInput) error {
	webhook, err := s.Get(ctx, id)
	if err != nil {
		return errors.Wrap(err, "while getting Webhook")
	}

	webhook = in.ToWebhook(id, webhook.ApplicationID)

	err = s.repo.Update(webhook)
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

	return s.repo.Delete(webhook)
}
