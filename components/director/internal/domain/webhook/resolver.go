package webhook

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.Webhook, error)
	ListAllApplicationWebhooks(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	Create(ctx context.Context, resourceID string, in model.WebhookInput, converterFunc model.WebhookConverterFunc) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore
type ApplicationTemplateService interface {
	Exists(ctx context.Context, id string) (bool, error)
}

//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

type webhookOwner struct {
	resource.Type
	id string
}

type existsFunc func(ctx context.Context, id string) (bool, error)

type Resolver struct {
	webhookSvc       WebhookService
	appSvc           ApplicationService
	appTemplateSvc   ApplicationTemplateService
	webhookConverter WebhookConverter
	transact         persistence.Transactioner
}

func NewResolver(transact persistence.Transactioner, webhookSvc WebhookService, applicationService ApplicationService, appTemplateService ApplicationTemplateService, webhookConverter WebhookConverter) *Resolver {
	return &Resolver{
		webhookSvc:       webhookSvc,
		appSvc:           applicationService,
		appTemplateSvc:   appTemplateService,
		webhookConverter: webhookConverter,
		transact:         transact,
	}
}

func (r *Resolver) AddWebhook(ctx context.Context, applicationID *string, applicationTemplateID *string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	appSpecified := applicationID != nil && applicationTemplateID == nil
	appTemplateSpecified := applicationID == nil && applicationTemplateID != nil

	if !(appSpecified || appTemplateSpecified) {
		return nil, apperrors.NewInvalidDataError("exactly one of applicationId and applicationTemplateID should be specified")
	}

	convertedIn, err := r.webhookConverter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the WebhookInput")
	}

	var owner webhookOwner
	if appSpecified {
		owner = webhookOwner{Type: resource.Application, id: *applicationID}
	} else if appTemplateSpecified {
		owner = webhookOwner{Type: resource.ApplicationTemplate, id: *applicationTemplateID}
	}

	id, err := r.checkForExistenceAndCreate(ctx, owner, *convertedIn)
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

func (r *Resolver) UpdateWebhook(ctx context.Context, webhookID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
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

func (r *Resolver) DeleteWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
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

func (r *Resolver) checkForExistenceAndCreate(ctx context.Context, owningResource webhookOwner, input model.WebhookInput) (string, error) {
	var (
		converterFunc model.WebhookConverterFunc
		existsFunc    existsFunc
	)

	switch owningResource.Type {
	case resource.Application:
		converterFunc = (*model.WebhookInput).ToApplicationWebhook
		existsFunc = r.appSvc.Exist
	case resource.ApplicationTemplate:
		converterFunc = (*model.WebhookInput).ToApplicationTemplateWebhook
		existsFunc = r.appTemplateSvc.Exists
	}
	err := r.genericCheckExistence(ctx, owningResource.id, string(owningResource.Type), existsFunc)
	if err != nil {
		return "", err
	}
	return r.webhookSvc.Create(ctx, owningResource.id, input, converterFunc)
}

func (r *Resolver) genericCheckExistence(ctx context.Context, resourceID, resourceName string, existsFunc existsFunc) error {
	found, err := existsFunc(ctx, resourceID)
	if err != nil {
		return errors.Wrapf(err, "while checking existence of %s", resourceName)
	}

	if !found {
		return apperrors.NewInvalidDataError(fmt.Sprintf("cannot add Webhook to not existing %s", resourceName))
	}
	return nil
}
