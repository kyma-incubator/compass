package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"

	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/jwt"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	directortime "github.com/kyma-incubator/compass/components/director/pkg/time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenantbusinesstype"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	oauth "github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const discoverSystemsOpMode = "DISCOVER_SYSTEMS"

type config struct {
	Address           string `envconfig:"default=127.0.0.1:8080"`
	AggregatorRootAPI string `envconfig:"APP_ROOT_API,default=/system-fetcher"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	SecurityConfig securityConfig

	ElectionConfig         cronjob.ElectionConfig
	SystemsSyncJobInterval time.Duration `envconfig:"APP_SYSTEM_SYNC_JOB_INTERVAL,default=24h"`

	APIConfig           systemfetcher.APIConfig
	OAuth2Config        oauth.Config
	SelfSignedJwtConfig jwt.Config
	SystemFetcher       systemfetcher.Config
	Database            persistence.DatabaseConfig
	TemplateConfig      appTemplateConfig

	Log log.Config

	Features features.Config

	ConfigurationFile string

	ConfigurationFileReload time.Duration `envconfig:"default=1m"`
	ClientTimeout           time.Duration `envconfig:"default=60s"`

	CertLoaderConfig credloader.CertConfig
	KeyLoaderConfig  credloader.KeysConfig

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`
}

type securityConfig struct {
	JwksEndpoint           string        `envconfig:"APP_JWKS_ENDPOINT"`
	JWKSSyncPeriod         time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone    bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	SystemFetcherSyncScope string        `envconfig:"APP_SYSTEM_FETCHER_SCOPE,default=system_fetcher:sync"`
}

type appTemplateConfig struct {
	LabelFilter                    string `envconfig:"APP_TEMPLATE_LABEL_FILTER"`
	OverrideApplicationInput       string `envconfig:"APP_TEMPLATE_OVERRIDE_APPLICATION_INPUT"`
	PlaceholderToSystemKeyMappings string `envconfig:"APP_TEMPLATE_PLACEHOLDER_TO_SYSTEM_KEY_MAPPINGS"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	handler := initHandler(ctx, cfg)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	if cfg.SystemFetcher.OperationalMode == discoverSystemsOpMode {
		go func() {
			if err := startSyncSystemsJob(ctx, cfg, tenantEntity.Customer); err != nil {
				log.C(ctx).WithError(err).Error("Failed to start sync systems for customers cronjob. Stopping app...")
			}
			cancel()
		}()
		if cfg.SystemFetcher.SyncGlobalAccounts {
			go func() {
				if err := startSyncSystemsJob(ctx, cfg, tenantEntity.Account); err != nil {
					log.C(ctx).WithError(err).Error("Failed to start sync systems for global accounts cronjob. Stopping app...")
				}
				cancel()
			}()
		}
	} else {
		log.C(ctx).Infof("The operatioal mode is set to %q, skipping systems discovery.", cfg.SystemFetcher.OperationalMode)
	}
	runMainSrv()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func startSyncSystemsJob(ctx context.Context, cfg config, tenantEntityType tenantEntity.Type) error {
	jobName := fmt.Sprintf("SyncSystemsForType_%s", tenantEntityType)
	resyncJob := cronjob.CronJob{
		Name: jobName,
		Fn: func(jobCtx context.Context) {
			syncSystemsOfType(ctx, cfg, tenantEntityType)
		},
		SchedulePeriod: cfg.SystemsSyncJobInterval,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, resyncJob)
}

func syncSystemsOfType(ctx context.Context, cfg config, tenantEntityType tenantEntity.Type) {
	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	certCache, err := credloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	keyCache, err := credloader.StartKeyLoader(ctx, cfg.KeyLoaderConfig)
	exitOnError(err, "Failed to initialize key loader")

	err = credloader.WaitForKeyCache(keyCache)
	exitOnError(err, "Failed to wait for key cache")

	err = credloader.WaitForCertCache(certCache)
	exitOnError(err, "Failed to wait for cert cache")

	httpClient := &http.Client{Timeout: cfg.ClientTimeout}
	securedHTTPClient := pkgAuth.PrepareHTTPClient(cfg.ClientTimeout)
	mtlsClient := pkgAuth.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExternalClientCertSecretName)
	extSvcMtlsClient := pkgAuth.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExtSvcClientCertSecretName)

	sf, err := createSystemFetcher(ctx, cfg, cfgProvider, transact, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient, certCache, keyCache)
	exitOnError(err, "Failed to initialize System Fetcher")

	log.C(ctx).Infof("Start syncing systems for %q ...", tenantEntityType)
	if err := sf.SyncSystems(ctx, tenantEntityType); err != nil {
		log.C(ctx).Errorf("Cannot sync systems for %q. Err: %v", tenantEntityType, err)
	}
	log.C(ctx).Infof("Step 1 of 2 - sync systems for %q is finished.", tenantEntityType)

	if err := sf.UpsertSystemsSyncTimestamps(ctx, transact); err != nil {
		log.C(ctx).Errorf("Cannot upsert systems synchronization timestamps in database for %q. Err: %v", tenantEntityType, err)
	}
	log.C(ctx).Infof("Step 2 of 2 - upsert systems synchronization timestamps for %q is finished.", tenantEntityType)
	log.C(ctx).Infof("Finished syncing systems for %q", tenantEntityType)
}

func initHandler(ctx context.Context, cfg config) http.Handler {
	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
		syncEndpoint    = "/sync"
	)
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.AggregatorRootAPI+healthzEndpoint, cfg.AggregatorRootAPI+readyzEndpoint))

	apiRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()

	httpClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	configureAuthMiddleware(ctx, httpClient, apiRouter, cfg, cfg.SecurityConfig.SystemFetcherSyncScope)
	apiRouter.HandleFunc(syncEndpoint, newSyncHandler(ctx, cfg)).Methods(http.MethodPost)

	healthCheckRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func newSyncHandler(ctx context.Context, cfg config) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		log.C(ctx).Infof("Start on demand syncing of systems")
		syncSystemsOfType(ctx, cfg, tenantEntity.Customer)
		if cfg.SystemFetcher.SyncGlobalAccounts {
			syncSystemsOfType(ctx, cfg, tenantEntity.Account)
		}
		log.C(ctx).Infof("Finished on demand syncing of systems")
		writer.WriteHeader(http.StatusOK)
	}
}

func configureAuthMiddleware(ctx context.Context, httpClient *http.Client, router *mux.Router, cfg config, requiredScopes ...string) {
	scopeValidator := claims.NewScopesValidator(requiredScopes)
	middleware := authmiddleware.New(httpClient, cfg.SecurityConfig.JwksEndpoint, cfg.SecurityConfig.AllowJWTSigningNone, "", scopeValidator)
	router.Use(middleware.Handler())

	log.C(ctx).Infof("JWKS synchronization enabled. Sync period: %v", cfg.SecurityConfig.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.SecurityConfig.JWKSSyncPeriod, func(ctx context.Context) {
		if err := middleware.SynchronizeJWKS(ctx); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring system fetcher handler")

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	runFn := func() {
		log.C(ctx).Infof("Running %s server on %s...", name, cfg.Address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.C(ctx).Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			log.C(ctx).Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func createSystemFetcher(ctx context.Context, cfg config, cfgProvider *configprovider.Provider, tx persistence.Transactioner, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient *http.Client, certCache credloader.CertCache, keyCache credloader.KeysCache) (*systemfetcher.SystemFetcher, error) {
	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	if err != nil {
		return nil, errors.Wrap(err, "failed while unmarshalling ord webhook mappings")
	}

	tenantConverter := tenant.NewConverter()
	tenantBusinessTypeConverter := tenantbusinesstype.NewConverter()
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	intSysConverter := integrationsystem.NewConverter()
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	bundleReferenceConverter := bundlereferences.NewConverter()
	runtimeContextConverter := runtimectx.NewConverter()
	formationConverter := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	assignmentConverter := scenarioassignment.NewConverter()
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	formationAssignmentConverter := formationassignment.NewConverter()
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	systemsSyncConverter := systemssync.NewConverter()
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)
	certSubjectMappingConv := certsubjectmapping.NewConverter()

	tenantRepo := tenant.NewRepository(tenantConverter)
	tenantBusinessTypeRepo := tenantbusinesstype.NewRepository(tenantBusinessTypeConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	applicationRepo := application.NewRepository(appConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	apiRepo := api.NewRepository(apiConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	specRepo := spec.NewRepository(specConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	bundleRepo := bundleutil.NewRepository(bundleConverter)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	formationRepo := formation.NewRepository(formationConverter)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	systemsSyncRepo := systemssync.NewRepository(systemsSyncConverter)
	bundleInstanceAuthRepo := bundleinstanceauth.NewRepository(bundleInstanceAuthConv)
	certSubjectMappingRepo := certsubjectmapping.NewRepository(certSubjectMappingConv)
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)

	timeSvc := directortime.NewService()
	uidSvc := uid.NewService()
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	tenantBusinessTypeSvc := tenantbusinesstype.NewService(tenantBusinessTypeRepo, uidSvc)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName))
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, bundleInstanceAuthSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo, applicationRepo, timeSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc, map[string]interface{}{}, "")
	webhookLabelBuilder := databuilder.NewWebhookLabelBuilder(labelRepo)
	webhookTenantBuilder := databuilder.NewWebhookTenantBuilder(webhookLabelBuilder, tenantRepo)
	certSubjectInputBuilder := databuilder.NewWebhookCertSubjectBuilder(certSubjectMappingRepo)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, webhookLabelBuilder, webhookTenantBuilder, certSubjectInputBuilder)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := operators.NewConstraintEngine(tx, formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, nil, nil, systemAuthSvc, formationRepo, labelRepo, labelSvc, applicationRepo, runtimeContextRepo, formationTemplateRepo, formationAssignmentRepo, nil, nil, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter, formationTemplateRepo)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationStatusSvc := formation.NewFormationStatusService(formationRepo, labelDefRepo, scenariosSvc, notificationSvc, constraintEngine)
	formationSvc := formation.NewService(tx, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, formationStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, cfg.SelfRegisterDistinguishLabelKey, ordWebhookMapping)
	systemsSyncSvc := systemssync.NewService(systemsSyncRepo)

	constraintEngine.SetFormationAssignmentNotificationService(faNotificationSvc)
	constraintEngine.SetFormationAssignmentService(formationAssignmentSvc)

	authProvider := pkgAuth.NewMtlsTokenAuthorizationProvider(cfg.OAuth2Config, cfg.ExternalClientCertSecretName, certCache, pkgAuth.DefaultMtlsClientCreator)
	jwtAuthProvider := pkgAuth.NewSelfSignedJWTTokenAuthorizationProvider(cfg.SelfSignedJwtConfig)
	client := &http.Client{
		Transport: httputil.NewSecuredTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)), authProvider, jwtAuthProvider),
		Timeout:   cfg.APIConfig.Timeout,
	}
	oauthMtlsClient := systemfetcher.NewOauthMtlsClient(cfg.OAuth2Config, certCache, client)
	jwtTokenClient := systemfetcher.NewJwtTokenClient(keyCache, cfg.KeyLoaderConfig.KeysSecretName, client)
	systemsAPIClient := systemfetcher.NewClient(cfg.APIConfig, oauthMtlsClient, jwtTokenClient)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SystemFetcher.DirectorSkipSSLValidation,
		},
	}

	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewErrorHandlerTransport(httputil.NewHTTPTransportWrapper(tr)))

	securedClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.SystemFetcher.DirectorRequestTimeout,
	}

	graphqlClient := gcli.NewClient(cfg.SystemFetcher.DirectorGraphqlURL, gcli.WithHTTPClient(securedClient))
	directorClient := &systemfetcher.DirectorGraphClient{
		Client:        graphqlClient,
		Authenticator: pkgAuth.NewServiceAccountTokenAuthorizationProvider(),
	}

	dataLoader := systemfetcher.NewDataLoader(tx, cfg.SystemFetcher, appTemplateSvc, intSysSvc, webhookSvc)
	if err := dataLoader.LoadData(ctx, os.ReadDir, os.ReadFile); err != nil {
		return nil, err
	}

	if err := loadSystemsSynchronizationTimestamps(ctx, tx, systemsSyncSvc); err != nil {
		return nil, errors.Wrap(err, "failed while loading systems synchronization timestamps")
	}

	var placeholdersMapping []systemfetcher.PlaceholderMapping
	if err := json.Unmarshal([]byte(cfg.TemplateConfig.PlaceholderToSystemKeyMappings), &placeholdersMapping); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling placeholders mapping")
	}

	templateRenderer, err := systemfetcher.NewTemplateRenderer(appTemplateSvc, appConverter, cfg.TemplateConfig.OverrideApplicationInput, placeholdersMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating template renderer")
	}

	if err := calculateTemplateMappings(ctx, cfg, tx, appTemplateSvc, placeholdersMapping, templateRenderer); err != nil {
		return nil, errors.Wrap(err, "failed while calculating application templates mappings")
	}

	return systemfetcher.NewSystemFetcher(tx, tenantSvc, appSvc, systemsSyncSvc, tenantBusinessTypeSvc, templateRenderer, systemsAPIClient, directorClient, cfg.SystemFetcher), nil
}

func createAndRunConfigProvider(ctx context.Context, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	err := provider.Load()
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "error on loading configuration file"))
	}
	executor.NewPeriodic(cfg.ConfigurationFileReload, func(ctx context.Context) {
		if err = provider.Load(); err != nil {
			if err != nil {
				log.D().Fatal(errors.Wrap(err, "error from Reloader watch"))
			}
		}
		log.C(ctx).Infof("Successfully reloaded configuration file.")
	}).Run(ctx)

	return provider
}

func calculateTemplateMappings(ctx context.Context, cfg config, transact persistence.Transactioner, appTemplateSvc apptemplate.ApplicationTemplateService, placeholdersMapping []systemfetcher.PlaceholderMapping, renderer systemfetcher.TemplateRenderer) error {
	applicationTemplates := make(map[systemfetcher.TemplateMappingKey]systemfetcher.TemplateMapping)

	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	appTemplates, err := appTemplateSvc.ListByFilters(ctx, []*labelfilter.LabelFilter{labelfilter.NewForKey(cfg.TemplateConfig.LabelFilter)})
	if err != nil {
		return errors.Wrapf(err, "while listing application templates by label filter %q", cfg.TemplateConfig.LabelFilter)
	}

	selectFilterProperties := make(map[string]bool, 0)
	for _, appTemplate := range appTemplates {
		templateMappingKey := systemfetcher.TemplateMappingKey{}

		labels, err := appTemplateSvc.ListLabels(ctx, appTemplate.ID)
		if err != nil {
			return errors.Wrapf(err, "while listing labels for application template with ID %q", appTemplate.ID)
		}

		regionModel, hasRegionLabel := labels[selfregmanager.RegionLabel]
		if hasRegionLabel {
			regionValue, ok := regionModel.Value.(string)
			if !ok {
				return errors.Errorf("%s label for Application Template with ID %s is not a string", selfregmanager.RegionLabel, appTemplate.ID)
			}

			templateMappingKey.Region = regionValue
		}

		systemRoleModel := labels[cfg.TemplateConfig.LabelFilter]
		appTemplateLblFilterArr, ok := systemRoleModel.Value.([]interface{})
		if !ok {
			continue
		}

		for _, systemRoleValue := range appTemplateLblFilterArr {
			systemRoleStrValue, ok := systemRoleValue.(string)
			if !ok {
				continue
			}

			templateMappingKey.Label = systemRoleStrValue

			applicationTemplates[templateMappingKey] = systemfetcher.TemplateMapping{AppTemplate: appTemplate, Labels: labels, Renderer: renderer}
		}

		addPropertiesFromAppTemplatePlaceholders(selectFilterProperties, appTemplate.Placeholders)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	systemfetcher.ApplicationTemplates = applicationTemplates
	systemfetcher.SelectFilter = createSelectFilter(selectFilterProperties, placeholdersMapping)
	systemfetcher.ApplicationTemplateLabelFilter = cfg.TemplateConfig.LabelFilter
	systemfetcher.SystemSourceKey = cfg.APIConfig.SystemSourceKey
	return nil
}

func loadSystemsSynchronizationTimestamps(ctx context.Context, transact persistence.Transactioner, systemSyncSvc systemfetcher.SystemsSyncService) error {
	systemSynchronizationTimestamps := make(map[string]map[string]systemfetcher.SystemSynchronizationTimestamp, 0)

	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	syncTimestamps, err := systemSyncSvc.List(ctx)
	if err != nil {
		return err
	}

	for _, s := range syncTimestamps {
		currentTimestamp := systemfetcher.SystemSynchronizationTimestamp{
			ID:                s.ID,
			LastSyncTimestamp: s.LastSyncTimestamp,
		}

		if _, ok := systemSynchronizationTimestamps[s.TenantID]; !ok {
			systemSynchronizationTimestamps[s.TenantID] = make(map[string]systemfetcher.SystemSynchronizationTimestamp, 0)
		}

		systemSynchronizationTimestamps[s.TenantID][s.ProductID] = currentTimestamp
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	systemfetcher.SystemSynchronizationTimestamps = systemSynchronizationTimestamps

	return nil
}

func getTopParentFromJSONPath(jsonPath string) string {
	prefix := "$."
	infix := "."

	topParent := strings.TrimPrefix(jsonPath, prefix)
	firstInfixIndex := strings.Index(topParent, infix)
	if firstInfixIndex == -1 {
		return topParent
	}

	return topParent[:firstInfixIndex]
}

func addPropertiesFromAppTemplatePlaceholders(selectFilterProperties map[string]bool, placeholders []model.ApplicationTemplatePlaceholder) {
	for _, placeholder := range placeholders {
		if placeholder.JSONPath != nil && len(*placeholder.JSONPath) > 0 {
			topParent := getTopParentFromJSONPath(*placeholder.JSONPath)
			if _, exists := selectFilterProperties[topParent]; !exists {
				selectFilterProperties[topParent] = true
			}
		}
	}
}

func createSelectFilter(selectFilterProperties map[string]bool, placeholdersMapping []systemfetcher.PlaceholderMapping) []string {
	selectFilter := make([]string, 0)

	for _, pm := range placeholdersMapping {
		topParent := getTopParentFromJSONPath(pm.SystemKey)
		if _, exists := selectFilterProperties[topParent]; !exists {
			selectFilterProperties[topParent] = true
		}
	}

	for property := range selectFilterProperties {
		selectFilter = append(selectFilter, property)
	}

	return selectFilter
}
