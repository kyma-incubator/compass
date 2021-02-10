package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/header"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/packagetobundles"

	"github.com/kyma-incubator/compass/components/director/internal/authnmappinghandler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/pkg/scenario"

	"github.com/kyma-incubator/compass/components/director/internal/error_presenter"

	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/kyma-incubator/compass/components/director/internal/panic_handler"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-incubator/compass/components/director/internal/metrics"

	"github.com/dlmiddlecote/sqlstats"

	mp_authenticator "github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-incubator/compass/components/director/internal/domain"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/runtimemapping"
	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

const envPrefix = "APP"

type config struct {
	Address         string `envconfig:"default=127.0.0.1:3000"`
	InternalAddress string `envconfig:"default=127.0.0.1:3002"`
	AppURL          string `envconfig:"APP_URL"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Database                      persistence.DatabaseConfig
	APIEndpoint                   string `envconfig:"default=/graphql"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping"`
	OperationEndpoint             string `envconfig:"default=/operations"`
	PlaygroundAPIEndpoint         string `envconfig:"default=/graphql"`
	ConfigurationFile             string
	ConfigurationFileReload       time.Duration `envconfig:"default=1m"`

	Log log.Config

	MetricsAddress string `envconfig:"default=127.0.0.1:3001"`

	JWKSEndpoint        string        `envconfig:"default=file://hack/default-jwks.json"`
	JWKSSyncPeriod      time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone bool          `envconfig:"default=true"`

	RuntimeJWKSCachePeriod time.Duration `envconfig:"default=5m"`

	StaticUsersSrc    string `envconfig:"default=/data/static-users.yaml"`
	StaticGroupsSrc   string `envconfig:"default=/data/static-groups.yaml"`
	PairingAdapterSrc string `envconfig:"optional"`

	OneTimeToken onetimetoken.Config
	OAuth20      oauth20.Config

	Features features.Config

	ProtectedLabelPattern string `envconfig:"default=.*_defaultEventing"`
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
	metricsCollector := metrics.NewCollector()
	dbStatsCollector := sqlstats.NewStatsCollector("director", transact)
	prometheus.MustRegister(metricsCollector, dbStatsCollector)

	pairingAdapters, err := getPairingAdaptersMapping(ctx, cfg.PairingAdapterSrc)
	exitOnError(err, "Error while reading Pairing Adapters Configuration")

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}

	appRepo := applicationRepo()

	gqlCfg := graphql.Config{
		Resolvers: domain.NewRootResolver(
			&normalizer.DefaultNormalizator{},
			transact,
			cfgProvider,
			cfg.OneTimeToken,
			cfg.OAuth20,
			pairingAdapters,
			cfg.Features,
			metricsCollector,
			httpClient,
			cfg.ProtectedLabelPattern,
		),
		Directives: graphql.DirectiveRoot{
			Async: operation.NewDirective(transact, webhookService().List, func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
				return appRepo.GetByID(ctx, tenantID, resourceID)
			}, tenant.LoadFromContext, operation.DefaultScheduler{}).HandleOperation,
			HasScenario: scenario.NewDirective(transact, label.NewRepository(label.NewConverter()), bundleRepo(), bundleInstanceAuthRepo()).HasScenario,
			HasScopes:   scope.NewDirective(cfgProvider).VerifyScopes,
			Validate:    inputvalidation.NewDirective().Validate,
		},
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	logger.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	authMiddleware := mp_authenticator.New(cfg.JWKSEndpoint, cfg.AllowJWTSigningNone)

	if cfg.JWKSSyncPeriod != 0 {
		logger.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
		periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
			err := authMiddleware.SynchronizeJWKS(ctx)
			if err != nil {
				logger.WithError(err).Error("An error has occurred while synchronizing JWKS")
			}
		})
		go periodicExecutor.Run(ctx)
	}

	packageToBundlesMiddleware := packagetobundles.NewHandler(transact)

	statusMiddleware := statusupdate.New(transact, statusupdate.NewRepository())

	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))

	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(), header.AttachHeadersToContext())
	presenter := error_presenter.NewPresenter(uid.NewService())

	operationMiddleware := operation.NewMiddleware(cfg.AppURL)

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.Use(packageToBundlesMiddleware.Handler())
	gqlAPIRouter.Use(statusMiddleware.Handler())
	gqlAPIRouter.HandleFunc("", metricsCollector.GraphQLHandlerWithInstrumentation(handler.GraphQL(executableSchema,
		handler.RequestMiddleware(operationMiddleware.ExtensionHandler),
		handler.ErrorPresenter(presenter.Do),
		handler.RecoverFunc(panic_handler.RecoverFn))))

	logger.Infof("Registering Tenant Mapping endpoint on %s...", cfg.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(transact, authenticators, cfg.StaticUsersSrc, cfg.StaticGroupsSrc, cfgProvider)
	exitOnError(err, "Error while configuring tenant mapping handler")

	mainRouter.HandleFunc(cfg.TenantMappingEndpoint, tenantMappingHandlerFunc)

	logger.Infof("Registering Runtime Mapping endpoint on %s...", cfg.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc, err := getRuntimeMappingHandlerFunc(transact, cfg.JWKSSyncPeriod, ctx, cfg.Features.DefaultScenarioEnabled, cfg.ProtectedLabelPattern)

	exitOnError(err, "Error while configuring runtime mapping handler")

	mainRouter.HandleFunc(cfg.RuntimeMappingEndpoint, runtimeMappingHandlerFunc)

	logger.Infof("Registering Authentication Mapping endpoint on %s...", cfg.AuthenticationMappingEndpoint)
	authnMappingHandlerFunc := authnmappinghandler.NewHandler(oathkeeper.NewReqDataParser(), httpClient, authnmappinghandler.DefaultTokenVerifierProvider, authenticators)

	mainRouter.HandleFunc(cfg.AuthenticationMappingEndpoint, authnMappingHandlerFunc.ServeHTTP)

	operationHandler := operation.NewHandler(transact, func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
		return appRepo.GetByID(ctx, tenantID, resourceID)
	}, tenant.LoadFromContext)

	operationsAPIRouter := mainRouter.PathPrefix(cfg.OperationEndpoint).Subrouter()
	operationsAPIRouter.Use(authMiddleware.Handler())
	operationsAPIRouter.HandleFunc("", operationHandler.ServeHTTP)

	logger.Infof("Registering readiness endpoint...")
	mainRouter.HandleFunc("/readyz", healthz.NewReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc("/healthz", healthz.NewLivenessHandler(transact))

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	internalRouter := mux.NewRouter()
	operationUpdaterHandler := operation.NewUpdateOperationHandler(transact, map[resource.Type]operation.ResourceUpdaterFunc{
		resource.Application: func(ctx context.Context, id string, ready bool, errorMsg string) error {
			app, err := appRepo.GetGlobalByID(ctx, id)
			if err != nil {
				return err
			}
			app.Ready = ready
			app.Error = &errorMsg
			return appRepo.Update(ctx, app)
		},
	}, map[resource.Type]operation.ResourceDeleterFunc{
		resource.Application: func(ctx context.Context, id string) error {
			return appRepo.DeleteGlobal(ctx, id)
		},
	})
	internalOperationsAPIRouter := internalRouter.PathPrefix(cfg.OperationEndpoint).Subrouter()
	internalOperationsAPIRouter.HandleFunc("", operationUpdaterHandler.ServeHTTP)

	runMetricsSrv, shutdownMetricsSrv := createServer(ctx, cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg.Address, mainRouter, "main", cfg.ServerTimeout)
	runInternalSrv, shutdownInternalSrv := createServer(ctx, cfg.InternalAddress, internalRouter, "internal", cfg.ServerTimeout)

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
		shutdownInternalSrv()
		shutdownMainSrv()
	}()

	go runMetricsSrv()
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

func getTenantMappingHandlerFunc(transact persistence.Transactioner, authenticators []authenticator.Config, staticUsersSrc string, staticGroupsSrc string, cfgProvider *configprovider.Provider) (func(writer http.ResponseWriter, request *http.Request), error) {
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
		tenantmapping.AuthenticatorObjectContextProvider: tenantmapping.NewAuthenticatorContextProvider(tenantRepo),
	}
	reqDataParser := oathkeeper.NewReqDataParser()

	return tenantmapping.NewHandler(authenticators, reqDataParser, transact, objectContextProviders).ServeHTTP, nil
}

func getRuntimeMappingHandlerFunc(transact persistence.Transactioner, cachePeriod time.Duration, ctx context.Context, defaultScenarioEnabled bool, protectedLabelPattern string) (func(writer http.ResponseWriter, request *http.Request), error) {
	uidSvc := uid.NewService()

	labelConv := label.NewConverter()
	labelRepo := label.NewRepository(labelConv)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	scenariosSvc := labeldef.NewScenariosService(labelDefRepo, uidSvc, defaultScenarioEnabled)
	labelUpsertSvc := label.NewLabelUpsertService(labelRepo, labelDefRepo, uidSvc)
	runtimeRepo := runtime.NewRepository()

	scenarioAssignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignmentConv)
	scenarioAssignmentEngine := scenarioassignment.NewEngine(labelUpsertSvc, labelRepo, scenarioAssignmentRepo)

	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelUpsertSvc, uidSvc, scenarioAssignmentEngine, protectedLabelPattern)

	tenantConv := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConv)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

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
		tenantSvc).ServeHTTP, nil
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
			log.C(ctx).WithError(err).Errorf("An error has occurred with %s HTTP server when ListenAndServe.", name)
		}
	}

	shutdownFn := func() {
		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while shutting down HTTP server %s.", name)
		}
	}

	return runFn, shutdownFn
}

func bundleInstanceAuthRepo() bundleinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return bundleinstanceauth.NewRepository(bundleinstanceauth.NewConverter(authConverter))
}

func bundleRepo() mp_bundle.BundleRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)

	return mp_bundle.NewRepository(mp_bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
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
	bundleConverter := mp_bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)

	appConverter := application.NewConverter(webhookConverter, bundleConverter)

	return application.NewRepository(appConverter)
}

func webhookService() webhook.WebhookService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()

	webhookConverter := webhook.NewConverter(authConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)

	return webhook.NewService(webhookRepo, uidSvc)
}
