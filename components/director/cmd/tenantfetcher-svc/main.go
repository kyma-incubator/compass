/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	auth "github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	authentication "github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	tenantfetcher "github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	Log log.Config

	TenantsRootAPI string `envconfig:"APP_ROOT_API,default=/tenants"`

	Handler   tenantfetcher.HandlerConfig
	EventsCfg tenantfetcher.EventsConfig

	SecurityConfig        securityConfig
	TenantOnDemandRegions []tenantfetcher.RegionDetails `envconfig:"APP_REGION_DETAILS"`
}

type securityConfig struct {
	JWKSSyncPeriod            time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone       bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	JwksEndpoint              string        `envconfig:"default=file://hack/default-jwks.json,APP_JWKS_ENDPOINT"`
	SubscriptionCallbackScope string        `envconfig:"APP_SUBSCRIPTION_CALLBACK_SCOPE"`
	FetchTenantOnDemandScope  string        `envconfig:"APP_FETCH_TENANT_ON_DEMAND_SCOPE"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, envPrefix)
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Handler.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	envVars := tenantfetcher.ReadFromEnvironment(os.Environ())
	jobNames := tenantfetcher.GetJobNames(envVars)
	log.C(ctx).Infof("Tenant fetcher jobs are: %s", strings.Join(jobNames, ","))

	dbCloseFunctions := make([]func() error, 0)
	dbCloseFunctions = append(dbCloseFunctions, closeFunc)
	defer func() {
		for _, fn := range dbCloseFunctions {
			err := fn()
			exitOnError(err, "Error while closing the connection to the database")
		}
	}()

	stopJobChannels := make([]chan bool, 0, len(jobNames))
	go func() {
		for _, job := range jobNames {
			stopJob := make(chan bool, 1)
			stopJobChannels = append(stopJobChannels, stopJob)

			jobConfig := readJobConfig(ctx, job, envVars)
			metricsReporter := createMetricsReporter(ctx, jobConfig)
			closeFn := runTenantFetcherJob(ctx, jobConfig, metricsReporter, stopJob)
			dbCloseFunctions = append(dbCloseFunctions, closeFn)
		}
	}()

	for _, stopJob := range stopJobChannels {
		<-stopJob
	}

	httpClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	handler := initAPIHandler(ctx, httpClient, cfg, transact)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	runMainSrv()
}

func readJobConfig(ctx context.Context, jobName string, environmentVars map[string]string) tenantfetcher.JobConfig {
	return tenantfetcher.NewTenantFetcherJobEnvironment(ctx, jobName, environmentVars).ReadJobConfig()
}

func runTenantFetcherJob(ctx context.Context, jobConfig tenantfetcher.JobConfig, metricsReporter metrics.MetricsReporter, stopJob chan bool) func() error {
	jobInterval := jobConfig.GetHandlerCgf().TenantFetcherJobIntervalMins
	ticker := time.NewTicker(jobInterval)
	jobName := jobConfig.JobName

	transact, closeFunc, err := persistence.Configure(ctx, jobConfig.GetHandlerCgf().Database)
	exitOnError(err, "Error while establishing the connection to the database")

	go func() {
		for {
			select {
			case <-ticker.C:
				log.C(ctx).Infof("Scheduled tenant fetcher job %s will be executed, job interval is %s", jobName, jobInterval)
				syncTenants(ctx, jobConfig, metricsReporter, transact)
			case <-ctx.Done():
				log.C(ctx).Errorf("Context is canceled and scheduled tenant fetcher job %s will be stopped", jobName)
				stopTenantFetcherJobTicker(ctx, ticker, jobName)
				stopJob <- true
				return
			}
		}
	}()

	return closeFunc
}

func syncTenants(ctx context.Context, jobConfig tenantfetcher.JobConfig, metricsReporter metrics.MetricsReporter, transact persistence.Transactioner) {
	tenantsFetcherSvc, err := createTenantsFetcherSvc(ctx, jobConfig, transact)
	exitOnError(err, "failed to create tenants fetcher service")

	err = tenantsFetcherSvc.SyncTenants()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error while running tenant fetcher job %s: %v", jobConfig.JobName, err)
		metricsReporter.ReportFailedSync(err, ctx)
	}
}

func createMetricsReporter(ctx context.Context, jobConfig tenantfetcher.JobConfig) metrics.MetricsReporter {
	pushEndpoint := jobConfig.GetEventsCgf().MetricsPushEndpoint
	if pushEndpoint == "" {
		log.C(ctx).Warnf("No metrics endpoint provided for tenant fetcher job %q, metric reporting will be skipped...", jobConfig.JobName)
		return metrics.MetricsReporter{}
	}

	metricsPusher := metrics.NewPusherPerJob(jobConfig.JobName, pushEndpoint, jobConfig.GetHandlerCgf().ClientTimeout)
	return metrics.NewMetricsReporter(metricsPusher)
}

func createTenantsFetcherSvc(ctx context.Context, jobConfig tenantfetcher.JobConfig, transact persistence.Transactioner) (tenantfetcher.TenantSyncService, error) {
	eventsCfg := jobConfig.GetEventsCgf()
	handlerCfg := jobConfig.GetHandlerCgf()

	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	tenantStorageConverter := tenant.NewConverter()
	labelConverter := label.NewConverter()
	authConverter := authentication.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	scenarioAssignConverter := scenarioassignment.NewConverter()
	runtimeContextConverter := runtimectx.NewConverter()
	tenantConverter := tenant.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter()

	webhookRepo := webhook.NewRepository(webhookConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelRepo := label.NewRepository(labelConverter)
	tenantStorageRepo := tenant.NewRepository(tenantStorageConverter)
	applicationRepo := application.NewRepository(appConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepo, labelSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantStorageRepo, uidSvc, handlerCfg.Features.DefaultScenarioEnabled)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, labelDefSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)
	formationSvc := formation.NewService(labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tenantSvc, runtimeRepo, runtimeContextRepo, nil, nil, applicationRepo, nil, webhookConverter, handlerCfg.Features.RuntimeTypeLabelKey, handlerCfg.Features.ApplicationTypeLabelKey)
	runtimeContextSvc := runtimectx.NewService(runtimeContextRepo, labelRepo, runtimeRepo, labelSvc, formationSvc, tenantSvc, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, labelDefSvc, labelSvc, uidSvc, formationSvc, tenantStorageSvc, webhookSvc, runtimeContextSvc, handlerCfg.Features.ProtectedLabelPattern, handlerCfg.Features.ImmutableLabelPattern, handlerCfg.Features.RuntimeTypeLabelKey, handlerCfg.Features.KymaRuntimeTypeLabelValue)

	kubeClient, err := tenantfetcher.NewKubernetesClient(ctx, handlerCfg.Kubernetes)
	exitOnError(err, "Failed to initialize Kubernetes client")

	eventAPIClient, err := tenantfetcher.NewClient(eventsCfg.OAuthConfig, eventsCfg.AuthMode, eventsCfg.APIConfig, handlerCfg.ClientTimeout)
	if nil != err {
		return nil, err
	}

	var metricsPusher *metrics.Pusher
	if eventsCfg.MetricsPushEndpoint != "" {
		metricsPusher = metrics.NewPusher(eventsCfg.MetricsPushEndpoint, handlerCfg.ClientTimeout)
	}

	if metricsPusher != nil {
		eventAPIClient.SetMetricsPusher(metricsPusher)
	}

	gqlClient := newInternalGraphQLClient(handlerCfg.DirectorGraphQLEndpoint, handlerCfg.ClientTimeout, handlerCfg.HTTPClientSkipSslValidation)
	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}
	directorClient := graphqlclient.NewDirector(gqlClient)

	if handlerCfg.ShouldSyncSubaccounts {
		return tenantfetcher.NewSubaccountService(eventsCfg.QueryConfig, transact, kubeClient, eventsCfg.TenantFieldMapping, eventsCfg.MovedSubaccountFieldMapping, handlerCfg.TenantProvider, eventsCfg.SubaccountRegions, eventAPIClient, tenantStorageSvc, runtimeSvc, labelRepo, handlerCfg.FullResyncInterval, directorClient, handlerCfg.TenantInsertChunkSize, tenantStorageConverter), nil
	}
	return tenantfetcher.NewGlobalAccountService(eventsCfg.QueryConfig, transact, kubeClient, eventsCfg.TenantFieldMapping, handlerCfg.TenantProvider, eventsCfg.AccountsRegion, eventAPIClient, tenantStorageSvc, handlerCfg.FullResyncInterval, directorClient, handlerCfg.TenantInsertChunkSize, tenantStorageConverter), nil
}

func stopTenantFetcherJobTicker(ctx context.Context, tenantFetcherJobTicker *time.Ticker, jobName string) {
	tenantFetcherJobTicker.Stop()
	log.C(ctx).Infof("Ticker for tenant fetcher job %s is stopped", jobName)
}

func initAPIHandler(ctx context.Context, httpClient *http.Client, cfg config, transact persistence.Transactioner) http.Handler {
	logger := log.C(ctx)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger())

	tenantsAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, tenantsAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.SubscriptionCallbackScope)
	registerTenantsHandler(ctx, tenantsAPIRouter, cfg.Handler)

	tenantsOnDemandAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, tenantsOnDemandAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.FetchTenantOnDemandScope)
	registerTenantsOnDemandHandler(ctx, tenantsOnDemandAPIRouter, cfg.EventsCfg, cfg.Handler, transact, cfg.TenantOnDemandRegions)

	healthCheckRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc("/readyz", newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc("/healthz", newReadinessHandler())

	return mainRouter
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	logger := log.C(ctx)

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	runFn := func() {
		logger.Infof("Running %s server on %s...", name, cfg.Address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func configureAuthMiddleware(ctx context.Context, httpClient *http.Client, router *mux.Router, cfg securityConfig, requiredScopes ...string) {
	scopeValidator := claims.NewScopesValidator(requiredScopes)
	middleware := auth.New(httpClient, cfg.JwksEndpoint, cfg.AllowJWTSigningNone, "", scopeValidator)
	router.Use(middleware.Handler())

	log.C(ctx).Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
		if err := middleware.SynchronizeJWKS(ctx); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)
}

func registerTenantsHandler(ctx context.Context, router *mux.Router, cfg tenantfetcher.HandlerConfig) {
	gqlClient := newInternalGraphQLClient(cfg.DirectorGraphQLEndpoint, cfg.ClientTimeout, cfg.HTTPClientSkipSslValidation)
	directorClient := graphqlclient.NewDirector(gqlClient)

	tenantConverter := tenant.NewConverter()

	provisioner := tenantfetcher.NewTenantProvisioner(directorClient, tenantConverter, cfg.TenantProvider)
	subscriber := tenantfetcher.NewSubscriber(directorClient, provisioner)
	tenantHandler := tenantfetcher.NewTenantsHTTPHandler(subscriber, cfg)

	log.C(ctx).Infof("Registering Regional Tenant Onboarding endpoint on %s...", cfg.RegionalHandlerEndpoint)
	router.HandleFunc(cfg.RegionalHandlerEndpoint, tenantHandler.SubscribeTenant).Methods(http.MethodPut)

	log.C(ctx).Infof("Registering Regional Tenant Decommissioning endpoint on %s...", cfg.RegionalHandlerEndpoint)
	router.HandleFunc(cfg.RegionalHandlerEndpoint, tenantHandler.UnSubscribeTenant).Methods(http.MethodDelete)

	log.C(ctx).Infof("Registering service dependencies endpoint on %s...", cfg.DependenciesEndpoint)
	router.HandleFunc(cfg.DependenciesEndpoint, tenantHandler.Dependencies).Methods(http.MethodGet)
}

func registerTenantsOnDemandHandler(ctx context.Context, router *mux.Router, eventsCfg tenantfetcher.EventsConfig, tenantHandlerCfg tenantfetcher.HandlerConfig, transact persistence.Transactioner, regionDetails []tenantfetcher.RegionDetails) {
	onDemandSvc, err := createTenantFetcherOnDemandSvc(eventsCfg, tenantHandlerCfg, transact, regionDetailsToMap(regionDetails))
	exitOnError(err, "failed to create tenant fetcher on-demand service")

	fetcher := tenantfetcher.NewTenantFetcher(*onDemandSvc)
	tenantHandler := tenantfetcher.NewTenantFetcherHTTPHandler(fetcher, tenantHandlerCfg)

	log.C(ctx).Infof("Registering fetch tenant on-demand endpoint on %s...", tenantHandlerCfg.TenantOnDemandHandlerEndpoint)
	router.HandleFunc(tenantHandlerCfg.TenantOnDemandHandlerEndpoint, tenantHandler.FetchTenantOnDemand).Methods(http.MethodPost)
}

func regionDetailsToMap(regionDetails []tenantfetcher.RegionDetails) map[string]tenantfetcher.RegionDetails {
	regionDetailsMap := make(map[string]tenantfetcher.RegionDetails)
	for _, region := range regionDetails {
		regionDetailsMap[region.Name] = region
	}
	return regionDetailsMap
}

func createTenantFetcherOnDemandSvc(eventsCfg tenantfetcher.EventsConfig, handlerCfg tenantfetcher.HandlerConfig, transact persistence.Transactioner, regionDetails map[string]tenantfetcher.RegionDetails) (*tenantfetcher.SubaccountOnDemandService, error) {
	eventAPIClient, err := tenantfetcher.NewClient(eventsCfg.OAuthConfig, eventsCfg.AuthMode, eventsCfg.APIConfig, handlerCfg.ClientTimeout)
	if nil != err {
		return nil, err
	}

	tenantStorageConv := tenant.NewConverter()
	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	labelDefRepository := labeldef.NewRepository(labelDefConverter)

	labelConverter := label.NewConverter()
	labelRepository := label.NewRepository(labelConverter)
	labelService := label.NewLabelService(labelRepository, labelDefRepository, uidSvc)

	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepository, labelService)

	gqlClient := newInternalGraphQLClient(handlerCfg.DirectorGraphQLEndpoint, handlerCfg.ClientTimeout, handlerCfg.HTTPClientSkipSslValidation)
	directorClient := graphqlclient.NewDirector(gqlClient)

	return tenantfetcher.NewSubaccountOnDemandService(eventsCfg.QueryConfig, eventsCfg.TenantFieldMapping, eventAPIClient, transact, tenantStorageSvc, directorClient, handlerCfg.TenantProvider, tenantStorageConv, regionDetails), nil
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func newInternalGraphQLClient(url string, timeout time.Duration, skipSSLValidation bool) *gcli.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}
	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(tr), "Authorization")),
		Timeout:   timeout,
	}

	gqlClient := gcli.NewClient(url, gcli.WithHTTPClient(client))

	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}

	return gqlClient
}
