package application

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
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

type APIService interface{}

type EventAPIService interface{}

type DocumentService interface{}

type WebhookService interface{}

type Resolver struct {
	svc       ApplicationService
	converter ApplicationConverter

	apiSvc      APIService
	eventAPISvc EventAPIService
	webhookSvc  WebhookService
	documentSvc DocumentService
}

func NewResolver(svc ApplicationService, apiSvc APIService, eventAPISvc EventAPIService, documentSvc DocumentService, webhookSvc WebhookService) *Resolver {
	return &Resolver{
		svc:         svc,
		apiSvc:      apiSvc,
		eventAPISvc: eventAPISvc,
		documentSvc: documentSvc,
		webhookSvc:  webhookSvc,
		converter:   &converter{},
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)

	var cursor string
	if after != nil {
		cursor = string(*after)
	}

	appPage, err := r.svc.List(ctx, labelFilter, first, &cursor)
	if err != nil {
		return nil, err
	}

	gqlApps := r.converter.MultipleToGraphQL(appPage.Data)
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
	app, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return r.converter.ToGraphQL(app), nil
}

func (r *Resolver) CreateApplication(ctx context.Context, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	id, err := r.svc.Create(ctx, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.converter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationInput) (*graphql.Application, error) {
	convertedIn := r.converter.InputFromGraphQL(in)

	err := r.svc.Update(ctx, id, convertedIn)
	if err != nil {
		return nil, err
	}

	app, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	gqlApp := r.converter.ToGraphQL(app)

	return gqlApp, nil
}
func (r *Resolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	app, err := r.svc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	deletedApp := r.converter.ToGraphQL(app)

	err = r.svc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	return deletedApp, nil
}
func (r *Resolver) AddApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	err := r.svc.AddLabel(ctx, applicationID, key, values)
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:    key,
		Values: values,
	}, nil
}

func (r *Resolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	err := r.svc.DeleteLabel(ctx, applicationID, key, values)
	if err != nil {
		return nil, err
	}

	app, err := r.svc.Get(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	return &graphql.Label{
		Key:    key,
		Values: app.Labels[key],
	}, nil
}

func (r *Resolver) AddApplicationAnnotation(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Annotation, error) {
	err := r.svc.AddAnnotation(ctx, applicationID, key, value)
	if err != nil {
		return nil, err
	}

	return &graphql.Annotation{
		Key:   key,
		Value: value,
	}, nil
}

func (r *Resolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, key string) (*graphql.Annotation, error) {
	app, err := r.svc.Get(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	value := app.Annotations[key]

	err = r.svc.DeleteAnnotation(ctx, applicationID, key)
	if err != nil {
		return nil, err
	}

	return &graphql.Annotation{
		Key:   key,
		Value: value,
	}, nil
}

func (r *Resolver) AddApplicationWebhook(ctx context.Context, applicationID string, in graphql.ApplicationWebhookInput) (*graphql.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in graphql.ApplicationWebhookInput) (*graphql.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*graphql.ApplicationWebhook, error) {
	panic("not implemented")
}

func (r *Resolver) Apis(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	panic("not implemented")
}
func (r *Resolver) EventAPIs(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.EventAPIDefinitionPage, error) {
	panic("not implemented")
}
func (r *Resolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	panic("not implemented")
}

func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.ApplicationWebhook, error) {
	panic("not implemented")
}
