package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/auth"
	httputilpkg "github.com/kyma-incubator/compass/components/director/pkg/http"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address           string `envconfig:"default=127.0.0.1:8080"`
	AggregatorRootAPI string `envconfig:"APP_ROOT_API,default=/ord-aggregator"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	SecurityConfig securityConfig
	Database       persistence.DatabaseConfig
	Log            log.Config
	Features       features.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	ClientTimeout     time.Duration `envconfig:"default=120s"`
	SkipSSLValidation bool          `envconfig:"default=false"`

	RetryConfig                    retry.Config
	CertLoaderConfig               certloader.Config
	GlobalRegistryConfig           ord.GlobalRegistryConfig
	ElectionConfig                 cronjob.ElectionConfig
	ORDAggregatorJobSchedulePeriod time.Duration `envconfig:"APP_ORD_AGGREGATOR_JOB_SCHEDULE_PERIOD,default=60m"`
	IsORDAggregatorJobScheduleable bool          `envconfig:"APP_ORD_AGGREGATOR_JOB_IS_SCHEDULABLE,default=true"`

	MaxParallelWebhookProcessors       int `envconfig:"APP_MAX_PARALLEL_WEBHOOK_PROCESSORS,default=1"`
	MaxParallelDocumentsPerApplication int `envconfig:"APP_MAX_PARALLEL_DOCUMENTS_PER_APPLICATION"`
	MaxParallelSpecificationProcessors int `envconfig:"APP_MAX_PARALLEL_SPECIFICATION_PROCESSORS,default=100"`

	OrdWebhookPartialProcessURL     string `envconfig:"optional,APP_ORD_WEBHOOK_PARTIAL_PROCESS_URL"`
	OrdWebhookPartialProcessMaxDays int    `envconfig:"APP_ORD_WEBHOOK_PARTIAL_PROCESS_MAX_DAYS,default=0"`
	OrdWebhookPartialProcessing     bool   `envconfig:"APP_ORD_WEBHOOK_PARTIAL_PROCESSING,default=false"`

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`

	MetricsConfig ord.MetricsConfig
}

type securityConfig struct {
	JwksEndpoint        string        `envconfig:"APP_JWKS_ENDPOINT"`
	JWKSSyncPeriod      time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	AggregatorSyncScope string        `envconfig:"APP_ORD_AGGREGATOR_SYNC_SCOPE,default=ord_aggregator:sync"`
}

// TenantService is responsible for service-layer tenant operations
type TenantService interface {
	GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error)
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

	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	exitOnError(err, "failed while unmarshalling ord webhook mappings")

	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	httpClient := &http.Client{
		Timeout: cfg.ClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
	}

	securedHTTPClient := httputil.PrepareHTTPClientWithSSLValidation(cfg.ClientTimeout, cfg.SkipSSLValidation)
	mtlsClient := httputil.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExternalClientCertSecretName)
	extSvcMtlsClient := httputil.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExtSvcClientCertSecretName)

	accessStrategyExecutorProviderWithTenant := accessstrategy.NewExecutorProviderWithTenant(certCache, ctxTenantProvider, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName)
	accessStrategyExecutorProviderWithoutTenant := accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName)
	retryHTTPExecutor := retry.NewHTTPExecutor(&cfg.RetryConfig)

	ordAggregator := createORDAggregatorSvc(cfgProvider, cfg, transact, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient, accessStrategyExecutorProviderWithTenant, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor, ordWebhookMapping)

	jwtHTTPClient := &http.Client{
		Transport: httputilpkg.NewCorrelationIDTransport(httputilpkg.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	handler := initHandler(ctx, jwtHTTPClient, ordAggregator, cfg, transact)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	if cfg.IsORDAggregatorJobScheduleable {
		go func() {
			if err := startSyncORDDocumentsJob(ctx, ordAggregator, cfg); err != nil {
				log.C(ctx).WithError(err).Error("Failed to start sync ORD documents cronjob. Stopping app...")
			}
			cancel()
		}()
	}

	runMainSrv()
}

func startSyncORDDocumentsJob(ctx context.Context, ordAggregator *ord.Service, cfg config) error {
	resyncJob := cronjob.CronJob{
		Name: "SyncORDDocuments",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Infof("Starting ORD documents aggregation...")
			if err := ordAggregator.SyncORDDocuments(ctx, cfg.MetricsConfig); err != nil {
				log.C(jobCtx).WithError(err).Errorf("error occurred while syncing Open Resource Discovery Documents")
			}
			log.C(jobCtx).Infof("ORD documents aggregation finished.")
		},
		SchedulePeriod: cfg.ORDAggregatorJobSchedulePeriod,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, resyncJob)
}

func ctxTenantProvider(ctx context.Context) (string, error) {
	localTenantID, err := tenant.LoadLocalTenantIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	return localTenantID, nil
}

func createORDAggregatorSvc(cfgProvider *configprovider.Provider, config config, transact persistence.Transactioner, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient *http.Client, accessStrategyExecutorProviderWithTenant *accessstrategy.Provider, accessStrategyExecutorProviderWithoutTenant *accessstrategy.Provider, retryHTTPExecutor *retry.HTTPExecutor, ordWebhookMapping []application.ORDWebhookMapping) *ord.Service {
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
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	pkgConverter := ordpackage.NewConverter()
	productConverter := product.NewConverter()
	vendorConverter := ordvendor.NewConverter()
	tombstoneConverter := tombstone.NewConverter()
	runtimeConverter := runtime.NewConverter(webhookConverter)
	bundleReferenceConv := bundlereferences.NewConverter()
	runtimeContextConv := runtimectx.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	assignmentConv := scenarioassignment.NewConverter()

	runtimeRepo := runtime.NewRepository(runtimeConverter)
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
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	bundleRepo := bundleutil.NewRepository(bundleConverter)
	pkgRepo := ordpackage.NewRepository(pkgConverter)
	productRepo := product.NewRepository(productConverter)
	vendorRepo := ordvendor.NewRepository(vendorConverter)
	tombstoneRepo := tombstone.NewRepository(tombstoneConverter)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	tenantRepo := tenant.NewRepository(tenant.NewConverter())

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewServiceWithRetry(fetchRequestRepo, httpClient, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)
	formationAssignmentConv := formationassignment.NewConverter()
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, formationRepo, labelRepo)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, config.Features.RuntimeTypeLabelKey, config.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc)
	formationSvc := formation.NewService(transact, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, config.Features.RuntimeTypeLabelKey, config.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, config.SelfRegisterDistinguishLabelKey, ordWebhookMapping)
	packageSvc := ordpackage.NewService(pkgRepo, uidSvc)
	productSvc := product.NewService(productRepo, uidSvc)
	vendorSvc := ordvendor.NewService(vendorRepo, uidSvc)
	tombstoneSvc := tombstone.NewService(tombstoneRepo, uidSvc)

	clientConfig := ord.NewClientConfig(config.MaxParallelDocumentsPerApplication)

	ordClientWithTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithTenant)
	ordClientWithoutTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithoutTenant)

	globalRegistrySvc := ord.NewGlobalRegistryService(transact, config.GlobalRegistryConfig, vendorSvc, productSvc, ordClientWithoutTenantExecutor)

	ordConfig := ord.NewServiceConfig(config.MaxParallelWebhookProcessors, config.MaxParallelSpecificationProcessors, config.OrdWebhookPartialProcessMaxDays, config.OrdWebhookPartialProcessURL, config.OrdWebhookPartialProcessing)
	return ord.NewAggregatorService(ordConfig, transact, appSvc, webhookSvc, bundleSvc, bundleReferenceSvc, apiSvc, eventAPISvc, specSvc, fetchRequestSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, ordClientWithTenantExecutor)
}

func createAndRunConfigProvider(ctx context.Context, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	err := provider.Load()
	exitOnError(err, "Error on loading configuration file")
	executor.NewPeriodic(cfg.ConfigurationFileReload, func(ctx context.Context) {
		if err := provider.Load(); err != nil {
			exitOnError(err, "Error from Reloader watch")
		}
		log.C(ctx).Infof("Successfully reloaded configuration file.")
	}).Run(ctx)

	return provider
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring ord aggregator handler")

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

func initHandler(ctx context.Context, httpClient *http.Client, svc *ord.Service, cfg config, transact persistence.Transactioner) http.Handler {
	const (
		healthzEndpoint   = "/healthz"
		readyzEndpoint    = "/readyz"
		aggregateEndpoint = "/aggregate"
	)
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.AggregatorRootAPI+healthzEndpoint, cfg.AggregatorRootAPI+readyzEndpoint))

	handler := ord.NewORDAggregatorHTTPHandler(svc, cfg.MetricsConfig)
	apiRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, apiRouter, cfg, cfg.SecurityConfig.AggregatorSyncScope)
	configureTenantContextMiddleware(apiRouter, transact)
	apiRouter.HandleFunc(aggregateEndpoint, handler.AggregateORDData).Methods(http.MethodPost)

	healthCheckRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func configureAuthMiddleware(ctx context.Context, httpClient *http.Client, router *mux.Router, cfg config, requiredScopes ...string) {
	scopeValidator := claims.NewScopesValidator(requiredScopes)
	middleware := authenticator.New(httpClient, cfg.SecurityConfig.JwksEndpoint, cfg.SecurityConfig.AllowJWTSigningNone, "", scopeValidator)
	router.Use(middleware.Handler())

	log.C(ctx).Infof("JWKS synchronization enabled. Sync period: %v", cfg.SecurityConfig.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.SecurityConfig.JWKSSyncPeriod, func(ctx context.Context) {
		if err := middleware.SynchronizeJWKS(ctx); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)
}

func configureTenantContextMiddleware(apiRouter *mux.Router, transact persistence.Transactioner) {
	tenantRepo := tenant.NewRepository(tenant.NewConverter())
	uidSvc := uid.NewService()
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

	apiRouter.Use(tenantContextHandler(transact, tenantSvc))
}

func tenantContextHandler(transact persistence.Transactioner, tenantSvc TenantService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			tx, err := transact.Begin()
			if err != nil {
				apperrors.WriteAppError(ctx, w, err, http.StatusInternalServerError)
				return
			}
			defer transact.RollbackUnlessCommitted(ctx, tx)

			ctx = persistence.SaveToContext(ctx, tx)

			tntFromCtx, err := tenant.LoadFromContext(ctx)
			if err != nil {
				apperrors.WriteAppError(ctx, w, err, http.StatusUnauthorized)
				return
			}

			tnt, err := tenantSvc.GetTenantByExternalID(ctx, tntFromCtx)
			if err != nil {
				apperrors.WriteAppError(ctx, w, err, http.StatusInternalServerError)
				return
			}

			if err := tx.Commit(); err != nil {
				apperrors.WriteAppError(ctx, w, err, http.StatusInternalServerError)
				return
			}

			ctx = tenant.SaveToContext(ctx, tnt.ID, tnt.ExternalTenant)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
