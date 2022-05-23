package runtime

import (
	"context"
	"strings"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	labelPkg "github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/timestamp"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// EventingService missing godoc
//go:generate mockery --name=EventingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventingService interface {
	GetForRuntime(ctx context.Context, runtimeID uuid.UUID) (*model.RuntimeEventingConfiguration, error)
}

// OAuth20Service missing godoc
//go:generate mockery --name=OAuth20Service --output=automock --outpkg=automock --case=underscore --disable-version-string
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []pkgmodel.SystemAuth) error
}

// RuntimeService missing godoc
//go:generate mockery --name=RuntimeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeService interface {
	CreateWithMandatoryLabels(ctx context.Context, in model.RuntimeRegisterInput, id string, mandatoryLabels map[string]interface{}) error
	Update(ctx context.Context, id string, in model.RuntimeUpdateInput) error
	Get(ctx context.Context, id string) (*model.Runtime, error)
	GetByTokenIssuer(ctx context.Context, issuer string) (*model.Runtime, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, runtimeID string, key string) error
}

// ScenarioAssignmentService missing godoc
//go:generate mockery --name=ScenarioAssignmentService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ScenarioAssignmentService interface {
	GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error)
	Delete(ctx context.Context, in model.AutomaticScenarioAssignment) error
}

// RuntimeConverter missing godoc
//go:generate mockery --name=RuntimeConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeConverter interface {
	ToGraphQL(in *model.Runtime) *graphql.Runtime
	MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime
	RegisterInputFromGraphQL(in graphql.RuntimeRegisterInput) (model.RuntimeRegisterInput, error)
	UpdateInputFromGraphQL(in graphql.RuntimeUpdateInput) model.RuntimeUpdateInput
}

// SystemAuthConverter missing godoc
//go:generate mockery --name=SystemAuthConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthConverter interface {
	ToGraphQL(in *pkgmodel.SystemAuth) (graphql.SystemAuth, error)
}

// SystemAuthService missing godoc
//go:generate mockery --name=SystemAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType pkgmodel.SystemAuthReferenceObjectType, objectID string) ([]pkgmodel.SystemAuth, error)
}

// BundleInstanceAuthService missing godoc
//go:generate mockery --name=BundleInstanceAuthService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleInstanceAuthService interface {
	ListByRuntimeID(ctx context.Context, runtimeID string) ([]*model.BundleInstanceAuth, error)
	Update(ctx context.Context, instanceAuth *model.BundleInstanceAuth) error
}

// SelfRegisterManager missing godoc
//go:generate mockery --name=SelfRegisterManager --output=automock --outpkg=automock --case=underscore --disable-version-string
type SelfRegisterManager interface {
	PrepareRuntimeForSelfRegistration(ctx context.Context, in model.RuntimeRegisterInput, id string) (map[string]interface{}, error)
	CleanupSelfRegisteredRuntime(ctx context.Context, selfRegisterLabelValue, region string) error
	GetSelfRegDistinguishingLabelKey() string
}

// SubscriptionService missing godoc
//go:generate mockery --name=SubscriptionService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SubscriptionService interface {
	SubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error)
	UnsubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error)
}

// TenantFetcher calls an API which fetches details for the given tenant from an external tenancy service, stores the tenant in the Compass DB and returns 200 OK if the tenant was successfully created.
//go:generate mockery --name=TenantFetcher --output=automock --outpkg=automock --case=underscore --disable-version-string
type TenantFetcher interface {
	FetchOnDemand(tenant, parentTenant string) error
}

// RuntimeContextService missing godoc
//go:generate mockery --name=RuntimeContextService --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextService interface {
	GetForRuntime(ctx context.Context, id, runtimeID string) (*model.RuntimeContext, error)
	ListByRuntimeIDs(ctx context.Context, runtimeIDs []string, pageSize int, cursor string) ([]*model.RuntimeContextPage, error)
}

// RuntimeContextConverter missing godoc
//go:generate mockery --name=RuntimeContextConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type RuntimeContextConverter interface {
	ToGraphQL(in *model.RuntimeContext) *graphql.RuntimeContext
	MultipleToGraphQL(in []*model.RuntimeContext) []*graphql.RuntimeContext
}

// WebhookService missing godoc
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore
type WebhookService interface {
	ListForRuntime(ctx context.Context, runtimeID string) ([]*model.Webhook, error)
	Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error)
}

// WebhookConverter missing godoc
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore
type WebhookConverter interface {
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// Resolver missing godoc
type Resolver struct {
	transact                  persistence.Transactioner
	runtimeService            RuntimeService
	scenarioAssignmentService ScenarioAssignmentService
	sysAuthSvc                SystemAuthService
	converter                 RuntimeConverter
	sysAuthConv               SystemAuthConverter
	oAuth20Svc                OAuth20Service
	eventingSvc               EventingService
	bundleInstanceAuthSvc     BundleInstanceAuthService
	selfRegManager            SelfRegisterManager
	uidService                uidService
	subscriptionSvc           SubscriptionService
	runtimeContextService     RuntimeContextService
	runtimeContextConverter   RuntimeContextConverter
	webhookService            WebhookService
	webhookConverter          WebhookConverter
	fetcher                   TenantFetcher
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, runtimeService RuntimeService, scenarioAssignmentService ScenarioAssignmentService,
	sysAuthSvc SystemAuthService, oAuthSvc OAuth20Service, conv RuntimeConverter, sysAuthConv SystemAuthConverter,
	eventingSvc EventingService, bundleInstanceAuthSvc BundleInstanceAuthService, selfRegManager SelfRegisterManager,
	uidService uidService, subscriptionSvc SubscriptionService, runtimeContextService RuntimeContextService, runtimeContextConverter RuntimeContextConverter, webhookService WebhookService, webhookConverter WebhookConverter, fetcher TenantFetcher) *Resolver {
	return &Resolver{
		transact:                  transact,
		runtimeService:            runtimeService,
		scenarioAssignmentService: scenarioAssignmentService,
		sysAuthSvc:                sysAuthSvc,
		oAuth20Svc:                oAuthSvc,
		converter:                 conv,
		sysAuthConv:               sysAuthConv,
		eventingSvc:               eventingSvc,
		bundleInstanceAuthSvc:     bundleInstanceAuthSvc,
		selfRegManager:            selfRegManager,
		uidService:                uidService,
		subscriptionSvc:           subscriptionSvc,
		runtimeContextService:     runtimeContextService,
		runtimeContextConverter:   runtimeContextConverter,
		webhookService:            webhookService,
		webhookConverter:          webhookConverter,
		fetcher:                   fetcher,
	}
}

// Runtimes missing godoc
// TODO: Proper error handling
func (r *Resolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)
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

	runtimesPage, err := r.runtimeService.List(ctx, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlRuntimes := r.converter.MultipleToGraphQL(runtimesPage.Data)

	return &graphql.RuntimePage{
		Data:       gqlRuntimes,
		TotalCount: runtimesPage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(runtimesPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(runtimesPage.PageInfo.EndCursor),
			HasNextPage: runtimesPage.PageInfo.HasNextPage,
		},
	}, nil
}

// Runtime missing godoc
func (r *Resolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtime), nil
}

// RuntimeByTokenIssuer returns a Runtime by a token issuer
func (r *Resolver) RuntimeByTokenIssuer(ctx context.Context, issuer string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.runtimeService.GetByTokenIssuer(ctx, issuer)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtime), nil
}

// RegisterRuntime missing godoc
func (r *Resolver) RegisterRuntime(ctx context.Context, in graphql.RuntimeRegisterInput) (*graphql.Runtime, error) {
	convertedIn, err := r.converter.RegisterInputFromGraphQL(in)
	if err != nil {
		return nil, err
	}

	id := r.uidService.Generate()

	labels, err := r.selfRegManager.PrepareRuntimeForSelfRegistration(ctx, convertedIn, id)
	if err != nil {
		return nil, err
	}

	if saVal, ok := in.Labels[scenarioassignment.SubaccountIDKey]; ok {
		sa, err := convertLabelValue(saVal)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting %s label", scenarioassignment.SubaccountIDKey)
		}

		parentTenant, err := tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if err := r.fetcher.FetchOnDemand(sa, parentTenant); err != nil {
			return nil, errors.Wrapf(err, "while trying to create if not exists subaccount %s", sa)
		}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		didRollback := r.transact.RollbackUnlessCommitted(ctx, tx)
		if didRollback {
			labelVal := str.CastOrEmpty(convertedIn.Labels[r.selfRegManager.GetSelfRegDistinguishingLabelKey()])
			if labelVal != "" {
				label, ok := in.Labels["region"].(string)
				if !ok {
					log.C(ctx).Errorf("An error occurred while casting region label value to string")
				} else {
					r.cleanupAndLogOnError(ctx, id, label)
				}
			}
		}
	}()

	ctx = persistence.SaveToContext(ctx, tx)

	if err = r.runtimeService.CreateWithMandatoryLabels(ctx, convertedIn, id, labels); err != nil {
		return nil, err
	}

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtime), nil
}

// UpdateRuntime missing godoc
func (r *Resolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeUpdateInput) (*graphql.Runtime, error) {
	convertedIn := r.converter.UpdateInputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = r.runtimeService.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(runtime), nil
}

// DeleteRuntime missing godoc
func (r *Resolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		r.transact.RollbackUnlessCommitted(ctx, tx)
	}()

	ctx = persistence.SaveToContext(ctx, tx)

	runtime, err := r.runtimeService.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	bundleInstanceAuths, err := r.bundleInstanceAuthSvc.ListByRuntimeID(ctx, runtime.ID)
	if err != nil {
		return nil, err
	}

	_, err = r.runtimeService.GetLabel(ctx, runtime.ID, r.selfRegManager.GetSelfRegDistinguishingLabelKey())

	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting self register info label")
		}
	} else {
		regionLabel, err := r.runtimeService.GetLabel(ctx, runtime.ID, "region")
		if err != nil {
			return nil, errors.Wrapf(err, "while getting region label")
		}

		// Committing transaction as the cleanup sends request to external service
		if err = tx.Commit(); err != nil {
			return nil, err
		}

		regionValue, ok := regionLabel.Value.(string)
		if !ok {
			return nil, errors.Wrap(err, "while casting region label value to string")
		}

		log.C(ctx).Infof("Executing clean-up for self-registered runtime with id %q", runtime.ID)
		if err := r.selfRegManager.CleanupSelfRegisteredRuntime(ctx, runtime.ID, regionValue); err != nil {
			return nil, errors.Wrap(err, "An error occurred during cleanup of self-registered runtime: ")
		}

		tx, err = r.transact.Begin()
		if err != nil {
			return nil, err
		}
		ctx = persistence.SaveToContext(ctx, tx)
	}

	currentTimestamp := timestamp.DefaultGenerator
	for _, auth := range bundleInstanceAuths {
		if auth.Status.Condition != model.BundleInstanceAuthStatusConditionUnused {
			if err := auth.SetDefaultStatus(model.BundleInstanceAuthStatusConditionUnused, currentTimestamp()); err != nil {
				log.C(ctx).WithError(err).Errorf("while update bundle instance auth status condition: %v", err)
				return nil, err
			}
			if err := r.bundleInstanceAuthSvc.Update(ctx, auth); err != nil {
				log.C(ctx).WithError(err).Errorf("Unable to update bundle instance auth with ID: %s for corresponding bundle with ID: %s: %v", auth.ID, auth.BundleID, err)
				return nil, err
			}
		}
	}

	auths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.RuntimeReference, runtime.ID)
	if err != nil {
		return nil, err
	}

	if err = r.deleteAssociatedScenarioAssignments(ctx, runtime.ID); err != nil {
		return nil, err
	}

	deletedRuntime := r.converter.ToGraphQL(runtime)

	if err = r.runtimeService.Delete(ctx, id); err != nil {
		return nil, err
	}

	if err = r.oAuth20Svc.DeleteMultipleClientCredentials(ctx, auths); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return deletedRuntime, nil
}

// GetLabel missing godoc
func (r *Resolver) GetLabel(ctx context.Context, runtimeID string, key string) (*graphql.Labels, error) {
	if runtimeID == "" {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}
	if key == "" {
		return nil, apperrors.NewInternalError("Runtime label key cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})
	resultLabels[key] = label.Value

	var gqlLabels graphql.Labels = resultLabels
	return &gqlLabels, nil
}

// SetRuntimeLabel missing godoc
func (r *Resolver) SetRuntimeLabel(ctx context.Context, runtimeID string, key string, value interface{}) (*graphql.Label, error) {
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

	err = r.runtimeService.SetLabel(ctx, &model.LabelInput{
		Key:        key,
		Value:      value,
		ObjectType: model.RuntimeLabelableObject,
		ObjectID:   runtimeID,
	})
	if err != nil {
		return nil, err
	}

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting label with key: [%s]", key)
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   label.Key,
		Value: label.Value,
	}, nil
}

// DeleteRuntimeLabel missing godoc
func (r *Resolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	label, err := r.runtimeService.GetLabel(ctx, runtimeID, key)
	if err != nil {
		return nil, err
	}

	if err = r.runtimeService.DeleteLabel(ctx, runtimeID, key); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:   key,
		Value: label.Value,
	}, nil
}

// Webhooks missing godoc
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.Runtime) ([]*graphql.Webhook, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := r.webhookService.ListForRuntime(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	gqlWebhooks, err := r.webhookConverter.MultipleToGraphQL(webhooks)
	if err != nil {
		return nil, err
	}

	return gqlWebhooks, nil
}

// Labels missing godoc
func (r *Resolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.runtimeService.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") { // TODO: Use custom error and check its type
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

// RuntimeContexts retrieves a page of RuntimeContexts for the specified Runtime
func (r *Resolver) RuntimeContexts(ctx context.Context, obj *graphql.Runtime, first *int, after *graphql.PageCursor) (*graphql.RuntimeContextPage, error) {
	param := dataloader.ParamRuntimeContext{ID: obj.ID, Ctx: ctx, First: first, After: after}
	return dataloader.RuntimeContextFor(ctx).RuntimeContextByID.Load(param)
}

// RuntimeContextsDataLoader retrieves a page of RuntimeContexts for each Runtime ID in the keys argument
func (r *Resolver) RuntimeContextsDataLoader(keys []dataloader.ParamRuntimeContext) ([]*graphql.RuntimeContextPage, []error) {
	if len(keys) == 0 {
		return nil, []error{apperrors.NewInternalError("No Runtimes found")}
	}

	ctx := keys[0].Ctx
	runtimeIDs := make([]string, 0, len(keys))
	for _, key := range keys {
		runtimeIDs = append(runtimeIDs, key.ID)
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

	runtimeContextPages, err := r.runtimeContextService.ListByRuntimeIDs(ctx, runtimeIDs, *keys[0].First, cursor)
	if err != nil {
		return nil, []error{err}
	}

	gqlRtmCtxs := make([]*graphql.RuntimeContextPage, 0, len(runtimeContextPages))
	for _, page := range runtimeContextPages {
		rtmCtxs := r.runtimeContextConverter.MultipleToGraphQL(page.Data)

		gqlRtmCtxs = append(gqlRtmCtxs, &graphql.RuntimeContextPage{Data: rtmCtxs, TotalCount: page.TotalCount, PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(page.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(page.PageInfo.EndCursor),
			HasNextPage: page.PageInfo.HasNextPage,
		}})
	}

	err = tx.Commit()
	if err != nil {
		return nil, []error{err}
	}

	return gqlRtmCtxs, nil
}

// RuntimeContext missing godoc
func (r *Resolver) RuntimeContext(ctx context.Context, obj *graphql.Runtime, id string) (*graphql.RuntimeContext, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	runtimeContext, err := r.runtimeContextService.GetForRuntime(ctx, id, obj.ID)
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

	return r.runtimeContextConverter.ToGraphQL(runtimeContext), nil
}

// Auths missing godoc
func (r *Resolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.RuntimeSystemAuth, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, pkgmodel.RuntimeReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out := make([]*graphql.RuntimeSystemAuth, 0, len(sysAuths))
	for _, sa := range sysAuths {
		c, err := r.sysAuthConv.ToGraphQL(&sa)
		if err != nil {
			return nil, err
		}
		out = append(out, c.(*graphql.RuntimeSystemAuth))
	}

	return out, nil
}

// EventingConfiguration missing godoc
func (r *Resolver) EventingConfiguration(ctx context.Context, obj *graphql.Runtime) (*graphql.RuntimeEventingConfiguration, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Runtime cannot be empty")
	}

	runtimeID, err := uuid.Parse(obj.ID)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing runtime ID as UUID")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventingCfg, err := r.eventingSvc.GetForRuntime(ctx, runtimeID)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing configuration for runtime")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while committing the transaction")
	}

	return eventing.RuntimeEventingConfigurationToGraphQL(eventingCfg), nil
}

// SubscribeTenant subscribes tenant to runtime labeled with `providerID` and `region`
func (r *Resolver) SubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	success, err := r.subscriptionSvc.SubscribeTenant(ctx, providerID, subaccountTenantID, providerSubaccountID, region)
	if err != nil {
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return success, nil
}

// UnsubscribeTenant unsubscribes tenant to runtime labeled with `providerID` and `region`
func (r *Resolver) UnsubscribeTenant(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, region string) (bool, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return false, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	success, err := r.subscriptionSvc.UnsubscribeTenant(ctx, providerID, subaccountTenantID, providerSubaccountID, region)
	if err != nil {
		return false, err
	}

	if err = tx.Commit(); err != nil {
		return false, err
	}

	return success, nil
}

// deleteAssociatedScenarioAssignments ensures that scenario assignments which are responsible for creation of certain runtime labels are deleted,
// if runtime doesn't have the scenarios label or is part of a scenario for which no scenario assignment exists => noop
func (r *Resolver) deleteAssociatedScenarioAssignments(ctx context.Context, runtimeID string) error {
	scenariosLbl, err := r.runtimeService.GetLabel(ctx, runtimeID, model.ScenariosKey)
	notFound := apperrors.IsNotFoundError(err)
	if err != nil && !notFound {
		return err
	}

	if notFound {
		return nil
	}

	scenarios, err := labelPkg.ValueToStringsSlice(scenariosLbl.Value)
	if err != nil {
		return err
	}

	for _, scenario := range scenarios {
		scenarioAssignment, err := r.scenarioAssignmentService.GetForScenarioName(ctx, scenario)
		notFound := apperrors.IsNotFoundError(err)
		if err != nil && !notFound {
			return err
		}

		if notFound {
			continue
		}

		if err := r.scenarioAssignmentService.Delete(ctx, scenarioAssignment); err != nil {
			return err
		}
	}

	return nil
}

func (r *Resolver) cleanupAndLogOnError(ctx context.Context, runtimeID, region string) {
	if err := r.selfRegManager.CleanupSelfRegisteredRuntime(ctx, runtimeID, region); err != nil {
		log.C(ctx).Errorf("An error occurred during cleanup of self-registered runtime: %v", err)
	}
}
