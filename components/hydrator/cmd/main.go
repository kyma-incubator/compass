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
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	"github.com/kyma-incubator/compass/components/director/pkg/namespacedname"
	"github.com/kyma-incubator/compass/components/hydrator/internal/istiocertresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/revocation"
	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"
	"net/http"
	"os"
	"time"

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

	StaticUsersSrc  string `envconfig:"default=/data/static-users.yaml"`
	StaticGroupsSrc string `envconfig:"default=/data/static-groups.yaml"`

	CSRSubject                   istiocertresolver.CSRSubjectConfig
	ExternalIssuerSubject        istiocertresolver.ExternalIssuerSubjectConfig
	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=compass-system/revocations-Config"`
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

	handler := initAPIHandlers(ctx, authenticators, cfg)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	runMainSrv()
}

func initAPIHandlers(ctx context.Context, authenticators []authenticator.Config, cfg config) http.Handler {
	logger := log.C(ctx)
	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger())

	router := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()
	healthCheckRouter := mainRouter.PathPrefix(cfg.RootAPI).Subrouter()

	registerHydratorHandlers(ctx, router, authenticators, cfg)

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

func registerHydratorHandlers(ctx context.Context, router *mux.Router, authenticators []authenticator.Config, cfg config) {
	logger := log.C(ctx)

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}

	directorClientProvider := director.NewClientProvider(cfg.Director.DirectorURL, cfg.Director.ClientTimeout)
	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	logger.Infof("Registering Authentication Mapping endpoint on %s...", cfg.Handler.AuthenticationMappingEndpoint)
	authnMappingHandlerFunc := authnmappinghandler.NewHandler(oathkeeper.NewReqDataParser(), httpClient, authnmappinghandler.DefaultTokenVerifierProvider, authenticators)

	logger.Infof("Registering Tenant Mapping endpoint on %s...", cfg.Handler.TenantMappingEndpoint)
	tenantMappingHandlerFunc, err := getTenantMappingHandlerFunc(authenticators, directorClientProvider, cfg.StaticUsersSrc, cfg.StaticGroupsSrc, cfgProvider)
	exitOnError(err, "Error while configuring tenant mapping handler")

	logger.Infof("Registering Runtime Mapping endpoint on %s...", cfg.Handler.RuntimeMappingEndpoint)
	runtimeMappingHandlerFunc := getRuntimeMappingHandlerFunc(ctx, directorClientProvider, cfg.JWKSSyncPeriod)

	logger.Infof("Registering Connector Token Resolver endpoint on %s...", cfg.Handler.TokenResolverEndpoint)
	connectorTokenResolverHandlerFunc := getTokenResolverHandler(directorClientProvider)

	logger.Infof("Registering Connector Certificate Resolver endpoint on %s...", cfg.Handler.TokenResolverEndpoint)
	connectorCertResolverHandlerFunc, revokedCertsLoader, err := getCertificateResolverHandler(ctx, cfg)
	exitOnError(err, "Error while configuring tenant mapping handler")

	router.HandleFunc(cfg.Handler.AuthenticationMappingEndpoint, authnMappingHandlerFunc.ServeHTTP)
	router.HandleFunc(cfg.Handler.TenantMappingEndpoint, tenantMappingHandlerFunc.ServeHTTP)
	router.HandleFunc(cfg.Handler.RuntimeMappingEndpoint, runtimeMappingHandlerFunc.ServeHTTP)
	router.HandleFunc(cfg.Handler.TokenResolverEndpoint, connectorTokenResolverHandlerFunc.ServeHTTP)
	router.HandleFunc(cfg.Handler.ValidationIstioCertEndpoint, connectorCertResolverHandlerFunc.ServeHTTP)

	go revokedCertsLoader.Run(ctx)
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func getTenantMappingHandlerFunc(authenticators []authenticator.Config, clientProvider director.ClientProvider, staticUsersSrc string, staticGroupsSrc string, cfgProvider *configprovider.Provider) (*tenantmapping.Handler, error) {
	staticUsersRepo, err := tenantmapping.NewStaticUserRepository(staticUsersSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticUser repository instance")
	}

	staticGroupsRepo, err := tenantmapping.NewStaticGroupRepository(staticGroupsSrc)
	if err != nil {
		return nil, errors.Wrap(err, "while creating StaticGroup repository instance")
	}

	objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
		tenantmapping.UserObjectContextProvider:          tenantmapping.NewUserContextProvider(clientProvider.Client(), staticUsersRepo, staticGroupsRepo),
		tenantmapping.SystemAuthObjectContextProvider:    tenantmapping.NewSystemAuthContextProvider(clientProvider.Client(), cfgProvider),
		tenantmapping.AuthenticatorObjectContextProvider: tenantmapping.NewAuthenticatorContextProvider(clientProvider.Client(), authenticators),
		tenantmapping.CertServiceObjectContextProvider:   tenantmapping.NewCertServiceContextProvider(clientProvider.Client(), cfgProvider),
		tenantmapping.TenantHeaderObjectContextProvider:  tenantmapping.NewAccessLevelContextProvider(clientProvider.Client()),
	}
	reqDataParser := oathkeeper.NewReqDataParser()

	return tenantmapping.NewHandler(reqDataParser, objectContextProviders), nil
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
