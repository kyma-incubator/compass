package application

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
)

type svc interface{}

type apiSvc interface{}

type eventAPISvc interface{}

type documentSvc interface{}

type Resolver struct {
	svc       svc
	converter *Converter

	apiSvc      apiSvc
	eventAPISvc eventAPISvc
	documentSvc documentSvc
}

func NewResolver(svc *Service, apiSvc apiSvc, eventAPISvc eventAPISvc, documentSvc documentSvc) *Resolver {
	return &Resolver{
		svc:         svc,
		apiSvc:      apiSvc,
		eventAPISvc: eventAPISvc,
		documentSvc: documentSvc,
		converter:   &Converter{},
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	return &graphql.ApplicationPage{
		Data:       []*graphql.Application{},
		TotalCount: 0,
		PageInfo: &graphql.PageInfo{
			HasNextPage: false,
		},
	}, nil
}

func (r *Resolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	panic("not implemented")
}

func (r *Resolver) CreateApplication(ctx context.Context, in graphql.ApplicationInput) (*graphql.Application, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationInput) (*graphql.Application, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	panic("not implemented")
}
func (r *Resolver) AddApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *Resolver) AddApplicationAnnotation(ctx context.Context, applicationID string, annotation string, value string) (string, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, annotation string) (*string, error) {
	panic("not implemented")
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
