package main

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/vektah/gqlparser/gqlerror"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"time"

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

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Database                      persistence.DatabaseConfig
	APIEndpoint                   string `envconfig:"default=/graphql"`
	TenantMappingEndpoint         string `envconfig:"default=/tenant-mapping"`
	RuntimeMappingEndpoint        string `envconfig:"default=/runtime-mapping"`
	AuthenticationMappingEndpoint string `envconfig:"default=/authn-mapping"`
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
			HasScenario: scenario.NewDirective(transact, label.NewRepository(label.NewConverter()), defaultBundleRepo(), defaultBundleInstanceAuthRepo()).HasScenario,
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

	statusMiddleware := statusupdate.New(transact, statusupdate.NewRepository())

	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))

	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger())
	presenter := error_presenter.NewPresenter(uid.NewService())

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.Use(PackageToBundleHandler())
	gqlAPIRouter.Use(statusMiddleware.Handler())
	gqlAPIRouter.HandleFunc("", metricsCollector.GraphQLHandlerWithInstrumentation(handler.GraphQL(executableSchema,
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

	logger.Infof("Registering readiness endpoint...")
	mainRouter.HandleFunc("/readyz", healthz.NewReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc("/healthz", healthz.NewLivenessHandler(transact))

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	runMetricsSrv, shutdownMetricsSrv := createServer(ctx, cfg.MetricsAddress, metricsHandler, "metrics", cfg.ServerTimeout)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg.Address, mainRouter, "main", cfg.ServerTimeout)

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
		shutdownMainSrv()
	}()

	go runMetricsSrv()
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

func defaultBundleInstanceAuthRepo() bundleinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return bundleinstanceauth.NewRepository(bundleinstanceauth.NewConverter(authConverter))
}

func defaultBundleRepo() mp_bundle.BundleRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	eventAPIConverter := eventdef.NewConverter(frConverter, versionConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(frConverter, versionConverter)

	return mp_bundle.NewRepository(mp_bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
}

func PackageToBundleHandler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			recorder := httptest.NewRecorder()

			reqBody, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.C(ctx).WithError(err).Error("Error reading request body")
				appErr := apperrors.InternalErrorFrom(err, "while reading request body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}

			body := string(reqBody)
			body = strings.ReplaceAll(body, "\\n", "") // removes unnecessary complexity from the next regexes

			reqPackagesJSONPattern := regexp.MustCompile(`(\s*)packages(\s*:\s*\[)`) // matches ` packages:  [`
			body = reqPackagesJSONPattern.ReplaceAllString(body, "${1}bundles${2}")

			reqPackagesGraphQLPattern := regexp.MustCompile(`(\s*)packages(\s*\{)`) // matches ` packages {`
			body = reqPackagesGraphQLPattern.ReplaceAllString(body, "${1}bundles${2}")

			reqPackageGraphQLPattern := regexp.MustCompile(`(\s*)package(\s*\(\s*id\s*:\s*)`) // matches ` package ( id : `
			body = reqPackageGraphQLPattern.ReplaceAllString(body, "${1}bundle${2}")

			reqPackageModeGraphQLPattern := regexp.MustCompile(`(\s*)mode(\s*):(\s*)PACKAGE(\s*)`) // matches ` mode: PACKAGE `
			body = reqPackageModeGraphQLPattern.ReplaceAllString(body, "${1}mode${2}:${3}BUNDLE${4}")

			body = strings.ReplaceAll(body, "addPackage", "addBundle")
			body = strings.ReplaceAll(body, "updatePackage", "updateBundle")
			body = strings.ReplaceAll(body, "deletePackage", "deleteBundle")
			body = strings.ReplaceAll(body, "addAPIDefinitionToPackage", "addAPIDefinitionToBundle")
			body = strings.ReplaceAll(body, "addEventDefinitionToPackage", "addEventDefinitionToBundle")
			body = strings.ReplaceAll(body, "addDocumentToPackage", "addDocumentToBundle")
			body = strings.ReplaceAll(body, "setPackageInstanceAuth", "setBundleInstanceAuth")
			body = strings.ReplaceAll(body, "deletePackageInstanceAuth", "deleteBundleInstanceAuth")
			body = strings.ReplaceAll(body, "requestPackageInstanceAuthCreation", "requestBundleInstanceAuthCreation")
			body = strings.ReplaceAll(body, "requestBundleInstanceAuthDeletion", "requestBundleInstanceAuthDeletion")

			r.Body = ioutil.NopCloser(strings.NewReader(body))
			r.ContentLength = int64(len(body))

			next.ServeHTTP(recorder, r)

			for key, values := range recorder.Header() {
				for _, v := range values {
					w.Header().Add(key, v)
				}
			}

			respBody, err := ioutil.ReadAll(recorder.Body)
			if err != nil {
				log.C(ctx).WithError(err).Error("Error reading response body")
				appErr := apperrors.InternalErrorFrom(err, "while reading response body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}

			body = string(respBody)

			respPackagesJSONPattern := regexp.MustCompile(`(\s*\")bundles(\"\s*:\s*\{)`) // matches ` "bundles":  {`
			body = respPackagesJSONPattern.ReplaceAllString(body, "${1}packages${2}")

			respPackageJSONPattern := regexp.MustCompile(`(\s*\")bundle(\"\s*:\s*\{)`) // matches ` "bundle":  {`
			body = respPackageJSONPattern.ReplaceAllString(body, "${1}package${2}")

			respPackageModeGraphQLPattern := regexp.MustCompile(`(\s*\")mode(\"\s*):(\s*\")BUNDLE(\"\s*)`) // matches ` "mode": "BUNDLE" `
			body = respPackageModeGraphQLPattern.ReplaceAllString(body, "${1}mode${2}:${3}PACKAGE${4}")

			w.WriteHeader(recorder.Code)
			if _, err := w.Write([]byte(body)); err != nil {
				log.C(ctx).WithError(err).Error("Error writing response body")
				appErr := apperrors.InternalErrorFrom(err, "while writing response body")
				writeAppError(ctx, w, appErr, http.StatusInternalServerError)
				return
			}
		})
	}
}

func writeAppError(ctx context.Context, w http.ResponseWriter, appErr error, statusCode int) {
	errCode := apperrors.ErrorCode(appErr)
	if errCode == apperrors.UnknownError || errCode == apperrors.InternalError {
		errCode = apperrors.InternalError
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")
	resp := gqlgen.Response{Errors: []*gqlerror.Error{{
		Message:    appErr.Error(),
		Extensions: map[string]interface{}{"error_code": errCode, "error": errCode.String()}}}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.C(ctx).WithError(err).Error("An error occurred while encoding data. ")
	}
}
