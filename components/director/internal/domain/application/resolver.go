package application

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ContextValueSetter -output=automock -outpkg=automock -case=underscore
type ContextValueSetter interface {
	WithValue(parent context.Context, key interface{}, val interface{}) context.Context
}

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationInput) error
	Get(ctx context.Context, id string) (*model.Application, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	ListByRuntimeID(ctx context.Context, runtimeID string, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	SetLabel(ctx context.Context, label *model.LabelInput) error
	GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error)
	ListLabels(ctx context.Context, applicationID string) (map[string]*model.Label, error)
	DeleteLabel(ctx context.Context, applicationID string, key string) error
}

//go:generate mockery -name=ApplicationConverter -output=automock -outpkg=automock -case=underscore
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	MultipleToGraphQL(in []*model.Application) []*graphql.Application
	InputFromGraphQL(in graphql.ApplicationInput) model.ApplicationInput
}

//go:generate mockery -name=APIService -output=automock -outpkg=automock -case=underscore
type APIService interface {
	List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.APIDefinitionPage, error)
	Create(ctx context.Context, applicationID string, in model.APIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.APIDefinitionInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=APIConverter -output=automock -outpkg=automock -case=underscore
type APIConverter interface {
	ToGraphQL(in *model.APIDefinition) *graphql.APIDefinition
	MultipleToGraphQL(in []*model.APIDefinition) []*graphql.APIDefinition
	MultipleInputFromGraphQL(in []*graphql.APIDefinitionInput) []*model.APIDefinitionInput
	InputFromGraphQL(in *graphql.APIDefinitionInput) *model.APIDefinitionInput
}

//go:generate mockery -name=EventAPIService -output=automock -outpkg=automock -case=underscore
type EventAPIService interface {
	List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.EventAPIDefinitionPage, error)
	Create(ctx context.Context, applicationID string, in model.EventAPIDefinitionInput) (string, error)
	Update(ctx context.Context, id string, in model.EventAPIDefinitionInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=EventAPIConverter -output=automock -outpkg=automock -case=underscore
type EventAPIConverter interface {
	ToGraphQL(in *model.EventAPIDefinition) *graphql.EventAPIDefinition
	MultipleToGraphQL(in []*model.EventAPIDefinition) []*graphql.EventAPIDefinition
	MultipleInputFromGraphQL(in []*graphql.EventAPIDefinitionInput) []*model.EventAPIDefinitionInput
	InputFromGraphQL(in *graphql.EventAPIDefinitionInput) *model.EventAPIDefinitionInput
}

//go:generate mockery -name=DocumentService -output=automock -outpkg=automock -case=underscore
type DocumentService interface {
	List(ctx context.Context, applicationID string, pageSize *int, cursor *string) (*model.DocumentPage, error)
}

//go:generate mockery -name=WebhookService -output=automock -outpkg=automock -case=underscore
type WebhookService interface {
	Get(ctx context.Context, id string) (*model.Webhook, error)
	List(ctx context.Context, applicationID string) ([]*model.Webhook, error)
	Create(ctx context.Context, applicationID string, in model.WebhookInput) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput) error
	Delete(ctx context.Context, id string) error
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

type Resolver struct {
	transact       persistence.Transactioner
	ctxValueSetter ContextValueSetter

	appSvc       ApplicationService
	appConverter ApplicationConverter

	apiSvc      APIService
	eventAPISvc EventAPIService
	webhookSvc  WebhookService
	documentSvc DocumentService

	documentConverter DocumentConverter
	webhookConverter  WebhookConverter
	apiConverter      APIConverter
	eventApiConverter EventAPIConverter
}

func NewResolver(transact persistence.Transactioner, ctxValueSetter ContextValueSetter, svc ApplicationService, apiSvc APIService, eventAPISvc EventAPIService, documentSvc DocumentService, webhookSvc WebhookService, appConverter ApplicationConverter, documentConverter DocumentConverter, webhookConverter WebhookConverter, apiConverter APIConverter, eventAPIConverter EventAPIConverter) *Resolver {
	return &Resolver{
		transact:          transact,
		ctxValueSetter:    ctxValueSetter,
		appSvc:            svc,
		apiSvc:            apiSvc,
		eventAPISvc:       eventAPISvc,
		documentSvc:       documentSvc,
		webhookSvc:        webhookSvc,
		appConverter:      appConverter,
		documentConverter: documentConverter,
		webhookConverter:  webhookConverter,
		apiConverter:      apiConverter,
		eventApiConverter: eventAPIConverter,
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	appPage, err := r.appSvc.List(ctx, labelFilter, first, &cursor)
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

func (r *Resolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	app, err := r.appSvc.Get(ctx, id)
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

	appPage, err := r.appSvc.ListByRuntimeID(ctx, runtimeID, first, &cursor)
	if err != nil {
		return nil, errors.Wrap(err, "while getting all Application for Runtime")
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

func (r *Resolver) CreateApplication(ctx context.Context, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.appConverter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

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
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.appConverter.InputFromGraphQL(in)

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

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
func (r *Resolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedApp := r.appConverter.ToGraphQL(app)

	err = r.appSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return deletedApp, nil
}
func (r *Resolver) SetApplicationLabel(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Label, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

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

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

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

func (r *Resolver) Apis(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	apisPage, err := r.apiSvc.List(ctx, obj.ID, first, &cursor)
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
func (r *Resolver) EventAPIs(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.EventAPIDefinitionPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	eventAPIPage, err := r.eventAPISvc.List(ctx, obj.ID, first, &cursor)
	if err != nil {
		return nil, err
	}

	gqlApis := r.eventApiConverter.MultipleToGraphQL(eventAPIPage.Data)
	totalCount := len(gqlApis)

	return &graphql.EventAPIDefinitionPage{
		Data:       gqlApis,
		TotalCount: totalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(eventAPIPage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(eventAPIPage.PageInfo.EndCursor),
			HasNextPage: eventAPIPage.PageInfo.HasNextPage,
		},
	}, nil
}

// TODO: Proper error handling
// TODO: Pagination
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	documentsPage, err := r.documentSvc.List(ctx, obj.ID, first, &cursor)
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
	webhooks, err := r.webhookSvc.List(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	gqlWebhooks := r.webhookConverter.MultipleToGraphQL(webhooks)

	return gqlWebhooks, nil
}

func (r *Resolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, errors.New("Application cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommited(tx)

	ctx = r.ctxValueSetter.WithValue(ctx, persistence.PersistenceCtxKey, tx)

	itemMap, err := r.appSvc.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return graphql.Labels{}, nil
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

	return resultLabels, nil
}
