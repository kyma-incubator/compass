package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

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

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error
	Get(ctx context.Context, id string) (*model.Application, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error)
	GetBySystemNumber(ctx context.Context, systemNumber string) (*model.Application, error)
	ListByRuntimeID(ctx context.Context, runtimeUUID uuid.UUID, pageSize int, cursor string) (*model.ApplicationPage, error)
	ListAll(ctx context.Context) ([]*model.Application, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, applicationID string, key string) error
	Unpair(ctx context.Context, id string) error
	Merge(ctx context.Context, destID, sourceID string) (*model.Application, error)
}

// ApplicationConverter missing godoc
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	MultipleToGraphQL(in []*model.Application) []*graphql.Application
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
	UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput
	GraphQLToModel(obj *graphql.Application, tenantID string) *model.Application
}

// EventingService missing godoc
//go:generate mockery --name=EventingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventingService interface {
	CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error)
	GetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error)
}

// WebhookService missing godoc
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListAllApplicationWebhooks(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
}

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string) ([]pkgmodel.SystemAuth, error)
	DeleteMultipleByIDForObject(ctx context.Context, systemAuths []pkgmodel.SystemAuth) error
}

// WebhookConverter missing godoc
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) (*graphql.Webhook, error)
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthConverter interface {
	ToGraphQL(in *pkgmodel.SystemAuth) (graphql.SystemAuth, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []pkgmodel.SystemAuth) error
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
}

// BundleService missing godoc
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error)
	ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error)
	CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleCreateInput) error
}

// BundleConverter missing godoc
//go:generate mockery --name=BundleConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleConverter interface {
	ToGraphQL(in *model.Bundle) (*graphql.Bundle, error)
	MultipleToGraphQL(in []*model.Bundle) ([]*graphql.Bundle, error)
	MultipleCreateInputFromGraphQL(in []*graphql.BundleCreateInput) ([]*model.BundleCreateInput, error)
}

// OneTimeTokenService missing godoc
//go:generate mockery --name=OneTimeTokenService --output=automock --outpkg=automock --case=underscore --disable-version-string
type OneTimeTokenService interface {
	IsTokenValid(systemAuth *pkgmodel.SystemAuth) (bool, error)
}

// ApplicationTemplateService missing godoc
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
}

// Resolver missing godoc
type Resolver struct {
	transact persistence.Transactioner

	appSvc       ApplicationService
	appConverter ApplicationConverter

	appTemplateSvc ApplicationTemplateService

	webhookSvc WebhookService
	oAuth20Svc OAuth20Service
	sysAuthSvc SystemAuthService
	bndlSvc    BundleService

	webhookConverter WebhookConverter
	sysAuthConv      SystemAuthConverter
	eventingSvc      EventingService
	bndlConv         BundleConverter

	selfRegisterDistinguishLabelKey string
	tokenPrefix                     string
}

// NewResolver missing godoc
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
	appTemplateSvc ApplicationTemplateService,
	selfRegisterDistinguishLabelKey, tokenPrefix string) *Resolver {
	return &Resolver{
		transact:                        transact,
		appSvc:                          svc,
		webhookSvc:                      webhookSvc,
		oAuth20Svc:                      oAuth20Svc,
		sysAuthSvc:                      sysAuthSvc,
		appConverter:                    appConverter,
		webhookConverter:                webhookConverter,
		sysAuthConv:                     sysAuthConv,
		eventingSvc:                     eventingSvc,
		bndlSvc:                         bndlSvc,
		bndlConv:                        bndlConverter,
		appTemplateSvc:                  appTemplateSvc,
		selfRegisterDistinguishLabelKey: selfRegisterDistinguishLabelKey,
		tokenPrefix:                     tokenPrefix,
	}
}

// Applications retrieves all tenant scoped applications.
// If this method is executed in a double authentication flow (i.e. consumerInfo.OnBehalfOf != nil)
// then it would return 0 or 1 tenant applications - it would return 1 if there exists a tenant application
// representing a tenant in an Application Provider and 0 if there is none such application.
func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if consumerInfo.OnBehalfOf != "" {
		log.C(ctx).Infof("External tenant with id %s is retrieving application on behalf of tenant with id %s", consumerInfo.ConsumerID, consumerInfo.OnBehalfOf)
		tenantApp, err := r.getApplicationProviderTenant(ctx, consumerInfo)
		if err != nil {
			return nil, err
		}

		err = tx.Commit()
		if err != nil {
			return nil, err
		}

		return &graphql.ApplicationPage{
			Data:       []*graphql.Application{tenantApp},
			TotalCount: 1,
			PageInfo: &graphql.PageInfo{
				StartCursor: "1",
				EndCursor:   "1",
				HasNextPage: false,
			},
		}, nil
	}

	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

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

func (r *Resolver) getApplication(ctx context.Context, get func(context.Context) (*model.Application, error)) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	app, err := get(ctx)

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

// ApplicationBySystemNumber missing godoc
func (r *Resolver) ApplicationBySystemNumber(ctx context.Context, systemNumber string) (*graphql.Application, error) {
	return r.getApplication(ctx, func(ctx context.Context) (*model.Application, error) {
		return r.appSvc.GetBySystemNumber(ctx, systemNumber)
	})
}

// Application missing godoc
func (r *Resolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	return r.getApplication(ctx, func(ctx context.Context) (*model.Application, error) {
		return r.appSvc.Get(ctx, id)
	})
}

// ApplicationsForRuntime missing godoc
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

// RegisterApplication missing godoc
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

// UpdateApplication missing godoc
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

// UnregisterApplication missing godoc
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

	auths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.ApplicationReference, app.ID)
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

// UnpairApplication Sets the UpdatedAt property for the given application, deletes associated []model.SystemAuth, deletes the hydra oauth clients.
func (r *Resolver) UnpairApplication(ctx context.Context, id string) (*graphql.Application, error) {
	log.C(ctx).Infof("Unpairing Application with id %s", id)

	if err := r.appSvc.Unpair(ctx, id); err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.ApplicationReference, app.ID)
	if err != nil {
		return nil, err
	}

	if err = r.sysAuthSvc.DeleteMultipleByIDForObject(ctx, auths); err != nil {
		return nil, err
	}

	if err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths); err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	log.C(ctx).Infof("Successfully Unpaired Application with id %s", id)
	return gqlApp, nil
}

// SetApplicationLabel missing godoc
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

// MergeApplications merges properties from Source Application into Destination Application, provided that the Destination's
// Application does not have a value set for a given property. Then the Source Application is being deleted.
func (r *Resolver) MergeApplications(ctx context.Context, destID string, sourceID string) (*graphql.Application, error) {
	log.C(ctx).Infof("Merging source app with id %s into destination app with id %s", sourceID, destID)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	mergedApp, err := r.appSvc.Merge(ctx, destID, sourceID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(mergedApp)

	return gqlApp, nil
}

// DeleteApplicationLabel missing godoc
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

// Webhooks missing godoc
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

// Labels missing godoc
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

// Auths missing godoc
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

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.ApplicationReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out := make([]*graphql.AppSystemAuth, 0, len(sysAuths))
	for _, sa := range sysAuths {
		c, err := r.sysAuthConv.ToGraphQL(&sa)
		if err != nil {
			return nil, err
		}

		out = append(out, c.(*graphql.AppSystemAuth))
	}

	return out, nil
}

// EventingConfiguration missing godoc
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
		return nil, errors.Wrap(err, "while committing the transaction")
	}

	return eventing.ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}

// Bundles missing godoc
func (r *Resolver) Bundles(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.BundlePage, error) {
	param := dataloader.ParamBundle{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.BundleFor(ctx).BundleByID.Load(param)
}

// BundlesDataLoader missing godoc
func (r *Resolver) BundlesDataLoader(keys []dataloader.ParamBundle) ([]*graphql.BundlePage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Applications found")}
	}

	ctx := keys[0].Ctx
	applicationIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		applicationIDs = append(applicationIDs, key.ID)
	}

	var cursor string
	if keys[0].After != nil {
		cursor = string(*keys[0].After)
	}

	if keys[0].First == nil {
		return nil, []error{apperrors.NewInvalidDataError("missing required parameter 'first'")}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, []error{err}
	}

	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	bndlPages, err := r.bndlSvc.ListByApplicationIDs(ctx, applicationIDs, *keys[0].First, cursor)
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

// Bundle missing godoc
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

// getApplicationProviderTenant should be used when making requests with double authentication, i.e. consumerInfo.OnBehalfOf != nil;
// The function leverages the knowledge of the provider tenant (it's in consumerInfo.ConsumerID) and consumer tenant (it's already set in the TenantCtx)
// in order to derive the application template representing an Application Provider and then finding an application among those of the consumer tenant
// which is associated with that application template.
// In this way the getApplicationProviderTenant function finds a specific tenant application of a given Application Provider - there should only be one or none.
func (r *Resolver) getApplicationProviderTenant(ctx context.Context, consumerInfo consumer.Consumer) (*graphql.Application, error) {
	tokenClientID := strings.TrimPrefix(consumerInfo.TokenClientID, r.tokenPrefix)
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(scenarioassignment.SubaccountIDKey, fmt.Sprintf("\"%s\"", consumerInfo.ConsumerID)),
		labelfilter.NewForKeyWithQuery(tenant.RegionLabelKey, fmt.Sprintf("\"%s\"", consumerInfo.Region)),
		labelfilter.NewForKeyWithQuery(r.selfRegisterDistinguishLabelKey, fmt.Sprintf("\"%s\"", tokenClientID)),
	}

	// Derive application provider's app template
	appTemplate, err := r.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil {
		log.C(ctx).Infof("No app template found with filter %q = %q, %q = %q, %q = %q", scenarioassignment.SubaccountIDKey, consumerInfo.ConsumerID, tenant.RegionLabelKey, consumerInfo.Region, r.selfRegisterDistinguishLabelKey, tokenClientID)
		return nil, errors.Wrapf(err, "no app template found with filter %q = %q, %q = %q, %q = %q", scenarioassignment.SubaccountIDKey, consumerInfo.ConsumerID, tenant.RegionLabelKey, consumerInfo.Region, r.selfRegisterDistinguishLabelKey, tokenClientID)
	}

	// Find the consuming tenant's applications
	tntApplications, err := r.appSvc.ListAll(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while listing applications for tenant")
	}

	// Try to find the application matching the derived provider app template
	var foundApp *model.Application
	for _, app := range tntApplications {
		if str.PtrStrToStr(app.ApplicationTemplateID) == appTemplate.ID {
			foundApp = app
			break
		}
	}

	if foundApp == nil {
		log.C(ctx).Infof("No application found for template with ID %q", appTemplate.ID)
		return nil, errors.Errorf("No application found for template with ID %q", appTemplate.ID)
	}

	return r.appConverter.ToGraphQL(foundApp), nil
}
