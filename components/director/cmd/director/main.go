package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/internal/info"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dlmiddlecote/sqlstats"
	"github.com/gorilla/mux"
	mp_authenticator "github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/authnmappinghandler"
	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/schema"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	errorpresenter "github.com/kyma-incubator/compass/components/director/internal/error_presenter"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/packagetobundles"
	panichandler "github.com/kyma-incubator/compass/components/director/internal/panic_handler"
	"github.com/kyma-incubator/compass/components/director/internal/runtimemapping"
	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/k8s"
	panicrecovery "github.com/kyma-incubator/compass/components/director/pkg/panic_recovery"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/scenario"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	directorTime "github.com/kyma-incubator/compass/components/director/pkg/time"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vrischmann/envconfig"
	cr "sigs.k8s.io/controller-runtime"
)

const envPrefix = "APP"

type config struct {
	Address         string `envconfig:"default=127.0.0.1:3000"`
	HydratorAddress string `envconfig:"default=127.0.0.1:8080"`

	InternalAddress string `envconfig:"default=127.0.0.1:3002"`
	AppURL          string `envconfig:"APP_URL"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Database                      persistence.DatabaseConfig
	APIEndpoint                   string `envconfig:"default=/graphql"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping/{authenticator}"`
	OperationPath                 string `envconfig:"default=/operation"`
	LastOperationPath             string `envconfig:"default=/last_operation"`
	PlaygroundAPIEndpoint         string `envconfig:"default=/graphql"`
	ConfigurationFile             string
	ConfigurationFileReload       time.Duration `envconfig:"default=1m"`

	Log log.Config

	MetricsAddress string `envconfig:"default=127.0.0.1:3003"`
	MetricsConfig  metrics.Config

	JWKSEndpoint          string        `envconfig:"default=file://hack/default-jwks.json"`
	JWKSSyncPeriod        time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone   bool          `envconfig:"default=false"`
	ClientIDHTTPHeaderKey string        `envconfig:"default=client_user,APP_CLIENT_ID_HTTP_HEADER"`

	RuntimeJWKSCachePeriod time.Duration `envconfig:"default=5m"`

	StaticUsersSrc    string `envconfig:"default=/data/static-users.yaml"`
	StaticGroupsSrc   string `envconfig:"default=/data/static-groups.yaml"`
	PairingAdapterSrc string `envconfig:"optional"`

	OneTimeToken onetimetoken.Config
	OAuth20      oauth20.Config

	Features features.Config

	SelfRegConfig runtime.SelfRegConfig

	OperationsNamespace string `envconfig:"default=compass-system"`

	DisableAsyncMode bool `envconfig:"default=false"`

	HealthConfig healthz.Config `envconfig:"APP_HEALTH_CONFIG_INDICATORS"`

	ReadyConfig healthz.ReadyConfig

	InfoConfig info.Config

	DataloaderMaxBatch int           `envconfig:"default=200"`
	DataloaderWait     time.Duration `envconfig:"default=10ms"`

	CertLoaderConfig certloader.Config
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(err, "Error while loading app config")

	authenticators, err := authenticator.InitFromEnv(envPrefix)
	exitOnError(err, "Failed to retrieve authenticators config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	logger.Infof("Registering metrics collectors...")
	metricsCollector := metrics.NewCollector(cfg.MetricsConfig)
	dbStatsCollector := sqlstats.NewStatsCollector("director", transact)
	prometheus.MustRegister(metricsCollector, dbStatsCollector)

	pairingAdapters, err := getPairingAdaptersMapping(ctx, cfg.PairingAdapterSrc)
	exitOnError(err, "Error while reading Pairing Adapters configuration")

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}
	cfg.SelfRegConfig.ClientTimeout = cfg.ClientTimeout

	internalHTTPClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(http.DefaultTransport)),
	}

	appRepo := applicationRepo()

	adminURL, err := url.Parse(cfg.OAuth20.URL)
	exitOnError(err, "Error while parsing Hydra URL")

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	accessStrategyExecutorProvider := accessstrategy.NewDefaultExecutorProvider(certCache)

	rootResolver := domain.NewRootResolver(
		&normalizer.DefaultNormalizator{},
		transact,
		cfgProvider,
		cfg.OneTimeToken,
		cfg.OAuth20,
		pairingAdapters,
		cfg.Features,
		metricsCollector,
		httpClient,
		internalHTTPClient,
		cfg.SelfRegConfig,
		cfg.OneTimeToken.Length,
		adminURL,
		accessStrategyExecutorProvider,
		certCache,
	)

	gqlCfg := graphql.Config{
		Resolvers: rootResolver,
		Directives: graphql.DirectiveRoot{
			Async:       getAsyncDirective(ctx, cfg, transact, appRepo),
			HasScenario: scenario.NewDirective(transact, label.NewRepository(label.NewConverter()), bundleRepo(), bundleInstanceAuthRepo()).HasScenario,
			HasScopes:   scope.NewDirective(cfgProvider, &scope.HasScopesErrorProvider{}).VerifyScopes,
			Sanitize:    scope.NewDirective(cfgProvider, &scope.SanitizeErrorProvider{}).VerifyScopes,
			Validate:    inputvalidation.NewDirective().Validate,
		},
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)
	claimsValidator := claims.NewValidator(runtimeSvc(cfg), intSystemSvc(), cfg.Features.SubscriptionProviderLabelKey, cfg.Features.ConsumerSubaccountIDsLabelKey)

	logger.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	authMiddleware := mp_authenticator.New(cfg.JWKSEndpoint, cfg.AllowJWTSigningNone, cfg.ClientIDHTTPHeaderKey, claimsValidator)

	if cfg.JWKSSyncPeriod != 0 {
		logger.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
		periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
			err := authMiddleware.SynchronizeJWKS(ctx)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
			}
		})
		go periodicExecutor.Run(ctx)
	}

	packageToBundlesMiddleware := packagetobundles.NewHandler(transact)

	statusMiddleware := statusupdate.New(transact, statusupdate.NewRepository())

	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", playground.Handler("Dataloader", cfg.PlaygroundAPIEndpoint))

	mainRouter.Use(panicrecovery.NewPanicRecoveryMiddleware(), correlation.AttachCorrelationIDToContext(), log.RequestLogger(), header.AttachHeadersToContext())
	presenter := errorpresenter.NewPresenter(uid.NewService())

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.Use(packageToBundlesMiddleware.Handler())
	gqlAPIRouter.Use(statusMiddleware.Handler())
	gqlAPIRouter.Use(dataloader.HandlerBundle(rootResolver.BundlesDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerAPIDef(rootResolver.APIDefinitionsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerEventDef(rootResolver.EventDefinitionsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerDocument(rootResolver.DocumentsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestAPIDef(rootResolver.FetchRequestAPIDefDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestEventDef(rootResolver.FetchRequestEventDefDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestDocument(rootResolver.FetchRequestDocumentDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))

	operationMiddleware := operation.NewMiddleware(cfg.AppURL + cfg.LastOperationPath)

	gqlServ := handler.NewDefaultServer(executableSchema)
	gqlServ.Use(log.NewGqlLoggingInterceptor())
	gqlServ.Use(operationMiddleware)
	gqlServ.SetErrorPresenter(presenter.Do)
	gqlServ.SetRecoverFunc(panichandler.RecoverFn)

	gqlAPIRouter.HandleFunc("", metricsCollector.GraphQLHandlerWithInstrumentation(gqlServ))

	logger.Infof("Registering Tenant Mapping endpoint on %s...", cfg.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(transact, authenticators, cfg.StaticUsersSrc, cfg.StaticGroupsSrc, cfgProvider, metricsCollector)
	exitOnError(err, "Error while configuring tenant mapping handler")

	mainRouter.HandleFunc(cfg.TenantMappingEndpoint, tenantMappingHandlerFunc)

	logger.Infof("Registering Runtime Mapping endpoint on %s...", cfg.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc := getRuntimeMappingHandlerFunc(ctx, transact, cfg.JWKSSyncPeriod, cfg.Features.DefaultScenarioEnabled, cfg.Features.ProtectedLabelPattern, cfg.Features.ImmutableLabelPattern)

	mainRouter.HandleFunc(cfg.RuntimeMappingEndpoint, runtimeMappingHandlerFunc)

	logger.Infof("Registering Authentication Mapping endpoint on %s...", cfg.AuthenticationMappingEndpoint)
	authnMappingHandlerFunc := authnmappinghandler.NewHandler(oathkeeper.NewReqDataParser(), httpClient, authnmappinghandler.DefaultTokenVerifierProvider, authenticators)

	mainRouter.HandleFunc(cfg.AuthenticationMappingEndpoint, authnMappingHandlerFunc.ServeHTTP)

	operationHandler := operation.NewHandler(transact, func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
		return appRepo.GetByID(ctx, tenantID, resourceID)
	}, tenant.LoadFromContext)

	operationsAPIRouter := mainRouter.PathPrefix(cfg.LastOperationPath).Subrouter()
	operationsAPIRouter.Use(authMiddleware.Handler())
	operationsAPIRouter.HandleFunc("/{resource_type}/{resource_id}", operationHandler.ServeHTTP)

	operationUpdaterHandler := operation.NewUpdateOperationHandler(transact, map[resource.Type]operation.ResourceUpdaterFunc{
		resource.Application: appUpdaterFunc(appRepo),
	}, map[resource.Type]operation.ResourceDeleterFunc{
		resource.Application: func(ctx context.Context, id string) error {
			return appRepo.DeleteGlobal(ctx, id)
		},
	})

	internalRouter := mux.NewRouter()
	internalRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(), header.AttachHeadersToContext())
	internalOperationsAPIRouter := internalRouter.PathPrefix(cfg.OperationPath).Subrouter()
	internalOperationsAPIRouter.HandleFunc("", operationUpdaterHandler.ServeHTTP)

	oneTimeTokenService := tokenService(cfg, cfgProvider, httpClient, internalHTTPClient, pairingAdapters, accessStrategyExecutorProvider)
	hydratorHandler, err := PrepareHydratorHandler(cfg, systemAuthSvc(), transact, oneTimeTokenService, correlation.AttachCorrelationIDToContext(), log.RequestLogger())
	exitOnError(err, "Failed configuring hydrator handler")

	logger.Infof("Registering readiness endpoint...")
	schemaRepo := schema.NewRepository()
	ready := healthz.NewReady(transact, cfg.ReadyConfig, schemaRepo)
	mainRouter.HandleFunc("/readyz", healthz.NewReadinessHandler(ready))

	logger.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc("/livez", healthz.NewLivenessHandler())

	logger.Infof("Registering health endpoint...")
	health, err := healthz.New(ctx, cfg.HealthConfig)
	exitOnError(err, "Could not initialize health")
	health.RegisterIndicator(healthz.NewIndicator(healthz.DBIndicatorName, healthz.NewDBIndicatorFunc(transact))).Start()
	mainRouter.HandleFunc("/healthz", healthz.NewHealthHandler(health))

	logger.Infof("Registering info endpoint...")
	mainRouter.HandleFunc(cfg.InfoConfig.APIEndpoint, info.NewInfoHandler(ctx, cfg.InfoConfig))

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	runMetricsSrv, shutdownMetricsSrv := createServer(ctx, cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg.Address, mainRouter, "main", cfg.ServerTimeout)
	runHydratorSrv, shutdownHydratorSrv := createServer(ctx, cfg.HydratorAddress, hydratorHandler, "hydrator", cfg.ServerTimeout)
	runInternalSrv, shutdownInternalSrv := createServer(ctx, cfg.InternalAddress, internalRouter, "internal", cfg.ServerTimeout)

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
		shutdownInternalSrv()
		shutdownMainSrv()
		shutdownHydratorSrv()
	}()

	go runMetricsSrv()
	go runHydratorSrv()
	go runInternalSrv()
	runMainSrv()
}

func getPairingAdaptersMapping(ctx context.Context, filePath string) (map[string]string, error) {
	logger := log.C(ctx)

	if filePath == "" {
		logger.Infof("No configuration for pairing adapters")
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "while opening pairing adapter configuration file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			logger.Warnf("Got error on closing file with pairing adapters configuration: %v", err)
		}
	}()

	decoder := json.NewDecoder(file)
	out := map[string]string{}
	err = decoder.Decode(&out)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding file [%s] to map[string]string", filePath)
	}
	logger.Infof("Successfully read pairing adapters configuration")
	return out, nil
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

func getTenantMappingHandlerFunc(transact persistence.Transactioner, authenticators []authenticator.Config, staticUsersSrc string, staticGroupsSrc string, cfgProvider *configprovider.Provider, metricsCollector *metrics.Collector) (func(writer http.ResponseWriter, request *http.Request), error) {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	staticUsersRepo, err := tenantmapping.NewStaticUserRepository(staticUsersSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticUser repository instance")
	}

	staticGroupsRepo, err := tenantmapping.NewStaticGroupRepository(staticGroupsSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticGroup repository instance")
	}

	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)

	objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
		tenantmapping.UserObjectContextProvider:          tenantmapping.NewUserContextProvider(staticUsersRepo, staticGroupsRepo, tenantRepo),
		tenantmapping.SystemAuthObjectContextProvider:    tenantmapping.NewSystemAuthContextProvider(systemAuthSvc, cfgProvider, tenantRepo),
		tenantmapping.AuthenticatorObjectContextProvider: tenantmapping.NewAuthenticatorContextProvider(tenantRepo, authenticators),
		tenantmapping.CertServiceObjectContextProvider:   tenantmapping.NewCertServiceContextProvider(tenantRepo, cfgProvider),
		tenantmapping.TenantHeaderObjectContextProvider:  tenantmapping.NewAccessLevelContextProvider(tenantRepo),
	}
	reqDataParser := oathkeeper.NewReqDataParser()

	return tenantmapping.NewHandler(reqDataParser, transact, objectContextProviders, metricsCollector).ServeHTTP, nil
}

func getRuntimeMappingHandlerFunc(ctx context.Context, transact persistence.Transactioner, cachePeriod time.Duration, defaultScenarioEnabled bool, protectedLabelPattern string, immutableLabelPattern string) func(writer http.ResponseWriter, request *http.Request) {
	uidSvc := uid.NewService()

	assignmentConv := scenarioassignment.NewConverter()
	labelConv := label.NewConverter()
	labelRepo := label.NewRepository(labelConv)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	tenantRepo := tenant.NewRepository(tenant.NewConverter())
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, defaultScenarioEnabled)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	runtimeConv := runtime.NewConverter()
	runtimeRepo := runtime.NewRepository(runtimeConv)

	scenarioAssignmentEngine := scenarioassignment.NewEngine(labelSvc, labelRepo, scenarioAssignmentRepo, runtimeRepo)

	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)

	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelSvc, uidSvc, scenarioAssignmentEngine, tenantSvc, protectedLabelPattern, immutableLabelPattern)

	reqDataParser := oathkeeper.NewReqDataParser()

	jwksFetch := runtimemapping.NewJWKsFetch()
	jwksCache := runtimemapping.NewJWKsCache(jwksFetch, cachePeriod)
	tokenVerifier := runtimemapping.NewTokenVerifier(jwksCache)

	executor.NewPeriodic(1*time.Minute, func(ctx context.Context) {
		jwksCache.Cleanup(ctx)
	}).Run(ctx)

	return runtimemapping.NewHandler(
		reqDataParser,
		transact,
		tokenVerifier,
		runtimeSvc,
		tenantSvc).ServeHTTP
}

func createServer(ctx context.Context, address string, handler http.Handler, name string, timeout time.Duration) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, timeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: timeout,
	}

	runFn := func() {
		log.C(ctx).Infof("Running %s server on %s...", name, address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.C(ctx).WithError(err).Errorf("An error has occurred with %s HTTP server when ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while shutting down HTTP server %s: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func bundleInstanceAuthRepo() bundleinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return bundleinstanceauth.NewRepository(bundleinstanceauth.NewConverter(authConverter))
}

func bundleRepo() bundle.BundleRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)

	return bundle.NewRepository(bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
}

func applicationRepo() application.ApplicationRepository {
	authConverter := auth.NewConverter()

	versionConverter := version.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)

	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)

	webhookConverter := webhook.NewConverter(authConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)

	appConverter := application.NewConverter(webhookConverter, bundleConverter)

	return application.NewRepository(appConverter)
}

func webhookService() webhook.WebhookService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()

	webhookConverter := webhook.NewConverter(authConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)

	return webhook.NewService(webhookRepo, applicationRepo(), uidSvc)
}

func tokenService(cfg config, cfgProvider *configprovider.Provider, httpClient, internalHTTPClient *http.Client, pairingAdapters map[string]string, accessStrategyExecutorProvider *accessstrategy.Provider) oathkeeper.OneTimeTokenService {
	uidSvc := uid.NewService()
	assignmentConv := scenarioassignment.NewConverter()
	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	labelConverter := label.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	bundleReferenceConv := bundlereferences.NewConverter()

	intSysConverter := integrationsystem.NewConverter()
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)
	webhookConverter := webhook.NewConverter(authConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	packageConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, packageConverter)
	applicationRepo := application.NewRepository(appConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	runtimeConverter := runtime.NewConverter()
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelUpsertSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, cfg.Features.DefaultScenarioEnabled)
	bundleRepo := bundle.NewRepository(packageConverter)
	apiRepo := api.NewRepository(apiConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, accessStrategyExecutorProvider)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	specRepo := spec.NewRepository(specConverter)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	documentSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleSvc := bundle.NewService(bundleRepo, apiSvc, eventAPISvc, documentSvc, uidSvc)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelUpsertSvc, scenariosSvc, bundleSvc, uidSvc)
	timeService := directorTime.NewService()
	return onetimetoken.NewTokenService(systemAuthSvc, appSvc, appConverter, tenantSvc, internalHTTPClient, onetimetoken.NewTokenGenerator(cfg.OneTimeToken.Length), cfg.OneTimeToken, pairingAdapters, timeService)
}

func systemAuthSvc() oathkeeper.SystemAuthService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)
	return systemauth.NewService(systemAuthRepo, uidSvc)
}

// PrepareHydratorHandler missing godoc
func PrepareHydratorHandler(cfg config, tokenService oathkeeper.SystemAuthService, transact persistence.Transactioner, oneTimeTokenService oathkeeper.OneTimeTokenService, middlewares ...mux.MiddlewareFunc) (http.Handler, error) {
	validationHydrator := oathkeeper.NewValidationHydrator(tokenService, transact, oneTimeTokenService)

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/tokens/resolve", validationHydrator.ResolveConnectorTokenHeader)

	router.Use(middlewares...)

	handlerWithTimeout, err := timeouthandler.WithTimeout(router, cfg.ServerTimeout)
	if err != nil {
		return nil, err
	}

	return handlerWithTimeout, nil
}

func getAsyncDirective(ctx context.Context, cfg config, transact persistence.Transactioner, appRepo application.ApplicationRepository) func(context.Context, interface{}, gqlgen.Resolver, graphql.OperationType, *graphql.WebhookType, *string) (res interface{}, err error) {
	resourceFetcherFunc := func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
		return appRepo.GetByID(ctx, tenantID, resourceID)
	}

	scheduler, err := buildScheduler(ctx, cfg)
	exitOnError(err, "Error while creating operations scheduler")

	return operation.NewDirective(transact, webhookService().ListAllApplicationWebhooks, resourceFetcherFunc, appUpdaterFunc(appRepo), tenant.LoadFromContext, scheduler).HandleOperation
}

func buildScheduler(ctx context.Context, config config) (operation.Scheduler, error) {
	if config.DisableAsyncMode {
		log.C(ctx).Info("Async operations are disabled")
		return &operation.DisabledScheduler{}, nil
	}

	cfg, err := cr.GetConfig()
	exitOnError(err, "Failed to get cluster config for operations k8s client")

	cfg.Timeout = config.ClientTimeout
	k8sClient, err := client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	operationsK8sClient := k8sClient.Operations(config.OperationsNamespace)

	return k8s.NewScheduler(operationsK8sClient), nil
}

func appUpdaterFunc(appRepo application.ApplicationRepository) operation.ResourceUpdaterFunc {
	return func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
		app, err := appRepo.GetGlobalByID(ctx, id)
		if err != nil {
			return err
		}
		app.Status = &model.ApplicationStatus{
			Condition: appStatusCondition,
			Timestamp: time.Now(),
		}
		app.Ready = ready
		app.Error = errorMsg
		return appRepo.TechnicalUpdate(ctx, app)
	}
}

func runtimeSvc(cfg config) claims.RuntimeService {
	uidSvc := uid.NewService()

	rtConverter := runtime.NewConverter()
	rtRepo := runtime.NewRepository(rtConverter)

	lblRepo := label.NewRepository(label.NewConverter())

	assignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)

	tenantRepo := tenant.NewRepository(tenant.NewConverter())

	labelDefRepo := labeldef.NewRepository(labeldef.NewConverter())
	labelDefSvc := labeldef.NewService(labelDefRepo, lblRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, cfg.Features.DefaultScenarioEnabled)

	labelSvc := label.NewLabelService(lblRepo, labelDefRepo, uidSvc)

	scenarioAssignmentEngine := scenarioassignment.NewEngine(labelSvc, lblRepo, scenarioAssignmentRepo, rtRepo)

	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, lblRepo, labelSvc)

	return runtime.NewService(rtRepo, lblRepo, labelDefSvc, labelSvc, uidSvc, scenarioAssignmentEngine, tenantSvc, cfg.Features.ProtectedLabelPattern, cfg.Features.ImmutableLabelPattern)
}

func intSystemSvc() claims.IntegrationSystemService {
	intSysConverter := integrationsystem.NewConverter()
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	return integrationsystem.NewService(intSysRepo, uid.NewService())
}
