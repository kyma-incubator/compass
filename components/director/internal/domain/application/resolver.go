package application

import (
	"context"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/tokens"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	Get(ctx context.Context, id string) (*model.Application, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error)
	ListByRuntimeID(ctx context.Context, runtimeUUID uuid.UUID, pageSize int, cursor string) (*model.ApplicationPage, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, applicationID string, key string) error
}

//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	MultipleToGraphQL(in []*model.Application) []*graphql.Application
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
	UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput
	GraphQLToModel(obj *graphql.Application, tenantID string) *model.Application
}

//go:generate mockery --name=EventingService --output=automock --outpkg=automock --case=underscore
type EventingService interface {
	CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error)
	GetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error)
}

//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.Webhook, error)
	ListAllApplicationWebhooks(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	Create(ctx context.Context, resourceID string, in model.WebhookInput, converterFunc model.WebhookConverterFunc) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
}

//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore
type SystemAuthConverter interface {
	ToGraphQL(in *model.SystemAuth) (graphql.SystemAuth, error)
}

//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []model.SystemAuth) error
}

//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore
type RuntimeService interface {
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
}

//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore
type BundleService interface {
	GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error)
	ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error)
	CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleCreateInput) error
}

//go:generate mockery --name=BundleConverter --output=automock --outpkg=automock --case=underscore
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	MultipleToGraphQL(in []*model.Bundle) ([]*graphql.Bundle, error)
	MultipleCreateInputFromGraphQL(in []*graphql.BundleCreateInput) ([]*model.BundleCreateInput, error)
}

//go:generate mockery --name=TokenConverter --output=automock --outpkg=automock --case=underscore
type TokenConverter interface {
	ToGraphQLForApplication(model model.OneTimeToken) (graphql.OneTimeTokenForApplication, error)
}

//go:generate mockery --name=TimeService --output=automock --outpkg=automock --case=underscore
type TimeService interface {
	Now() time.Time
}

type Resolver struct {
	transact persistence.Transactioner

	appSvc       ApplicationService
	appConverter ApplicationConverter

	webhookSvc WebhookService
	oAuth20Svc OAuth20Service
	sysAuthSvc SystemAuthService
	bndlSvc    BundleService

	webhookConverter WebhookConverter
	sysAuthConv      SystemAuthConverter
	eventingSvc      EventingService
	bndlConv         BundleConverter
	oneTimeTokenConv TokenConverter

	oneTimeTokenCfg onetimetoken.Config

	timeService TimeService
}

func NewResolver(transact persistence.Transactioner,
	svc ApplicationService,
	webhookSvc WebhookService,
	oAuth20Svc OAuth20Service,
	sysAuthSvc SystemAuthService,
	appConverter ApplicationConverter,
	webhookConverter WebhookConverter,
	sysAuthConv SystemAuthConverter,
	eventingSvc EventingService,
	bndlSvc BundleService,
	bndlConverter BundleConverter,
	oneTimeTokenConv TokenConverter,
	oneTimeTokenCfg onetimetoken.Config,
	timeService TimeService) *Resolver {
	return &Resolver{
		transact:         transact,
		appSvc:           svc,
		webhookSvc:       webhookSvc,
		oAuth20Svc:       oAuth20Svc,
		sysAuthSvc:       sysAuthSvc,
		appConverter:     appConverter,
		webhookConverter: webhookConverter,
		sysAuthConv:      sysAuthConv,
		eventingSvc:      eventingSvc,
		bndlSvc:          bndlSvc,
		bndlConv:         bndlConverter,
		oneTimeTokenConv: oneTimeTokenConv,
		oneTimeTokenCfg:  oneTimeTokenCfg,
		timeService:      timeService,
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appPage, err := r.appSvc.List(ctx, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApps := r.appConverter.MultipleToGraphQL(appPage.Data)

	return &graphql.ApplicationPage{
		Data:       gqlApps,
		TotalCount: appPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appPage.PageInfo.EndCursor),
			HasNextPage: appPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.appConverter.ToGraphQL(app), nil
}

func (r *Resolver) ApplicationsForRuntime(ctx context.Context, runtimeID string, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	runtimeUUID, err := uuid.Parse(runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while converting runtimeID to UUID")
	}

	appPage, err := r.appSvc.ListByRuntimeID(ctx, runtimeUUID, *first, cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while getting all Application for Runtime")
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApps := r.appConverter.MultipleToGraphQL(appPage.Data)

	return &graphql.ApplicationPage{
		Data:       gqlApps,
		TotalCount: appPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appPage.PageInfo.EndCursor),
			HasNextPage: appPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) RegisterApplication(ctx context.Context, in graphql.ApplicationRegisterInput) (*graphql.Application, error) {
	log.C(ctx).Infof("Registering Application with name %s", in.Name)

	convertedIn, err := r.appConverter.CreateInputFromGraphQL(ctx, in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting ApplicationRegister input")
	}

	if convertedIn.Labels == nil {
		convertedIn.Labels = make(map[string]interface{})
	}
	convertedIn.Labels["managed"] = "false"

	id, err := r.appSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	log.C(ctx).Infof("Application with name %s and id %s successfully registered", in.Name, id)
	return gqlApp, nil
}

func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	log.C(ctx).Infof("Updating Application with id %s", id)

	convertedIn := r.appConverter.UpdateInputFromGraphQL(in)
	err := r.appSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	log.C(ctx).Infof("Application with id %s successfully updated", id)

	return gqlApp, nil
}

func (r *Resolver) UnregisterApplication(ctx context.Context, id string) (*graphql.Application, error) {
	log.C(ctx).Infof("Unregistering Application with id %s", id)

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	appID, err := uuid.Parse(app.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing application ID as UUID")
	}

	if _, err = r.eventingSvc.CleanupAfterUnregisteringApplication(ctx, appID); err != nil {
		return nil, err
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, model.ApplicationReference, app.ID)
	if err != nil {
		return nil, err
	}

	err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths)
	if err != nil {
		return nil, err
	}
	err = r.appSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedApp := r.appConverter.ToGraphQL(app)

	log.C(ctx).Infof("Successfully unregistered Application with id %s", id)
	return deletedApp, nil
}
func (r *Resolver) SetApplicationLabel(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Label, error) {
	// TODO: Use @validation directive on input type instead, after resolving https://github.com/kyma-incubator/compass/issues/515
	gqlLabel := graphql.LabelInput{Key: key, Value: value}
	if err := inputvalidation.Validate(&gqlLabel); err != nil {
		return nil, errors.Wrap(err, "validation error for type LabelInput")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.appSvc.SetLabel(ctx, &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectType: model.ApplicationLabelableObject,
		ObjectID:   applicationID,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: value,
	}, nil
}

func (r *Resolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.appSvc.GetLabel(ctx, applicationID, key)
	if err != nil {
		return nil, err
	}

	err = r.appSvc.DeleteLabel(ctx, applicationID, key)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: label.Value,
	}, nil
}

// TODO: Proper error handling
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := r.webhookSvc.ListAllApplicationWebhooks(ctx, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	gqlWebhooks, err := r.webhookConverter.MultipleToGraphQL(webhooks)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return gqlWebhooks, nil
}

func (r *Resolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.appSvc.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	var gqlLabels graphql.Labels = resultLabels
	return gqlLabels, nil
}

func (r *Resolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.AppSystemAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, model.ApplicationReference, obj.ID)
	if err != nil {
		return nil, err
	}

	var out []*graphql.AppSystemAuth
	for _, sa := range sysAuths {
		if shouldSkip := r.checkApplicationOneTimeTokenIsInvalid(sa.Value, r.oneTimeTokenCfg); shouldSkip {
			continue
		}

		c, err := r.sysAuthConv.ToGraphQL(&sa)
		if err != nil {
			return nil, err
		}

		if sa.Value.OneTimeToken != nil && sa.Value.OneTimeToken.Type == tokens.ApplicationToken {
			oneTimeTokenForApplication, err := r.oneTimeTokenConv.ToGraphQLForApplication(*sa.Value.OneTimeToken)
			if err != nil {
				return nil, errors.Wrap(err, "while converting one-time token to graphql")
			}

			c.(*graphql.AppSystemAuth).Auth.OneTimeToken = &oneTimeTokenForApplication
		}
		out = append(out, c.(*graphql.AppSystemAuth))
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Resolver) checkApplicationOneTimeTokenIsInvalid(auth *model.Auth, config onetimetoken.Config) bool {
	if auth == nil || auth.OneTimeToken == nil {
		return false
	}

	if auth.OneTimeToken.Type != tokens.ApplicationToken || auth.OneTimeToken.Used {
		return true
	}

	isExpired := auth.OneTimeToken.CreatedAt.Add(config.ApplicationExpiration).Before(r.timeService.Now())
	if isExpired {
		return true
	}

	return false
}

func (r *Resolver) EventingConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventingConfiguration, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Application cannot be empty")
	}
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, apperrors.NewCannotReadTenantError()
	}

	app := r.appConverter.GraphQLToModel(obj, tenantID)
	if app == nil {
		return nil, apperrors.NewInternalError("application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventingCfg, err := r.eventingSvc.GetForApplication(ctx, *app)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return eventing.ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}

func (r *Resolver) Bundles(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.BundlePage, error) {
	param := dataloader.ParamBundle{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.BundleFor(ctx).BundleById.Load(param)
}

func (r *Resolver) BundlesDataLoader(keys []dataloader.ParamBundle) ([]*graphql.BundlePage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Applications found")}
	}

	ctx := keys[0].Ctx
	first := keys[0].First
	after := keys[0].After

	applicationIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		applicationIDs = append(applicationIDs, key.ID)
	}

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndlPages, err := r.bndlSvc.ListByApplicationIDs(ctx, applicationIDs, *first, cursor)
	if err != nil {
		return nil, []error{err}
	}

	gqlBndls := make([]*graphql.BundlePage, 0, len(bndlPages))
	for _, page := range bndlPages {
		bndls, err := r.bndlConv.MultipleToGraphQL(page.Data)
		if err != nil {
			return nil, []error{err}
		}

		gqlBndls = append(gqlBndls, &graphql.BundlePage{Data: bndls, TotalCount: page.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		}})
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	return gqlBndls, nil
}

func (r *Resolver) Bundle(ctx context.Context, obj *graphql.Application, id string) (*graphql.Bundle, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndl, err := r.bndlSvc.GetForApplication(ctx, id, obj.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	gqlBundle, err := r.bndlConv.ToGraphQL(bndl)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return gqlBundle, nil
}
