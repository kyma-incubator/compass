package domain

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/event"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/healthcheck"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/graphql_client"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	app         *application.Resolver
	appTemplate *apptemplate.Resolver
	api         *api.Resolver
	eventAPI    *eventapi.Resolver
	doc         *document.Resolver
	runtime     *runtime.Resolver
	healthCheck *healthcheck.Resolver
	webhook     *webhook.Resolver
	labelDef    *labeldef.Resolver
	token       *onetimetoken.Resolver
	systemAuth  *systemauth.Resolver
	oAuth20     *oauth20.Resolver
	intSys      *integrationsystem.Resolver
}

func NewRootResolver(transact persistence.Transactioner, scopeCfgProvider *scope.Provider, oneTimeTokenCfg onetimetoken.Config, oAuth20Cfg oauth20.Config, eventCfg event.Config) *RootResolver {
	authConverter := auth.NewConverter()
	apiRtmAuthConverter := apiruntimeauth.NewConverter(authConverter)
	runtimeConverter := runtime.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	apiConverter := api.NewConverter(authConverter, frConverter, versionConverter)
	eventAPIConverter := eventapi.NewConverter(frConverter, versionConverter)
	appConverter := application.NewConverter(webhookConverter, apiConverter, eventAPIConverter, docConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	tokenConverter := onetimetoken.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	intSysConverter := integrationsystem.NewConverter()
	appTemplateConverter := apptemplate.NewConverter(appConverter)

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewRepository()
	applicationRepo := application.NewRepository(appConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	apiRepo := api.NewRepository(apiConverter)
	eventAPIRepo := eventapi.NewRepository(eventAPIConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	apiRtmAuthRepo := apiruntimeauth.NewRepository(apiRtmAuthConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)

	connectorGCLI := graphql_client.NewGraphQLClient(oneTimeTokenCfg.OneTimeTokenURL)

	uidSvc := uid.NewService()
	apiRtmAuthSvc := apiruntimeauth.NewService(apiRtmAuthRepo, uidSvc)
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewScenariosService(labelDefRepo, uidSvc)
	appSvc := application.NewService(applicationRepo, webhookRepo, apiRepo, eventAPIRepo, docRepo, runtimeRepo, labelRepo, fetchRequestRepo, intSysRepo, labelUpsertSvc, scenariosSvc, uidSvc)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, fetchRequestRepo, uidSvc)
	eventAPISvc := eventapi.NewService(eventAPIRepo, fetchRequestRepo, uidSvc)
	webhookSvc := webhook.NewService(webhookRepo, uidSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelUpsertSvc, uidSvc)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, uidSvc)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	tokenSvc := onetimetoken.NewTokenService(connectorGCLI, systemAuthSvc, oneTimeTokenCfg.ConnectorURL)
	oAuth20Svc := oauth20.NewService(scopeCfgProvider, uidSvc, oAuth20Cfg)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)

	return &RootResolver{
		app:         application.NewResolver(transact, appSvc, apiSvc, eventAPISvc, docSvc, webhookSvc, systemAuthSvc, oAuth20Svc, appConverter, docConverter, webhookConverter, apiConverter, eventAPIConverter, systemAuthConverter, eventCfg.DefaultEventURL),
		appTemplate: apptemplate.NewResolver(transact, appSvc, appConverter, appTemplateSvc, appTemplateConverter),
		api:         api.NewResolver(transact, apiSvc, appSvc, runtimeSvc, apiRtmAuthSvc, apiConverter, authConverter, frConverter, apiRtmAuthConverter),
		eventAPI:    eventapi.NewResolver(transact, eventAPISvc, appSvc, eventAPIConverter, frConverter),
		doc:         document.NewResolver(transact, docSvc, appSvc, frConverter),
		runtime:     runtime.NewResolver(transact, runtimeSvc, systemAuthSvc, oAuth20Svc, runtimeConverter, systemAuthConverter),
		healthCheck: healthcheck.NewResolver(healthCheckSvc),
		webhook:     webhook.NewResolver(transact, webhookSvc, appSvc, webhookConverter),
		labelDef:    labeldef.NewResolver(transact, labelDefSvc, labelDefConverter),
		token:       onetimetoken.NewTokenResolver(transact, tokenSvc, tokenConverter),
		systemAuth:  systemauth.NewResolver(transact, systemAuthSvc, oAuth20Svc, systemAuthConverter),
		oAuth20:     oauth20.NewResolver(transact, oAuth20Svc, appSvc, runtimeSvc, intSysSvc, systemAuthSvc, systemAuthConverter),
		intSys:      integrationsystem.NewResolver(transact, intSysSvc, systemAuthSvc, oAuth20Svc, intSysConverter, systemAuthConverter),
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
func (r *RootResolver) APIDefinition() graphql.APIDefinitionResolver {
	return &apiDefinitionResolver{r}
}
func (r *RootResolver) APISpec() graphql.APISpecResolver {
	return &apiSpecResolver{r}
}
func (r *RootResolver) Document() graphql.DocumentResolver {
	return &documentResolver{r}
}
func (r *RootResolver) EventAPISpec() graphql.EventAPISpecResolver {
	return &eventAPISpecResolver{r}
}

func (r *RootResolver) IntegrationSystem() graphql.IntegrationSystemResolver {
	return &integrationSystemResolver{r}
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
func (r *queryResolver) ApplicationTemplates(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.ApplicationTemplatePage, error) {
	return r.appTemplate.ApplicationTemplates(ctx, first, after)
}
func (r *queryResolver) ApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.ApplicationTemplate(ctx, id)
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
func (r *queryResolver) IntegrationSystems(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.IntegrationSystemPage, error) {
	return r.intSys.IntegrationSystems(ctx, first, after)
}
func (r *queryResolver) IntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	return r.intSys.IntegrationSystem(ctx, id)
}

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) CreateApplication(ctx context.Context, in graphql.ApplicationCreateInput) (*graphql.Application, error) {
	return r.app.CreateApplication(ctx, in)
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	return r.app.UpdateApplication(ctx, id, in)
}
func (r *mutationResolver) DeleteApplication(ctx context.Context, id string) (*graphql.Application, error) {
	return r.app.DeleteApplication(ctx, id)
}
func (r *mutationResolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.CreateApplicationTemplate(ctx, in)
}
func (r *mutationResolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	return r.appTemplate.RegisterApplicationFromTemplate(ctx, in)
}
func (r *mutationResolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.UpdateApplicationTemplate(ctx, id, in)
}
func (r *mutationResolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.DeleteApplicationTemplate(ctx, id)
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
func (r *mutationResolver) SetAPIAuth(ctx context.Context, apiID string, runtimeID string, in graphql.AuthInput) (*graphql.APIRuntimeAuth, error) {
	return r.api.SetAPIAuth(ctx, apiID, runtimeID, in)
}
func (r *mutationResolver) DeleteAPIAuth(ctx context.Context, apiID string, runtimeID string) (*graphql.APIRuntimeAuth, error) {
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
func (r *mutationResolver) GenerateOneTimeTokenForApplication(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	return r.token.GenerateOneTimeTokenForApplication(ctx, id)
}
func (r *mutationResolver) GenerateOneTimeTokenForRuntime(ctx context.Context, id string) (*graphql.OneTimeToken, error) {
	return r.token.GenerateOneTimeTokenForRuntime(ctx, id)
}
func (r *mutationResolver) GenerateClientCredentialsForRuntime(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.GenerateClientCredentialsForRuntime(ctx, id)
}
func (r *mutationResolver) GenerateClientCredentialsForApplication(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.GenerateClientCredentialsForApplication(ctx, id)
}
func (r *mutationResolver) GenerateClientCredentialsForIntegrationSystem(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.GenerateClientCredentialsForIntegrationSystem(ctx, id)
}
func (r *mutationResolver) DeleteSystemAuthForRuntime(ctx context.Context, authID string) (*graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.RuntimeReference)
	return fn(ctx, authID)
}
func (r *mutationResolver) DeleteSystemAuthForApplication(ctx context.Context, authID string) (*graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.ApplicationReference)
	return fn(ctx, authID)
}
func (r *mutationResolver) DeleteSystemAuthForIntegrationSystem(ctx context.Context, authID string) (*graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.IntegrationSystemReference)
	return fn(ctx, authID)
}
func (r *mutationResolver) CreateIntegrationSystem(ctx context.Context, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.CreateIntegrationSystem(ctx, in)
}
func (r *mutationResolver) UpdateIntegrationSystem(ctx context.Context, id string, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.UpdateIntegrationSystem(ctx, id, in)
}
func (r *mutationResolver) DeleteIntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	return r.intSys.DeleteIntegrationSystem(ctx, id)
}

type applicationResolver struct {
	*RootResolver
}

func (r *applicationResolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.SystemAuth, error) {
	return r.app.Auths(ctx, obj)
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
func (r *applicationResolver) API(ctx context.Context, obj *graphql.Application, id string) (*graphql.APIDefinition, error) {
	return r.app.API(ctx, id, obj)
}
func (r *applicationResolver) EventAPI(ctx context.Context, obj *graphql.Application, id string) (*graphql.EventAPIDefinition, error) {
	return r.app.EventAPI(ctx, id, obj)
}
func (r *applicationResolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.app.Documents(ctx, obj, first, after)
}

func (r *applicationResolver) EventConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventConfiguration, error) {
	return r.app.EventConfiguration(ctx, obj)
}

type runtimeResolver struct {
	*RootResolver
}

func (r *runtimeResolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	return r.runtime.Labels(ctx, obj, key)
}

func (r *runtimeResolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.SystemAuth, error) {
	return r.runtime.Auths(ctx, obj)
}

type apiDefinitionResolver struct {
	*RootResolver
}

func (r *apiDefinitionResolver) Auth(ctx context.Context, obj *graphql.APIDefinition, runtimeID string) (*graphql.APIRuntimeAuth, error) {
	return r.api.Auth(ctx, obj, runtimeID)
}
func (r *apiDefinitionResolver) Auths(ctx context.Context, obj *graphql.APIDefinition) ([]*graphql.APIRuntimeAuth, error) {
	return r.api.Auths(ctx, obj)
}

type apiSpecResolver struct{ *RootResolver }

func (r *apiSpecResolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	return r.api.FetchRequest(ctx, obj)
}

type documentResolver struct{ *RootResolver }

func (r *documentResolver) FetchRequest(ctx context.Context, obj *graphql.Document) (*graphql.FetchRequest, error) {
	return r.doc.FetchRequest(ctx, obj)
}

type eventAPISpecResolver struct{ *RootResolver }

func (r *eventAPISpecResolver) FetchRequest(ctx context.Context, obj *graphql.EventAPISpec) (*graphql.FetchRequest, error) {
	return r.eventAPI.FetchRequest(ctx, obj)
}

type integrationSystemResolver struct{ *RootResolver }

func (r *integrationSystemResolver) Auths(ctx context.Context, obj *graphql.IntegrationSystem) ([]*graphql.SystemAuth, error) {
	return r.intSys.Auths(ctx, obj)
}
