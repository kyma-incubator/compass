package application

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

//go:generate mockery -name=ApplicationService -output=automock -outpkg=automock -case=underscore
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationInput) error
	Get(ctx context.Context, id string) (*model.Application, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error)
	AddLabel(ctx context.Context, applicationID string, key string, values []string) error
	DeleteLabel(ctx context.Context, applicationID string, key string, values []string) error
	AddAnnotation(ctx context.Context, applicationID string, key string, value interface{}) error
	DeleteAnnotation(ctx context.Context, applicationID string, key string) error
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
	Create(ctx context.Context, id string, applicationID string, in model.APIDefinitionInput) (string, error)
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
	Create(ctx context.Context, id string, applicationID string, in model.EventAPIDefinitionInput) (string, error)
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
	Get(ctx context.Context, id string) (*model.ApplicationWebhook, error)
	List(ctx context.Context, applicationID string) ([]*model.ApplicationWebhook, error)
	Create(ctx context.Context, applicationID string, in model.ApplicationWebhookInput) (string, error)
	Update(ctx context.Context, id string, in model.ApplicationWebhookInput) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=DocumentConverter -output=automock -outpkg=automock -case=underscore
type DocumentConverter interface {
	MultipleToGraphQL(in []*model.Document) []*graphql.Document
	MultipleInputFromGraphQL(in []*graphql.DocumentInput) []*model.DocumentInput
}

//go:generate mockery -name=WebhookConverter -output=automock -outpkg=automock -case=underscore
type WebhookConverter interface {
	ToGraphQL(in *model.ApplicationWebhook) *graphql.ApplicationWebhook
	MultipleToGraphQL(in []*model.ApplicationWebhook) []*graphql.ApplicationWebhook
	InputFromGraphQL(in *graphql.ApplicationWebhookInput) *model.ApplicationWebhookInput
	MultipleInputFromGraphQL(in []*graphql.ApplicationWebhookInput) []*model.ApplicationWebhookInput
}

type Resolver struct {
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

func NewResolver(svc ApplicationService, apiSvc APIService, eventAPISvc EventAPIService, documentSvc DocumentService, webhookSvc WebhookService, appConverter ApplicationConverter, documentConverter DocumentConverter, webhookConverter WebhookConverter, apiConverter APIConverter, eventAPIConverter EventAPIConverter) *Resolver {
	return &Resolver{
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

func (r *Resolver) CreateApplication(ctx context.Context, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.appConverter.InputFromGraphQL(in)

	id, err := r.appSvc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.appConverter.InputFromGraphQL(in)

	err := r.appSvc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedApp := r.appConverter.ToGraphQL(app)

	err = r.appSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return deletedApp, nil
}
func (r *Resolver) AddApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	err := r.appSvc.AddLabel(ctx, applicationID, key, values)
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:    key,
		Values: values,
	}, nil
}

func (r *Resolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	err := r.appSvc.DeleteLabel(ctx, applicationID, key, values)
	if err != nil {
		return nil, err
	}

	app, err := r.appSvc.Get(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:    key,
		Values: app.Labels[key],
	}, nil
}

func (r *Resolver) AddApplicationAnnotation(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Annotation, error) {
	err := r.appSvc.AddAnnotation(ctx, applicationID, key, value)
	if err != nil {
		return nil, err
	}

	return &graphql.Annotation{
		Key:   key,
		Value: value,
	}, nil
}

func (r *Resolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, key string) (*graphql.Annotation, error) {
	app, err := r.appSvc.Get(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	value := app.Annotations[key]

	err = r.appSvc.DeleteAnnotation(ctx, applicationID, key)
	if err != nil {
		return nil, err
	}

	return &graphql.Annotation{
		Key:   key,
		Value: value,
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
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.ApplicationWebhook, error) {
	webhooks, err := r.webhookSvc.List(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	gqlWebhooks := r.webhookConverter.MultipleToGraphQL(webhooks)

	return gqlWebhooks, nil
}

func (r *Resolver) AddApplicationWebhook(ctx context.Context, applicationID string, in graphql.ApplicationWebhookInput) (*graphql.ApplicationWebhook, error) {
	convertedIn := r.webhookConverter.InputFromGraphQL(&in)

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
