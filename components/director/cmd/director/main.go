package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/packageinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	"github.com/kyma-incubator/compass/components/director/pkg/scenario"

	"github.com/kyma-incubator/compass/components/director/internal/error_presenter"

	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/kyma-incubator/compass/components/director/internal/panic_handler"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kyma-incubator/compass/components/director/internal/metrics"

	"github.com/dlmiddlecote/sqlstats"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
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
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Database                persistence.DatabaseConfig
	APIEndpoint             string `envconfig:"default=/graphql"`
	TenantMappingEndpoint   string `envconfig:"default=/tenant-mapping"`
	RuntimeMappingEndpoint  string `envconfig:"default=/runtime-mapping"`
	PlaygroundAPIEndpoint   string `envconfig:"default=/graphql"`
	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

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
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	stopCh := signal.SetupChannel()
	cfgProvider := createAndRunConfigProvider(stopCh, cfg)

	log.Infof("Registering metrics collectors...")
	metricsCollector := metrics.NewCollector()
	dbStatsCollector := sqlstats.NewStatsCollector("director", transact)
	prometheus.MustRegister(metricsCollector, dbStatsCollector)

	pairingAdapters, err := getPairingAdaptersMapping(cfg.PairingAdapterSrc)
	exitOnError(err, "Error while reading Pairing Adapters Configuration")

	gqlCfg := graphql.Config{
		Resolvers: domain.NewRootResolver(
			transact,
			cfgProvider,
			cfg.OneTimeToken,
			cfg.OAuth20,
			pairingAdapters,
			cfg.Features,
			metricsCollector,
			cfg.ClientTimeout,
		),
		Directives: graphql.DirectiveRoot{
			HasScenario: scenario.NewDirective(label.NewRepository(label.NewConverter()), defaultPackageRepo(), defaultPackageInstanceAuthRepo()).HasScenario,
			HasScopes:   scope.NewDirective(cfgProvider).VerifyScopes,
			Validate:    inputvalidation.NewDirective().Validate,
		},
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)

	log.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	authMiddleware := authenticator.New(cfg.JWKSEndpoint, cfg.AllowJWTSigningNone)

	if cfg.JWKSSyncPeriod != 0 {
		log.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
		periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(stopCh <-chan struct{}) {
			err := authMiddleware.SynchronizeJWKS()
			if err != nil {
				log.Error(errors.Wrap(err, "while synchronizing JWKS"))
			}
		})
		go periodicExecutor.Run(stopCh)
	}

	statusMiddleware := statusupdate.New(transact, statusupdate.NewRepository(), log.New())

	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))

	contextEnricher := correlation.NewContextEnrichMiddleware()
	mainRouter.Use(contextEnricher.AttachCorrelationIDToContext)

	presenter := error_presenter.NewPresenter(log.StandardLogger(), uid.NewService())

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.Use(statusMiddleware.Handler())
	gqlAPIRouter.HandleFunc("", metricsCollector.GraphQLHandlerWithInstrumentation(handler.GraphQL(executableSchema,
		handler.ErrorPresenter(presenter.Do),
		handler.RecoverFunc(panic_handler.RecoverFn))))

	log.Infof("Registering Tenant Mapping endpoint on %s...", cfg.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(transact, cfg.StaticUsersSrc, cfg.StaticGroupsSrc, cfgProvider)
	exitOnError(err, "Error while configuring tenant mapping handler")

	mainRouter.HandleFunc(cfg.TenantMappingEndpoint, tenantMappingHandlerFunc)

	log.Infof("Registering Runtime Mapping endpoint on %s...", cfg.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc, err := getRuntimeMappingHandlerFunc(transact, cfg.JWKSSyncPeriod, stopCh, cfg.Features.DefaultScenarioEnabled)
	exitOnError(err, "Error while configuring runtime mapping handler")

	mainRouter.HandleFunc(cfg.RuntimeMappingEndpoint, runtimeMappingHandlerFunc)

	log.Infof("Registering readiness endpoint...")
	mainRouter.HandleFunc("/readyz", healthz.NewReadinessHandler())

	log.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc("/healthz", healthz.NewLivenessHandler(transact, log.StandardLogger()))

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	runMetricsSrv, shutdownMetricsSrv := createServer(cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	runMainSrv, shutdownMainSrv := createServer(cfg.Address, mainRouter, "main", cfg.ServerTimeout)

	go func() {
		<-stopCh
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
		shutdownMainSrv()
	}()

	go runMetricsSrv()
	runMainSrv()
}

func getPairingAdaptersMapping(filePath string) (map[string]string, error) {
	if filePath == "" {
		log.Infof("No configuration for pairing adapters")
		return nil, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "while opening pairing adapter configuration file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Warnf("Got error on closing file with pairing adapters configuration: %v", err)
		}
	}()

	decoder := json.NewDecoder(file)
	out := map[string]string{}
	err = decoder.Decode(&out)
	if err != nil {
		return nil, errors.Wrapf(err, "while decoding file [%s] to map[string]string", filePath)
	}
	log.Infof("Successfully read pairing adapters configuration")
	return out, nil
}

func createAndRunConfigProvider(stopCh <-chan struct{}, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	err := provider.Load()
	exitOnError(err, "Error on loading configuration file")
	executor.NewPeriodic(cfg.ConfigurationFileReload, func(stopCh <-chan struct{}) {
		if err := provider.Load(); err != nil {
			exitOnError(err, "Error from Reloader watch")
		}
		log.Infof("Successfully reloaded configuration file")

	}).Run(stopCh)

	return provider
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}

func getTenantMappingHandlerFunc(transact persistence.Transactioner, staticUsersSrc string, staticGroupsSrc string, cfgProvider *configprovider.Provider) (func(writer http.ResponseWriter, request *http.Request), error) {
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

	mapperForUser := tenantmapping.NewMapperForUser(staticUsersRepo, staticGroupsRepo, tenantRepo)
	mapperForSystemAuth := tenantmapping.NewMapperForSystemAuth(systemAuthSvc, cfgProvider, tenantRepo)

	reqDataParser := oathkeeper.NewReqDataParser()

	return tenantmapping.NewHandler(reqDataParser, transact, mapperForUser, mapperForSystemAuth).ServeHTTP, nil
}

func getRuntimeMappingHandlerFunc(transact persistence.Transactioner, cachePeriod time.Duration, stopCh <-chan struct{}, defaultScenarioEnabled bool) (func(writer http.ResponseWriter, request *http.Request), error) {
	logger := log.WithField("component", "runtime-mapping-handler").Logger

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

	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, scenariosSvc, labelUpsertSvc, uidSvc, scenarioAssignmentEngine)

	tenantConv := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConv)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

	reqDataParser := oathkeeper.NewReqDataParser()

	jwksFetch := runtimemapping.NewJWKsFetch(logger)
	jwksCache := runtimemapping.NewJWKsCache(logger, jwksFetch, cachePeriod)
	tokenVerifier := runtimemapping.NewTokenVerifier(logger, jwksCache)

	executor.NewPeriodic(1*time.Minute, func(stopCh <-chan struct{}) {
		jwksCache.Cleanup()
	}).Run(stopCh)

	return runtimemapping.NewHandler(
		logger,
		reqDataParser,
		transact,
		tokenVerifier,
		runtimeSvc,
		tenantSvc).ServeHTTP, nil
}

func createServer(address string, handler http.Handler, name string, timeout time.Duration) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, timeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: timeout,
	}

	runFn := func() {
		log.Infof("Running %s server on %s...", name, address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		log.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func defaultPackageInstanceAuthRepo() packageinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return packageinstanceauth.NewRepository(packageinstanceauth.NewConverter(authConverter))
}

func defaultPackageRepo() mp_package.PackageRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(frConverter, versionConverter)

	return mp_package.NewRepository(mp_package.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
}
