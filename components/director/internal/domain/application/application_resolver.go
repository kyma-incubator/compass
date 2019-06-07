package application

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type Resolver struct {
	svc       *Service
	converter *Converter
}

func NewResolver(svc *Service) *Resolver {
	return &Resolver{
		svc:       svc,
		converter: &Converter{},
	}
}

func (r *Resolver) Applications(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.ApplicationPage, error) {
	return &gqlschema.ApplicationPage{
		Data:       []*gqlschema.Application{},
		TotalCount: 0,
		PageInfo: &gqlschema.PageInfo{
			HasNextPage: false,
		},
	}, nil
}
func (r *Resolver) Application(ctx context.Context, id string) (*gqlschema.Application, error) {
	panic("not implemented")
}

func (r *Resolver) CreateApplication(ctx context.Context, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateApplication(ctx context.Context, id string, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplication(ctx context.Context, id string) (*gqlschema.Application, error) {
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
func (r *Resolver) AddApplicationWebhook(ctx context.Context, applicationID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *Resolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *Resolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
