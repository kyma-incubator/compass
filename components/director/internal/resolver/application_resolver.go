package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type applicationResolver struct {

}

func (r *applicationResolver) Applications(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.ApplicationPage, error) {
	return &gqlschema.ApplicationPage{
		Data:       []*gqlschema.Application{},
		TotalCount: 0,
		PageInfo: &gqlschema.PageInfo{
			HasNextPage: false,
		},
	}, nil
}
func (r *applicationResolver) Application(ctx context.Context, id string) (*gqlschema.Application, error) {
	panic("not implemented")
}

func (a *applicationResolver) CreateApplication(ctx context.Context, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {

}
func (a *applicationResolver) UpdateApplication(ctx context.Context, id string, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {

}
func (a *applicationResolver) DeleteApplication(ctx context.Context, id string) (*gqlschema.Application, error) {

}
func (a *applicationResolver) AddApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {

}
func (a *applicationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {

}
func (a *applicationResolver) AddApplicationAnnotation(ctx context.Context, applicationID string, annotation string, value string) (string, error) {

}
func (a *applicationResolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, annotation string) (*string, error) {

}
func (a *applicationResolver) AddApplicationWebhook(ctx context.Context, applicationID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {

}
func (a *applicationResolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {

}
func (a *applicationResolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*gqlschema.ApplicationWebhook, error) {

}

