package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	mp_authenticator "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"

	"github.com/kyma-incubator/compass/components/director/pkg/applicationtenancy"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	authpkg "github.com/kyma-incubator/compass/components/director/pkg/auth"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"

	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/subscription"

	kube "github.com/kyma-incubator/compass/components/director/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/internal/info"

	gqlgen "github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/dlmiddlecote/sqlstats"
	"github.com/gorilla/mux"
	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	"github.com/kyma-incubator/compass/components/director/internal/domain"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/onetimetoken"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/schema"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	errorpresenter "github.com/kyma-incubator/compass/components/director/internal/error_presenter"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/healthz"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/packagetobundles"
	panichandler "github.com/kyma-incubator/compass/components/director/internal/panic_handler"
	"github.com/kyma-incubator/compass/components/director/internal/statusupdate"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	pkgadapters "github.com/kyma-incubator/compass/components/director/pkg/adapters"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/header"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/kyma-incubator/compass/components/director/pkg/operation/k8s"
	panicrecovery "github.com/kyma-incubator/compass/components/director/pkg/panic_recovery"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/scenario"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/kyma-incubator/compass/components/operations-controller/client"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vrischmann/envconfig"
	cr "sigs.k8s.io/controller-runtime"
)

const envPrefix = "APP"

type config struct {
	Address string `envconfig:"default=127.0.0.1:3000"`

	InternalAddress string `envconfig:"default=127.0.0.1:3002"`
	AppURL          string `envconfig:"APP_URL"`

	ClientTimeout time.Duration `envconfig:"default=105s"`
	ServerTimeout time.Duration `envconfig:"default=110s"`

	Database                persistence.DatabaseConfig
	APIEndpoint             string `envconfig:"default=/graphql"`
	OperationPath           string `envconfig:"default=/operation"`
	LastOperationPath       string `envconfig:"default=/last_operation"`
	PlaygroundAPIEndpoint   string `envconfig:"default=/graphql"`
	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	Log log.Config

	MetricsAddress string `envconfig:"default=127.0.0.1:3003"`
	MetricsConfig  metrics.Config

	JWKSEndpoint          string        `envconfig:"default=file://hack/default-jwks.json"`
	JWKSSyncPeriod        time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone   bool          `envconfig:"default=false"`
	ClientIDHTTPHeaderKey string        `envconfig:"default=client_user,APP_CLIENT_ID_HTTP_HEADER"`

	RuntimeJWKSCachePeriod time.Duration `envconfig:"default=5m"`

	PairingAdapterCfg configprovider.PairingAdapterConfig

	OneTimeToken onetimetoken.Config
	OAuth20      oauth20.Config

	Features features.Config

	SelfRegConfig configprovider.SelfRegConfig

	OperationsNamespace string `envconfig:"default=compass-system"`

	DisableAsyncMode bool `envconfig:"default=false"`

	HealthConfig healthz.Config `envconfig:"APP_HEALTH_CONFIG_INDICATORS"`

	ReadyConfig healthz.ReadyConfig

	InfoConfig info.Config

	FormationMappingCfg formationmapping.Config

	DataloaderMaxBatch int           `envconfig:"default=200"`
	DataloaderWait     time.Duration `envconfig:"default=10ms"`

	CertLoaderConfig certloader.Config

	SubscriptionConfig subscription.Config

	TenantOnDemandConfig tenant.FetchOnDemandAPIConfig

	RetryConfig retry.Config

	SkipSSLValidation bool `envconfig:"default=false,APP_HTTP_CLIENT_SKIP_SSL_VALIDATION"`

	ORDWebhookMappings       string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`
	TenantMappingConfigPath  string `envconfig:"APP_TENANT_MAPPING_CONFIG_PATH"`
	TenantMappingCallbackURL string `envconfig:"APP_TENANT_MAPPING_CALLBACK_URL"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`

	ApplicationTemplateProductLabel string `envconfig:"APP_APPLICATION_TEMPLATE_PRODUCT_LABEL"`
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
	logger := log.C(ctx)

	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	exitOnError(err, "Error while loading ORD Webhook Mappings")

	tenantMappingConfig, err := apptemplate.UnmarshalTenantMappingConfig(cfg.TenantMappingConfigPath)
	exitOnError(err, "Error while loading Tenant mapping config")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	logger.Infof("Registering metrics collectors...")
	metricsCollector := metrics.NewCollector(cfg.MetricsConfig)
	dbStatsCollector := sqlstats.NewStatsCollector("director", transact)
	prometheus.MustRegister(metricsCollector, dbStatsCollector)

	k8sClient, err := kube.NewKubernetesClientSet(ctx, time.Second, time.Minute, time.Minute)
	exitOnError(err, "Error while creating kubernetes client")

	pa, err := getPairingAdaptersMapping(ctx, k8sClient, cfg.PairingAdapterCfg)
	exitOnError(err, "Error while getting pairing adapters configuration")

	startPairingAdaptersWatcher(ctx, k8sClient, pa, cfg.PairingAdapterCfg)

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	securedHTTPClient := authpkg.PrepareHTTPClientWithSSLValidation(cfg.ClientTimeout, cfg.SkipSSLValidation)

	cfg.SelfRegConfig.ClientTimeout = cfg.ClientTimeout

	internalClientTransport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SkipSSLValidation,
		},
	}

	internalFQDNHTTPClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)))),
	}

	internalGatewayHTTPClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(httputil.NewHTTPTransportWrapper(internalClientTransport), mp_authenticator.AuthorizationHeaderKey)),
	}

	appRepo := applicationRepo()

	adminURL, err := url.Parse(cfg.OAuth20.URL)
	exitOnError(err, "Error while parsing Hydra URL")

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	accessStrategyExecutorProvider := accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName)
	retryHTTPExecutor := retry.NewHTTPExecutor(&cfg.RetryConfig)

	mtlsHTTPClient := authpkg.PrepareMTLSClientWithSSLValidation(cfg.ClientTimeout, certCache, cfg.SkipSSLValidation, cfg.ExternalClientCertSecretName)
	extSvcMtlsHTTPClient := authpkg.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExtSvcClientCertSecretName)
	rootResolver, err := domain.NewRootResolver(
		&normalizer.DefaultNormalizator{},
		transact,
		cfgProvider,
		cfg.OneTimeToken,
		cfg.OAuth20,
		pa,
		cfg.Features,
		metricsCollector,
		retryHTTPExecutor,
		httpClient,
		internalFQDNHTTPClient,
		internalGatewayHTTPClient,
		securedHTTPClient,
		mtlsHTTPClient,
		extSvcMtlsHTTPClient,
		cfg.SelfRegConfig,
		cfg.OneTimeToken.Length,
		adminURL,
		accessStrategyExecutorProvider,
		cfg.SubscriptionConfig,
		cfg.TenantOnDemandConfig,
		ordWebhookMapping,
		tenantMappingConfig,
		cfg.TenantMappingCallbackURL,
		cfg.ApplicationTemplateProductLabel,
	)
	exitOnError(err, "Failed to initialize root resolver")

	gqlCfg := graphql.Config{
		Resolvers: rootResolver,
		Directives: graphql.DirectiveRoot{
			Async:                         getAsyncDirective(ctx, cfg, transact, appRepo, tenantMappingConfig),
			HasScenario:                   scenario.NewDirective(transact, label.NewRepository(label.NewConverter()), bundleRepo(), bundleInstanceAuthRepo()).HasScenario,
			HasScopes:                     scope.NewDirective(cfgProvider, &scope.HasScopesErrorProvider{}).VerifyScopes,
			Sanitize:                      scope.NewDirective(cfgProvider, &scope.SanitizeErrorProvider{}).VerifyScopes,
			Validate:                      inputvalidation.NewDirective().Validate,
			SynchronizeApplicationTenancy: applicationtenancy.NewDirective(transact, tenant.NewService(tenant.NewRepository(tenant.NewConverter()), uid.NewService(), tenant.NewConverter()), applicationSvc(transact, cfg, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient, certCache, ordWebhookMapping)).SynchronizeApplicationTenancy,
		},
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)
	claimsValidator := claims.NewValidator(transact, runtimeSvc(transact, cfg, tenantMappingConfig, httpClient, mtlsHTTPClient, extSvcMtlsHTTPClient), runtimeCtxSvc(transact, cfg, httpClient, mtlsHTTPClient, extSvcMtlsHTTPClient), appTemplateSvc(), applicationSvc(transact, cfg, httpClient, mtlsHTTPClient, extSvcMtlsHTTPClient, certCache, ordWebhookMapping), intSystemSvc(), cfg.Features.SubscriptionProviderLabelKey, cfg.Features.ConsumerSubaccountLabelKey, cfg.Features.TokenPrefix)

	logger.Infof("Registering GraphQL endpoint on %s...", cfg.APIEndpoint)
	authMiddleware := mp_authenticator.New(httpClient, cfg.JWKSEndpoint, cfg.AllowJWTSigningNone, cfg.ClientIDHTTPHeaderKey, claimsValidator)

	if cfg.JWKSSyncPeriod != 0 {
		logger.Infof("JWKS synchronization enabled. Sync period: %v", cfg.JWKSSyncPeriod)
		periodicExecutor := executor.NewPeriodic(cfg.JWKSSyncPeriod, func(ctx context.Context) {
			err := authMiddleware.SynchronizeJWKS(ctx)
			if err != nil {
				logger.WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
			}
		})
		go periodicExecutor.Run(ctx)
	}

	packageToBundlesMiddleware := packagetobundles.NewHandler(transact)

	statusMiddleware := statusupdate.New(transact, statusupdate.NewRepository())

	const (
		healthzEndpoint = "/healthz"
		livezEndpoint   = "/livez"
		readyzEndpoint  = "/readyz"
	)

	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", playground.Handler("Dataloader", cfg.PlaygroundAPIEndpoint))

	mainRouter.Use(panicrecovery.NewPanicRecoveryMiddleware(), correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		healthzEndpoint, livezEndpoint, readyzEndpoint,
	), header.AttachHeadersToContext())
	presenter := errorpresenter.NewPresenter(uid.NewService())

	gqlAPIRouter := mainRouter.PathPrefix(cfg.APIEndpoint).Subrouter()
	gqlAPIRouter.Use(authMiddleware.Handler())
	gqlAPIRouter.Use(packageToBundlesMiddleware.Handler())
	gqlAPIRouter.Use(statusMiddleware.Handler())
	gqlAPIRouter.Use(dataloader.HandlerBundle(rootResolver.BundlesDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerAPIDef(rootResolver.APIDefinitionsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerEventDef(rootResolver.EventDefinitionsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerDocument(rootResolver.DocumentsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestAPIDef(rootResolver.FetchRequestAPIDefDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestEventDef(rootResolver.FetchRequestEventDefDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFetchRequestDocument(rootResolver.FetchRequestDocumentDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerRuntimeContext(rootResolver.RuntimeContextsDataloader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFormationAssignment(rootResolver.FormationAssignmentsDataLoader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFormationStatus(rootResolver.StatusDataLoader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	gqlAPIRouter.Use(dataloader.HandlerFormationConstraint(rootResolver.FormationConstraintsDataLoader, cfg.DataloaderMaxBatch, cfg.DataloaderWait))
	operationMiddleware := operation.NewMiddleware(cfg.AppURL + cfg.LastOperationPath)

	gqlServ := handler.NewDefaultServer(executableSchema)
	gqlServ.Use(log.NewGqlLoggingInterceptor())
	gqlServ.Use(metrics.NewInstrumentGraphqlRequestInterceptor(metricsCollector))

	gqlServ.Use(operationMiddleware)
	gqlServ.SetErrorPresenter(presenter.Do)
	gqlServ.SetRecoverFunc(panichandler.RecoverFn)

	gqlAPIRouter.HandleFunc("", metricsCollector.GraphQLHandlerWithInstrumentation(gqlServ))

	operationHandler := operation.NewHandler(transact, func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
		return appRepo.GetByID(ctx, tenantID, resourceID)
	}, tenant.LoadFromContext)

	operationsAPIRouter := mainRouter.PathPrefix(cfg.LastOperationPath).Subrouter()
	operationsAPIRouter.Use(authMiddleware.Handler())
	operationsAPIRouter.HandleFunc("/{resource_type}/{resource_id}", operationHandler.ServeHTTP)

	operationUpdaterHandler := operation.NewUpdateOperationHandler(transact, map[resource.Type]operation.ResourceUpdaterFunc{
		resource.Application: appUpdaterFunc(appRepo),
	}, map[resource.Type]operation.ResourceDeleterFunc{
		resource.Application: func(ctx context.Context, id string) error {
			return appRepo.DeleteGlobal(ctx, id)
		},
	})

	internalRouter := mux.NewRouter()
	internalRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(), header.AttachHeadersToContext())
	internalOperationsAPIRouter := internalRouter.PathPrefix(cfg.OperationPath).Subrouter()
	internalOperationsAPIRouter.HandleFunc("", operationUpdaterHandler.ServeHTTP)

	logger.Infof("Registering readiness endpoint...")
	schemaRepo := schema.NewRepository()
	ready := healthz.NewReady(transact, cfg.ReadyConfig, schemaRepo)
	mainRouter.HandleFunc(readyzEndpoint, healthz.NewReadinessHandler(ready))

	logger.Infof("Registering liveness endpoint...")
	mainRouter.HandleFunc(livezEndpoint, healthz.NewLivenessHandler())

	logger.Infof("Registering health endpoint...")
	health, err := healthz.New(ctx, cfg.HealthConfig)
	exitOnError(err, "Could not initialize health")
	health.RegisterIndicator(healthz.NewIndicator(healthz.DBIndicatorName, healthz.NewDBIndicatorFunc(transact))).Start()
	mainRouter.HandleFunc(healthzEndpoint, healthz.NewHealthHandler(health))

	logger.Infof("Registering info endpoint...")
	mainRouter.HandleFunc(cfg.InfoConfig.APIEndpoint, info.NewInfoHandler(ctx, cfg.InfoConfig, certCache))

	fmAuthMiddleware := createFormationMappingAuthenticator(transact, cfg, appRepo, httpClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	fmHandler := createFormationMappingHandler(transact, appRepo, cfg, httpClient, mtlsHTTPClient, extSvcMtlsHTTPClient)

	asyncFormationAssignmentStatusRouter := mainRouter.PathPrefix(cfg.FormationMappingCfg.AsyncAPIPathPrefix).Subrouter()
	asyncFormationAssignmentStatusRouter.Use(authMiddleware.Handler(), fmAuthMiddleware.FormationAssignmentHandler()) // order is important

	asyncFormationStatusRouter := mainRouter.PathPrefix(cfg.FormationMappingCfg.AsyncAPIPathPrefix).Subrouter()
	asyncFormationStatusRouter.Use(authMiddleware.Handler(), fmAuthMiddleware.FormationHandler()) // order is important

	logger.Infof("Registering formation tenant mapping endpoints...")
	asyncFormationAssignmentStatusRouter.HandleFunc(cfg.FormationMappingCfg.AsyncFormationAssignmentStatusAPIEndpoint, fmHandler.UpdateFormationAssignmentStatus).Methods(http.MethodPatch)
	asyncFormationStatusRouter.HandleFunc(cfg.FormationMappingCfg.AsyncFormationStatusAPIEndpoint, fmHandler.UpdateFormationStatus).Methods(http.MethodPatch)

	examplesServer := http.FileServer(http.Dir("./examples/"))
	mainRouter.PathPrefix("/examples/").Handler(http.StripPrefix("/examples/", examplesServer))

	metricsHandler := http.NewServeMux()
	metricsHandler.Handle("/metrics", promhttp.Handler())

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

func getPairingAdaptersMapping(ctx context.Context, k8sClient *kubernetes.Clientset, adaptersCfg configprovider.PairingAdapterConfig) (*pkgadapters.Adapters, error) {
	logger := log.C(ctx)
	logger.Infof("Getting pairing adapter configuration from the cluster...")
	cm, err := k8sClient.CoreV1().ConfigMaps(adaptersCfg.ConfigmapNamespace).Get(ctx, adaptersCfg.ConfigmapName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	adaptersMap := make(map[string]string)
	if err = json.Unmarshal([]byte(cm.Data[adaptersCfg.ConfigmapKey]), &adaptersMap); err != nil {
		return nil, err
	}

	logger.Infof("Successfully read pairing adapters configuration from the cluster")

	a := pkgadapters.NewAdapters()
	a.Update(adaptersMap)

	logger.Infof("Successfully updated pairing adapters configuration")

	return a, nil
}

func startPairingAdaptersWatcher(ctx context.Context, k8sClient *kubernetes.Clientset, adapters *pkgadapters.Adapters, adaptersCfg configprovider.PairingAdapterConfig) {
	processEventsFunc := func(ctx context.Context, events <-chan watch.Event) {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					return
				}
				switch ev.Type {
				case watch.Added:
					fallthrough
				case watch.Modified:
					log.C(ctx).Info("Updating pairing adapter configuration...")
					cm, ok := ev.Object.(*v1.ConfigMap)
					if !ok {
						log.C(ctx).Error("Unexpected error: object is not configmap. Try again")
						continue
					}
					aCfg, found := cm.Data[adaptersCfg.ConfigmapKey]
					if !found {
						log.C(ctx).Errorf("Did not find the expected key: %s in the pairing adapter configmap", adaptersCfg.ConfigmapKey)
						return
					}
					adaptersCM := make(map[string]string)
					if err := json.Unmarshal([]byte(aCfg), &adaptersCM); err != nil {
						log.C(ctx).Error("error while unmarshalling adapters configuration")
						return
					}
					adapters.Update(adaptersCM)
					log.C(ctx).Info("Successfully updated in memory pairing adapter configuration")
				case watch.Deleted:
					log.C(ctx).Info("Delete event is received, removing pairing adapter configuration")
					adapters.Update(nil)
				case watch.Error:
					log.C(ctx).Error("Error event is received, stop pairing adapter configmap watcher and try again...")
					return
				}
			}
		}
	}

	cmManager := k8sClient.CoreV1().ConfigMaps(adaptersCfg.ConfigmapNamespace)
	w := kube.NewWatcher(ctx, cmManager, processEventsFunc, time.Second, adaptersCfg.ConfigmapName, adaptersCfg.WatcherCorrelationID)
	go w.Run(ctx)
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
			log.C(ctx).WithError(err).Errorf("An error has occurred with %s HTTP server when ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(context.Background()); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while shutting down HTTP server %s: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func bundleInstanceAuthRepo() bundleinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return bundleinstanceauth.NewRepository(bundleinstanceauth.NewConverter(authConverter))
}

func bundleRepo() bundle.BundleRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)

	return bundle.NewRepository(bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
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
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)

	appConverter := application.NewConverter(webhookConverter, bundleConverter)

	return application.NewRepository(appConverter)
}

func webhookService(tenantMappingConfig map[string]interface{}, callbackURL string) webhook.WebhookService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()

	webhookConverter := webhook.NewConverter(authConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)

	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)

	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)
	labelDefinitionConverter := labeldef.NewConverter()
	labelDefinitionRepo := labeldef.NewRepository(labelDefinitionConverter)
	labelSvc := label.NewLabelService(labelRepo, labelDefinitionRepo, uidSvc)

	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	return webhook.NewService(webhookRepo, applicationRepo(), uidSvc, tenantSvc, tenantMappingConfig, callbackURL)
}

func getAsyncDirective(ctx context.Context, cfg config, transact persistence.Transactioner, appRepo application.ApplicationRepository, tenantMappingConfig map[string]interface{}) func(context.Context, interface{}, gqlgen.Resolver, graphql.OperationType, *graphql.WebhookType, *string) (res interface{}, err error) {
	resourceFetcherFunc := func(ctx context.Context, tenantID, resourceID string) (model.Entity, error) {
		return appRepo.GetByID(ctx, tenantID, resourceID)
	}

	scheduler, err := buildScheduler(ctx, cfg)
	exitOnError(err, "Error while creating operations scheduler")

	return operation.NewDirective(transact, webhookService(tenantMappingConfig, cfg.TenantMappingCallbackURL).ListAllApplicationWebhooks, resourceFetcherFunc, appUpdaterFunc(appRepo), tenant.LoadFromContext, scheduler).HandleOperation
}

func buildScheduler(ctx context.Context, config config) (operation.Scheduler, error) {
	if config.DisableAsyncMode {
		log.C(ctx).Info("Async operations are disabled")
		return &operation.DisabledScheduler{}, nil
	}

	cfg, err := cr.GetConfig()
	exitOnError(err, "Failed to get cluster config for operations k8s client")

	cfg.Timeout = config.ClientTimeout
	k8sClient, err := client.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	operationsK8sClient := k8sClient.Operations(config.OperationsNamespace)

	return k8s.NewScheduler(operationsK8sClient), nil
}

func appUpdaterFunc(appRepo application.ApplicationRepository) operation.ResourceUpdaterFunc {
	return func(ctx context.Context, id string, ready bool, errorMsg *string, appStatusCondition model.ApplicationStatusCondition) error {
		app, err := appRepo.GetGlobalByID(ctx, id)
		if err != nil {
			return err
		}
		app.Status = &model.ApplicationStatus{
			Condition: appStatusCondition,
			Timestamp: time.Now(),
		}
		app.Ready = ready
		app.Error = errorMsg
		return appRepo.TechnicalUpdate(ctx, app)
	}
}

func runtimeSvc(transact persistence.Transactioner, cfg config, tenantMappingConfig map[string]interface{}, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient *http.Client) claims.RuntimeService {
	asaConverter := scenarioassignment.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appRepo := application.NewRepository(appConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)

	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelConverter := label.NewConverter()
	labelDefinitionConverter := labeldef.NewConverter()
	runtimeContextConverter := runtimectx.NewConverter()
	runtimeConverter := runtime.NewConverter(webhookConverter)
	tenantConverter := tenant.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()

	asaRepo := scenarioassignment.NewRepository(asaConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefinitionRepo := labeldef.NewRepository(labelDefinitionConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefinitionRepo, uidSvc)
	labelDefinitionSvc := labeldef.NewService(labelDefinitionRepo, labelRepo, asaRepo, tenantRepo, uidSvc)
	asaSvc := scenarioassignment.NewService(asaRepo, labelDefinitionSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)

	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)

	formationAssignmentConv := formationassignment.NewConverter()
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, asaSvc, formationRepo, labelRepo, labelSvc, appRepo, runtimeContextRepo, formationTemplateRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, formationRepo, formationTemplateRepo, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, appRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationSvc := formation.NewService(transact, appRepo, labelDefinitionRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefinitionSvc, asaRepo, asaSvc, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	runtimeContextSvc := runtimectx.NewService(runtimeContextRepo, labelRepo, runtimeRepo, labelSvc, formationSvc, tenantSvc, uidSvc)

	return runtime.NewService(runtimeRepo, labelRepo, labelSvc, uidSvc, formationSvc, tenantSvc, webhookService(tenantMappingConfig, cfg.TenantMappingCallbackURL), runtimeContextSvc, cfg.Features.ProtectedLabelPattern, cfg.Features.ImmutableLabelPattern, cfg.Features.RuntimeTypeLabelKey, cfg.Features.KymaRuntimeTypeLabelValue, cfg.Features.KymaApplicationNamespaceValue)
}

func runtimeCtxSvc(transact persistence.Transactioner, cfg config, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient *http.Client) claims.RuntimeCtxService {
	runtimeContextConverter := runtimectx.NewConverter()
	labelConverter := label.NewConverter()
	labelDefinitionConverter := labeldef.NewConverter()
	asaConverter := scenarioassignment.NewConverter()
	tenantConverter := tenant.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appRepo := application.NewRepository(appConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()

	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefinitionRepo := labeldef.NewRepository(labelDefinitionConverter)
	asaRepo := scenarioassignment.NewRepository(asaConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefinitionRepo, uidSvc)
	labelDefinitionSvc := labeldef.NewService(labelDefinitionRepo, labelRepo, asaRepo, tenantRepo, uidSvc)
	asaSvc := scenarioassignment.NewService(asaRepo, labelDefinitionSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	formationAssignmentConv := formationassignment.NewConverter()
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, asaSvc, formationRepo, labelRepo, labelSvc, appRepo, runtimeContextRepo, formationTemplateRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, formationRepo, formationTemplateRepo, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, appRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationSvc := formation.NewService(transact, appRepo, labelDefinitionRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefinitionSvc, asaRepo, asaSvc, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)

	return runtimectx.NewService(runtimeContextRepo, labelRepo, runtimeRepo, labelSvc, formationSvc, tenantSvc, uidSvc)
}

func appTemplateSvc() claims.ApplicationTemplateService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	versionConverter := version.NewConverter()

	frConverter := fetchrequest.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConv := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConv)

	webhookRepo := webhook.NewRepository(webhookConverter)

	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	appRepo := application.NewRepository(appConverter)

	return apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo, appRepo)
}

func applicationSvc(transact persistence.Transactioner, cfg config, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient *http.Client, certCache certloader.Cache, ordWebhookMapping []application.ORDWebhookMapping) claims.ApplicationService {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)

	intSysConverter := integrationsystem.NewConverter()
	intSysRepo := integrationsystem.NewRepository(intSysConverter)

	webhookRepo := webhook.NewRepository(webhookConverter)

	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	assignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)

	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)

	frConverter := fetchrequest.NewConverter(authConverter)
	docConverter := document.NewConverter(frConverter)
	docRepo := document.NewRepository(docConverter)

	fetchRequestRepo := fetchrequest.NewRepository(frConverter)

	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	apiRepo := api.NewRepository(apiConverter)
	specRepo := spec.NewRepository(specConverter)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, securedHTTPClient, accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName))
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceConv := bundlereferences.NewConverter()
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)

	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	bundleRepo := bundle.NewRepository(bundleConverter)

	bundleSvc := bundle.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)

	runtimeContextConv := runtimectx.NewConverter()
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)

	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	applicationRepo := application.NewRepository(appConverter)

	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)

	formationConstraintConverter := formationconstraint.NewConverter()
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)

	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)

	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	formationAssignmentConv := formationassignment.NewConverter()
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tntSvc, scenarioAssignmentSvc, formationRepo, labelRepo, labelSvc, applicationRepo, runtimeContextRepo, formationTemplateRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, formationRepo, formationTemplateRepo, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationSvc := formation.NewService(transact, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)

	return application.NewService(&normalizer.DefaultNormalizator{}, nil, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, cfg.SelfRegConfig.SelfRegisterDistinguishLabelKey, ordWebhookMapping)
}

func intSystemSvc() claims.IntegrationSystemService {
	intSysConverter := integrationsystem.NewConverter()
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	return integrationsystem.NewService(intSysRepo, uid.NewService())
}

func createFormationMappingAuthenticator(transact persistence.Transactioner, cfg config, appRepo application.ApplicationRepository, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient *http.Client) *formationmapping.Authenticator {
	formationAssignmentConv := formationassignment.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	docConverter := document.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	tenantConverter := tenant.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)

	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelRepo := label.NewRepository(label.NewConverter())
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimectx.NewConverter())
	webhookRepo := webhook.NewRepository(webhookConverter)
	runtimeRepo := runtime.NewRepository(runtime.NewConverter(webhook.NewConverter(auth.NewConverter())))
	tenantRepo := tenant.NewRepository(tenantConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)

	labelDefinitionConverter := labeldef.NewConverter()
	asaConverter := scenarioassignment.NewConverter()
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	labelDefinitionRepo := labeldef.NewRepository(labelDefinitionConverter)
	asaRepo := scenarioassignment.NewRepository(asaConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	uidSvc := uid.NewService()
	labelDefinitionSvc := labeldef.NewService(labelDefinitionRepo, labelRepo, asaRepo, tenantRepo, uidSvc)
	asaSvc := scenarioassignment.NewService(asaRepo, labelDefinitionSvc)
	labelSvc := label.NewLabelService(labelRepo, labelDefinitionRepo, uidSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, asaSvc, formationRepo, labelRepo, labelSvc, appRepo, runtimeContextRepo, formationTemplateRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, nil, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, formationRepo, formationTemplateRepo, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uid.NewService(), appRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)

	return formationmapping.NewFormationMappingAuthenticator(transact, formationAssignmentSvc, runtimeRepo, runtimeContextRepo, appRepo, appTemplateRepo, labelRepo, formationRepo, formationTemplateRepo, tenantRepo, cfg.SubscriptionConfig.ConsumerSubaccountLabelKey)
}

func createFormationMappingHandler(transact persistence.Transactioner, appRepo application.ApplicationRepository, cfg config, securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient *http.Client) *formationmapping.Handler {
	formationAssignmentConv := formationassignment.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	labelDefinitionConverter := labeldef.NewConverter()
	asaConverter := scenarioassignment.NewConverter()
	tenantConverter := tenant.NewConverter()
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()

	labelRepo := label.NewRepository(label.NewConverter())
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	runtimeRepo := runtime.NewRepository(runtime.NewConverter(webhook.NewConverter(auth.NewConverter())))
	runtimeContextRepo := runtimectx.NewRepository(runtimectx.NewConverter())
	webhookRepo := webhook.NewRepository(webhookConverter)
	labelDefRepo := labeldef.NewRepository(labeldef.NewConverter())
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	labelDefinitionRepo := labeldef.NewRepository(labelDefinitionConverter)
	asaRepo := scenarioassignment.NewRepository(asaConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)

	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsHTTPClient, extSvcMtlsHTTPClient)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)

	uidSvc := uid.NewService()
	labelDefinitionSvc := labeldef.NewService(labelDefinitionRepo, labelRepo, asaRepo, tenantRepo, uidSvc)
	asaSvc := scenarioassignment.NewService(asaRepo, labelDefinitionSvc)
	labelSvc := label.NewLabelService(labelRepo, labelDefinitionRepo, uidSvc)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := formationconstraint.NewConstraintEngine(formationConstraintSvc, tenantSvc, asaSvc, formationRepo, labelRepo, labelSvc, appRepo, runtimeContextRepo, formationTemplateRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(appRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, formationRepo, formationTemplateRepo, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uid.NewService(), appRepo, runtimeRepo, runtimeContextRepo, formationAssignmentConv, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationSvc := formation.NewService(transact, appRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefinitionSvc, asaRepo, asaSvc, tenantSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)

	fmHandler := formationmapping.NewFormationMappingHandler(transact, formationAssignmentSvc, formationAssignmentStatusSvc, faNotificationSvc, formationSvc)

	return fmHandler
}
