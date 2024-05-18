package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	assignmentOp "github.com/kyma-incubator/compass/components/director/internal/domain/assignmentoperation"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"

	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/jwt"

	directortime "github.com/kyma-incubator/compass/components/director/pkg/time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"

	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/credloader"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemauth"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	oauth "github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const (
	discoverSystemsOpMode = "DISCOVER_SYSTEMS"
)

type config struct {
	Address           string `envconfig:"default=127.0.0.1:8080"`
	AggregatorRootAPI string `envconfig:"APP_ROOT_API,default=/system-fetcher"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	SecurityConfig securityConfig

	ElectionConfig                cronjob.ElectionConfig
	OperationsManagerConfig       operationsmanager.OperationsManagerConfig
	ParallelOperationProcessors   int           `envconfig:"APP_PARALLEL_OPERATION_PROCESSORS,default=10"`
	OperationProcessorQuietPeriod time.Duration `envconfig:"APP_OPERATION_PROCESSORS_QUIET_PERIOD,default=5s"`
	MaintainOperationsJobInterval time.Duration `envconfig:"APP_MAINTAIN_OPERATIONS_JOB_INTERVAL,default=60m"`

	APIConfig           systemfetcher.APIConfig
	OAuth2Config        oauth.Config
	SelfSignedJwtConfig jwt.Config
	SystemFetcher       systemfetcher.Config
	Database            persistence.DatabaseConfig
	TemplateConfig      appTemplateConfig

	Log log.Config

	Features features.Config

	ClientTimeout           time.Duration `envconfig:"default=60s"`

	CertLoaderConfig credloader.CertConfig
	KeyLoaderConfig  credloader.KeysConfig

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
}

type securityConfig struct {
	JwksEndpoint           string        `envconfig:"APP_JWKS_ENDPOINT"`
	JWKSSyncPeriod         time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone    bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	SystemFetcherSyncScope string        `envconfig:"APP_SYSTEM_FETCHER_SCOPE,default=system_fetcher:sync"`
}

type appTemplateConfig struct {
	LabelFilter                    string `envconfig:"APP_TEMPLATE_LABEL_FILTER"`
	OverrideApplicationInput       string `envconfig:"APP_TEMPLATE_OVERRIDE_APPLICATION_INPUT"`
	PlaceholderToSystemKeyMappings string `envconfig:"APP_TEMPLATE_PLACEHOLDER_TO_SYSTEM_KEY_MAPPINGS"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")
	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	uidSvc := uid.NewService()
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	businessTenantMappingSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)

	opRepo := operation.NewRepository(operation.NewConverter())
	opSvc := operation.NewService(opRepo, uidSvc)

	certCache, err := credloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	keyCache, err := credloader.StartKeyLoader(ctx, cfg.KeyLoaderConfig)
	exitOnError(err, "Failed to initialize key loader")

	err = credloader.WaitForKeyCache(keyCache)
	exitOnError(err, "Failed to wait for key cache")

	err = credloader.WaitForCertCache(certCache)
	exitOnError(err, "Failed to wait for cert cache")

	httpClient := &http.Client{Timeout: cfg.ClientTimeout}
	securedHTTPClient := pkgAuth.PrepareHTTPClient(cfg.ClientTimeout)
	mtlsClient := pkgAuth.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExternalClientCertSecretName)

	systemFetcherSvc, err := createSystemFetcher(ctx, cfg, transact, httpClient, securedHTTPClient, mtlsClient, certCache, keyCache)
	exitOnError(err, "Failed to initialize System Fetcher")

	operationsManager := operationsmanager.NewOperationsManager(transact, opSvc, model.OperationTypeSystemFetching, cfg.OperationsManagerConfig)

	onDemandChannel := make(chan string, 100)
	handler := initHandler(ctx, operationsManager, businessTenantMappingSvc, transact, onDemandChannel, cfg)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	if cfg.SystemFetcher.OperationalMode == discoverSystemsOpMode {
		systemFetcherOperationProcessor := &systemfetcher.OperationsProcessor{
			SystemFetcherSvc: systemFetcherSvc,
		}
		systemFetcherOperationMaintainer := systemfetcher.NewOperationMaintainer(model.OperationTypeSystemFetching, transact, opSvc, businessTenantMappingSvc)
		var mutex sync.Mutex
		for i := 0; i < cfg.ParallelOperationProcessors; i++ {
			go func(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *systemfetcher.OperationsProcessor, mutex *sync.Mutex, executorIndex int) {
				for {
					select {
					case <-onDemandChannel:
					default:
					}
					mutex.Lock()
					templateRenderer, err := reloadTemplates(ctx, cfg, transact)
					mutex.Unlock()
					if err != nil {
						log.C(ctx).Errorf("Failed to reload templates by executor %d . Err: %v", executorIndex, err)
					}
					opProcessor.SystemFetcherSvc.SetTemplateRenderer(templateRenderer)

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
			}(ctx, operationsManager, systemFetcherOperationProcessor, &mutex, i)
		}

		go func() {
			if err := startSyncSystemFetcherOperationsJob(ctx, systemFetcherOperationMaintainer, cfg); err != nil {
				log.C(ctx).WithError(err).Error("Failed to start sync System Fetcher cronjob. Stopping app...")
			}
			cancel()
		}()

		go func() {
			if err := operationsManager.StartRescheduleOperationsJob(ctx); err != nil {
				log.C(ctx).WithError(err).Error("Failed to run  RescheduleOperationsJob. Stopping app...")
				cancel()
			}
		}()

		go func() {
			if err := operationsManager.StartRescheduleHangedOperationsJob(ctx); err != nil {
				log.C(ctx).WithError(err).Error("Failed to run RescheduleHangedOperationsJob. Stopping app...")
				cancel()
			}
		}()
	}

	runMainSrv()
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func reloadTemplates(ctx context.Context, cfg config, transact persistence.Transactioner) (*systemfetcher.Renderer, error) {
	uidSvc := uid.NewService()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	versionConverter := version.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)
	labelDefConverter := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	applicationRepo := application.NewRepository(appConverter)
	timeSvc := directortime.NewService()
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo, applicationRepo, timeSvc)
	intSysConverter := integrationsystem.NewConverter()
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc, map[string]interface{}{}, "")

	dataLoader := systemfetcher.NewDataLoader(transact, cfg.SystemFetcher, appTemplateSvc, intSysSvc, webhookSvc)
	if err := dataLoader.LoadData(ctx, os.ReadDir, os.ReadFile); err != nil {
		return nil, errors.Wrapf(err, "while loading template data")
	}

	var placeholdersMapping []systemfetcher.PlaceholderMapping
	err := json.Unmarshal([]byte(cfg.TemplateConfig.PlaceholderToSystemKeyMappings), &placeholdersMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling placeholders mapping")
	}

	templateRenderer, err := systemfetcher.NewTemplateRenderer(appTemplateSvc, appConverter, cfg.TemplateConfig.OverrideApplicationInput, placeholdersMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating template renderer")
	}

	err = calculateTemplateMappings(ctx, cfg, transact, appTemplateSvc, placeholdersMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while calculating application templates mappings")
	}

	return templateRenderer, nil
}

func claimAndProcessOperation(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *systemfetcher.OperationsProcessor) (string, error) {
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
		processingError := &ord.ProcessingError{
			RuntimeError: &ord.RuntimeError{
				Message: errProcess.Error(),
			},
		}
		if errMarkAsFailed := opManager.MarkOperationFailed(ctx, op.ID, processingError); errMarkAsFailed != nil {
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

func startSyncSystemFetcherOperationsJob(ctx context.Context, systemFetcherOperationMaintainer systemfetcher.OperationMaintainer, cfg config) error {
	resyncJob := cronjob.CronJob{
		Name: "SyncSystemFetcherOperations",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Infof("Start syncing System Fetcher operations...")

			if err := systemFetcherOperationMaintainer.Maintain(ctx); err != nil {
				log.C(jobCtx).Errorf("Cannot sync System Fetcher operations. Err: %v", err)
			}

			log.C(jobCtx).Infof("System Fetcher systems sync finished.")
		},
		SchedulePeriod: cfg.MaintainOperationsJobInterval,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, resyncJob)
}

func initHandler(ctx context.Context, opMgr *operationsmanager.OperationsManager, businessTenantMappingSvc systemfetcher.BusinessTenantMappingService, transact persistence.Transactioner, onDemandChannel chan string, cfg config) http.Handler {
	const (
		healthzEndpoint = "/healthz"
		readyzEndpoint  = "/readyz"
		syncEndpoint    = "/sync"
	)
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.AggregatorRootAPI+healthzEndpoint, cfg.AggregatorRootAPI+readyzEndpoint))

	apiRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()

	httpClient := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	configureAuthMiddleware(ctx, httpClient, apiRouter, cfg, cfg.SecurityConfig.SystemFetcherSyncScope)
	logger.Infof("Registering sync endpoint...")
	if cfg.SystemFetcher.OperationalMode == discoverSystemsOpMode {
		logger.Infof("Sync endpoint is enabled.")
		handler := systemfetcher.NewSystemFetcherAggregatorHTTPHandler(opMgr, businessTenantMappingSvc, transact, onDemandChannel, make(chan struct{}, cfg.SystemFetcher.AsyncRequestProcessors))
		apiRouter.HandleFunc(syncEndpoint, handler.ScheduleAggregationForSystemFetcherData).Methods(http.MethodPost)
	} else {
		logger.Infof("Sync endpoint is not enabled.")
		apiRouter.HandleFunc(syncEndpoint, newNotSupportedHandler()).Methods(http.MethodPost)
	}

	healthCheckRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func configureAuthMiddleware(ctx context.Context, httpClient *http.Client, router *mux.Router, cfg config, requiredScopes ...string) {
	scopeValidator := claims.NewScopesValidator(requiredScopes)
	middleware := authmiddleware.New(httpClient, cfg.SecurityConfig.JwksEndpoint, cfg.SecurityConfig.AllowJWTSigningNone, "", scopeValidator)
	router.Use(middleware.Handler())

	log.C(ctx).Infof("JWKS synchronization enabled. Sync period: %v", cfg.SecurityConfig.JWKSSyncPeriod)
	periodicExecutor := executor.NewPeriodic(cfg.SecurityConfig.JWKSSyncPeriod, func(ctx context.Context) {
		if err := middleware.SynchronizeJWKS(ctx); err != nil {
			log.C(ctx).WithError(err).Errorf("An error has occurred while synchronizing JWKS: %v", err)
		}
	})
	go periodicExecutor.Run(ctx)
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
}

func newNotSupportedHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusNotAcceptable)
	}
}

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring system fetcher handler")

	srv := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	runFn := func() {
		log.C(ctx).Infof("Running %s server on %s...", name, cfg.Address)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.C(ctx).Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := srv.Shutdown(ctx); err != nil {
			log.C(ctx).Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}

func createSystemFetcher(ctx context.Context, cfg config, tx persistence.Transactioner, httpClient, securedHTTPClient, mtlsClient *http.Client, certCache credloader.CertCache, keyCache credloader.KeysCache) (*systemfetcher.SystemFetcher, error) {
	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	if err != nil {
		return nil, errors.Wrap(err, "failed while unmarshalling ord webhook mappings")
	}

	tenantConverter := tenant.NewConverter()
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	docConverter := document.NewConverter(frConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	intSysConverter := integrationsystem.NewConverter()
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	bundleReferenceConverter := bundlereferences.NewConverter()
	runtimeContextConverter := runtimectx.NewConverter()
	formationConverter := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	assignmentConverter := scenarioassignment.NewConverter()
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	formationAssignmentConverter := formationassignment.NewConverter()
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	systemsSyncConverter := systemssync.NewConverter()
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)
	certSubjectMappingConv := certsubjectmapping.NewConverter()

	tenantRepo := tenant.NewRepository(tenantConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	applicationRepo := application.NewRepository(appConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	apiRepo := api.NewRepository(apiConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	specRepo := spec.NewRepository(specConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	bundleRepo := bundleutil.NewRepository(bundleConverter)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	formationRepo := formation.NewRepository(formationConverter)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	systemsSyncRepo := systemssync.NewRepository(systemsSyncConverter)
	bundleInstanceAuthRepo := bundleinstanceauth.NewRepository(bundleInstanceAuthConv)
	certSubjectMappingRepo := certsubjectmapping.NewRepository(certSubjectMappingConv)
	systemAuthConverter := systemauth.NewConverter(authConverter)
	systemAuthRepo := systemauth.NewRepository(systemAuthConverter)

	uidSvc := uid.NewService()
	assignmentOperationConv := assignmentOp.NewConverter()
	assignmentOperationRepo := assignmentOp.NewRepository(assignmentOperationConv)
	assignmentOperationSvc := assignmentOp.NewService(assignmentOperationRepo, uidSvc)
	systemAuthSvc := systemauth.NewService(systemAuthRepo, uidSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName))
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, bundleInstanceAuthSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient)
	webhookLabelBuilder := databuilder.NewWebhookLabelBuilder(labelRepo)
	webhookTenantBuilder := databuilder.NewWebhookTenantBuilder(webhookLabelBuilder, tenantRepo)
	certSubjectInputBuilder := databuilder.NewWebhookCertSubjectBuilder(certSubjectMappingRepo)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, webhookLabelBuilder, webhookTenantBuilder, certSubjectInputBuilder)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := operators.NewConstraintEngine(tx, formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, nil, nil, systemAuthSvc, formationRepo, labelRepo, labelSvc, applicationRepo, runtimeContextRepo, formationTemplateRepo, formationAssignmentRepo, nil, nil, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter, formationTemplateRepo, formationAssignmentRepo, formationRepo)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, notificationSvc, faNotificationSvc, assignmentOperationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationStatusSvc := formation.NewFormationStatusService(formationRepo, labelDefRepo, scenariosSvc, notificationSvc, constraintEngine)
	formationSvc := formation.NewService(tx, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, assignmentOperationSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, formationStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, cfg.SelfRegisterDistinguishLabelKey, ordWebhookMapping)
	systemsSyncSvc := systemssync.NewService(systemsSyncRepo)

	constraintEngine.SetFormationAssignmentNotificationService(faNotificationSvc)
	constraintEngine.SetFormationAssignmentService(formationAssignmentSvc)

	authProvider := pkgAuth.NewMtlsTokenAuthorizationProvider(cfg.OAuth2Config, cfg.ExternalClientCertSecretName, certCache, pkgAuth.DefaultMtlsClientCreator)
	jwtAuthProvider := pkgAuth.NewSelfSignedJWTTokenAuthorizationProvider(cfg.SelfSignedJwtConfig)
	client := &http.Client{
		Transport: httputil.NewSecuredTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)), authProvider, jwtAuthProvider),
		Timeout:   cfg.APIConfig.Timeout,
	}
	oauthMtlsClient := systemfetcher.NewOauthMtlsClient(cfg.OAuth2Config, certCache, client)
	jwtTokenClient := systemfetcher.NewJwtTokenClient(keyCache, cfg.KeyLoaderConfig.KeysSecretName, client)
	systemsAPIClient := systemfetcher.NewClient(cfg.APIConfig, oauthMtlsClient, jwtTokenClient)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SystemFetcher.DirectorSkipSSLValidation,
		},
	}

	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewErrorHandlerTransport(httputil.NewHTTPTransportWrapper(tr)))

	securedClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.SystemFetcher.DirectorRequestTimeout,
	}

	graphqlClient := gcli.NewClient(cfg.SystemFetcher.DirectorGraphqlURL, gcli.WithHTTPClient(securedClient))
	directorClient := &systemfetcher.DirectorGraphClient{
		Client:        graphqlClient,
		Authenticator: pkgAuth.NewServiceAccountTokenAuthorizationProvider(),
	}

	templateRenderer, err := reloadTemplates(ctx, cfg, tx)
	if err != nil {
		return nil, errors.Wrapf(err, "while reload templates")
	}

	return systemfetcher.NewSystemFetcher(tx, tenantSvc, appSvc, systemsSyncSvc, templateRenderer, systemsAPIClient, directorClient, cfg.SystemFetcher), nil
}

func calculateTemplateMappings(ctx context.Context, cfg config, transact persistence.Transactioner, appTemplateSvc apptemplate.ApplicationTemplateService, placeholdersMapping []systemfetcher.PlaceholderMapping) error {
	applicationTemplates := make([]systemfetcher.TemplateMapping, 0)

	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	appTemplates, err := appTemplateSvc.ListByFilters(ctx, []*labelfilter.LabelFilter{labelfilter.NewForKey(cfg.TemplateConfig.LabelFilter)})
	if err != nil {
		return errors.Wrapf(err, "while listing application templates by label filter %q", cfg.TemplateConfig.LabelFilter)
	}

	selectFilterProperties := make(map[string]bool, 0)
	for _, appTemplate := range appTemplates {
		labels, err := appTemplateSvc.ListLabels(ctx, appTemplate.ID)
		if err != nil {
			return errors.Wrapf(err, "while listing labels for application template with ID %q", appTemplate.ID)
		}

		slisFilterLabel, slisFilterLabelExists := labels[systemfetcher.SlisFilterLabelKey]
		if !slisFilterLabelExists {
			return errors.Errorf("missing slis filter label for application template with ID %q", appTemplate.ID)
		}

		applicationTemplates = append(applicationTemplates, systemfetcher.TemplateMapping{AppTemplate: appTemplate, Labels: labels})

		productIDFilterMappings := make([]systemfetcher.ProductIDFilterMapping, 0)

		slisFilterLabelJSON, err := json.Marshal(slisFilterLabel.Value)
		if err != nil {
			return err
		}

		err = json.Unmarshal(slisFilterLabelJSON, &productIDFilterMappings)
		if err != nil {
			return err
		}

		for _, mapping := range productIDFilterMappings {
			for _, filter := range mapping.Filter {
				topParent := getTopParentFromJSONPath(filter.Key)
				selectFilterProperties[topParent] = true
			}
		}

		addPropertiesFromAppTemplatePlaceholders(selectFilterProperties, appTemplate.Placeholders)
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	systemfetcher.ApplicationTemplates = applicationTemplates
	systemfetcher.SelectFilter = createSelectFilter(selectFilterProperties, placeholdersMapping)
	systemfetcher.ApplicationTemplateLabelFilter = cfg.TemplateConfig.LabelFilter
	systemfetcher.SystemSourceKey = cfg.APIConfig.SystemSourceKey
	return nil
}

func getTopParentFromJSONPath(jsonPath string) string {
	trimmedJSONPath := strings.TrimPrefix(jsonPath, systemfetcher.TrimPrefix)

	regexForTopParent := regexp.MustCompile(`^[^\[.]+`)
	topParent := regexForTopParent.FindStringSubmatch(trimmedJSONPath)
	if len(topParent) > 0 {
		return topParent[0]
	}

	return trimmedJSONPath
}

func addPropertiesFromAppTemplatePlaceholders(selectFilterProperties map[string]bool, placeholders []model.ApplicationTemplatePlaceholder) {
	for _, placeholder := range placeholders {
		if placeholder.JSONPath != nil && len(*placeholder.JSONPath) > 0 {
			topParent := getTopParentFromJSONPath(*placeholder.JSONPath)
			if _, exists := selectFilterProperties[topParent]; !exists {
				selectFilterProperties[topParent] = true
			}
		}
	}
}

func createSelectFilter(selectFilterProperties map[string]bool, placeholdersMapping []systemfetcher.PlaceholderMapping) []string {
	selectFilter := make([]string, 0)

	for _, pm := range placeholdersMapping {
		topParent := getTopParentFromJSONPath(pm.SystemKey)
		if _, exists := selectFilterProperties[topParent]; !exists {
			selectFilterProperties[topParent] = true
		}
	}

	for property := range selectFilterProperties {
		selectFilter = append(selectFilter, property)
	}

	return selectFilter
}
