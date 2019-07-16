package domain

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/uid"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/healthcheck"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	app         *application.Resolver
	api         *api.Resolver
	eventAPI    *eventapi.Resolver
	doc         *document.Resolver
	runtime     *runtime.Resolver
	healthCheck *healthcheck.Resolver
	webhook     *webhook.Resolver
}

func NewRootResolver() *RootResolver {
	authConverter := auth.NewConverter()

	runtimeConverter := runtime.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	apiConverter := api.NewConverter(authConverter, frConverter, versionConverter)
	eventAPIConverter := eventapi.NewConverter(frConverter, versionConverter)
	appConverter := application.NewConverter(webhookConverter, apiConverter, eventAPIConverter, docConverter)

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewRepository()
	applicationRepo := application.NewRepository()
	webhookRepo := webhook.NewRepository()
	apiRepo := api.NewAPIRepository()
	eventAPIRepo := eventapi.NewRepository()
	docRepo := document.NewRepository()

	uidService := uid.NewService()
	appSvc := application.NewService(applicationRepo, webhookRepo, apiRepo, eventAPIRepo, docRepo, uidService)
	apiSvc := api.NewService(apiRepo, uidService)
	eventAPISvc := eventapi.NewService(eventAPIRepo, uidService)
	webhookSvc := webhook.NewService(webhookRepo, uidService)
	docSvc := document.NewService(docRepo, uidService)
	runtimeSvc := runtime.NewService(runtimeRepo, uidService)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)

	return &RootResolver{
		app:         application.NewResolver(appSvc, apiSvc, eventAPISvc, docSvc, webhookSvc, appConverter, docConverter, webhookConverter, apiConverter, eventAPIConverter),
		api:         api.NewResolver(apiSvc, appSvc, apiConverter, authConverter),
		eventAPI:    eventapi.NewResolver(eventAPISvc, appSvc, eventAPIConverter),
		doc:         document.NewResolver(docSvc, appSvc, frConverter),
		runtime:     runtime.NewResolver(runtimeSvc, runtimeConverter),
		healthCheck: healthcheck.NewResolver(healthCheckSvc),
		webhook:     webhook.NewResolver(webhookSvc, appSvc, webhookConverter),
	}
}

func (r *RootResolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}
func (r *RootResolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

func (r *RootResolver) Application() graphql.ApplicationResolver {
	return &applicationResolver{r}
}

type queryResolver struct {
	*RootResolver
}

func (r *queryResolver) Applications(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	return r.app.Applications(ctx, filter, first, after)
}
func (r *queryResolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	return r.app.Application(ctx, id)
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	return r.runtime.Runtimes(ctx, filter, first, after)
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.Runtime(ctx, id)
}
func (r *queryResolver) HealthChecks(ctx context.Context, types []graphql.HealthCheckType, origin *string, first *int, after *graphql.PageCursor) (*graphql.HealthCheckPage, error) {
	return r.healthCheck.HealthChecks(ctx, types, origin, first, after)
}

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) CreateApplication(ctx context.Context, in graphql.ApplicationInput) (*graphql.Application, error) {
	return r.app.CreateApplication(ctx, in)
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationInput) (*graphql.Application, error) {
	return r.app.UpdateApplication(ctx, id, in)
}
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	return r.app.DeleteApplication(ctx, id)
}
func (r *mutationResolver) AddApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	return r.app.AddApplicationLabel(ctx, applicationID, key, values)
}
func (r *mutationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string, values []string) (*graphql.Label, error) {
	return r.app.DeleteApplicationLabel(ctx, applicationID, key, values)
}
func (r *mutationResolver) AddWebhook(ctx context.Context, applicationID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.AddApplicationWebhook(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateWebhook(ctx context.Context, webhookID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.UpdateApplicationWebhook(ctx, webhookID, in)
}
func (r *mutationResolver) DeleteWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
	return r.webhook.DeleteApplicationWebhook(ctx, webhookID)
}
func (r *mutationResolver) AddAPI(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.AddAPI(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateAPI(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.UpdateAPI(ctx, id, in)
}
func (r *mutationResolver) DeleteAPI(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	return r.api.DeleteAPI(ctx, id)
}
func (r *mutationResolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	return r.api.RefetchAPISpec(ctx, apiID)
}
func (r *mutationResolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.RuntimeAuth, error) {
	return r.api.SetAPIAuth(ctx, apiID, runtimeID, in)
}
func (r *mutationResolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.RuntimeAuth, error) {
	return r.api.DeleteAPIAuth(ctx, apiID, runtimeID)
}
func (r *mutationResolver) AddEventAPI(ctx context.Context, applicationID string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	return r.eventAPI.AddEventAPI(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateEventAPI(ctx context.Context, id string, in graphql.EventAPIDefinitionInput) (*graphql.EventAPIDefinition, error) {
	return r.eventAPI.UpdateEventAPI(ctx, id, in)
}
func (r *mutationResolver) DeleteEventAPI(ctx context.Context, id string) (*graphql.EventAPIDefinition, error) {
	return r.eventAPI.DeleteEventAPI(ctx, id)
}
func (r *mutationResolver) RefetchEventAPISpec(ctx context.Context, eventID string) (*graphql.EventAPISpec, error) {
	return r.eventAPI.RefetchEventAPISpec(ctx, eventID)
}
func (r *mutationResolver) CreateRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	return r.runtime.CreateRuntime(ctx, in)
}
func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	return r.runtime.UpdateRuntime(ctx, id, in)
}
func (r *mutationResolver) DeleteRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.DeleteRuntime(ctx, id)
}
func (r *mutationResolver) AddRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) (*graphql.Label, error) {
	return r.runtime.AddRuntimeLabel(ctx, runtimeID, key, values)
}
func (r *mutationResolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string, values []string) (*graphql.Label, error) {
	return r.runtime.DeleteRuntimeLabel(ctx, runtimeID, key, values)
}
func (r *mutationResolver) AddDocument(ctx context.Context, applicationID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return r.doc.AddDocument(ctx, applicationID, in)
}
func (r *mutationResolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	return r.doc.DeleteDocument(ctx, id)
}

type applicationResolver struct {
	*RootResolver
}

func (r *applicationResolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	return r.app.Webhooks(ctx, obj)
}
func (r *applicationResolver) Apis(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return r.app.Apis(ctx, obj, group, first, after)
}
func (r *applicationResolver) EventAPIs(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.EventAPIDefinitionPage, error) {
	return r.app.EventAPIs(ctx, obj, group, first, after)
}
func (r *applicationResolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.app.Documents(ctx, obj, first, after)
}
