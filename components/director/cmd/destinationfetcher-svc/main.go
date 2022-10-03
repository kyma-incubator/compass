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
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator"
	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	destinationfetcher "github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uuid"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout              time.Duration `envconfig:"default=110s"`
	ShutdownTimeout            time.Duration `envconfig:"default=10s"`
	DestinationFetcherSchedule time.Duration `envconfig:"APP_DESTINATION_FETCHER_SCHEDULE,default=10m"`
	ParallelTenantSyncs        int64         `envconfig:"APP_DESTINATION_FETCHER_PARALLEL_TENANTS,default=10"`
	DestinationsRootAPI        string        `envconfig:"APP_ROOT_API,default=/destinations"`

	HandlerConfig               destinationfetcher.HandlerConfig
	DestinationServiceAPIConfig destinationfetcher.DestinationServiceAPIConfig
	DestinationsConfig          configprovider.DestinationsConfig
	Database                    persistence.DatabaseConfig
	Log                         log.Config
	SecurityConfig              securityConfig
	ElectionConfig              cronjob.ElectionConfig
}

type securityConfig struct {
	JWKSSyncPeriod                 time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone            bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	JwksEndpoint                   string        `envconfig:"APP_JWKS_ENDPOINT,default=file://hack/default-jwks.json"`
	DestinationsOnDemandScope      string        `envconfig:"APP_DESTINATIONS_SYNC_SCOPE,default=destinations:sync"`
	DestinationsSensitiveDataScope string        `envconfig:"APP_DESTINATIONS_SENSITIVE_DATA_SCOPE,default=destinations_sensitive_data:read"`
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

	transactioner, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "error while closing the connection to the database")
	}()

	httpClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	destinationService := getDestinationService(cfg, transactioner)
	handler := initAPIHandler(ctx, httpClient, cfg, destinationService)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	syncJobConfig := destinationfetcher.SyncJobConfig{
		ElectionCfg:       cfg.ElectionConfig,
		JobSchedulePeriod: cfg.DestinationFetcherSchedule,
		ParallelTenants:   cfg.ParallelTenantSyncs,
	}
	go func() {
		err := destinationfetcher.StartDestinationFetcherSyncJob(ctx, syncJobConfig, destinationService)
		if err != nil {
			log.C(ctx).WithError(err).Error("Failed to start destination fetcher cronjob. Stopping app...")
			cancel()
		}
	}()

	runMainSrv()
}

func configureAuthMiddleware(ctx context.Context, httpClient *http.Client, router *mux.Router, cfg securityConfig, requiredScopes ...string) {
	scopeValidator := claims.NewScopesValidator(requiredScopes)
	middleware := authenticator.New(httpClient, cfg.JwksEndpoint, cfg.AllowJWTSigningNone, "", scopeValidator)
	router.Use(middleware.Handler())

	log.C(ctx).Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
		if err := middleware.SynchronizeJWKS(ctx); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func getDestinationService(cfg config, transact persistence.Transactioner) *destinationfetcher.DestinationService {
	uuidSvc := uuid.NewService()
	destRepo := destination.NewRepository()
	bundleRepo := bundleRepo()

	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)

	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)

	err := cfg.DestinationsConfig.MapInstanceConfigs()
	exitOnError(err, "error while loading destination instances config")

	return &destinationfetcher.DestinationService{
		Transactioner:      transact,
		UUIDSvc:            uuidSvc,
		DestinationRepo:    destRepo,
		BundleRepo:         bundleRepo,
		LabelRepo:          labelRepo,
		DestinationsConfig: cfg.DestinationsConfig,
		APIConfig:          cfg.DestinationServiceAPIConfig,
		TenantRepo:         tenantRepo,
	}
}

func initAPIHandler(ctx context.Context, httpClient *http.Client,
	cfg config, destService *destinationfetcher.DestinationService) http.Handler {
	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
	)
	logger := log.C(ctx)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.DestinationsRootAPI+healthzEndpoint, cfg.DestinationsRootAPI+readyzEndpoint))

	syncDestinationsAPIRouter := mainRouter.PathPrefix(cfg.DestinationsRootAPI).Subrouter()
	destinationHandler := destinationfetcher.NewDestinationsHTTPHandler(destService, cfg.HandlerConfig)
	sensitiveDataAPIRouter := syncDestinationsAPIRouter

	log.C(ctx).Infof("Registering service destinations endpoint on %s...", cfg.HandlerConfig.SyncDestinationsEndpoint)
	configureAuthMiddleware(ctx, httpClient, syncDestinationsAPIRouter,
		cfg.SecurityConfig, cfg.SecurityConfig.DestinationsOnDemandScope)

	syncDestinationsAPIRouter.HandleFunc(cfg.HandlerConfig.SyncDestinationsEndpoint, destinationHandler.SyncTenantDestinations).
		Methods(http.MethodPut)

	log.C(ctx).Infof("Registering service destinations endpoint on %s...", cfg.HandlerConfig.DestinationsSensitiveEndpoint)
	configureAuthMiddleware(ctx, httpClient, sensitiveDataAPIRouter,
		cfg.SecurityConfig, cfg.SecurityConfig.DestinationsSensitiveDataScope)
	sensitiveDataAPIRouter.HandleFunc(cfg.HandlerConfig.DestinationsSensitiveEndpoint, destinationHandler.FetchDestinationsSensitiveData).
		Methods(http.MethodGet)

	healthCheckRouter := mainRouter.PathPrefix(cfg.DestinationsRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func bundleRepo() destinationfetcher.BundleRepo {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)

	return bundle.NewRepository(bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
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
