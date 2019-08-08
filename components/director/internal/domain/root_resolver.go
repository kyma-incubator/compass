package domain

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"

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
	labelDef    *labeldef.Resolver
}

func NewRootResolver(transact persistence.Transactioner) *RootResolver {
	authConverter := auth.NewConverter()

	runtimeConverter := runtime.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	apiConverter := api.NewConverter(authConverter, frConverter, versionConverter)
	eventAPIConverter := eventapi.NewConverter(frConverter, versionConverter)
	appConverter := application.NewConverter(webhookConverter, apiConverter, eventAPIConverter, docConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewPostgresRepository()
	applicationRepo := application.NewRepository()
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)

	webhookRepo := webhook.NewRepository()
	apiRepo := api.NewAPIRepository()
	eventAPIRepo := eventapi.NewRepository()
	docRepo := document.NewRepository()

	uidService := uid.NewService()
	labelUpsertService := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidService)
	scenariosService := labeldef.NewScenariosService(labelDefRepo, uidService)
	appSvc := application.NewService(applicationRepo, webhookRepo, apiRepo, eventAPIRepo, docRepo, runtimeRepo, labelRepo, labelUpsertService, scenariosService, uidService)
	apiSvc := api.NewService(apiRepo, uidService)
	eventAPISvc := eventapi.NewService(eventAPIRepo, uidService)
	webhookSvc := webhook.NewService(webhookRepo, uidService)
	docSvc := document.NewService(docRepo, uidService)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosService, labelUpsertService, uidService)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)
	labelDefService := labeldef.NewService(labelDefRepo, labelRepo, uidService)

	return &RootResolver{
		app:         application.NewResolver(transact, appSvc, apiSvc, eventAPISvc, docSvc, webhookSvc, appConverter, docConverter, webhookConverter, apiConverter, eventAPIConverter),
		api:         api.NewResolver(apiSvc, appSvc, apiConverter, authConverter),
		eventAPI:    eventapi.NewResolver(eventAPISvc, appSvc, eventAPIConverter),
		doc:         document.NewResolver(docSvc, appSvc, frConverter),
		runtime:     runtime.NewResolver(transact, runtimeSvc, runtimeConverter),
		healthCheck: healthcheck.NewResolver(healthCheckSvc),
		webhook:     webhook.NewResolver(webhookSvc, appSvc, webhookConverter),
		labelDef:    labeldef.NewResolver(labelDefService, labelDefConverter, transact),
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
func (r *RootResolver) Runtime() graphql.RuntimeResolver {
	return &runtimeResolver{r}
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
func (r *queryResolver) ApplicationsForRuntime(ctx context.Context, runtimeID string, first *int, after *graphql.PageCursor) (*graphql.ApplicationPage, error) {
	return r.app.ApplicationsForRuntime(ctx, runtimeID, first, after)
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	return r.runtime.Runtimes(ctx, filter, first, after)
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.Runtime(ctx, id)
}
func (r *queryResolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinitions(ctx)
}
func (r *queryResolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinition(ctx, key)
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
func (r *mutationResolver) AddDocument(ctx context.Context, applicationID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return r.doc.AddDocument(ctx, applicationID, in)
}
func (r *mutationResolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	return r.doc.DeleteDocument(ctx, id)
}
func (r *mutationResolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	return r.labelDef.CreateLabelDefinition(ctx, in)
}
func (r *mutationResolver) UpdateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	return r.labelDef.UpdateLabelDefinition(ctx, in)
}
func (r *mutationResolver) DeleteLabelDefinition(ctx context.Context, key string, deleteRelatedLabels *bool) (*graphql.LabelDefinition, error) {
	return r.labelDef.DeleteLabelDefinition(ctx, key, deleteRelatedLabels)
}
func (r *mutationResolver) SetApplicationLabel(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Label, error) {
	return r.app.SetApplicationLabel(ctx, applicationID, key, value)
}
func (r *mutationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string) (*graphql.Label, error) {
	return r.app.DeleteApplicationLabel(ctx, applicationID, key)
}
func (r *mutationResolver) SetRuntimeLabel(ctx context.Context, runtimeID string, key string, value interface{}) (*graphql.Label, error) {
	return r.runtime.SetRuntimeLabel(ctx, runtimeID, key, value)
}
func (r *mutationResolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string) (*graphql.Label, error) {
	return r.runtime.DeleteRuntimeLabel(ctx, runtimeID, key)
}

type applicationResolver struct {
	*RootResolver
}

func (r *applicationResolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (graphql.Labels, error) {
	return r.app.Labels(ctx, obj, key)
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

type runtimeResolver struct {
	*RootResolver
}

func (r *runtimeResolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	return r.runtime.Labels(ctx, obj, key)
}
