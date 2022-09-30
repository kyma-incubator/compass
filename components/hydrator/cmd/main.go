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

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"
	"github.com/kyma-incubator/compass/components/hydrator/internal/certresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/metrics"
	"github.com/kyma-incubator/compass/components/hydrator/internal/revocation"
	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler"
	"github.com/kyma-incubator/compass/components/hydrator/internal/connectortokenresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/director"
	"github.com/kyma-incubator/compass/components/hydrator/internal/runtimemapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	tenantmappingconst "github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"

	cfg "github.com/kyma-incubator/compass/components/hydrator/internal/config"

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

	Handler cfg.HandlerConfig

	Director director.Config

	JWKSSyncPeriod time.Duration `envconfig:"default=5m"`

	KubeConfig kubernetes.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	StaticGroupsSrc string `envconfig:"default=/data/static-groups.yaml"`

	MetricsConfig metrics.Config

	CSRSubject            subject.CSRSubjectConfig
	ExternalIssuerSubject subject.ExternalIssuerSubjectConfig

	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=compass-system/revocations-config"`
	SubjectConsumerMappingConfig string `envconfig:"default=[]"`

	ConsumerClaimsKeys cfg.ConsumerClaimsKeysConfig

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
	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
	)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthzEndpoint, readyzEndpoint))

	router := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()
	healthCheckRouter := mainRouter.NewRoute().Subrouter()

	registerHydratorHandlers(ctx, router, authenticators, cfg, metricsCollector)

	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())

	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func registerHydratorHandlers(ctx context.Context, router *mux.Router, authenticators []authenticator.Config, cfg config, metricsCollector *metrics.Collector) {
	logger := log.C(ctx)

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
	}

	internalDirectorClientProvider := director.NewClientProvider(cfg.Director.InternalURL, cfg.Director.ClientTimeout, cfg.Director.SkipSSLValidation)
	internalGatewayClientProvider := director.NewClientProvider(cfg.Director.InternalGatewayURL, cfg.Director.ClientTimeout, cfg.Director.SkipSSLValidation)
	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	logger.Infof("Registering Runtime Mapping endpoint on %s...", cfg.Handler.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc := getRuntimeMappingHandlerFunc(ctx, internalDirectorClientProvider, cfg.JWKSSyncPeriod)

	logger.Infof("Registering Authentication Mapping endpoint on %s...", cfg.Handler.AuthenticationMappingEndpoint)
	authnMappingHandlerFunc := authnmappinghandler.NewHandler(oathkeeper.NewReqDataParser(), httpClient, authnmappinghandler.DefaultTokenVerifierProvider, authenticators)

	logger.Infof("Registering Tenant Mapping endpoint on %s...", cfg.Handler.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(authenticators, internalDirectorClientProvider, internalGatewayClientProvider, cfg.StaticGroupsSrc, cfgProvider, cfg.ConsumerClaimsKeys, metricsCollector)
	exitOnError(err, "Error while configuring tenant mapping handler")

	logger.Infof("Registering Certificate Resolver endpoint on %s...", cfg.Handler.CertResolverEndpoint)
	certResolverHandlerFunc, revokedCertsLoader, err := getCertificateResolverHandler(ctx, cfg)
	exitOnError(err, "Error while configuring tenant mapping handler")

	logger.Infof("Registering Connector Token Resolver endpoint on %s...", cfg.Handler.TokenResolverEndpoint)
	connectorTokenResolverHandlerFunc := getTokenResolverHandler(internalDirectorClientProvider)

	router.HandleFunc(cfg.Handler.RuntimeMappingEndpoint, metricsCollector.HandlerInstrumentation(runtimeMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.AuthenticationMappingEndpoint, metricsCollector.HandlerInstrumentation(authnMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.TenantMappingEndpoint, metricsCollector.HandlerInstrumentation(tenantMappingHandlerFunc))
	router.HandleFunc(cfg.Handler.CertResolverEndpoint, metricsCollector.HandlerInstrumentation(certResolverHandlerFunc))
	router.HandleFunc(cfg.Handler.TokenResolverEndpoint, metricsCollector.HandlerInstrumentation(connectorTokenResolverHandlerFunc))

	go revokedCertsLoader.Run(ctx)
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

func getTenantMappingHandlerFunc(authenticators []authenticator.Config, internalDirectorClientProvider, internalGatewayClientProvider director.ClientProvider, staticGroupsSrc string, cfgProvider *configprovider.Provider, consumerClaimsKeysConfig cfg.ConsumerClaimsKeysConfig, metricsCollector *metrics.Collector) (*tenantmapping.Handler, error) {
	staticGroupsRepo, err := tenantmapping.NewStaticGroupRepository(staticGroupsSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticGroup repository instance")
	}

	objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
		tenantmappingconst.UserObjectContextProvider:             tenantmapping.NewUserContextProvider(internalDirectorClientProvider.Client(), staticGroupsRepo),
		tenantmappingconst.SystemAuthObjectContextProvider:       tenantmapping.NewSystemAuthContextProvider(internalDirectorClientProvider.Client(), cfgProvider),
		tenantmappingconst.AuthenticatorObjectContextProvider:    tenantmapping.NewAuthenticatorContextProvider(internalDirectorClientProvider.Client(), authenticators),
		tenantmappingconst.CertServiceObjectContextProvider:      tenantmapping.NewCertServiceContextProvider(internalDirectorClientProvider.Client(), cfgProvider),
		tenantmappingconst.TenantHeaderObjectContextProvider:     tenantmapping.NewAccessLevelContextProvider(internalDirectorClientProvider.Client()),
		tenantmappingconst.ConsumerProviderObjectContextProvider: tenantmapping.NewConsumerContextProvider(internalGatewayClientProvider.Client(), consumerClaimsKeysConfig),
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

func getCertificateResolverHandler(ctx context.Context, cfg config) (certresolver.ValidationHydrator, revocation.Loader, error) {
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

	connectorCertHeaderParser := certresolver.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ConnectorIssuer,
		subject.ConnectorCertificateSubjectMatcher(cfg.CSRSubject), cert.GetCommonName, subjectProcessor.EmptyAuthSessionExtraFunc())
	externalCertHeaderParser := certresolver.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ExternalIssuer,
		subject.ExternalCertIssuerSubjectMatcher(cfg.ExternalIssuerSubject), subjectProcessor.AuthIDFromSubjectFunc(), subjectProcessor.AuthSessionExtraFromSubjectFunc())

	return certresolver.NewValidationHydrator(revokedCertsCache, connectorCertHeaderParser, externalCertHeaderParser), revokedCertsLoader, nil
}

func getTokenResolverHandler(clientProvider director.ClientProvider) http.Handler {
	return connectortokenresolver.NewValidationHydrator(clientProvider.Client())
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

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
