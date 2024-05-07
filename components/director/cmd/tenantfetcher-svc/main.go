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
	"github.com/davecgh/go-spew/spew"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
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
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	systemfielddiscoveryengine "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine"
	systemfielddiscoveryenginecfg "github.com/kyma-incubator/compass/components/director/internal/system-field-discovery-engine/config"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
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

	SecurityConfig                   securityConfig
	SystemFieldDiscoveryEngineConfig systemfielddiscoveryenginecfg.SystemFieldDiscoveryEngineConfig

	OperationsManagerConfig       operationsmanager.OperationsManagerConfig
	ParallelOperationProcessors   int           `envconfig:"APP_PARALLEL_OPERATION_PROCESSORS,default=10"` // add env vars
	OperationProcessorQuietPeriod time.Duration `envconfig:"APP_OPERATION_PROCESSORS_QUIET_PERIOD,default=5s"`
}

type securityConfig struct {
	JWKSSyncPeriod            time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone       bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	JwksEndpoint              string        `envconfig:"default=file://hack/default-jwks.json,APP_JWKS_ENDPOINT"`
	SubscriptionCallbackScope string        `envconfig:"APP_SUBSCRIPTION_CALLBACK_SCOPE"`
	FetchTenantOnDemandScope  string        `envconfig:"APP_FETCH_TENANT_ON_DEMAND_SCOPE"`
	SystemFieldDiscoveryScope string        `envconfig:"APP_SYSTEM_FIELD_DISCOVERY_SCOPE"`
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

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Handler.Database)
	exitOnError(err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(err, "error while closing the connection to the database")
	}()
	opRepo := operation.NewRepository(operation.NewConverter())
	opSvc := operation.NewService(opRepo, uid.NewService())

	saasRegistryOperationsManager := operationsmanager.NewOperationsManager(transact, opSvc, model.OperationTypeSaasRegistryDiscovery, cfg.OperationsManagerConfig)

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

	applicationRepo := application.NewRepository(appConverter)
	labelRepo := label.NewRepository(label.NewConverter())
	tenantRepo := tenant.NewRepository(tenantConverter)

	uidSvc := uid.NewService()
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, nil, applicationRepo, nil, nil, labelRepo, nil, nil, nil, uidSvc, nil, "", nil)

	systemFieldDiscoveryClient := pkgAuth.PrepareHTTPClient(cfg.Handler.ClientTimeout)

	systemFieldDiscoverySvc, err := systemfielddiscoveryengine.NewSystemFieldDiscoverEngineService(cfg.SystemFieldDiscoveryEngineConfig, systemFieldDiscoveryClient, transact, appSvc, tenantSvc)
	exitOnError(err, "error while creating system field discovery engine")

	systemFieldDiscoveryProcessor := &systemfielddiscoveryengine.OperationsProcessor{
		SystemFieldDiscoverySvc: systemFieldDiscoverySvc,
	}
	onDemandChannel := make(chan string, 100)

	handler := initAPIHandler(ctx, httpClient, cfg, tenantSynchronizers, saasRegistryOperationsManager, onDemandChannel)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	spew.Dump("ParallelOperationProcessors: ", cfg.ParallelOperationProcessors)
	for i := 0; i < cfg.ParallelOperationProcessors; i++ {
		go func(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *systemfielddiscoveryengine.OperationsProcessor, executorIndex int) {
			for {
				select {
				case <-onDemandChannel:
				default:
				}

				processedOperationID, err := claimAndProcessOperation(ctx, opManager, opProcessor)
				if err != nil {
					log.C(ctx).Errorf("Failed during claim and process operation %q by executor %d . Err: %v", processedOperationID, executorIndex, err)
				}
				if len(processedOperationID) > 0 {
					log.C(ctx).Infof("Processed Operation: %s by executor %d", processedOperationID, executorIndex)
				} else {
					// Queue is empty - no operation claimed
					log.C(ctx).Infof("No Processed Operation by executor %d", executorIndex)

					select {
					case operationID := <-onDemandChannel:
						log.C(ctx).Infof("Operation %q send for processing through OnDemand channel to executor %d", operationID, executorIndex)
					case <-time.After(cfg.OperationProcessorQuietPeriod):
						log.C(ctx).Infof("Quiet period finished for executor %d", executorIndex)
					}
				}
			}
		}(ctx, saasRegistryOperationsManager, systemFieldDiscoveryProcessor, i)
	}

	go func() {
		if err := saasRegistryOperationsManager.StartRescheduleOperationsJob(ctx, []string{"COMPLETED"}); err != nil {
			log.C(ctx).WithError(err).Error("Failed to run  RescheduleOperationsJob. Stopping app...")
			cancel()
		}
	}()

	go func() {
		if err := saasRegistryOperationsManager.StartDeleteOperationsJob(ctx); err != nil {
			log.C(ctx).WithError(err).Error("Failed to run  RescheduleOperationsJob. Stopping app...")
			cancel()
		}
	}()

	runMainSrv()
}

func stopTenantFetcherJobTicker(ctx context.Context, tenantFetcherJobTicker *time.Ticker, jobName string) {
	tenantFetcherJobTicker.Stop()
	log.C(ctx).Infof("Ticker for tenant fetcher job %s is stopped", jobName)
}

func initAPIHandler(ctx context.Context, httpClient *http.Client, cfg config, synchronizers []*resync.TenantsSynchronizer, opMgr *operationsmanager.OperationsManager, onDemandChannel chan string) http.Handler {
	const (
		healthzEndpoint              = "/healthz"
		readyzEndpoint               = "/readyz"
		systemFieldDiscoveryEndpoint = "/system-field-discovery"
	)
	logger := log.C(ctx)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.TenantsRootAPI+healthzEndpoint, cfg.TenantsRootAPI+readyzEndpoint))

	tenantsAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, tenantsAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.SubscriptionCallbackScope)
	registerTenantsHandler(ctx, tenantsAPIRouter, cfg.Handler)

	handler := systemfielddiscoveryengine.NewSystemFieldDiscoveryHTTPHandler(opMgr, onDemandChannel)
	systemFieldDiscoveryAPIRouter := mainRouter.PathPrefix(cfg.TenantsRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, systemFieldDiscoveryAPIRouter, cfg.SecurityConfig, cfg.SecurityConfig.SystemFieldDiscoveryScope)
	systemFieldDiscoveryAPIRouter.HandleFunc(systemFieldDiscoveryEndpoint, handler.ScheduleSaaSRegistryDiscoveryForSystemFieldDiscoveryData).Methods(http.MethodPost)

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

func claimAndProcessOperation(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *systemfielddiscoveryengine.OperationsProcessor) (string, error) {
	op, errGetOperation := opManager.GetOperation(ctx)
	if errGetOperation != nil {
		if apperrors.IsNoScheduledOperationsError(errGetOperation) {
			log.C(ctx).Infof("There aro no scheduled operations for processing. Err: %v", errGetOperation)
			return "", nil
		} else {
			log.C(ctx).Errorf("Cannot get operation from OperationsManager. Err: %v", errGetOperation)
			return "", errGetOperation
		}
	}
	log.C(ctx).Infof("Taken operation for processing: %s", op.ID)
	if errProcess := opProcessor.Process(ctx, op); errProcess != nil {
		log.C(ctx).Infof("Error while processing operation with id %q. Err: %v", op.ID, errProcess)

		if errMarkAsFailed := opManager.MarkOperationFailed(ctx, op.ID, errProcess); errMarkAsFailed != nil {
			log.C(ctx).Errorf("Error while marking operation with id %q as failed. Err: %v", op.ID, errMarkAsFailed)
			return op.ID, errMarkAsFailed
		}
		return op.ID, errProcess
	}
	if errMarkAsCompleted := opManager.MarkOperationCompleted(ctx, op.ID, nil); errMarkAsCompleted != nil {
		log.C(ctx).Errorf("Error while marking operation with id %q as completed. Err: %v", op.ID, errMarkAsCompleted)
		return op.ID, errMarkAsCompleted
	}

	return op.ID, nil
}
