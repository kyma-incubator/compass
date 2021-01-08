package webhook

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=WebhookService -output=automock -outpkg=automock -case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.Webhook, error)
	List(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	Create(ctx context.Context, applicationID string, in model.WebhookInput) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery -name=WebhookConverter -output=automock -outpkg=automock -case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

type Resolver struct {
	webhookSvc       WebhookService
	appSvc           ApplicationService
	webhookConverter WebhookConverter
	transact         persistence.Transactioner
}

func NewResolver(transact persistence.Transactioner, webhookSvc WebhookService, applicationService ApplicationService, webhookConverter WebhookConverter) *Resolver {
	return &Resolver{
		webhookSvc:       webhookSvc,
		appSvc:           applicationService,
		webhookConverter: webhookConverter,
		transact:         transact,
	}
}

func (r *Resolver) AddApplicationWebhook(ctx context.Context, applicationID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.webhookConverter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the WebhookInput")
	}

	found, err := r.appSvc.Exist(ctx, applicationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking existence of Application")
	}

	if !found {
		return nil, apperrors.NewInvalidDataError("cannot add Webhook to not existing Application")
	}

	id, err := r.webhookSvc.Create(ctx, applicationID, *convertedIn)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.webhookConverter.ToGraphQL(webhook)
}

func (r *Resolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn, err := r.webhookConverter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the WebhookInput")
	}

	err = r.webhookSvc.Update(ctx, webhookID, *convertedIn)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.webhookConverter.ToGraphQL(webhook)
}

func (r *Resolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	webhook, err := r.webhookSvc.Get(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	deletedWebhook, err := r.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the Webhook model to GraphQL")
	}

	err = r.webhookSvc.Delete(ctx, webhookID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return deletedWebhook, nil
}
