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
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	tenantfetcher "github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	tenant2 "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	Log log.Config

	TenantsRootAPI string `envconfig:"APP_ROOT_API,default=/tenants"`

	Handler tenantfetcher.HandlerConfig

	Features features.Config

	SecurityConfig securityConfig

	WebhookProcessorElectionConfig cronjob.ElectionConfig
	WebhookProcessorJobInterval    time.Duration `envconfig:"APP_WEBHOOK_PROCESSOR_JOB_INTERVAL,default=3m"`

	SystemFieldDiscoveryWebhookPartialProcessing     bool `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_WEBHOOK_PARTIAL_PROCESSING"`
	SystemFieldDiscoveryWebhookPartialProcessMaxDays int  `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_WEBHOOK_PARTIAL_PROCESS_MAX_DAYS"`
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

	tenantSynchronizers, dbCloseFuncs := tenantSynchronizers(ctx, cfg.Handler, cfg.Features)
	defer func() {
		for _, fn := range dbCloseFuncs {
			if err := fn(); err != nil {
				log.D().WithError(err).Error("Error while closing the connection to the database")
			}
		}
	}()

	for _, sync := range tenantSynchronizers {
		log.C(ctx).Infof("Starting tenant synchronizer %s...", sync.Name())
		go func(synchronizer *resync.TenantsSynchronizer) {
			synchronizeTenants(synchronizer, ctx)
		}(sync)
	}

	httpClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	handler := initAPIHandler(ctx, httpClient, cfg, tenantSynchronizers)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)

	webhookRepo := webhook.NewRepository(webhookConverter)
	applicationRepo := application.NewRepository(appConverter)
	uidSvc := uid.NewService()
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, nil, applicationRepo, nil, nil, nil, nil, nil, nil, uidSvc, nil, "", nil)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc, map[string]interface{}{}, "")

	webhookClient := pkgAuth.PrepareHTTPClient(cfg.Handler.ClientTimeout)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Handler.Database)
	exitOnError(err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(err, "error while closing the connection to the database")
	}()

	webhookProcessor := tenantfetcher.NewWebhookProcessor(transact, webhookSvc, tenantSvc, appSvc, webhookClient, cfg.WebhookProcessorElectionConfig, cfg.WebhookProcessorJobInterval, cfg.SystemFieldDiscoveryWebhookPartialProcessing, cfg.SystemFieldDiscoveryWebhookPartialProcessMaxDays)
	go func() {
		if err := webhookProcessor.StartWebhookProcessorJob(ctx, tenantfetcher.SaaSRegistryLabelValue); err != nil {
			log.C(ctx).WithError(err).Error("Failed to run WebhookProcessorJob. Stopping app...")
			cancel()
		}
	}()

	runMainSrv()
}

func stopTenantFetcherJobTicker(ctx context.Context, tenantFetcherJobTicker *time.Ticker, jobName string) {
	tenantFetcherJobTicker.Stop()
	log.C(ctx).Infof("Ticker for tenant fetcher job %s is stopped", jobName)
}

func initAPIHandler(ctx context.Context, httpClient *http.Client, cfg config, synchronizers []*resync.TenantsSynchronizer) http.Handler {
	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
	)
	logger := log.C(ctx)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.TenantsRootAPI+healthzEndpoint, cfg.TenantsRootAPI+readyzEndpoint))

	tenantsAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, tenantsAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.SubscriptionCallbackScope)
	registerTenantsHandler(ctx, tenantsAPIRouter, cfg.Handler)

	tenantsOnDemandAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, tenantsOnDemandAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.FetchTenantOnDemandScope)
	registerTenantsOnDemandHandler(ctx, tenantsOnDemandAPIRouter, cfg.Handler, synchronizers)

	healthCheckRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

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
	middleware := authmiddleware.New(httpClient, cfg.JwksEndpoint, cfg.AllowJWTSigningNone, "", scopeValidator)
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

	dependencies, err := dependenciesConfigToMap(cfg)
	exitOnError(err, "failed to read service dependencies")
	cfg.RegionToDependenciesConfig = dependencies

	tenantHandler := tenantfetcher.NewTenantsHTTPHandler(subscriber, cfg)

	log.C(ctx).Infof("Registering Regional Tenant Onboarding endpoint on %s...", cfg.RegionalHandlerEndpoint)
	router.HandleFunc(cfg.RegionalHandlerEndpoint, tenantHandler.SubscribeTenant).Methods(http.MethodPut)

	log.C(ctx).Infof("Registering Regional Tenant Decommissioning endpoint on %s...", cfg.RegionalHandlerEndpoint)
	router.HandleFunc(cfg.RegionalHandlerEndpoint, tenantHandler.UnSubscribeTenant).Methods(http.MethodDelete)

	log.C(ctx).Infof("Registering service dependencies endpoint on %s...", cfg.DependenciesEndpoint)
	router.HandleFunc(cfg.DependenciesEndpoint, tenantHandler.Dependencies).Methods(http.MethodGet)
}

func tenantSynchronizers(ctx context.Context, taConfig tenantfetcher.HandlerConfig, featuresConfig features.Config) ([]*resync.TenantsSynchronizer, []func() error) {
	envVars := resync.ReadFromEnvironment(os.Environ())
	jobNames := resync.GetJobNames(envVars)
	log.C(ctx).Infof("Tenant fetcher jobs are: %s", strings.Join(jobNames, ","))

	jobConfigs := make([]resync.JobConfig, 0)
	for _, job := range jobNames {
		jobConfig, err := resync.NewTenantFetcherJobEnvironment(ctx, job, envVars).ReadJobConfig()
		exitOnError(err, fmt.Sprintf("Error while reading job config for job %s", job))
		jobConfigs = append(jobConfigs, *jobConfig)
	}

	gqlClient := newInternalGraphQLClient(taConfig.DirectorGraphQLEndpoint, taConfig.ClientTimeout, taConfig.HTTPClientSkipSslValidation)
	directorClient := graphqlclient.NewDirector(gqlClient)

	dbCloseFunctions := make([]func() error, 0)
	synchronizers := make([]*resync.TenantsSynchronizer, 0)
	for _, jobConfig := range jobConfigs {
		transact, closeFunc, err := persistence.Configure(ctx, taConfig.Database)
		exitOnError(err, "Error while establishing the connection to the database")

		metricsCfg := metrics.PusherConfig{
			Enabled:    len(taConfig.MetricsPushEndpoint) > 0,
			Endpoint:   taConfig.MetricsPushEndpoint,
			MetricName: strings.ReplaceAll(strings.ToLower(jobConfig.JobName), "-", "_") + "_job_sync_failure_number",
			Timeout:    taConfig.ClientTimeout,
			Subsystem:  metrics.TenantFetcherSubsystem,
			Labels:     []string{metrics.ErrorMetricLabel},
		}
		metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
		builder := resync.NewSynchronizerBuilder(jobConfig, featuresConfig, transact, directorClient, metricsPusher)
		log.C(ctx).Infof("Creating tenant synchronizer %s for tenants of type %s", jobConfig.JobName, jobConfig.TenantType)
		synchronizer, err := builder.Build(ctx)
		exitOnError(err, fmt.Sprintf("Error while creating tenant synchronizer %s for tenants of type %s", jobConfig.JobName, jobConfig.TenantType))

		synchronizers = append(synchronizers, synchronizer)
		dbCloseFunctions = append(dbCloseFunctions, closeFunc)
	}

	return synchronizers, dbCloseFunctions
}

func synchronizeTenants(synchronizer *resync.TenantsSynchronizer, ctx context.Context) {
	ticker := time.NewTicker(synchronizer.ResyncInterval())
	for {
		select {
		case <-ticker.C:
			logger := log.C(ctx)
			logger = logger.WithField(log.FieldRequestID, uuid.New().String()).WithField("tenant-synchronizer-name", synchronizer.Name())
			resyncCtx := log.ContextWithLogger(ctx, logger)
			log.C(resyncCtx).Infof("Scheduled tenant resync job %s will be executed, job interval is %s", synchronizer.Name(), synchronizer.ResyncInterval())
			if err := synchronizer.Synchronize(resyncCtx); err != nil {
				log.C(resyncCtx).WithError(err).Errorf("Tenant fetcher resync %s failed with error: %v", synchronizer.Name(), err)
			}
		case <-ctx.Done():
			log.C(ctx).Errorf("Context is canceled and scheduled tenant fetcher job %s will be stopped", synchronizer.Name())
			stopTenantFetcherJobTicker(ctx, ticker, synchronizer.Name())
			return
		}
	}
}

func registerTenantsOnDemandHandler(ctx context.Context, router *mux.Router, handlerCfg tenantfetcher.HandlerConfig, synchronizers []*resync.TenantsSynchronizer) {
	var subaccountSynchronizer *resync.TenantsSynchronizer
	for _, tenantSynchronizer := range synchronizers {
		if tenantSynchronizer.TenantType() == tenant2.Subaccount {
			subaccountSynchronizer = tenantSynchronizer
			break
		}
	}

	if subaccountSynchronizer == nil {
		log.C(ctx).Infof("Subaccount synchronizer not found, tenant on-demand API won't be enabled")
		return
	}

	log.C(ctx).Infof("Registering fetch tenant with parent on-demand endpoint on %s...", handlerCfg.TenantWithParentOnDemandHandlerEndpoint)
	log.C(ctx).Infof("Registering fetch tenant without parent on-demand endpoint on %s...", handlerCfg.TenantWithoutParentOnDemandHandlerEndpoint)
	tenantHandler := tenantfetcher.NewTenantFetcherHTTPHandler(subaccountSynchronizer, handlerCfg)
	router.HandleFunc(handlerCfg.TenantWithParentOnDemandHandlerEndpoint, tenantHandler.FetchTenantOnDemand).Methods(http.MethodPost)
	router.HandleFunc(handlerCfg.TenantWithoutParentOnDemandHandlerEndpoint, tenantHandler.FetchTenantOnDemand).Methods(http.MethodPost)
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

func dependenciesConfigToMap(cfg tenantfetcher.HandlerConfig) (map[string][]tenantfetcher.Dependency, error) {
	secretData, err := configprovider.ReadConfigFile(cfg.TenantDependenciesConfigPath)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading tenant service dependencies config file")
	}

	dependenciesConfig := make(map[string][]tenantfetcher.Dependency)
	config, err := configprovider.ParseConfigToJSONMap(secretData)
	if err != nil {
		return nil, errors.Wrapf(err, "while parsing tenant service dependencies config file")
	}

	for region, dependencies := range config {
		for _, dependency := range dependencies.Array() {
			xsappName := gjson.Get(dependency.String(), cfg.XsAppNamePathParam)
			dependenciesConfig[region] = append(dependenciesConfig[region], tenantfetcher.Dependency{Xsappname: xsappName.String()})
		}
	}

	return dependenciesConfig, nil
}
