package domain

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"

	pkgadapters "github.com/kyma-incubator/compass/components/director/pkg/adapters"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	httptransport "github.com/go-openapi/runtime/client"
	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventing"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/healthcheck"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/viewer"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/time"
	hydraClient "github.com/ory/hydra-client-go/client"
)

var _ graphql.ResolverRoot = &RootResolver{}

// RootResolver missing godoc
type RootResolver struct {
	appNameNormalizer   normalizer.Normalizator
	app                 *application.Resolver
	appTemplate         *apptemplate.Resolver
	api                 *api.Resolver
	eventAPI            *eventdef.Resolver
	eventing            *eventing.Resolver
	doc                 *document.Resolver
	formation           *formation.Resolver
	runtime             *runtime.Resolver
	runtimeContext      *runtimectx.Resolver
	healthCheck         *healthcheck.Resolver
	webhook             *webhook.Resolver
	labelDef            *labeldef.Resolver
	token               *onetimetoken.Resolver
	systemAuth          *systemauth.Resolver
	oAuth20             *oauth20.Resolver
	intSys              *integrationsystem.Resolver
	viewer              *viewer.Resolver
	tenant              *tenant.Resolver
	mpBundle            *bundleutil.Resolver
	bundleInstanceAuth  *bundleinstanceauth.Resolver
	scenarioAssignment  *scenarioassignment.Resolver
	subscription        *subscription.Resolver
	formationTemplate   *formationtemplate.Resolver
	formationConstraint *formationconstraint.Resolver
	constraintReference *formationtemplateconstraintreferences.Resolver
	certSubjectMapping  *certsubjectmapping.Resolver
}

// NewRootResolver missing godoc
func NewRootResolver(
	appNameNormalizer normalizer.Normalizator,
	transact persistence.Transactioner,
	cfgProvider *config.Provider,
	oneTimeTokenCfg onetimetoken.Config,
	oAuth20Cfg oauth20.Config,
	pairingAdapters *pkgadapters.Adapters,
	featuresConfig features.Config,
	metricsCollector *metrics.Collector,
	retryHTTPExecutor *retry.HTTPExecutor,
	httpClient, internalFQDNHTTPClient, internalGatewayHTTPClient, securedHTTPClient, mtlsHTTPClient, extSvcMtlsClient *http.Client,
	selfRegConfig config.SelfRegConfig,
	tokenLength int,
	hydraURL *url.URL,
	accessStrategyExecutorProvider *accessstrategy.Provider,
	subscriptionConfig subscription.Config,
	tenantOnDemandAPIConfig tenant.FetchOnDemandAPIConfig,
	ordWebhookMappings []application.ORDWebhookMapping,
	tenantMappingConfig map[string]interface{},
	callbackURL string,
	appTemplateProductLabel string,
) (*RootResolver, error) {
	timeService := time.NewService()

	oAuth20HTTPClient := &http.Client{
		Timeout:   oAuth20Cfg.HTTPClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)))),
	}

	transport := httptransport.NewWithClient(hydraURL.Host, hydraURL.Path, []string{hydraURL.Scheme}, oAuth20HTTPClient)
	hydra := hydraClient.New(transport, nil)

	metricsCollector.InstrumentOAuth20HTTPClient(oAuth20HTTPClient)

	tokenConverter := onetimetoken.NewConverter(oneTimeTokenCfg.LegacyConnectorURL)
	authConverter := auth.NewConverterWithOTT(tokenConverter)
	runtimeContextConverter := runtimectx.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	intSysConverter := integrationsystem.NewConverter()
	tenantConverter := tenant.NewConverter()
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)
	assignmentConv := scenarioassignment.NewConverter()
	bundleReferenceConv := bundlereferences.NewConverter()
	formationConv := formation.NewConverter()
	runtimeConverter := runtime.NewConverter(webhookConverter)
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationAssignmentConv := formationassignment.NewConverter()
	formationConstraintConverter := formationconstraint.NewConverter()
	appTemplateConv := apptemplate.NewConverter(appConverter, webhookConverter)
	constraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	certSubjectMappingConv := certsubjectmapping.NewConverter()

	healthcheckRepo := healthcheck.NewRepository()
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimectx.NewConverter())
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
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	constraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(constraintReferencesConverter)
	certSubjectMappingRepo := certsubjectmapping.NewRepository(certSubjectMappingConv)

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewServiceWithRetry(fetchRequestRepo, httpClient, accessStrategyExecutorProvider, retryHTTPExecutor)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, labelDefSvc)
	healthCheckSvc := healthcheck.NewService(healthcheckRepo)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	oAuth20Svc := oauth20.NewService(cfgProvider, uidSvc, oAuth20Cfg.PublicAccessTokenEndpoint, hydra.Admin)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	eventingSvc := eventing.NewService(appNameNormalizer, runtimeRepo, labelRepo)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsClient)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, constraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, formationRepo, labelRepo, labelSvc, applicationRepo)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, featuresConfig.RuntimeTypeLabelKey, featuresConfig.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, labelSvc, constraintEngine, formationRepo, formationTemplateRepo, featuresConfig.RuntimeTypeLabelKey, featuresConfig.ApplicationTypeLabelKey)
	formationSvc := formation.NewService(transact, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, featuresConfig.RuntimeTypeLabelKey, featuresConfig.ApplicationTypeLabelKey)
	appSvc := application.NewService(appNameNormalizer, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, selfRegConfig.SelfRegisterDistinguishLabelKey, ordWebhookMappings)
	runtimeContextSvc := runtimectx.NewService(runtimeContextRepo, labelRepo, runtimeRepo, labelSvc, formationSvc, tenantSvc, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, labelSvc, uidSvc, formationSvc, tenantSvc, webhookSvc, runtimeContextSvc, featuresConfig.ProtectedLabelPattern, featuresConfig.ImmutableLabelPattern, featuresConfig.RuntimeTypeLabelKey, featuresConfig.KymaRuntimeTypeLabelValue, featuresConfig.KymaApplicationNamespaceValue)
	tokenSvc := onetimetoken.NewTokenService(systemAuthSvc, appSvc, appConverter, tenantSvc, internalFQDNHTTPClient, onetimetoken.NewTokenGenerator(tokenLength), oneTimeTokenCfg, pairingAdapters, timeService)
	subscriptionSvc := subscription.NewService(runtimeSvc, runtimeContextSvc, tenantSvc, labelSvc, appTemplateSvc, appConverter, appTemplateConv, appSvc, uidSvc, subscriptionConfig.ConsumerSubaccountLabelKey, subscriptionConfig.SubscriptionLabelKey, subscriptionConfig.RuntimeTypeLabelKey, subscriptionConfig.ProviderLabelKey)
	tenantOnDemandSvc := tenant.NewFetchOnDemandService(internalGatewayHTTPClient, tenantOnDemandAPIConfig)
	formationTemplateSvc := formationtemplate.NewService(formationTemplateRepo, uidSvc, formationTemplateConverter, tenantSvc, webhookRepo, webhookSvc)
	constraintReferenceSvc := formationtemplateconstraintreferences.NewService(constraintReferencesRepo, constraintReferencesConverter)
	certSubjectMappingSvc := certsubjectmapping.NewService(certSubjectMappingRepo)

	selfRegisterManager, err := selfregmanager.NewSelfRegisterManager(selfRegConfig, &selfregmanager.CallerProvider{})
	if err != nil {
		return nil, err
	}

	return &RootResolver{
		appNameNormalizer:   appNameNormalizer,
		app:                 application.NewResolver(transact, appSvc, webhookSvc, oAuth20Svc, systemAuthSvc, appConverter, webhookConverter, systemAuthConverter, eventingSvc, bundleSvc, bundleConverter, appTemplateSvc, appTemplateConverter, selfRegConfig.SelfRegisterDistinguishLabelKey, featuresConfig.TokenPrefix),
		appTemplate:         apptemplate.NewResolver(transact, appSvc, appConverter, appTemplateSvc, appTemplateConverter, webhookSvc, webhookConverter, selfRegisterManager, uidSvc, tenantMappingConfig, callbackURL, appTemplateProductLabel),
		api:                 api.NewResolver(transact, apiSvc, runtimeSvc, bundleSvc, bundleReferenceSvc, apiConverter, frConverter, specSvc, specConverter, appSvc),
		eventAPI:            eventdef.NewResolver(transact, eventAPISvc, bundleSvc, bundleReferenceSvc, eventAPIConverter, frConverter, specSvc, specConverter),
		eventing:            eventing.NewResolver(transact, eventingSvc, appSvc),
		doc:                 document.NewResolver(transact, docSvc, appSvc, bundleSvc, frConverter),
		formation:           formation.NewResolver(transact, formationSvc, formationConv, formationAssignmentSvc, formationAssignmentConv, tenantOnDemandSvc),
		runtime:             runtime.NewResolver(transact, runtimeSvc, scenarioAssignmentSvc, systemAuthSvc, oAuth20Svc, runtimeConverter, systemAuthConverter, eventingSvc, bundleInstanceAuthSvc, selfRegisterManager, uidSvc, subscriptionSvc, runtimeContextSvc, runtimeContextConverter, webhookSvc, webhookConverter, tenantOnDemandSvc, formationSvc),
		runtimeContext:      runtimectx.NewResolver(transact, runtimeContextSvc, runtimeContextConverter),
		healthCheck:         healthcheck.NewResolver(healthCheckSvc),
		webhook:             webhook.NewResolver(transact, webhookSvc, appSvc, appTemplateSvc, runtimeSvc, formationTemplateSvc, webhookConverter),
		labelDef:            labeldef.NewResolver(transact, labelDefSvc, formationSvc, labelDefConverter),
		token:               onetimetoken.NewTokenResolver(transact, tokenSvc, tokenConverter, oneTimeTokenCfg.SuggestTokenHeaderKey),
		systemAuth:          systemauth.NewResolver(transact, systemAuthSvc, oAuth20Svc, tokenSvc, systemAuthConverter, authConverter),
		oAuth20:             oauth20.NewResolver(transact, oAuth20Svc, appSvc, runtimeSvc, intSysSvc, systemAuthSvc, systemAuthConverter),
		intSys:              integrationsystem.NewResolver(transact, intSysSvc, systemAuthSvc, oAuth20Svc, intSysConverter, systemAuthConverter),
		viewer:              viewer.NewViewerResolver(),
		tenant:              tenant.NewResolver(transact, tenantSvc, tenantConverter, tenantOnDemandSvc),
		mpBundle:            bundleutil.NewResolver(transact, bundleSvc, bundleInstanceAuthSvc, bundleReferenceSvc, apiSvc, eventAPISvc, docSvc, bundleConverter, bundleInstanceAuthConv, apiConverter, eventAPIConverter, docConverter, specSvc, appSvc),
		bundleInstanceAuth:  bundleinstanceauth.NewResolver(transact, bundleInstanceAuthSvc, bundleSvc, bundleInstanceAuthConv, bundleConverter),
		scenarioAssignment:  scenarioassignment.NewResolver(transact, scenarioAssignmentSvc, assignmentConv, tenantSvc),
		subscription:        subscription.NewResolver(transact, subscriptionSvc),
		formationTemplate:   formationtemplate.NewResolver(transact, formationTemplateConverter, formationTemplateSvc, webhookConverter, formationConstraintSvc, formationConstraintConverter),
		formationConstraint: formationconstraint.NewResolver(transact, formationConstraintConverter, formationConstraintSvc),
		constraintReference: formationtemplateconstraintreferences.NewResolver(transact, constraintReferencesConverter, constraintReferenceSvc),
		certSubjectMapping:  certsubjectmapping.NewResolver(transact, certSubjectMappingConv, certSubjectMappingSvc, uidSvc),
	}, nil
}

// BundlesDataloader missing godoc
func (r *RootResolver) BundlesDataloader(ids []dataloader.ParamBundle) ([]*graphql.BundlePage, []error) {
	return r.app.BundlesDataLoader(ids)
}

// APIDefinitionsDataloader missing godoc
func (r *RootResolver) APIDefinitionsDataloader(ids []dataloader.ParamAPIDef) ([]*graphql.APIDefinitionPage, []error) {
	return r.mpBundle.APIDefinitionsDataLoader(ids)
}

// EventDefinitionsDataloader missing godoc
func (r *RootResolver) EventDefinitionsDataloader(ids []dataloader.ParamEventDef) ([]*graphql.EventDefinitionPage, []error) {
	return r.mpBundle.EventDefinitionsDataLoader(ids)
}

// DocumentsDataloader missing godoc
func (r *RootResolver) DocumentsDataloader(ids []dataloader.ParamDocument) ([]*graphql.DocumentPage, []error) {
	return r.mpBundle.DocumentsDataLoader(ids)
}

// FetchRequestAPIDefDataloader missing godoc
func (r *RootResolver) FetchRequestAPIDefDataloader(ids []dataloader.ParamFetchRequestAPIDef) ([]*graphql.FetchRequest, []error) {
	return r.api.FetchRequestAPIDefDataLoader(ids)
}

// FetchRequestEventDefDataloader missing godoc
func (r *RootResolver) FetchRequestEventDefDataloader(ids []dataloader.ParamFetchRequestEventDef) ([]*graphql.FetchRequest, []error) {
	return r.eventAPI.FetchRequestEventDefDataLoader(ids)
}

// FetchRequestDocumentDataloader missing godoc
func (r *RootResolver) FetchRequestDocumentDataloader(ids []dataloader.ParamFetchRequestDocument) ([]*graphql.FetchRequest, []error) {
	return r.doc.FetchRequestDocumentDataLoader(ids)
}

// RuntimeContextsDataloader missing godoc
func (r *RootResolver) RuntimeContextsDataloader(ids []dataloader.ParamRuntimeContext) ([]*graphql.RuntimeContextPage, []error) {
	return r.runtime.RuntimeContextsDataLoader(ids)
}

// FormationAssignmentsDataLoader missing godoc
func (r *RootResolver) FormationAssignmentsDataLoader(ids []dataloader.ParamFormationAssignment) ([]*graphql.FormationAssignmentPage, []error) {
	return r.formation.FormationAssignmentsDataLoader(ids)
}

// StatusDataLoader is the FormationStatus dataloader used in the graphql API router
func (r *RootResolver) StatusDataLoader(ids []dataloader.ParamFormationStatus) ([]*graphql.FormationStatus, []error) {
	return r.formation.StatusDataLoader(ids)
}

// FormationConstraintsDataLoader is the FormationConstraint dataloader used in the graphql API router
func (r *RootResolver) FormationConstraintsDataLoader(ids []dataloader.ParamFormationConstraint) ([][]*graphql.FormationConstraint, []error) {
	return r.formationTemplate.FormationConstraintsDataLoader(ids)
}

// Mutation missing godoc
func (r *RootResolver) Mutation() graphql.MutationResolver {
	return &mutationResolver{r}
}

// Query missing godoc
func (r *RootResolver) Query() graphql.QueryResolver {
	return &queryResolver{r}
}

// Application missing godoc
func (r *RootResolver) Application() graphql.ApplicationResolver {
	return &applicationResolver{r}
}

// ApplicationTemplate missing godoc
func (r *RootResolver) ApplicationTemplate() graphql.ApplicationTemplateResolver {
	return &applicationTemplateResolver{r}
}

// Runtime missing godoc
func (r *RootResolver) Runtime() graphql.RuntimeResolver {
	return &runtimeResolver{r}
}

// RuntimeContext missing godoc
func (r *RootResolver) RuntimeContext() graphql.RuntimeContextResolver {
	return &runtimeContextResolver{r}
}

// Formation missing godoc
func (r *RootResolver) Formation() graphql.FormationResolver {
	return &formationResolver{r}
}

// APISpec missing godoc
func (r *RootResolver) APISpec() graphql.APISpecResolver {
	return &apiSpecResolver{r}
}

// Document missing godoc
func (r *RootResolver) Document() graphql.DocumentResolver {
	return &documentResolver{r}
}

// EventSpec missing godoc
func (r *RootResolver) EventSpec() graphql.EventSpecResolver {
	return &eventSpecResolver{r}
}

// Bundle missing godoc
func (r *RootResolver) Bundle() graphql.BundleResolver {
	return &BundleResolver{r}
}

// IntegrationSystem missing godoc
func (r *RootResolver) IntegrationSystem() graphql.IntegrationSystemResolver {
	return &integrationSystemResolver{r}
}

// OneTimeTokenForApplication missing godoc
func (r *RootResolver) OneTimeTokenForApplication() graphql.OneTimeTokenForApplicationResolver {
	return &oneTimeTokenForApplicationResolver{r}
}

// OneTimeTokenForRuntime missing godoc
func (r *RootResolver) OneTimeTokenForRuntime() graphql.OneTimeTokenForRuntimeResolver {
	return &oneTimeTokenForRuntimeResolver{r}
}

// Tenant missing godoc
func (r *RootResolver) Tenant() graphql.TenantResolver {
	return &tenantResolver{r}
}

type queryResolver struct {
	*RootResolver
}

func (r *queryResolver) FormationConstraint(ctx context.Context, id string) (*graphql.FormationConstraint, error) {
	return r.formationConstraint.FormationConstraint(ctx, id)
}

func (r *queryResolver) FormationConstraints(ctx context.Context) ([]*graphql.FormationConstraint, error) {
	return r.formationConstraint.FormationConstraints(ctx)
}

func (r *queryResolver) FormationConstraintsByFormationType(ctx context.Context, formationTemplateID string) ([]*graphql.FormationConstraint, error) {
	return r.formationConstraint.FormationConstraintsByFormationType(ctx, formationTemplateID)
}

func (r *queryResolver) Formation(ctx context.Context, id string) (*graphql.Formation, error) {
	return r.formation.Formation(ctx, id)
}

// FormationByName returns a formation retrieved by name
func (r *queryResolver) FormationByName(ctx context.Context, name string) (*graphql.Formation, error) {
	return r.formation.FormationByName(ctx, name)
}

func (r *queryResolver) Formations(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.FormationPage, error) {
	return r.formation.Formations(ctx, first, after)
}

func (r *queryResolver) FormationTemplate(ctx context.Context, id string) (*graphql.FormationTemplate, error) {
	return r.formationTemplate.FormationTemplate(ctx, id)
}

func (r *queryResolver) FormationTemplates(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.FormationTemplatePage, error) {
	return r.formationTemplate.FormationTemplates(ctx, first, after)
}

// Viewer missing godoc
func (r *queryResolver) Viewer(ctx context.Context) (*graphql.Viewer, error) {
	return r.viewer.Viewer(ctx)
}

// Applications missing godoc
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

// Application missing godoc
func (r *queryResolver) Application(ctx context.Context, id string) (*graphql.Application, error) {
	return r.app.Application(ctx, id)
}

// ApplicationBySystemNumber returns an application retrieved by systemNumber
func (r *queryResolver) ApplicationBySystemNumber(ctx context.Context, systemNumber string) (*graphql.Application, error) {
	return r.app.ApplicationBySystemNumber(ctx, systemNumber)
}

// ApplicationTemplates missing godoc
func (r *queryResolver) ApplicationTemplates(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationTemplatePage, error) {
	return r.appTemplate.ApplicationTemplates(ctx, filter, first, after)
}

// ApplicationTemplate missing godoc
func (r *queryResolver) ApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.ApplicationTemplate(ctx, id)
}

// ApplicationsForRuntime missing godoc
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
	for i := range apps.Data {
		if apps.Data[i].SystemNumber != nil {
			apps.Data[i].Name = fmt.Sprintf("%s-%s", apps.Data[i].Name, *apps.Data[i].SystemNumber)
		}
	}

	return apps, nil
}

// Runtimes missing godoc
func (r *queryResolver) Runtimes(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.RuntimePage, error) {
	return r.runtime.Runtimes(ctx, filter, first, after)
}

// Runtime missing godoc
func (r *queryResolver) Runtime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.Runtime(ctx, id)
}

// RuntimeByTokenIssuer missing godoc
func (r *queryResolver) RuntimeByTokenIssuer(ctx context.Context, issuer string) (*graphql.Runtime, error) {
	return r.runtime.RuntimeByTokenIssuer(ctx, issuer)
}

// LabelDefinitions missing godoc
func (r *queryResolver) LabelDefinitions(ctx context.Context) ([]*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinitions(ctx)
}

// LabelDefinition missing godoc
func (r *queryResolver) LabelDefinition(ctx context.Context, key string) (*graphql.LabelDefinition, error) {
	return r.labelDef.LabelDefinition(ctx, key)
}

// BundleByInstanceAuth missing godoc
func (r *queryResolver) BundleByInstanceAuth(ctx context.Context, authID string) (*graphql.Bundle, error) {
	return r.bundleInstanceAuth.BundleByInstanceAuth(ctx, authID)
}

// BundleInstanceAuth missing godoc
func (r *queryResolver) BundleInstanceAuth(ctx context.Context, id string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.BundleInstanceAuth(ctx, id)
}

// HealthChecks missing godoc
func (r *queryResolver) HealthChecks(ctx context.Context, types []graphql.HealthCheckType, origin *string, first *int, after *graphql.PageCursor) (*graphql.HealthCheckPage, error) {
	return r.healthCheck.HealthChecks(ctx, types, origin, first, after)
}

// IntegrationSystems missing godoc
func (r *queryResolver) IntegrationSystems(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.IntegrationSystemPage, error) {
	return r.intSys.IntegrationSystems(ctx, first, after)
}

// IntegrationSystem missing godoc
func (r *queryResolver) IntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	return r.intSys.IntegrationSystem(ctx, id)
}

// Tenants fetches tenants by page and search term
func (r *queryResolver) Tenants(ctx context.Context, first *int, after *graphql.PageCursor, searchTerm *string) (*graphql.TenantPage, error) {
	return r.tenant.Tenants(ctx, first, after, searchTerm)
}

// AutomaticScenarioAssignmentForScenario missing godoc
func (r *queryResolver) AutomaticScenarioAssignmentForScenario(ctx context.Context, scenarioName string) (*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.GetAutomaticScenarioAssignmentForScenarioName(ctx, scenarioName)
}

// AutomaticScenarioAssignmentsForSelector missing godoc
func (r *queryResolver) AutomaticScenarioAssignmentsForSelector(ctx context.Context, selector graphql.LabelSelectorInput) ([]*graphql.AutomaticScenarioAssignment, error) {
	return r.scenarioAssignment.AutomaticScenarioAssignmentsForSelector(ctx, selector)
}

// AutomaticScenarioAssignments missing godoc
func (r *queryResolver) AutomaticScenarioAssignments(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.AutomaticScenarioAssignmentPage, error) {
	return r.scenarioAssignment.AutomaticScenarioAssignments(ctx, first, after)
}

// SystemAuth missing godoc
func (r *queryResolver) SystemAuth(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.systemAuth.SystemAuth(ctx, id)
}

// SystemAuthByToken missing godoc
func (r *queryResolver) SystemAuthByToken(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.systemAuth.SystemAuthByToken(ctx, id)
}

func (r *queryResolver) CertificateSubjectMapping(ctx context.Context, id string) (*graphql.CertificateSubjectMapping, error) {
	return r.certSubjectMapping.CertificateSubjectMapping(ctx, id)
}

func (r *queryResolver) CertificateSubjectMappings(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.CertificateSubjectMappingPage, error) {
	return r.certSubjectMapping.CertificateSubjectMappings(ctx, first, after)
}

type mutationResolver struct {
	*RootResolver
}

func (r *mutationResolver) AddTenantAccess(ctx context.Context, in graphql.TenantAccessInput) (*graphql.TenantAccess, error) {
	return r.tenant.AddTenantAccess(ctx, in)
}

func (r *mutationResolver) RemoveTenantAccess(ctx context.Context, tenantID string, resourceID string, resourceType graphql.TenantAccessObjectType) (*graphql.TenantAccess, error) {
	return r.tenant.RemoveTenantAccess(ctx, tenantID, resourceID, resourceType)
}

func (r *mutationResolver) UpdateFormationConstraint(ctx context.Context, id string, in graphql.FormationConstraintUpdateInput) (*graphql.FormationConstraint, error) {
	return r.formationConstraint.UpdateFormationConstraint(ctx, id, in)
}

func (r *mutationResolver) ResynchronizeFormationNotifications(ctx context.Context, formationID string) (*graphql.Formation, error) {
	return r.formation.ResynchronizeFormationNotifications(ctx, formationID)
}

func (r *mutationResolver) AttachConstraintToFormationTemplate(ctx context.Context, constraintID string, formationTemplateID string) (*graphql.ConstraintReference, error) {
	return r.constraintReference.AttachConstraintToFormationTemplate(ctx, constraintID, formationTemplateID)
}

func (r *mutationResolver) DetachConstraintFromFormationTemplate(ctx context.Context, constraintID string, formationTemplateID string) (*graphql.ConstraintReference, error) {
	return r.constraintReference.DetachConstraintFromFormationTemplate(ctx, constraintID, formationTemplateID)
}

func (r *mutationResolver) CreateFormationConstraint(ctx context.Context, formationConstraint graphql.FormationConstraintInput) (*graphql.FormationConstraint, error) {
	return r.formationConstraint.CreateFormationConstraint(ctx, formationConstraint)
}

func (r *mutationResolver) DeleteFormationConstraint(ctx context.Context, id string) (*graphql.FormationConstraint, error) {
	return r.formationConstraint.DeleteFormationConstraint(ctx, id)
}

func (r *mutationResolver) CreateFormationTemplate(ctx context.Context, in graphql.FormationTemplateInput) (*graphql.FormationTemplate, error) {
	return r.formationTemplate.CreateFormationTemplate(ctx, in)
}

func (r *mutationResolver) DeleteFormationTemplate(ctx context.Context, id string) (*graphql.FormationTemplate, error) {
	return r.formationTemplate.DeleteFormationTemplate(ctx, id)
}

func (r *mutationResolver) UpdateFormationTemplate(ctx context.Context, id string, in graphql.FormationTemplateInput) (*graphql.FormationTemplate, error) {
	return r.formationTemplate.UpdateFormationTemplate(ctx, id, in)
}

func (r *mutationResolver) AssignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	return r.formation.AssignFormation(ctx, objectID, objectType, formation)
}

func (r *mutationResolver) UnassignFormation(ctx context.Context, objectID string, objectType graphql.FormationObjectType, formation graphql.FormationInput) (*graphql.Formation, error) {
	return r.formation.UnassignFormation(ctx, objectID, objectType, formation)
}

func (r *mutationResolver) CreateFormation(ctx context.Context, formationInput graphql.FormationInput) (*graphql.Formation, error) {
	return r.formation.CreateFormation(ctx, formationInput)
}

func (r *mutationResolver) DeleteFormation(ctx context.Context, formation graphql.FormationInput) (*graphql.Formation, error) {
	return r.formation.DeleteFormation(ctx, formation)
}

// RegisterApplication missing godoc
func (r *mutationResolver) RegisterApplication(ctx context.Context, in graphql.ApplicationRegisterInput, _ *graphql.OperationMode) (*graphql.Application, error) {
	return r.app.RegisterApplication(ctx, in)
}

// UpdateApplication missing godoc
func (r *mutationResolver) UpdateApplication(ctx context.Context, id string, in graphql.ApplicationUpdateInput) (*graphql.Application, error) {
	return r.app.UpdateApplication(ctx, id, in)
}

// UnregisterApplication missing godoc
func (r *mutationResolver) UnregisterApplication(ctx context.Context, id string, _ *graphql.OperationMode) (*graphql.Application, error) {
	return r.app.UnregisterApplication(ctx, id)
}

// UnpairApplication removes system auths, hydra client credentials and performs the "unpair" flow in the database
func (r *mutationResolver) UnpairApplication(ctx context.Context, id string, _ *graphql.OperationMode) (*graphql.Application, error) {
	return r.app.UnpairApplication(ctx, id)
}

// MergeApplications Merges properties from Source Application into Destination Application, provided that the Destination's
// Application does not have a value set for a given property. Then the Source Application is being deleted.
func (r *mutationResolver) MergeApplications(ctx context.Context, destID, srcID string) (*graphql.Application, error) {
	return r.app.MergeApplications(ctx, destID, srcID)
}

// CreateApplicationTemplate missing godoc
func (r *mutationResolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.CreateApplicationTemplate(ctx, in)
}

// RegisterApplicationFromTemplate missing godoc
func (r *mutationResolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	return r.appTemplate.RegisterApplicationFromTemplate(ctx, in)
}

// UpdateApplicationTemplate missing godoc
func (r *mutationResolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateUpdateInput) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.UpdateApplicationTemplate(ctx, id, in)
}

// DeleteApplicationTemplate missing godoc
func (r *mutationResolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	return r.appTemplate.DeleteApplicationTemplate(ctx, id)
}

// AddWebhook missing godoc
func (r *mutationResolver) AddWebhook(ctx context.Context, applicationID *string, applicationTemplateID *string, runtimeID *string, formationTemplateID *string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.AddWebhook(ctx, applicationID, applicationTemplateID, runtimeID, formationTemplateID, in)
}

// UpdateWebhook missing godoc
func (r *mutationResolver) UpdateWebhook(ctx context.Context, webhookID string, in graphql.WebhookInput) (*graphql.Webhook, error) {
	return r.webhook.UpdateWebhook(ctx, webhookID, in)
}

// DeleteWebhook missing godoc
func (r *mutationResolver) DeleteWebhook(ctx context.Context, webhookID string) (*graphql.Webhook, error) {
	return r.webhook.DeleteWebhook(ctx, webhookID)
}

// UpdateAPIDefinition missing godoc
func (r *mutationResolver) UpdateAPIDefinition(ctx context.Context, id string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.UpdateAPIDefinition(ctx, id, in)
}

// DeleteAPIDefinition missing godoc
func (r *mutationResolver) DeleteAPIDefinition(ctx context.Context, id string) (*graphql.APIDefinition, error) {
	return r.api.DeleteAPIDefinition(ctx, id)
}

// RefetchAPISpec missing godoc
func (r *mutationResolver) RefetchAPISpec(ctx context.Context, apiID string) (*graphql.APISpec, error) {
	return r.api.RefetchAPISpec(ctx, apiID)
}

// UpdateEventDefinition missing godoc
func (r *mutationResolver) UpdateEventDefinition(ctx context.Context, id string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.UpdateEventDefinition(ctx, id, in)
}

// DeleteEventDefinition missing godoc
func (r *mutationResolver) DeleteEventDefinition(ctx context.Context, id string) (*graphql.EventDefinition, error) {
	return r.eventAPI.DeleteEventDefinition(ctx, id)
}

// RefetchEventDefinitionSpec missing godoc
func (r *mutationResolver) RefetchEventDefinitionSpec(ctx context.Context, eventID string) (*graphql.EventSpec, error) {
	return r.eventAPI.RefetchEventDefinitionSpec(ctx, eventID)
}

// RegisterRuntime missing godoc
func (r *mutationResolver) RegisterRuntime(ctx context.Context, in graphql.RuntimeRegisterInput) (*graphql.Runtime, error) {
	return r.runtime.RegisterRuntime(ctx, in)
}

// UpdateRuntime missing godoc
func (r *mutationResolver) UpdateRuntime(ctx context.Context, id string, in graphql.RuntimeUpdateInput) (*graphql.Runtime, error) {
	return r.runtime.UpdateRuntime(ctx, id, in)
}

// UnregisterRuntime missing godoc
func (r *mutationResolver) UnregisterRuntime(ctx context.Context, id string) (*graphql.Runtime, error) {
	return r.runtime.DeleteRuntime(ctx, id)
}

// RegisterRuntimeContext missing godoc
func (r *mutationResolver) RegisterRuntimeContext(ctx context.Context, runtimeID string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.RegisterRuntimeContext(ctx, runtimeID, in)
}

// UpdateRuntimeContext missing godoc
func (r *mutationResolver) UpdateRuntimeContext(ctx context.Context, id string, in graphql.RuntimeContextInput) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.UpdateRuntimeContext(ctx, id, in)
}

// UnregisterRuntimeContext missing godoc
func (r *mutationResolver) UnregisterRuntimeContext(ctx context.Context, id string) (*graphql.RuntimeContext, error) {
	return r.runtimeContext.DeleteRuntimeContext(ctx, id)
}

// DeleteDocument missing godoc
func (r *mutationResolver) DeleteDocument(ctx context.Context, id string) (*graphql.Document, error) {
	return r.doc.DeleteDocument(ctx, id)
}

// CreateLabelDefinition missing godoc
func (r *mutationResolver) CreateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	return r.labelDef.CreateLabelDefinition(ctx, in)
}

// UpdateLabelDefinition missing godoc
func (r *mutationResolver) UpdateLabelDefinition(ctx context.Context, in graphql.LabelDefinitionInput) (*graphql.LabelDefinition, error) {
	return r.labelDef.UpdateLabelDefinition(ctx, in)
}

// SetApplicationLabel missing godoc
func (r *mutationResolver) SetApplicationLabel(ctx context.Context, applicationID string, key string, value interface{}) (*graphql.Label, error) {
	return r.app.SetApplicationLabel(ctx, applicationID, key, value)
}

// DeleteApplicationLabel missing godoc
func (r *mutationResolver) DeleteApplicationLabel(ctx context.Context, applicationID string, key string) (*graphql.Label, error) {
	return r.app.DeleteApplicationLabel(ctx, applicationID, key)
}

// SetRuntimeLabel missing godoc
func (r *mutationResolver) SetRuntimeLabel(ctx context.Context, runtimeID string, key string, value interface{}) (*graphql.Label, error) {
	return r.runtime.SetRuntimeLabel(ctx, runtimeID, key, value)
}

// DeleteRuntimeLabel missing godoc
func (r *mutationResolver) DeleteRuntimeLabel(ctx context.Context, runtimeID string, key string) (*graphql.Label, error) {
	return r.runtime.DeleteRuntimeLabel(ctx, runtimeID, key)
}

// RequestOneTimeTokenForApplication missing godoc
func (r *mutationResolver) RequestOneTimeTokenForApplication(ctx context.Context, id string, systemAuthID *string) (*graphql.OneTimeTokenForApplication, error) {
	return r.token.RequestOneTimeTokenForApplication(ctx, id, systemAuthID)
}

// RequestOneTimeTokenForRuntime missing godoc
func (r *mutationResolver) RequestOneTimeTokenForRuntime(ctx context.Context, id string, systemAuthID *string) (*graphql.OneTimeTokenForRuntime, error) {
	return r.token.RequestOneTimeTokenForRuntime(ctx, id, systemAuthID)
}

// RequestClientCredentialsForRuntime missing godoc
func (r *mutationResolver) RequestClientCredentialsForRuntime(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForRuntime(ctx, id)
}

// RequestClientCredentialsForApplication missing godoc
func (r *mutationResolver) RequestClientCredentialsForApplication(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForApplication(ctx, id)
}

// RequestClientCredentialsForIntegrationSystem missing godoc
func (r *mutationResolver) RequestClientCredentialsForIntegrationSystem(ctx context.Context, id string) (graphql.SystemAuth, error) {
	return r.oAuth20.RequestClientCredentialsForIntegrationSystem(ctx, id)
}

// DeleteSystemAuthForRuntime missing godoc
func (r *mutationResolver) DeleteSystemAuthForRuntime(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.RuntimeReference)
	return fn(ctx, authID)
}

// DeleteSystemAuthForApplication missing godoc
func (r *mutationResolver) DeleteSystemAuthForApplication(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.ApplicationReference)
	return fn(ctx, authID)
}

// DeleteSystemAuthForIntegrationSystem missing godoc
func (r *mutationResolver) DeleteSystemAuthForIntegrationSystem(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	fn := r.systemAuth.GenericDeleteSystemAuth(model.IntegrationSystemReference)
	return fn(ctx, authID)
}

// UpdateSystemAuth missing godoc
func (r *mutationResolver) UpdateSystemAuth(ctx context.Context, authID string, in graphql.AuthInput) (graphql.SystemAuth, error) {
	return r.systemAuth.UpdateSystemAuth(ctx, authID, in)
}

// InvalidateSystemAuthOneTimeToken missing godoc
func (r *mutationResolver) InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) (graphql.SystemAuth, error) {
	return r.systemAuth.InvalidateSystemAuthOneTimeToken(ctx, authID)
}

// RegisterIntegrationSystem missing godoc
func (r *mutationResolver) RegisterIntegrationSystem(ctx context.Context, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.RegisterIntegrationSystem(ctx, in)
}

// UpdateIntegrationSystem missing godoc
func (r *mutationResolver) UpdateIntegrationSystem(ctx context.Context, id string, in graphql.IntegrationSystemInput) (*graphql.IntegrationSystem, error) {
	return r.intSys.UpdateIntegrationSystem(ctx, id, in)
}

// UnregisterIntegrationSystem missing godoc
func (r *mutationResolver) UnregisterIntegrationSystem(ctx context.Context, id string) (*graphql.IntegrationSystem, error) {
	return r.intSys.UnregisterIntegrationSystem(ctx, id)
}

// SetDefaultEventingForApplication missing godoc
func (r *mutationResolver) SetDefaultEventingForApplication(ctx context.Context, appID string, runtimeID string) (*graphql.ApplicationEventingConfiguration, error) {
	return r.eventing.SetEventingForApplication(ctx, appID, runtimeID)
}

// DeleteDefaultEventingForApplication missing godoc
func (r *mutationResolver) DeleteDefaultEventingForApplication(ctx context.Context, appID string) (*graphql.ApplicationEventingConfiguration, error) {
	return r.eventing.UnsetEventingForApplication(ctx, appID)
}

// AddAPIDefinitionToBundle missing godoc
func (r *mutationResolver) AddAPIDefinitionToBundle(ctx context.Context, bundleID string, in graphql.APIDefinitionInput) (*graphql.APIDefinition, error) {
	return r.api.AddAPIDefinitionToBundle(ctx, bundleID, in)
}

// AddEventDefinitionToBundle missing godoc
func (r *mutationResolver) AddEventDefinitionToBundle(ctx context.Context, bundleID string, in graphql.EventDefinitionInput) (*graphql.EventDefinition, error) {
	return r.eventAPI.AddEventDefinitionToBundle(ctx, bundleID, in)
}

// AddDocumentToBundle missing godoc
func (r *mutationResolver) AddDocumentToBundle(ctx context.Context, bundleID string, in graphql.DocumentInput) (*graphql.Document, error) {
	return r.doc.AddDocumentToBundle(ctx, bundleID, in)
}

// SetBundleInstanceAuth missing godoc
func (r *mutationResolver) SetBundleInstanceAuth(ctx context.Context, authID string, in graphql.BundleInstanceAuthSetInput) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.SetBundleInstanceAuth(ctx, authID, in)
}

// DeleteBundleInstanceAuth missing godoc
func (r *mutationResolver) DeleteBundleInstanceAuth(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.DeleteBundleInstanceAuth(ctx, authID)
}

// RequestBundleInstanceAuthCreation missing godoc
func (r *mutationResolver) RequestBundleInstanceAuthCreation(ctx context.Context, bundleID string, in graphql.BundleInstanceAuthRequestInput) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.RequestBundleInstanceAuthCreation(ctx, bundleID, in)
}

// RequestBundleInstanceAuthDeletion missing godoc
func (r *mutationResolver) RequestBundleInstanceAuthDeletion(ctx context.Context, authID string) (*graphql.BundleInstanceAuth, error) {
	return r.bundleInstanceAuth.RequestBundleInstanceAuthDeletion(ctx, authID)
}

// AddBundle missing godoc
func (r *mutationResolver) AddBundle(ctx context.Context, applicationID string, in graphql.BundleCreateInput) (*graphql.Bundle, error) {
	return r.mpBundle.AddBundle(ctx, applicationID, in)
}

// UpdateBundle missing godoc
func (r *mutationResolver) UpdateBundle(ctx context.Context, id string, in graphql.BundleUpdateInput) (*graphql.Bundle, error) {
	return r.mpBundle.UpdateBundle(ctx, id, in)
}

// DeleteBundle missing godoc
func (r *mutationResolver) DeleteBundle(ctx context.Context, id string) (*graphql.Bundle, error) {
	return r.mpBundle.DeleteBundle(ctx, id)
}

// WriteTenants creates tenants of type customer, account, subaccount, organization, folder, or resource-group
func (r *mutationResolver) WriteTenants(ctx context.Context, in []*graphql.BusinessTenantMappingInput) ([]string, error) {
	return r.tenant.Write(ctx, in)
}

// WriteTenant creates tenant of type customer, account, subaccount, organization, folder, or resource-group
func (r *mutationResolver) WriteTenant(ctx context.Context, in graphql.BusinessTenantMappingInput) (string, error) {
	return r.tenant.WriteSingle(ctx, in)
}

// DeleteTenants deletes multiple tenants by external tenant id
func (r *mutationResolver) DeleteTenants(ctx context.Context, in []string) (int, error) {
	return r.tenant.Delete(ctx, in)
}

// SubscribeTenant subscribes given tenant
func (r *mutationResolver) SubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionAppName string, subscriptionPayload string) (bool, error) {
	return r.subscription.SubscribeTenant(ctx, providerID, subaccountID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionPayload)
}

// UnsubscribeTenant unsubscribes given tenant
func (r *mutationResolver) UnsubscribeTenant(ctx context.Context, providerID, subaccountID, providerSubaccountID, consumerTenantID, region string) (bool, error) {
	return r.subscription.UnsubscribeTenant(ctx, providerID, subaccountID, providerSubaccountID, consumerTenantID, region)
}

func (r *mutationResolver) UpdateTenant(ctx context.Context, id string, in graphql.BusinessTenantMappingInput) (*graphql.Tenant, error) {
	return r.tenant.Update(ctx, id, in)
}

func (r *mutationResolver) CreateCertificateSubjectMapping(ctx context.Context, in graphql.CertificateSubjectMappingInput) (*graphql.CertificateSubjectMapping, error) {
	return r.certSubjectMapping.CreateCertificateSubjectMapping(ctx, in)
}

func (r *mutationResolver) UpdateCertificateSubjectMapping(ctx context.Context, id string, in graphql.CertificateSubjectMappingInput) (*graphql.CertificateSubjectMapping, error) {
	return r.certSubjectMapping.UpdateCertificateSubjectMapping(ctx, id, in)
}

func (r *mutationResolver) DeleteCertificateSubjectMapping(ctx context.Context, id string) (*graphql.CertificateSubjectMapping, error) {
	return r.certSubjectMapping.DeleteCertificateSubjectMapping(ctx, id)
}

type applicationResolver struct {
	*RootResolver
}

// Auths missing godoc
func (r *applicationResolver) Auths(ctx context.Context, obj *graphql.Application) ([]*graphql.AppSystemAuth, error) {
	return r.app.Auths(ctx, obj)
}

// Labels missing godoc
func (r *applicationResolver) Labels(ctx context.Context, obj *graphql.Application, key *string) (graphql.Labels, error) {
	return r.app.Labels(ctx, obj, key)
}

// Webhooks missing godoc
func (r *applicationResolver) Webhooks(ctx context.Context, obj *graphql.Application) ([]*graphql.Webhook, error) {
	return r.app.Webhooks(ctx, obj)
}

// EventingConfiguration missing godoc
func (r *applicationResolver) EventingConfiguration(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationEventingConfiguration, error) {
	return r.app.EventingConfiguration(ctx, obj)
}

// Bundles missing godoc
func (r *applicationResolver) Bundles(ctx context.Context, obj *graphql.Application, first *int, after *graphql.PageCursor) (*graphql.BundlePage, error) {
	return r.app.Bundles(ctx, obj, first, after)
}

// Bundle missing godoc
func (r *applicationResolver) Bundle(ctx context.Context, obj *graphql.Application, id string) (*graphql.Bundle, error) {
	return r.app.Bundle(ctx, obj, id)
}

func (r *applicationResolver) ApplicationTemplate(ctx context.Context, obj *graphql.Application) (*graphql.ApplicationTemplate, error) {
	return r.app.ApplicationTemplate(ctx, obj)
}

type applicationTemplateResolver struct {
	*RootResolver
}

// Webhooks missing godoc
func (r applicationTemplateResolver) Webhooks(ctx context.Context, obj *graphql.ApplicationTemplate) ([]*graphql.Webhook, error) {
	return r.appTemplate.Webhooks(ctx, obj)
}

// Labels missing godoc
func (r applicationTemplateResolver) Labels(ctx context.Context, obj *graphql.ApplicationTemplate, key *string) (graphql.Labels, error) {
	return r.appTemplate.Labels(ctx, obj, key)
}

type formationTemplateResolver struct {
	*RootResolver
}

// FormationTemplate represents the resolver for Formation Template
func (r *RootResolver) FormationTemplate() graphql.FormationTemplateResolver {
	return &formationTemplateResolver{r}
}

// Webhooks missing godoc
func (r *formationTemplateResolver) Webhooks(ctx context.Context, obj *graphql.FormationTemplate) ([]*graphql.Webhook, error) {
	return r.formationTemplate.Webhooks(ctx, obj)
}

// FormationConstraints missing godoc
func (r *formationTemplateResolver) FormationConstraints(ctx context.Context, obj *graphql.FormationTemplate) ([]*graphql.FormationConstraint, error) {
	return r.formationTemplate.FormationConstraint(ctx, obj)
}

type runtimeResolver struct {
	*RootResolver
}

func (r *runtimeResolver) Webhooks(ctx context.Context, obj *graphql.Runtime) ([]*graphql.Webhook, error) {
	return r.runtime.Webhooks(ctx, obj)
}

// Labels missing godoc
func (r *runtimeResolver) Labels(ctx context.Context, obj *graphql.Runtime, key *string) (graphql.Labels, error) {
	return r.runtime.Labels(ctx, obj, key)
}

// Auths missing godoc
func (r *runtimeResolver) Auths(ctx context.Context, obj *graphql.Runtime) ([]*graphql.RuntimeSystemAuth, error) {
	return r.runtime.Auths(ctx, obj)
}

// EventingConfiguration missing godoc
func (r *runtimeResolver) EventingConfiguration(ctx context.Context, obj *graphql.Runtime) (*graphql.RuntimeEventingConfiguration, error) {
	return r.runtime.EventingConfiguration(ctx, obj)
}

// RuntimeContexts missing godoc
func (r *runtimeResolver) RuntimeContexts(ctx context.Context, obj *graphql.Runtime, first *int, after *graphql.PageCursor) (*graphql.RuntimeContextPage, error) {
	return r.runtime.RuntimeContexts(ctx, obj, first, after)
}

// RuntimeContext missing godoc
func (r *runtimeResolver) RuntimeContext(ctx context.Context, obj *graphql.Runtime, id string) (*graphql.RuntimeContext, error) {
	return r.runtime.RuntimeContext(ctx, obj, id)
}

type apiSpecResolver struct{ *RootResolver }

// FetchRequest missing godoc
func (r *apiSpecResolver) FetchRequest(ctx context.Context, obj *graphql.APISpec) (*graphql.FetchRequest, error) {
	return r.api.FetchRequest(ctx, obj)
}

type documentResolver struct{ *RootResolver }

// FetchRequest missing godoc
func (r *documentResolver) FetchRequest(ctx context.Context, obj *graphql.Document) (*graphql.FetchRequest, error) {
	return r.doc.FetchRequest(ctx, obj)
}

type eventSpecResolver struct{ *RootResolver }

// FetchRequest missing godoc
func (r *eventSpecResolver) FetchRequest(ctx context.Context, obj *graphql.EventSpec) (*graphql.FetchRequest, error) {
	return r.eventAPI.FetchRequest(ctx, obj)
}

type integrationSystemResolver struct{ *RootResolver }

// Auths missing godoc
func (r *integrationSystemResolver) Auths(ctx context.Context, obj *graphql.IntegrationSystem) ([]*graphql.IntSysSystemAuth, error) {
	return r.intSys.Auths(ctx, obj)
}

type oneTimeTokenForApplicationResolver struct{ *RootResolver }

// RawEncoded missing godoc
func (r *oneTimeTokenForApplicationResolver) RawEncoded(ctx context.Context, obj *graphql.OneTimeTokenForApplication) (*string, error) {
	return r.token.RawEncoded(ctx, &obj.TokenWithURL)
}

// Raw missing godoc
func (r *oneTimeTokenForApplicationResolver) Raw(ctx context.Context, obj *graphql.OneTimeTokenForApplication) (*string, error) {
	return r.token.Raw(ctx, &obj.TokenWithURL)
}

type oneTimeTokenForRuntimeResolver struct{ *RootResolver }

// RawEncoded missing godoc
func (r *oneTimeTokenForRuntimeResolver) RawEncoded(ctx context.Context, obj *graphql.OneTimeTokenForRuntime) (*string, error) {
	return r.token.RawEncoded(ctx, &obj.TokenWithURL)
}

// Raw missing godoc
func (r *oneTimeTokenForRuntimeResolver) Raw(ctx context.Context, obj *graphql.OneTimeTokenForRuntime) (*string, error) {
	return r.token.Raw(ctx, &obj.TokenWithURL)
}

type runtimeContextResolver struct {
	*RootResolver
}

// Labels missing godoc
func (r *runtimeContextResolver) Labels(ctx context.Context, obj *graphql.RuntimeContext, key *string) (graphql.Labels, error) {
	return r.runtimeContext.Labels(ctx, obj, key)
}

type formationResolver struct {
	*RootResolver
}

// FormationAssignment missing godoc
func (r *formationResolver) FormationAssignment(ctx context.Context, obj *graphql.Formation, id string) (*graphql.FormationAssignment, error) {
	return r.formation.FormationAssignment(ctx, obj, id)
}

// FormationAssignments missing godoc
func (r *formationResolver) FormationAssignments(ctx context.Context, obj *graphql.Formation, first *int, after *graphql.PageCursor) (*graphql.FormationAssignmentPage, error) {
	return r.formation.FormationAssignments(ctx, obj, first, after)
}

// Status missing godoc
func (r *formationResolver) Status(ctx context.Context, obj *graphql.Formation) (*graphql.FormationStatus, error) {
	return r.formation.Status(ctx, obj)
}

// BundleResolver missing godoc
type BundleResolver struct{ *RootResolver }

// CorrelationIDs missing godoc
func (r *BundleResolver) CorrelationIDs(_ context.Context, obj *graphql.Bundle) ([]string, error) {
	return obj.CorrelationIDs, nil
}

// InstanceAuth missing godoc
func (r *BundleResolver) InstanceAuth(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.BundleInstanceAuth, error) {
	return r.mpBundle.InstanceAuth(ctx, obj, id)
}

// InstanceAuths missing godoc
func (r *BundleResolver) InstanceAuths(ctx context.Context, obj *graphql.Bundle) ([]*graphql.BundleInstanceAuth, error) {
	return r.mpBundle.InstanceAuths(ctx, obj)
}

// APIDefinitions missing godoc
func (r *BundleResolver) APIDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.APIDefinitionPage, error) {
	return r.mpBundle.APIDefinitions(ctx, obj, group, first, after)
}

// EventDefinitions missing godoc
func (r *BundleResolver) EventDefinitions(ctx context.Context, obj *graphql.Bundle, group *string, first *int, after *graphql.PageCursor) (*graphql.EventDefinitionPage, error) {
	return r.mpBundle.EventDefinitions(ctx, obj, group, first, after)
}

// Documents missing godoc
func (r *BundleResolver) Documents(ctx context.Context, obj *graphql.Bundle, first *int, after *graphql.PageCursor) (*graphql.DocumentPage, error) {
	return r.mpBundle.Documents(ctx, obj, first, after)
}

// APIDefinition missing godoc
func (r *BundleResolver) APIDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.APIDefinition, error) {
	return r.mpBundle.APIDefinition(ctx, obj, id)
}

// EventDefinition missing godoc
func (r *BundleResolver) EventDefinition(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.EventDefinition, error) {
	return r.mpBundle.EventDefinition(ctx, obj, id)
}

// Document missing godoc
func (r *BundleResolver) Document(ctx context.Context, obj *graphql.Bundle, id string) (*graphql.Document, error) {
	return r.mpBundle.Document(ctx, obj, id)
}

type tenantResolver struct {
	*RootResolver
}

// Labels missing godoc
func (r *tenantResolver) Labels(ctx context.Context, obj *graphql.Tenant, key *string) (graphql.Labels, error) {
	return r.tenant.Labels(ctx, obj, key)
}

// TenantByExternalID returns a tenant by external ID
func (r *queryResolver) TenantByExternalID(ctx context.Context, id string) (*graphql.Tenant, error) {
	return r.tenant.Tenant(ctx, id)
}

// TenantByInternalID returns a tenant by an internal ID
func (r *queryResolver) TenantByInternalID(ctx context.Context, id string) (*graphql.Tenant, error) {
	return r.tenant.TenantByID(ctx, id)
}

// TenantByLowestOwnerForResource returns the lowest tenant in the hierarchy that is owner of a given resource.
func (r *queryResolver) TenantByLowestOwnerForResource(ctx context.Context, resource, objectID string) (string, error) {
	return r.tenant.TenantByLowestOwnerForResource(ctx, resource, objectID)
}
