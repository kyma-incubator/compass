package webhook

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=WebhookService -output=automock -outpkg=automock -case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.ApplicationWebhook, error)
	List(ctx context.Context, applicationID string) ([]*model.ApplicationWebhook, error)
	Create(ctx context.Context, applicationID string, in model.ApplicationWebhookInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationWebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=WebhookConverter -output=automock -outpkg=automock -case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.ApplicationWebhook) *graphql.ApplicationWebhook
	MultipleToGraphQL(in []*model.ApplicationWebhook) []*graphql.ApplicationWebhook
	InputFromGraphQL(in *graphql.ApplicationWebhookInput) *model.ApplicationWebhookInput
	MultipleInputFromGraphQL(in []*graphql.ApplicationWebhookInput) []*model.ApplicationWebhookInput
}

type Resolver struct {
	webhookSvc       WebhookService
	appSvc           ApplicationService
	webhookConverter WebhookConverter
}

func NewResolver(webhookSvc WebhookService, applicationService ApplicationService, webhookConverter WebhookConverter) *Resolver {
	return &Resolver{
		webhookSvc:       webhookSvc,
		appSvc:           applicationService,
		webhookConverter: webhookConverter,
	}
}

func (r *Resolver) AddApplicationWebhook(ctx context.Context, applicationID string, in graphql.ApplicationWebhookInput) (*graphql.ApplicationWebhook, error) {
	convertedIn := r.webhookConverter.InputFromGraphQL(&in)

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if found == false {
		return nil, errors.New("Cannot add document to not existing application")
	}

	id, err := r.webhookSvc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlWebhook := r.webhookConverter.ToGraphQL(webhook)

	return gqlWebhook, nil
}

func (r *Resolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in graphql.ApplicationWebhookInput) (*graphql.ApplicationWebhook, error) {
	convertedIn := r.webhookConverter.InputFromGraphQL(&in)

	err := r.webhookSvc.Update(ctx, webhookID, *convertedIn)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	gqlWebhook := r.webhookConverter.ToGraphQL(webhook)

	return gqlWebhook, nil
}

func (r *Resolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*graphql.ApplicationWebhook, error) {
	webhook, err := r.webhookSvc.Get(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	deletedWebhook := r.webhookConverter.ToGraphQL(webhook)

	err = r.webhookSvc.Delete(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	return deletedWebhook, nil
}
