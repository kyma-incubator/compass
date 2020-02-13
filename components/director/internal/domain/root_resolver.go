package domain

import (
	"context"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/viewer"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apiruntimeauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
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
	app                 *application.Resolver
	appTemplate         *apptemplate.Resolver
	api                 *api.Resolver
	eventAPI            *eventdef.Resolver
	eventing            *eventing.Resolver
	doc                 *document.Resolver
	runtime             *runtime.Resolver
	healthCheck         *healthcheck.Resolver
	webhook             *webhook.Resolver
	labelDef            *labeldef.Resolver
	token               *onetimetoken.Resolver
	systemAuth          *systemauth.Resolver
	oAuth20             *oauth20.Resolver
	intSys              *integrationsystem.Resolver
	viewer              *viewer.Resolver
	tenant              *tenant.Resolver
	mpPackage           *mp_package.Resolver
	packageInstanceAuth *packageinstanceauth.Resolver
}

func NewRootResolver(transact persistence.Transactioner, scopeCfgProvider *scope.Provider, oneTimeTokenCfg onetimetoken.Config, oAuth20Cfg oauth20.Config) *RootResolver {
	authConverter := auth.NewConverter()
	apiRtmAuthConverter := apiruntimeauth.NewConverter(authConverter)
	runtimeConverter := runtime.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	apiConverter := api.NewConverter(authConverter, frConverter, versionConverter)
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	appConverter := application.NewConverter(webhookConverter, apiConverter, eventAPIConverter, docConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	tokenConverter := onetimetoken.NewConverter(oneTimeTokenCfg.LegacyConnectorURL)
	systemAuthConverter := systemauth.NewConverter(authConverter)
	intSysConverter := integrationsystem.NewConverter()
	appTemplateConverter := apptemplate.NewConverter(appConverter)
	tenantConverter := tenant.NewConverter()

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewRepository()
	applicationRepo := application.NewRepository(appConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	apiRepo := api.NewRepository(apiConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	apiRtmAuthRepo := apiruntimeauth.NewRepository(apiRtmAuthConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)

	connectorGCLI := graphql_client.NewGraphQLClient(oneTimeTokenCfg.OneTimeTokenURL)

	uidSvc := uid.NewService()
	apiRtmAuthSvc := apiruntimeauth.NewService(apiRtmAuthRepo, uidSvc)
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewScenariosService(labelDefRepo, uidSvc)
	appSvc := application.NewService(applicationRepo, webhookRepo, apiRepo, eventAPIRepo, docRepo, runtimeRepo, labelRepo, fetchRequestRepo, intSysRepo, labelUpsertSvc, scenariosSvc, uidSvc)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, fetchRequestRepo, uidSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, fetchRequestRepo, uidSvc)
	webhookSvc := webhook.NewService(webhookRepo, uidSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelUpsertSvc, uidSvc)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, uidSvc)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	tokenSvc := onetimetoken.NewTokenService(connectorGCLI, systemAuthSvc, oneTimeTokenCfg.ConnectorURL)
	oAuth20Svc := oauth20.NewService(scopeCfgProvider, uidSvc, oAuth20Cfg)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	eventingSvc := eventing.NewService(runtimeRepo, labelRepo)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

	return &RootResolver{
		app:                 application.NewResolver(transact, appSvc, apiSvc, eventAPISvc, docSvc, webhookSvc, oAuth20Svc, systemAuthSvc, appConverter, docConverter, webhookConverter, apiConverter, eventAPIConverter, systemAuthConverter, eventingSvc),
		appTemplate:         apptemplate.NewResolver(transact, appSvc, appConverter, appTemplateSvc, appTemplateConverter),
		api:                 api.NewResolver(transact, apiSvc, appSvc, runtimeSvc, apiRtmAuthSvc, apiConverter, authConverter, frConverter, apiRtmAuthConverter),
		eventAPI:            eventdef.NewResolver(transact, eventAPISvc, appSvc, eventAPIConverter, frConverter),
		eventing:            eventing.NewResolver(transact, eventingSvc, appSvc),
		doc:                 document.NewResolver(transact, docSvc, appSvc, frConverter),
		runtime:             runtime.NewResolver(transact, runtimeSvc, systemAuthSvc, oAuth20Svc, runtimeConverter, systemAuthConverter, eventingSvc),
		healthCheck:         healthcheck.NewResolver(healthCheckSvc),
		webhook:             webhook.NewResolver(transact, webhookSvc, appSvc, webhookConverter),
		labelDef:            labeldef.NewResolver(transact, labelDefSvc, labelDefConverter),
		token:               onetimetoken.NewTokenResolver(transact, tokenSvc, tokenConverter),
		systemAuth:          systemauth.NewResolver(transact, systemAuthSvc, oAuth20Svc, systemAuthConverter),
		oAuth20:             oauth20.NewResolver(transact, oAuth20Svc, appSvc, runtimeSvc, intSysSvc, systemAuthSvc, systemAuthConverter),
		intSys:              integrationsystem.NewResolver(transact, intSysSvc, systemAuthSvc, oAuth20Svc, intSysConverter, systemAuthConverter),
		viewer:              viewer.NewViewerResolver(),
		tenant:              tenant.NewResolver(transact, tenantSvc, tenantConverter),
		mpPackage:           mp_package.NewResolver(),
		packageInstanceAuth: packageinstanceauth.NewResolver(),
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
func (r *RootResolver) EventSpec() graphql.EventSpecResolver {
	return &eventSpecResolver{r}
}

func (r *RootResolver) Package() graphql.PackageResolver {
	return &PackageResolver{r}
}

func (r *RootResolver) IntegrationSystem() graphql.IntegrationSystemResolver {
	return &integrationSystemResolver{r}
}

func (r *RootResolver) OneTimeTokenForApplication() graphql.OneTimeTokenForApplicationResolver {
	return &oneTimeTokenForApplicationResolver{r}
}

func (r *RootResolver) OneTimeTokenForRuntime() graphql.OneTimeTokenForRuntimeResolver {
	return &oneTimeTokenForRuntimeResolver{r}
}

type queryResolver struct {
	*RootResolver
}

func (r *queryResolver) Viewer(ctx context.Context) (*graphql.Viewer, error) {
	return r.viewer.Viewer(ctx)
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

func (r *queryResolver) Tenants(ctx context.Context) ([]*graphql.Tenant, error) {
	return r.tenant.Tenants(ctx)
}

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) RegisterApplication(ctx context.Context, in graphql.ApplicationRegisterInput) (*graphql.Application, error) {
	return r.app.RegisterApplication(ctx, in)
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	return r.app.UpdateApplication(ctx, id, in)
}
func (r *mutationResolver) UnregisterApplication(ctx context.Context, id string) (*graphql.Application, error) {
	return r.app.UnregisterApplication(ctx, id)
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
func (r *mutationResolver) AddAPIDefinition(ctx context.Context, applicationID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.AddAPIDefinition(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateAPIDefinition(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.UpdateAPIDefinition(ctx, id, in)
}
func (r *mutationResolver) DeleteAPIDefinition(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	return r.api.DeleteAPIDefinition(ctx, id)
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
func (r *mutationResolver) AddEventDefinition(ctx context.Context, applicationID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.AddEventDefinition(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.UpdateEventDefinition(ctx, id, in)
}
func (r *mutationResolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	return r.eventAPI.DeleteEventDefinition(ctx, id)
}
func (r *mutationResolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	return r.eventAPI.RefetchEventDefinitionSpec(ctx, eventID)
}
func (r *mutationResolver) RegisterRuntime(ctx context.Context, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	return r.runtime.RegisterRuntime(ctx, in)
}
func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeInput) (*graphql.Runtime, error) {
	return r.runtime.UpdateRuntime(ctx, id, in)
}
func (r *mutationResolver) UnregisterRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
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
func (r *mutationResolver) RequestOneTimeTokenForApplication(ctx context.Context, id string) (*graphql.OneTimeTokenForApplication, error) {
	return r.token.RequestOneTimeTokenForApplication(ctx, id)
}
func (r *mutationResolver) RequestOneTimeTokenForRuntime(ctx context.Context, id string) (*graphql.OneTimeTokenForRuntime, error) {
	return r.token.RequestOneTimeTokenForRuntime(ctx, id)
}
func (r *mutationResolver) RequestClientCredentialsForRuntime(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForRuntime(ctx, id)
}
func (r *mutationResolver) RequestClientCredentialsForApplication(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForApplication(ctx, id)
}
func (r *mutationResolver) RequestClientCredentialsForIntegrationSystem(ctx context.Context, id string) (*graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForIntegrationSystem(ctx, id)
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
func (r *mutationResolver) RegisterIntegrationSystem(ctx context.Context, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.RegisterIntegrationSystem(ctx, in)
}
func (r *mutationResolver) UpdateIntegrationSystem(ctx context.Context, id string, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.UpdateIntegrationSystem(ctx, id, in)
}
func (r *mutationResolver) UnregisterIntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	return r.intSys.UnregisterIntegrationSystem(ctx, id)
}

func (r *mutationResolver) SetDefaultEventingForApplication(ctx context.Context, appID string, runtimeID string) (*graphql.ApplicationEventingConfiguration, error) {
	return r.eventing.SetEventingForApplication(ctx, appID, runtimeID)
}

func (r *mutationResolver) DeleteDefaultEventingForApplication(ctx context.Context, appID string) (*graphql.ApplicationEventingConfiguration, error) {
	return r.eventing.UnsetEventingForApplication(ctx, appID)
}

func (r *mutationResolver) AddAPIDefinitionToPackage(ctx context.Context, packageID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.AddAPIDefinitionToPackage(ctx, packageID, in)
}
func (r *mutationResolver) AddEventDefinitionToPackage(ctx context.Context, packageID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.AddEventDefinitionToPackage(ctx, packageID, in)
}
func (r *mutationResolver) AddDocumentToPackage(ctx context.Context, packageID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return r.doc.AddDocumentToPackage(ctx, packageID, in)
}
func (r *mutationResolver) SetPackageInstanceAuth(ctx context.Context, packageID string, authID string, in graphql.PackageInstanceAuthSetInput) (*graphql.PackageInstanceAuth, error) {
	return r.packageInstanceAuth.SetPackageInstanceAuth(ctx, packageID, authID, in)
}
func (r *mutationResolver) DeletePackageInstanceAuth(ctx context.Context, packageID string, authID string) (*graphql.PackageInstanceAuth, error) {
	return r.packageInstanceAuth.DeletePackageInstanceAuth(ctx, packageID, authID)
}
func (r *mutationResolver) RequestPackageInstanceAuthCreation(ctx context.Context, packageID string, in graphql.PackageInstanceAuthRequestInput) (*graphql.PackageInstanceAuth, error) {
	return r.packageInstanceAuth.RequestPackageInstanceAuthCreation(ctx, packageID, in)
}
func (r *mutationResolver) RequestPackageInstanceAuthDeletion(ctx context.Context, packageID string, authID string) (*graphql.PackageInstanceAuth, error) {
	return r.packageInstanceAuth.RequestPackageInstanceAuthDeletion(ctx, packageID, authID)
}

func (r *mutationResolver) AddPackage(ctx context.Context, applicationID string, in graphql.PackageCreateInput) (*graphql.Package, error) {
	return r.mpPackage.AddPackage(ctx, applicationID, in)
}
func (r *mutationResolver) UpdatePackage(ctx context.Context, id string, in graphql.PackageUpdateInput) (*graphql.Package, error) {
	return r.mpPackage.UpdatePackage(ctx, id, in)
}
func (r *mutationResolver) DeletePackage(ctx context.Context, id string) (*graphql.Package, error) {
	return r.mpPackage.DeletePackage(ctx, id)
}

type applicationResolver struct {
	*RootResolver
}

func (r *applicationResolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.SystemAuth, error) {
	return r.app.Auths(ctx, obj)
}

func (r *applicationResolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (*graphql.Labels, error) {
	return r.app.Labels(ctx, obj, key)
}
func (r *applicationResolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	return r.app.Webhooks(ctx, obj)
}
func (r *applicationResolver) APIDefinitions(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return r.app.ApiDefinitions(ctx, obj, group, first, after)
}
func (r *applicationResolver) EventDefinitions(ctx context.Context, obj *graphql.Application, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return r.app.EventDefinitions(ctx, obj, group, first, after)
}
func (r *applicationResolver) APIDefinition(ctx context.Context, obj *graphql.Application, id string) (*graphql.APIDefinition, error) {
	return r.app.APIDefinition(ctx, id, obj)
}
func (r *applicationResolver) EventDefinition(ctx context.Context, obj *graphql.Application, id string) (*graphql.EventDefinition, error) {
	return r.app.EventDefinition(ctx, id, obj)
}
func (r *applicationResolver) Documents(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.app.Documents(ctx, obj, first, after)
}
func (r *applicationResolver) EventingConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventingConfiguration, error) {
	return r.app.EventingConfiguration(ctx, obj)
}
func (r *applicationResolver) Packages(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.PackagePage, error) {
	return r.app.Packages(ctx, obj, first, after)
}
func (r *applicationResolver) Package(ctx context.Context, obj *graphql.Application, id string) (*graphql.Package, error) {
	return r.app.Package(ctx, obj, id)
}

type runtimeResolver struct {
	*RootResolver
}

func (r *runtimeResolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (*graphql.Labels, error) {
	return r.runtime.Labels(ctx, obj, key)
}

func (r *runtimeResolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.SystemAuth, error) {
	return r.runtime.Auths(ctx, obj)
}

func (r *runtimeResolver) EventingConfiguration(ctx context.Context, obj *graphql.Runtime) (*graphql.RuntimeEventingConfiguration, error) {
	return r.runtime.EventingConfiguration(ctx, obj)
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

type eventSpecResolver struct{ *RootResolver }

func (r *eventSpecResolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	return r.eventAPI.FetchRequest(ctx, obj)
}

type integrationSystemResolver struct{ *RootResolver }

func (r *integrationSystemResolver) Auths(ctx context.Context, obj *graphql.IntegrationSystem) ([]*graphql.SystemAuth, error) {
	return r.intSys.Auths(ctx, obj)
}

type oneTimeTokenForApplicationResolver struct{ *RootResolver }

func (r *oneTimeTokenForApplicationResolver) RawEncoded(ctx context.Context, obj *graphql.OneTimeTokenForApplication) (*string, error) {
	return r.token.RawEncoded(ctx, &obj.TokenWithURL)
}

func (r *oneTimeTokenForApplicationResolver) Raw(ctx context.Context, obj *graphql.OneTimeTokenForApplication) (*string, error) {
	return r.token.Raw(ctx, &obj.TokenWithURL)
}

type oneTimeTokenForRuntimeResolver struct{ *RootResolver }

func (r *oneTimeTokenForRuntimeResolver) RawEncoded(ctx context.Context, obj *graphql.OneTimeTokenForRuntime) (*string, error) {
	return r.token.RawEncoded(ctx, &obj.TokenWithURL)
}

func (r *oneTimeTokenForRuntimeResolver) Raw(ctx context.Context, obj *graphql.OneTimeTokenForRuntime) (*string, error) {
	return r.token.Raw(ctx, &obj.TokenWithURL)
}

type PackageResolver struct{ *RootResolver }

func (r *PackageResolver) InstanceAuth(ctx context.Context, obj *graphql.Package, id string) (*graphql.PackageInstanceAuth, error) {
	return r.mpPackage.InstanceAuth(ctx, obj, id)
}
func (r *PackageResolver) InstanceAuths(ctx context.Context, obj *graphql.Package) ([]*graphql.PackageInstanceAuth, error) {
	return r.mpPackage.InstanceAuths(ctx, obj)
}
func (r *PackageResolver) APIDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return r.mpPackage.APIDefinitions(ctx, obj, group, first, after)
}
func (r *PackageResolver) EventDefinitions(ctx context.Context, obj *graphql.Package, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return r.mpPackage.EventDefinitions(ctx, obj, group, first, after)
}
func (r *PackageResolver) Documents(ctx context.Context, obj *graphql.Package, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.mpPackage.Documents(ctx, obj, first, after)
}
func (r *PackageResolver) APIDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.APIDefinition, error) {
	return r.mpPackage.APIDefinition(ctx, obj, id)
}
func (r *PackageResolver) EventDefinition(ctx context.Context, obj *graphql.Package, id string) (*graphql.EventDefinition, error) {
	return r.mpPackage.EventDefinition(ctx, obj, id)
}
func (r *PackageResolver) Document(ctx context.Context, obj *graphql.Package, id string) (*graphql.Document, error) {
	return r.mpPackage.Document(ctx, obj, id)
}
