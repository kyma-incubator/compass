package domain

import (
	"context"
	"net/http"
	"net/url"

	httptransport "github.com/go-openapi/runtime/client"
	hydraClient "github.com/ory/hydra-client-go/client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/healthcheck"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/viewer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/time"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	appNameNormalizer  normalizer.Normalizator
	app                *application.Resolver
	appTemplate        *apptemplate.Resolver
	api                *api.Resolver
	eventAPI           *eventdef.Resolver
	eventing           *eventing.Resolver
	doc                *document.Resolver
	runtime            *runtime.Resolver
	runtimeContext     *runtime_context.Resolver
	healthCheck        *healthcheck.Resolver
	webhook            *webhook.Resolver
	labelDef           *labeldef.Resolver
	token              *onetimetoken.Resolver
	systemAuth         *systemauth.Resolver
	oAuth20            *oauth20.Resolver
	intSys             *integrationsystem.Resolver
	viewer             *viewer.Resolver
	tenant             *tenant.Resolver
	mpBundle           *bundleutil.Resolver
	bundleInstanceAuth *bundleinstanceauth.Resolver
	scenarioAssignment *scenarioassignment.Resolver
}

func NewRootResolver(
	appNameNormalizer normalizer.Normalizator,
	transact persistence.Transactioner,
	cfgProvider *configprovider.Provider,
	oneTimeTokenCfg onetimetoken.Config,
	oAuth20Cfg oauth20.Config,
	pairingAdaptersMapping map[string]string,
	featuresConfig features.Config,
	metricsCollector *metrics.Collector,
	httpClient *http.Client,
	tokenLength int,
	hydraURL *url.URL,
) *RootResolver {
	oAuth20HTTPClient := &http.Client{
		Timeout:   oAuth20Cfg.HTTPClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}

	transport := httptransport.NewWithClient(hydraURL.Host, hydraURL.Path, []string{hydraURL.Scheme}, oAuth20HTTPClient)
	hydra := hydraClient.New(transport, nil)

	metricsCollector.InstrumentOAuth20HTTPClient(oAuth20HTTPClient)

	authConverter := auth.NewConverter()
	runtimeConverter := runtime.NewConverter()
	runtimeContextConverter := runtime_context.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	tokenConverter := onetimetoken.NewConverter(oneTimeTokenCfg.LegacyConnectorURL)
	systemAuthConverter := systemauth.NewConverter(authConverter)
	intSysConverter := integrationsystem.NewConverter()
	tenantConverter := tenant.NewConverter()
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)
	assignmentConv := scenarioassignment.NewConverter()

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	runtimeContextRepo := runtime_context.NewRepository()
	applicationRepo := application.NewRepository(appConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	apiRepo := api.NewRepository(apiConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	specRepo := spec.NewRepository(specConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	bundleRepo := bundleutil.NewRepository(bundleConverter)
	bundleInstanceAuthRepo := bundleinstanceauth.NewRepository(bundleInstanceAuthConv)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)

	uidSvc := uid.NewService()
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewScenariosService(labelDefRepo, uidSvc, featuresConfig.DefaultScenarioEnabled)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc)

	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	scenarioAssignmentEngine := scenarioassignment.NewEngine(labelUpsertSvc, labelRepo, scenarioAssignmentRepo)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc, scenarioAssignmentEngine)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelUpsertSvc, uidSvc, scenarioAssignmentEngine, featuresConfig.ProtectedLabelPattern)
	runtimeCtxSvc := runtime_context.NewService(runtimeContextRepo, labelRepo, labelUpsertSvc, uidSvc)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, scenariosSvc, uidSvc)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)
	oAuth20Svc := oauth20.NewService(cfgProvider, uidSvc, oAuth20Cfg.PublicAccessTokenEndpoint, hydra.Admin)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	eventingSvc := eventing.NewService(appNameNormalizer, runtimeRepo, labelRepo)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	appSvc := application.NewService(appNameNormalizer, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelUpsertSvc, scenariosSvc, bundleSvc, uidSvc)
	timeService := time.NewService()
	tokenSvc := onetimetoken.NewTokenService(systemAuthSvc, appSvc, appConverter, tenantSvc, httpClient, onetimetoken.NewTokenGenerator(tokenLength), oneTimeTokenCfg.ConnectorURL, pairingAdaptersMapping, timeService)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)

	return &RootResolver{
		appNameNormalizer:  appNameNormalizer,
		app:                application.NewResolver(transact, appSvc, webhookSvc, oAuth20Svc, systemAuthSvc, appConverter, webhookConverter, systemAuthConverter, eventingSvc, bundleSvc, bundleConverter),
		appTemplate:        apptemplate.NewResolver(transact, appSvc, appConverter, appTemplateSvc, appTemplateConverter, webhookSvc, webhookConverter),
		api:                api.NewResolver(transact, apiSvc, runtimeSvc, bundleSvc, apiConverter, frConverter, specSvc, specConverter),
		eventAPI:           eventdef.NewResolver(transact, eventAPISvc, bundleSvc, eventAPIConverter, frConverter, specSvc, specConverter),
		eventing:           eventing.NewResolver(transact, eventingSvc, appSvc),
		doc:                document.NewResolver(transact, docSvc, appSvc, bundleSvc, frConverter),
		runtime:            runtime.NewResolver(transact, runtimeSvc, scenarioAssignmentSvc, systemAuthSvc, oAuth20Svc, runtimeConverter, systemAuthConverter, eventingSvc, bundleInstanceAuthSvc),
		runtimeContext:     runtime_context.NewResolver(transact, runtimeCtxSvc, runtimeContextConverter),
		healthCheck:        healthcheck.NewResolver(healthCheckSvc),
		webhook:            webhook.NewResolver(transact, webhookSvc, appSvc, appTemplateSvc, webhookConverter),
		labelDef:           labeldef.NewResolver(transact, labelDefSvc, labelDefConverter),
		token:              onetimetoken.NewTokenResolver(transact, tokenSvc, tokenConverter),
		systemAuth:         systemauth.NewResolver(transact, systemAuthSvc, oAuth20Svc, systemAuthConverter),
		oAuth20:            oauth20.NewResolver(transact, oAuth20Svc, appSvc, runtimeSvc, intSysSvc, systemAuthSvc, systemAuthConverter),
		intSys:             integrationsystem.NewResolver(transact, intSysSvc, systemAuthSvc, oAuth20Svc, intSysConverter, systemAuthConverter),
		viewer:             viewer.NewViewerResolver(),
		tenant:             tenant.NewResolver(transact, tenantSvc, tenantConverter),
		mpBundle:           bundleutil.NewResolver(transact, bundleSvc, bundleInstanceAuthSvc, apiSvc, eventAPISvc, docSvc, bundleConverter, bundleInstanceAuthConv, apiConverter, eventAPIConverter, docConverter, specSvc),
		bundleInstanceAuth: bundleinstanceauth.NewResolver(transact, bundleInstanceAuthSvc, bundleSvc, bundleInstanceAuthConv, bundleConverter),
		scenarioAssignment: scenarioassignment.NewResolver(transact, scenarioAssignmentSvc, assignmentConv),
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

func (r *RootResolver) ApplicationTemplate() graphql.ApplicationTemplateResolver {
	return &applicationTemplateResolver{r}
}

func (r *RootResolver) Runtime() graphql.RuntimeResolver {
	return &runtimeResolver{r}
}
func (r *RootResolver) RuntimeContext() graphql.RuntimeContextResolver {
	return &runtimeContextResolver{r}
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

func (r *RootResolver) Bundle() graphql.BundleResolver {
	return &BundleResolver{r}
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
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if consumerInfo.ConsumerType == consumer.Runtime {
		log.C(ctx).Debugf("Consumer type is of type %v. Filtering response based on scenarios...", consumer.Runtime)
		return r.app.ApplicationsForRuntime(ctx, consumerInfo.ConsumerID, first, after)
	}

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
	apps, err := r.app.ApplicationsForRuntime(ctx, runtimeID, first, after)
	if err != nil {
		return nil, err
	}

	labels, err := r.runtime.GetLabel(ctx, runtimeID, runtime.IsNormalizedLabel)
	if err != nil {
		return nil, err
	}

	shouldNormalize := true
	if labels != nil {
		labelsMap := (map[string]interface{})(*labels)
		shouldNormalize = labelsMap[runtime.IsNormalizedLabel] == nil || labelsMap[runtime.IsNormalizedLabel] == "true"
	}

	if shouldNormalize {
		for i := range apps.Data {
			apps.Data[i].Name = r.appNameNormalizer.Normalize(apps.Data[i].Name)
		}
	}

	return apps, nil
}
func (r *queryResolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	return r.runtime.Runtimes(ctx, filter, first, after)
}
func (r *queryResolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.Runtime(ctx, id)
}
func (r *queryResolver) RuntimeContexts(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimeContextPage, error) {
	return r.runtimeContext.RuntimeContexts(ctx, filter, first, after)
}
func (r *queryResolver) RuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.RuntimeContext(ctx, id)
}
func (r *queryResolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinitions(ctx)
}
func (r *queryResolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinition(ctx, key)
}
func (r *queryResolver) BundleByInstanceAuth(ctx context.Context, authID string) (*graphql.Bundle, error) {
	return r.bundleInstanceAuth.BundleByInstanceAuth(ctx, authID)
}
func (r *queryResolver) BundleInstanceAuth(ctx context.Context, id string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.BundleInstanceAuth(ctx, id)
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

func (r *queryResolver) AutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.GetAutomaticScenarioAssignmentForScenarioName(ctx, scenarioName)
}

func (r *queryResolver) AutomaticScenarioAssignmentsForSelector(ctx context.Context, selector graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.AutomaticScenarioAssignmentsForSelector(ctx, selector)
}

func (r *queryResolver) AutomaticScenarioAssignments(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.AutomaticScenarioAssignmentPage, error) {
	return r.scenarioAssignment.AutomaticScenarioAssignments(ctx, first, after)
}

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) RegisterApplication(ctx context.Context, in graphql.ApplicationRegisterInput, _ *graphql.OperationMode) (*graphql.Application, error) {
	return r.app.RegisterApplication(ctx, in)
}
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	return r.app.UpdateApplication(ctx, id, in)
}
func (r *mutationResolver) UnregisterApplication(ctx context.Context, id string, _ *graphql.OperationMode) (*graphql.Application, error) {
	return r.app.UnregisterApplication(ctx, id)
}
func (r *mutationResolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.CreateApplicationTemplate(ctx, in)
}
func (r *mutationResolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	return r.appTemplate.RegisterApplicationFromTemplate(ctx, in)
}
func (r *mutationResolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateUpdateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.UpdateApplicationTemplate(ctx, id, in)
}
func (r *mutationResolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.DeleteApplicationTemplate(ctx, id)
}
func (r *mutationResolver) AddWebhook(ctx context.Context, applicationID *string, applicationTemplateID *string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.AddWebhook(ctx, applicationID, applicationTemplateID, in)
}
func (r *mutationResolver) UpdateWebhook(ctx context.Context, webhookID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.UpdateWebhook(ctx, webhookID, in)
}
func (r *mutationResolver) DeleteWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
	return r.webhook.DeleteWebhook(ctx, webhookID)
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
func (r *mutationResolver) RegisterRuntimeContext(ctx context.Context, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.RegisterRuntimeContext(ctx, in)
}
func (r *mutationResolver) UpdateRuntimeContext(ctx context.Context, id string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.UpdateRuntimeContext(ctx, id, in)
}
func (r *mutationResolver) UnregisterRuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.DeleteRuntimeContext(ctx, id)
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
func (r *mutationResolver) RequestClientCredentialsForRuntime(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForRuntime(ctx, id)
}
func (r *mutationResolver) RequestClientCredentialsForApplication(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForApplication(ctx, id)
}
func (r *mutationResolver) RequestClientCredentialsForIntegrationSystem(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForIntegrationSystem(ctx, id)
}
func (r *mutationResolver) DeleteSystemAuthForRuntime(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.RuntimeReference)
	return fn(ctx, authID)
}
func (r *mutationResolver) DeleteSystemAuthForApplication(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.ApplicationReference)
	return fn(ctx, authID)
}
func (r *mutationResolver) DeleteSystemAuthForIntegrationSystem(ctx context.Context, authID string) (graphql.SystemAuth, error) {
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

func (r *mutationResolver) AddAPIDefinitionToBundle(ctx context.Context, bundleID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.AddAPIDefinitionToBundle(ctx, bundleID, in)
}
func (r *mutationResolver) AddEventDefinitionToBundle(ctx context.Context, bundleID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.AddEventDefinitionToBundle(ctx, bundleID, in)
}
func (r *mutationResolver) AddDocumentToBundle(ctx context.Context, bundleID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return r.doc.AddDocumentToBundle(ctx, bundleID, in)
}
func (r *mutationResolver) SetBundleInstanceAuth(ctx context.Context, authID string, in graphql.BundleInstanceAuthSetInput) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.SetBundleInstanceAuth(ctx, authID, in)
}
func (r *mutationResolver) DeleteBundleInstanceAuth(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.DeleteBundleInstanceAuth(ctx, authID)
}
func (r *mutationResolver) RequestBundleInstanceAuthCreation(ctx context.Context, bundleID string, in graphql.BundleInstanceAuthRequestInput) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.RequestBundleInstanceAuthCreation(ctx, bundleID, in)
}
func (r *mutationResolver) RequestBundleInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.RequestBundleInstanceAuthDeletion(ctx, authID)
}

func (r *mutationResolver) AddBundle(ctx context.Context, applicationID string, in graphql.BundleCreateInput) (*graphql.Bundle, error) {
	return r.mpBundle.AddBundle(ctx, applicationID, in)
}
func (r *mutationResolver) UpdateBundle(ctx context.Context, id string, in graphql.BundleUpdateInput) (*graphql.Bundle, error) {
	return r.mpBundle.UpdateBundle(ctx, id, in)
}
func (r *mutationResolver) DeleteBundle(ctx context.Context, id string) (*graphql.Bundle, error) {
	return r.mpBundle.DeleteBundle(ctx, id)
}

func (r *mutationResolver) DeleteAutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.DeleteAutomaticScenarioAssignmentForScenario(ctx, scenarioName)
}

func (r *mutationResolver) DeleteAutomaticScenarioAssignmentsForSelector(ctx context.Context, selector graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.DeleteAutomaticScenarioAssignmentsForSelector(ctx, selector)
}
func (r *mutationResolver) CreateAutomaticScenarioAssignment(ctx context.Context, in graphql.AutomaticScenarioAssignmentSetInput) (*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.CreateAutomaticScenarioAssignment(ctx, in)
}

type applicationResolver struct {
	*RootResolver
}

func (r *applicationResolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.AppSystemAuth, error) {
	return r.app.Auths(ctx, obj)
}

func (r *applicationResolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (graphql.Labels, error) {
	return r.app.Labels(ctx, obj, key)
}
func (r *applicationResolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	return r.app.Webhooks(ctx, obj)
}
func (r *applicationResolver) EventingConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventingConfiguration, error) {
	return r.app.EventingConfiguration(ctx, obj)
}
func (r *applicationResolver) Bundles(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.BundlePage, error) {
	return r.app.Bundles(ctx, obj, first, after)
}
func (r *applicationResolver) Bundle(ctx context.Context, obj *graphql.Application, id string) (*graphql.Bundle, error) {
	return r.app.Bundle(ctx, obj, id)
}

type applicationTemplateResolver struct {
	*RootResolver
}

func (r applicationTemplateResolver) Webhooks(ctx context.Context, obj *graphql.ApplicationTemplate) ([]*graphql.Webhook, error) {
	return r.appTemplate.Webhooks(ctx, obj)
}

type runtimeResolver struct {
	*RootResolver
}

func (r *runtimeResolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	return r.runtime.Labels(ctx, obj, key)
}

func (r *runtimeResolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.RuntimeSystemAuth, error) {
	return r.runtime.Auths(ctx, obj)
}

func (r *runtimeResolver) EventingConfiguration(ctx context.Context, obj *graphql.Runtime) (*graphql.RuntimeEventingConfiguration, error) {
	return r.runtime.EventingConfiguration(ctx, obj)
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

func (r *integrationSystemResolver) Auths(ctx context.Context, obj *graphql.IntegrationSystem) ([]*graphql.IntSysSystemAuth, error) {
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

type runtimeContextResolver struct {
	*RootResolver
}

func (r *runtimeContextResolver) Labels(ctx context.Context, obj *graphql.RuntimeContext, key *string) (graphql.Labels, error) {
	return r.runtimeContext.Labels(ctx, obj, key)
}

type BundleResolver struct{ *RootResolver }

func (r *BundleResolver) InstanceAuth(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.BundleInstanceAuth, error) {
	return r.mpBundle.InstanceAuth(ctx, obj, id)
}
func (r *BundleResolver) InstanceAuths(ctx context.Context, obj *graphql.Bundle) ([]*graphql.BundleInstanceAuth, error) {
	return r.mpBundle.InstanceAuths(ctx, obj)
}
func (r *BundleResolver) APIDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return r.mpBundle.APIDefinitions(ctx, obj, group, first, after)
}
func (r *BundleResolver) EventDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return r.mpBundle.EventDefinitions(ctx, obj, group, first, after)
}
func (r *BundleResolver) Documents(ctx context.Context, obj *graphql.Bundle, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.mpBundle.Documents(ctx, obj, first, after)
}
func (r *BundleResolver) APIDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.APIDefinition, error) {
	return r.mpBundle.APIDefinition(ctx, obj, id)
}
func (r *BundleResolver) EventDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.EventDefinition, error) {
	return r.mpBundle.EventDefinition(ctx, obj, id)
}
func (r *BundleResolver) Document(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.Document, error) {
	return r.mpBundle.Document(ctx, obj, id)
}
