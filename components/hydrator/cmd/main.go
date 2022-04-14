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

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"
	"github.com/kyma-incubator/compass/components/hydrator/internal/istiocertresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/metrics"
	"github.com/kyma-incubator/compass/components/hydrator/internal/revocation"
	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler"
	"github.com/kyma-incubator/compass/components/hydrator/internal/connectortokenresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/internal/runtimemapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	tenantmappingconst "github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	handlerConfig "github.com/kyma-incubator/compass/components/hydrator/internal/config"

	"github.com/gorilla/mux"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`
	RootAPI string `envconfig:"APP_ROOT_API,default=/hydrators"`

	ClientTimeout   time.Duration `envconfig:"default=105s"`
	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	Handler handlerConfig.HandlerConfig

	Director director.Config

	JWKSSyncPeriod time.Duration `envconfig:"default=5m"`

	KubeConfig kubernetes.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	StaticGroupsSrc string `envconfig:"default=/data/static-groups.yaml"`

	MetricsConfig metrics.Config

	CSRSubject            istiocertresolver.CSRSubjectConfig
	ExternalIssuerSubject istiocertresolver.ExternalIssuerSubjectConfig

	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=compass-system/revocations-config"`
	SubjectConsumerMappingConfig string `envconfig:"default=[]"`

	Log log.Config
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

	if err := cfg.Handler.Validate(); err != nil {
		exitOnError(errors.New("missing handler endpoint"), "Error while loading app handler config")
	}

	logger := log.C(ctx)

	logger.Infof("Registering metrics collectors...")
	metricsCollector := metrics.NewCollector(cfg.MetricsConfig)
	prometheus.MustRegister(metricsCollector)

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

	handler := initAPIHandlers(ctx, logger, authenticators, cfg, metricsCollector)

	runMetricsSrv, shutdownMetricsSrv := createServer(ctx, metricsHandler, cfg.MetricsConfig.Address, "metrics", cfg.ServerTimeout, cfg.ShutdownTimeout)
	runMainSrv, shutdownMainSrv := createServer(ctx, handler, cfg.Address, "main", cfg.ServerTimeout, cfg.ShutdownTimeout)

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMetricsSrv()
		shutdownMainSrv()
	}()

	go runMetricsSrv()
	runMainSrv()
}

func initAPIHandlers(ctx context.Context, logger *logrus.Entry, authenticators []authenticator.Config, cfg config, metricsCollector *metrics.Collector) http.Handler {
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger())

	router := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()
	healthCheckRouter := mainRouter.NewRoute().Subrouter()

	registerHydratorHandlers(ctx, router, authenticators, cfg, metricsCollector)

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

func createServer(ctx context.Context, handler http.Handler, serverAddress, name string, serverTimeout, shutdownTimeout time.Duration) (func(), func()) {
	logger := log.C(ctx)

	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, serverTimeout)
	exitOnError(err, "Error while configuring tenant mapping handler")

	srv := &http.Server{
		Addr:              serverAddress,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: serverTimeout,
	}

	runFn := func() {
		logger.Infof("Running %s server on %s...", name, serverAddress)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		logger.Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			logger.Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func registerHydratorHandlers(ctx context.Context, router *mux.Router, authenticators []authenticator.Config, cfg config, metricsCollector *metrics.Collector) {
	logger := log.C(ctx)

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}

	directorClientProvider := director.NewClientProvider(cfg.Director.URL, cfg.Director.ClientTimeout, cfg.Director.SkipSSLValidation)
	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	logger.Infof("Registering Authentication Mapping endpoint on %s...", cfg.Handler.AuthenticationMappingEndpoint)
	authnMappingHandlerFunc := authnmappinghandler.NewHandler(oathkeeper.NewReqDataParser(), httpClient, authnmappinghandler.DefaultTokenVerifierProvider, authenticators)

	logger.Infof("Registering Tenant Mapping endpoint on %s...", cfg.Handler.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(authenticators, directorClientProvider, cfg.StaticGroupsSrc, cfgProvider, metricsCollector)
	exitOnError(err, "Error while configuring tenant mapping handler")

	logger.Infof("Registering Runtime Mapping endpoint on %s...", cfg.Handler.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc := getRuntimeMappingHandlerFunc(ctx, directorClientProvider, cfg.JWKSSyncPeriod)

	logger.Infof("Registering Connector Token Resolver endpoint on %s...", cfg.Handler.TokenResolverEndpoint)
	connectorTokenResolverHandlerFunc := getTokenResolverHandler(directorClientProvider)

	logger.Infof("Registering Connector Certificate Resolver endpoint on %s...", cfg.Handler.TokenResolverEndpoint)
	connectorCertResolverHandlerFunc, revokedCertsLoader, err := getCertificateResolverHandler(ctx, cfg)
	exitOnError(err, "Error while configuring tenant mapping handler")

	router.HandleFunc(cfg.Handler.AuthenticationMappingEndpoint, metricsCollector.HandlerInstrumentation(authnMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.TenantMappingEndpoint, metricsCollector.HandlerInstrumentation(tenantMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.RuntimeMappingEndpoint, metricsCollector.HandlerInstrumentation(runtimeMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.TokenResolverEndpoint, metricsCollector.HandlerInstrumentation(connectorTokenResolverHandlerFunc))
	router.HandleFunc(cfg.Handler.ValidationIstioCertEndpoint, metricsCollector.HandlerInstrumentation(connectorCertResolverHandlerFunc))

	go revokedCertsLoader.Run(ctx)
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func getTenantMappingHandlerFunc(authenticators []authenticator.Config, clientProvider director.ClientProvider, staticGroupsSrc string, cfgProvider *configprovider.Provider, metricsCollector *metrics.Collector) (*tenantmapping.Handler, error) {
	staticGroupsRepo, err := tenantmapping.NewStaticGroupRepository(staticGroupsSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticGroup repository instance")
	}

	objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
		tenantmappingconst.UserObjectContextProvider:          tenantmapping.NewUserContextProvider(clientProvider.Client(), staticGroupsRepo),
		tenantmappingconst.SystemAuthObjectContextProvider:    tenantmapping.NewSystemAuthContextProvider(clientProvider.Client(), cfgProvider),
		tenantmappingconst.AuthenticatorObjectContextProvider: tenantmapping.NewAuthenticatorContextProvider(clientProvider.Client(), authenticators),
		tenantmappingconst.CertServiceObjectContextProvider:   tenantmapping.NewCertServiceContextProvider(clientProvider.Client(), cfgProvider),
		tenantmappingconst.TenantHeaderObjectContextProvider:  tenantmapping.NewAccessLevelContextProvider(clientProvider.Client()),
	}
	reqDataParser := oathkeeper.NewReqDataParser()

	return tenantmapping.NewHandler(reqDataParser, objectContextProviders, metricsCollector), nil
}

func getRuntimeMappingHandlerFunc(ctx context.Context, clientProvider director.ClientProvider, cachePeriod time.Duration) *runtimemapping.Handler {
	reqDataParser := oathkeeper.NewReqDataParser()

	jwksFetch := runtimemapping.NewJWKsFetch()
	jwksCache := runtimemapping.NewJWKsCache(jwksFetch, cachePeriod)
	tokenVerifier := runtimemapping.NewTokenVerifier(jwksCache)

	executor.NewPeriodic(1*time.Minute, func(ctx context.Context) {
		jwksCache.Cleanup(ctx)
	}).Run(ctx)

	return runtimemapping.NewHandler(
		reqDataParser,
		clientProvider.Client(),
		tokenVerifier)
}

func getCertificateResolverHandler(ctx context.Context, cfg config) (istiocertresolver.ValidationHydrator, revocation.Loader, error) {
	k8sClientSet, err := kubernetes.NewKubernetesClientSet(ctx, cfg.KubeConfig.PollInterval, cfg.KubeConfig.PollTimeout, cfg.KubeConfig.Timeout)
	if err != nil {
		return nil, nil, err
	}

	revokedCertsCache := revocation.NewCache()

	revokedCertsConfigMap, err := namespacedname.Parse(cfg.RevocationConfigMapName)
	if err != nil {
		return nil, nil, err
	}

	revokedCertsLoader := revocation.NewRevokedCertificatesLoader(
		revokedCertsCache,
		k8sClientSet.CoreV1().ConfigMaps(revokedCertsConfigMap.Namespace),
		revokedCertsConfigMap.Name,
		time.Second,
	)

	subjectProcessor, err := subject.NewProcessor(cfg.SubjectConsumerMappingConfig, cfg.ExternalIssuerSubject.OrganizationalUnitPattern)
	if err != nil {
		return nil, nil, err
	}

	externalCertHeaderParser := istiocertresolver.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ExternalIssuer,
		istiocertresolver.ExternalCertIssuerSubjectMatcher(cfg.ExternalIssuerSubject), subjectProcessor.AuthIDFromSubjectFunc(), subjectProcessor.AuthSessionExtraFromSubjectFunc())
	connectorCertHeaderParser := istiocertresolver.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ConnectorIssuer,
		istiocertresolver.ConnectorCertificateSubjectMatcher(cfg.CSRSubject), cert.GetCommonName, subjectProcessor.EmptyAuthSessionExtraFunc())

	return istiocertresolver.NewValidationHydrator(revokedCertsCache, connectorCertHeaderParser, externalCertHeaderParser), revokedCertsLoader, nil
}

func getTokenResolverHandler(clientProvider director.ClientProvider) http.Handler {
	return connectortokenresolver.NewValidationHydrator(clientProvider.Client())
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
