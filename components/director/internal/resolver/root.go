package resolver

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/gqlschema"
)

type Resolver struct{
	app *applicationResolver
	api *apiResolver
	eventAPI *eventAPIResolver
	runtime *runtimeResolver
}

func New() gqlschema.ResolverRoot {
	return &Resolver{}
}

func (r *Resolver) Mutation() gqlschema.MutationResolver {
	return &mutationResolver{r}
}
func (r *Resolver) Query() gqlschema.QueryResolver {
	return &queryResolver{r}
}

type mutationResolver struct{ *Resolver }

func (r *mutationResolver) CreateApplication(ctx context.Context, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in gqlschema.ApplicationInput) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, label string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationAnnotation(ctx context.Context, applicationID string, annotation string, value string) (string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationAnnotation(ctx context.Context, applicationID string, annotation string) (*string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddApplicationWebhook(ctx context.Context, applicationID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateApplicationWebhook(ctx context.Context, webhookID string, in gqlschema.ApplicationWebhookInput) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteApplicationWebhook(ctx context.Context, webhookID string) (*gqlschema.ApplicationWebhook, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddAPI(ctx context.Context, applicationID string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateAPI(ctx context.Context, id string, in gqlschema.APIDefinitionInput) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAPI(ctx context.Context, id string) (*gqlschema.APIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) RefetchAPISpec(ctx context.Context, apiID string) (*gqlschema.APISpec, error) {
	panic("not implemented")
}
func (r *mutationResolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in gqlschema.AuthInput) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*gqlschema.RuntimeAuth, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddEventAPI(ctx context.Context, applicationID string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateEventAPI(ctx context.Context, id string, in gqlschema.EventAPIDefinitionInput) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteEventAPI(ctx context.Context, id string) (*gqlschema.EventAPIDefinition, error) {
	panic("not implemented")
}
func (r *mutationResolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*gqlschema.EventAPISpec, error) {
	panic("not implemented")
}
func (r *mutationResolver) CreateRuntime(ctx context.Context, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in gqlschema.RuntimeInput) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntime(ctx context.Context, id string) (*gqlschema. Runtime, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntimeLabel(ctx context.Context, id string, key string, values []string) ([]string, error) {
	panic("not implemented")
}
func (r *mutationResolver) AddRuntimeAnnotation(ctx context.Context, runtimeID string, key string, value string) (string, error) {
	panic("not implemented")
}
func (r *mutationResolver) DeleteRuntimeAnnotation(ctx context.Context, id string, key string) (*string, error) {
	panic("not implemented")
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Applications(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.ApplicationPage, error) {
	return &gqlschema.ApplicationPage{
		Data:       []*gqlschema.Application{},
		TotalCount: 0,
		PageInfo: &gqlschema.PageInfo{
			HasNextPage: false,
		},
	}, nil
}
func (r *queryResolver) Application(ctx context.Context, id string) (*gqlschema.Application, error) {
	panic("not implemented")
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*gqlschema.LabelFilter, first *int, after *string) (*gqlschema.RuntimePage, error) {
	panic("not implemented")
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*gqlschema.Runtime, error) {
	panic("not implemented")
}
func (r *queryResolver) HealthChecks(ctx context.Context, types []gqlschema.HealthCheckType, origin *string, first *int, after *string) (*gqlschema.HealthCheckPage, error) {
	panic("not implemented")
}
