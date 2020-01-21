package application

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
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

//go:generate mockery -name=ApplicationConverter -output=automock -outpkg=automock -case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	MultipleToGraphQL(in []*model.Application) []*graphql.Application
	CreateInputFromGraphQL(in graphql.ApplicationRegisterInput) model.ApplicationRegisterInput
	UpdateInputFromGraphQL(in graphql.ApplicationUpdateInput) model.ApplicationUpdateInput
	ConvertToModel(obj *graphql.Application) (model.Application, error)
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.APIDefinitionPage, error)
	Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*model.APIDefinition, error)
	GetForApplication(ctx context.Context, id string, applicationID string) (*model.APIDefinition, error)
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

//go:generate mockery -name=EventDefinitionService -output=automock -outpkg=automock -case=underscore
type EventDefinitionService interface {
	Get(ctx context.Context, id string) (*model.EventDefinition, error)
	GetForApplication(ctx context.Context, id string, applicationID string) (*model.EventDefinition, error)
	List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.EventDefinitionPage, error)
	Create(ctx context.Context, applicationID string, in model.EventDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventDefinitionInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=EventAPIConverter -output=automock -outpkg=automock -case=underscore
type EventAPIConverter interface {
	ToGraphQL(in *model.EventDefinition) *graphql.EventDefinition
	MultipleToGraphQL(in []*model.EventDefinition) []*graphql.EventDefinition
	MultipleInputFromGraphQL(in []*graphql.EventDefinitionInput) []*model.EventDefinitionInput
	InputFromGraphQL(in *graphql.EventDefinitionInput) *model.EventDefinitionInput
}

//go:generate mockery -name=EventingService -output=automock -outpkg=automock -case=underscore
type EventingService interface {
	CleanupAfterUnregisteringApplication(ctx context.Context, appID uuid.UUID) (*model.ApplicationEventingConfiguration, error)
	GetForApplication(ctx context.Context, app model.Application) (*model.ApplicationEventingConfiguration, error)
}

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	List(ctx context.Context, applicationID string, pageSize int, cursor string) (*model.DocumentPage, error)
}

//go:generate mockery -name=WebhookService -output=automock -outpkg=automock -case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.Webhook, error)
	List(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	Create(ctx context.Context, applicationID string, in model.WebhookInput) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=SystemAuthService -output=automock -outpkg=automock -case=underscore
type SystemAuthService interface {
	ListForObject(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error)
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	MultipleToGraphQL(in []*model.Document) []*graphql.Document
	MultipleInputFromGraphQL(in []*graphql.DocumentInput) []*model.DocumentInput
}

//go:generate mockery -name=WebhookConverter -output=automock -outpkg=automock -case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.Webhook) *graphql.Webhook
	MultipleToGraphQL(in []*model.Webhook) []*graphql.Webhook
	InputFromGraphQL(in *graphql.WebhookInput) *model.WebhookInput
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) []*model.WebhookInput
}

//go:generate mockery -name=SystemAuthConverter -output=automock -outpkg=automock -case=underscore
type SystemAuthConverter interface {
	ToGraphQL(in *model.SystemAuth) *graphql.SystemAuth
}

//go:generate mockery -name=OAuth20Service -output=automock -outpkg=automock -case=underscore
type OAuth20Service interface {
	DeleteMultipleClientCredentials(ctx context.Context, auths []model.SystemAuth) error
}

//go:generate mockery -name=RuntimeService -output=automock -outpkg=automock -case=underscore
type RuntimeService interface {
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error)
	GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error)
}

type Resolver struct {
	transact persistence.Transactioner

	appSvc       ApplicationService
	appConverter ApplicationConverter

	apiSvc      APIService
	eventDefSvc EventDefinitionService
	webhookSvc  WebhookService
	documentSvc DocumentService
	oAuth20Svc  OAuth20Service
	sysAuthSvc  SystemAuthService

	documentConverter DocumentConverter
	webhookConverter  WebhookConverter
	apiConverter      APIConverter
	eventApiConverter EventAPIConverter
	sysAuthConv       SystemAuthConverter
	eventingSvc       EventingService
}

func NewResolver(transact persistence.Transactioner,
	svc ApplicationService,
	apiSvc APIService,
	eventDefSrv EventDefinitionService,
	documentSvc DocumentService,
	webhookSvc WebhookService,
	oAuth20Svc OAuth20Service,
	sysAuthSvc SystemAuthService,
	appConverter ApplicationConverter,
	documentConverter DocumentConverter,
	webhookConverter WebhookConverter,
	apiConverter APIConverter,
	eventAPIConverter EventAPIConverter,
	sysAuthConv SystemAuthConverter,
	eventingSvc EventingService) *Resolver {
	return &Resolver{
		transact:          transact,
		appSvc:            svc,
		apiSvc:            apiSvc,
		eventDefSvc:       eventDefSrv,
		documentSvc:       documentSvc,
		webhookSvc:        webhookSvc,
		oAuth20Svc:        oAuth20Svc,
		sysAuthSvc:        sysAuthSvc,
		appConverter:      appConverter,
		documentConverter: documentConverter,
		webhookConverter:  webhookConverter,
		apiConverter:      apiConverter,
		eventApiConverter: eventAPIConverter,
		sysAuthConv:       sysAuthConv,
		eventingSvc:       eventingSvc,
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

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
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
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
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
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
	totalCount := len(gqlApps)

	return &graphql.ApplicationPage{
		Data:       gqlApps,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appPage.PageInfo.EndCursor),
			HasNextPage: appPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) RegisterApplication(ctx context.Context, in graphql.ApplicationRegisterInput) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.appConverter.CreateInputFromGraphQL(in)
	id, err := r.appSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	convertedIn := r.appConverter.UpdateInputFromGraphQL(in)
	err = r.appSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) UnregisterApplication(ctx context.Context, id string) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

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

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	deletedApp := r.appConverter.ToGraphQL(app)

	return deletedApp, nil
}
func (r *Resolver) SetApplicationLabel(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Label, error) {
	// TODO: Use @validation directive on input type instead, after resolving https://github.com/kyma-incubator/compass/issues/515
	gqlLabel := graphql.LabelInput{Key: key, Value: value}
	if err := gqlLabel.Validate(); err != nil {
		return nil, errors.Wrap(err, "validation error for type LabelInput")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

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
	defer r.transact.RollbackUnlessCommited(tx)

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

func (r *Resolver) ApiDefinitions(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	apisPage, err := r.apiSvc.List(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApis := r.apiConverter.MultipleToGraphQL(apisPage.Data)
	totalCount := len(gqlApis)

	return &graphql.APIDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(apisPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(apisPage.PageInfo.EndCursor),
			HasNextPage: apisPage.PageInfo.HasNextPage,
		},
	}, nil
}
func (r *Resolver) EventDefinitions(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	eventAPIPage, err := r.eventDefSvc.List(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApis := r.eventApiConverter.MultipleToGraphQL(eventAPIPage.Data)
	totalCount := len(gqlApis)

	return &graphql.EventDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventAPIPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventAPIPage.PageInfo.EndCursor),
			HasNextPage: eventAPIPage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) APIDefinition(ctx context.Context, id string, application *graphql.Application) (*graphql.APIDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	api, err := r.apiSvc.GetForApplication(ctx, id, application.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.apiConverter.ToGraphQL(api), nil
}

func (r *Resolver) EventDefinition(ctx context.Context, id string, application *graphql.Application) (*graphql.EventDefinition, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventAPI, err := r.eventDefSvc.GetForApplication(ctx, id, application.ID)
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return nil, nil
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return r.eventApiConverter.ToGraphQL(eventAPI), nil
}

// TODO: Proper error handling
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	if first == nil {
		return nil, errors.New("missing required parameter 'first'")
	}

	documentsPage, err := r.documentSvc.List(ctx, obj.ID, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlDocuments := r.documentConverter.MultipleToGraphQL(documentsPage.Data)
	totalCount := len(gqlDocuments)

	return &graphql.DocumentPage{
		Data:       gqlDocuments,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(documentsPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(documentsPage.PageInfo.EndCursor),
			HasNextPage: documentsPage.PageInfo.HasNextPage,
		},
	}, nil
}

// TODO: Proper error handling
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := r.webhookSvc.List(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	gqlWebhooks := r.webhookConverter.MultipleToGraphQL(webhooks)

	return gqlWebhooks, nil
}

func (r *Resolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (*graphql.Labels, error) {
	if obj == nil {
		return nil, errors.New("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.appSvc.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, nil
		}

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		resultLabels[label.Key] = label.Value
	}

	var gqlLabels graphql.Labels = resultLabels

	return &gqlLabels, nil
}

func (r *Resolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.SystemAuth, error) {
	if obj == nil {
		return nil, errors.New("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)
	ctx = persistence.SaveToContext(ctx, tx)

	sysAuths, err := r.sysAuthSvc.ListForObject(ctx, model.ApplicationReference, obj.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	var out []*graphql.SystemAuth
	for _, sa := range sysAuths {
		c := r.sysAuthConv.ToGraphQL(&sa)
		out = append(out, c)
	}

	return out, nil
}

func (r *Resolver) EventingConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventingConfiguration, error) {
	if obj == nil {
		return nil, errors.New("Application cannot be empty")
	}
	app, err := r.appConverter.ConvertToModel(obj)
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, errors.Wrap(err, "while opening the transaction")
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = persistence.SaveToContext(ctx, tx)

	eventingCfg, err := r.eventingSvc.GetForApplication(ctx, app)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching eventing cofiguration for application")
	}

	if err = tx.Commit(); err != nil {
		return nil, errors.Wrap(err, "while commiting the transaction")
	}

	return eventing.ApplicationEventingConfigurationToGraphQL(eventingCfg), nil
}
