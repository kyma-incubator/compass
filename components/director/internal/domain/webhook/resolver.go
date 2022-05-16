package webhook

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

// WebhookService missing godoc
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) (*model.Webhook, error)
	ListAllApplicationWebhooks(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	ListForRuntime(ctx context.Context, runtimeID string) ([]*model.Webhook, error)
	Create(ctx context.Context, resourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) error
	Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error
}

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

// ApplicationTemplateService missing godoc
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore
type ApplicationTemplateService interface {
	Exists(ctx context.Context, id string) (bool, error)
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	Exist(ctx context.Context, id string) (bool, error)
}

// WebhookConverter missing godoc
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

type existsFunc func(ctx context.Context, id string) (bool, error)

// Resolver missing godoc
type Resolver struct {
	webhookSvc       WebhookService
	appSvc           ApplicationService
	appTemplateSvc   ApplicationTemplateService
	runtimeSvc       RuntimeService
	webhookConverter WebhookConverter
	transact         persistence.Transactioner
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, webhookSvc WebhookService, applicationService ApplicationService, appTemplateService ApplicationTemplateService, runtimeService RuntimeService, webhookConverter WebhookConverter) *Resolver {
	return &Resolver{
		webhookSvc:       webhookSvc,
		appSvc:           applicationService,
		appTemplateSvc:   appTemplateService,
		runtimeSvc:       runtimeService,
		webhookConverter: webhookConverter,
		transact:         transact,
	}
}

// AddWebhook missing godoc
func (r *Resolver) AddWebhook(ctx context.Context, applicationID *string, applicationTemplateID *string, runtimeID *string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	appSpecified := applicationID != nil && applicationTemplateID == nil && runtimeID == nil
	appTemplateSpecified := applicationID == nil && applicationTemplateID != nil && runtimeID == nil
	runtimeSpecified := applicationID == nil && applicationTemplateID == nil && runtimeID != nil

	if !(appSpecified || appTemplateSpecified || runtimeSpecified) {
		return nil, apperrors.NewInvalidDataError("exactly one of applicationID, applicationTemplateID or runtimeID should be specified")
	}

	convertedIn, err := r.webhookConverter.InputFromGraphQL(&in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the WebhookInput")
	}

	var objectID string
	var objectType model.WebhookReferenceObjectType
	if appSpecified {
		objectID = *applicationID
		objectType = model.ApplicationWebhookReference
	} else if appTemplateSpecified {
		objectID = *applicationTemplateID
		objectType = model.ApplicationTemplateWebhookReference
	} else if runtimeSpecified {
		objectID = *runtimeID
		objectType = model.RuntimeWebhookReference
	}

	id, err := r.checkForExistenceAndCreate(ctx, *convertedIn, objectID, objectType)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, id, objectType)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.webhookConverter.ToGraphQL(webhook)
}

// UpdateWebhook missing godoc
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

	err = r.webhookSvc.Update(ctx, webhookID, *convertedIn, model.UnknownWebhookReference)
	if err != nil {
		return nil, err
	}

	webhook, err := r.webhookSvc.Get(ctx, webhookID, model.UnknownWebhookReference)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.webhookConverter.ToGraphQL(webhook)
}

// DeleteWebhook missing godoc
func (r *Resolver) DeleteWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	webhook, err := r.webhookSvc.Get(ctx, webhookID, model.UnknownWebhookReference)
	if err != nil {
		return nil, err
	}

	deletedWebhook, err := r.webhookConverter.ToGraphQL(webhook)
	if err != nil {
		return nil, errors.Wrap(err, "while converting the Webhook model to GraphQL")
	}

	err = r.webhookSvc.Delete(ctx, webhookID, model.UnknownWebhookReference)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return deletedWebhook, nil
}

func (r *Resolver) checkForExistenceAndCreate(ctx context.Context, input model.WebhookInput, objectID string, objectType model.WebhookReferenceObjectType) (string, error) {
	var (
		existsFunc existsFunc
	)

	switch objectType {
	case model.ApplicationWebhookReference:
		existsFunc = r.appSvc.Exist
	case model.ApplicationTemplateWebhookReference:
		existsFunc = r.appTemplateSvc.Exists
	case model.RuntimeWebhookReference:
		existsFunc = r.runtimeSvc.Exist
	}

	err := r.genericCheckExistence(ctx, objectID, objectType, existsFunc)
	if err != nil {
		return "", err
	}
	return r.webhookSvc.Create(ctx, objectID, input, objectType)
}

func (r *Resolver) genericCheckExistence(ctx context.Context, resourceID string, objectType model.WebhookReferenceObjectType, existsFunc existsFunc) error {
	found, err := existsFunc(ctx, resourceID)
	if err != nil {
		return errors.Wrapf(err, "while checking existence of %s", objectType)
	}

	if !found {
		return apperrors.NewInvalidDataError(fmt.Sprintf("cannot add %s due to not existing reference entity", objectType))
	}
	return nil
}
