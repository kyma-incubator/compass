package main

import (
	"context"
	"crypto/tls"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
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
	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/retry"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/schema"
	"github.com/kyma-incubator/compass/components/director/internal/healthz"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address            string        `envconfig:"default=127.0.0.1:8080"`
	ShutdownTimeout    time.Duration `envconfig:"default=10s"`
	ReadHeadersTimeout time.Duration `envconfig:"APP_READ_REQUEST_HEADERS_TIMEOUT,default=30s"`
	ClientTimeout      time.Duration `envconfig:"APP_CLIENT_TIMEOUT,default=30s"`

	ORDOpCreationJobSchedulePeriod  time.Duration `envconfig:"APP_ORD_OPERATIONS_CREATION_JOB_SCHEDULE_PERIOD,default=168h"`
	ORDOpDeletionJobSchedulePeriod  time.Duration `envconfig:"APP_ORD_OPERATIONS_DELETION_JOB_SCHEDULE_PERIOD,default=24h"`
	DeleteCompletedOpsOlderThanDays int           `envconfig:"APP_DELETE_COMPLETED_OPERATIONS_OLDER_THAN_DAYS,default=5"`
	DeleteFailedOpsOlderThanDays    int           `envconfig:"APP_DELETE_FAILED_OPERATIONS_OLDER_THAN_DAYS,default=10"`

	SkipSSLValidation               bool          `envconfig:"default=false"`
	ConfigurationFileReload         time.Duration `envconfig:"default=1m"`
	SelfRegisterDistinguishLabelKey string        `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`
	RuntimeTypeLabelKey             string        `envconfig:"APP_RUNTIME_TYPE_LABEL_KEY,default=runtimeType"`
	ApplicationTypeLabelKey         string        `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`

	ORDWebhookMappings       string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`
	TenantMappingConfigPath  string `envconfig:"APP_TENANT_MAPPING_CONFIG_PATH"`
	TenantMappingCallbackURL string `envconfig:"APP_TENANT_MAPPING_CALLBACK_URL"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`

	Log               *log.Config
	Database          persistence.DatabaseConfig
	CertLoaderConfig  certloader.Config
	ReadyConfig       healthz.ReadyConfig
	RetryConfig       retry.Config
	ConfigurationFile string
	Features          features.Config
	ElectionConfig    cronjob.ElectionConfig
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	conf := config{}
	err := envconfig.InitWithPrefix(&conf, "APP")
	exitOnError(err, "while reading operations-manager configuration")

	ctx, err = log.Configure(ctx, conf.Log)
	exitOnError(err, "while configuring logger")

	transact, closeDBConn, err := persistence.Configure(ctx, conf.Database)
	exitOnError(err, "Error while establishing the connection to the database")
	defer func() {
		err := closeDBConn()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	router := mux.NewRouter()

	log.C(ctx).Info("Registering health endpoint...")
	router.Use(correlation.AttachCorrelationIDToContext())
	router.HandleFunc("/healthz", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	log.C(ctx).Info("Registering readiness endpoint...")
	schemaRepo := schema.NewRepository()
	ready := healthz.NewReady(transact, conf.ReadyConfig, schemaRepo)
	router.HandleFunc("/readyz", healthz.NewReadinessHandler(ready))

	cfgProvider := createAndRunConfigProvider(ctx, conf)

	ordWebhookMapping, err := application.UnmarshalMappings(conf.ORDWebhookMappings)
	exitOnError(err, "failed while unmarshalling ord webhook mappings")

	tenantMappingConfig, err := apptemplate.UnmarshalTenantMappingConfig(conf.TenantMappingConfigPath)
	exitOnError(err, "Error while loading Tenant mapping config")

	svc, err := createOperationsManagerService(ctx, cfgProvider, transact, ordWebhookMapping, conf, tenantMappingConfig, conf.TenantMappingCallbackURL)
	if err != nil {
		exitOnError(err, "failed while creating operations manager service")
	}

	runMainSrv, shutdownMainSrv := createServer(ctx, conf, router, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	go func() {
		if err := startCreateORDOperationsJob(ctx, svc, conf); err != nil {
			log.C(ctx).WithError(err).Error("Failed to start create ORD operations cronjob. Stopping app...")
		}
		cancel()
	}()

	go func() {
		if err := startDeleteOldORDOperationsJob(ctx, svc, conf); err != nil {
			log.C(ctx).WithError(err).Error("Failed to start delete old ORD operations cronjob. Stopping app...")
		}
		cancel()
	}()

	log.C(ctx).Infof("Operations Manager has started")
	runMainSrv()
}

func startCreateORDOperationsJob(ctx context.Context, opManager *operationsmanager.Service, cfg config) error {
	job := cronjob.CronJob{
		Name: "CreateORDOperations",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Infof("Starting creation of ORD operations...")
			if err := opManager.CreateORDOperations(ctx); err != nil {
				log.C(jobCtx).WithError(err).Errorf("error occurred while creating Open Resource Discovery operations")
			}
			log.C(jobCtx).Infof("Creation of ORD operations finished.")
		},
		SchedulePeriod: cfg.ORDOpCreationJobSchedulePeriod,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, job)
}

func startDeleteOldORDOperationsJob(ctx context.Context, opManager *operationsmanager.Service, cfg config) error {
	job := cronjob.CronJob{
		Name: "DeleteOldORDOperations",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Infof("Starting deletion of old ORD operations...")
			if err := opManager.DeleteOldOperations(ctx, operationsmanager.OrdAggregationOpType, cfg.DeleteCompletedOpsOlderThanDays, cfg.DeleteFailedOpsOlderThanDays); err != nil {
				log.C(jobCtx).WithError(err).Errorf("error occurred while deleting old Open Resource Discovery operations")
			}
			log.C(jobCtx).Infof("Deletion of old ORD operations finished.")
		},
		SchedulePeriod: cfg.ORDOpDeletionJobSchedulePeriod,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, job)
}

func createOperationsManagerService(ctx context.Context, cfgProvider *configprovider.Provider, transact persistence.Transactioner, ordWebhookMapping []application.ORDWebhookMapping, conf config, tenantMappingConfig map[string]interface{}, callbackURL string) (*operationsmanager.Service, error) {
	retryHTTPExecutor := retry.NewHTTPExecutor(&conf.RetryConfig)

	httpClient := &http.Client{
		Timeout: conf.ClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: conf.SkipSSLValidation,
			},
		},
	}

	certCache, err := certloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		return nil, err
	}
	accessStrategyExecutorProviderWithoutTenant := accessstrategy.NewDefaultExecutorProvider(certCache, conf.ExternalClientCertSecretName, conf.ExtSvcClientCertSecretName)

	securedHTTPClient := httputil.PrepareHTTPClientWithSSLValidation(conf.ClientTimeout, conf.SkipSSLValidation)
	mtlsClient := httputil.PrepareMTLSClient(conf.ClientTimeout, certCache, conf.ExternalClientCertSecretName)
	extSvcMtlsClient := httputil.PrepareMTLSClient(conf.ClientTimeout, certCache, conf.ExtSvcClientCertSecretName)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)

	opConv := operation.NewConverter()
	tenantConverter := tenant.NewConverter()
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
	runtimeConverter := runtime.NewConverter(webhookConverter)
	labelConverter := label.NewConverter()
	intSysConverter := integrationsystem.NewConverter()
	labelDefConverter := labeldef.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationConstraintConverter := formationconstraint.NewConverter()
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	formationConv := formation.NewConverter()
	runtimeContextConv := runtimectx.NewConverter()
	bundleReferenceConv := bundlereferences.NewConverter()
	assignmentConv := scenarioassignment.NewConverter()
	formationAssignmentConv := formationassignment.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)

	opRepo := operation.NewRepository(opConv)
	applicationRepo := application.NewRepository(appConverter)
	webhookRepo := webhook.NewRepository(webhookConverter)
	tenantRepo := tenant.NewRepository(tenantConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	intSysRepo := integrationsystem.NewRepository(intSysConverter)
	apiRepo := api.NewRepository(apiConverter)
	specRepo := spec.NewRepository(specConverter)
	docRepo := document.NewRepository(docConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	bundleRepo := bundleutil.NewRepository(bundleConverter)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
	bundleInstanceAuthRepo := bundleinstanceauth.NewRepository(bundleInstanceAuthConv)

	uidSvc := uid.NewService()
	opSvc := operation.NewService(opRepo, uidSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc, tenantMappingConfig, callbackURL)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewServiceWithRetry(fetchRequestRepo, httpClient, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, bundleInstanceAuthSvc, uidSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := operators.NewConstraintEngine(transact, formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, nil, formationRepo, labelRepo, labelSvc, applicationRepo, runtimeContextRepo, formationTemplateRepo, formationAssignmentRepo, conf.RuntimeTypeLabelKey, conf.ApplicationTypeLabelKey)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, conf.Features.RuntimeTypeLabelKey, conf.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter, formationTemplateRepo)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, conf.Features.RuntimeTypeLabelKey, conf.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, conf.Features.RuntimeTypeLabelKey, conf.Features.ApplicationTypeLabelKey)
	formationStatusSvc := formation.NewFormationStatusService(formationRepo, labelDefRepo, scenariosSvc, notificationSvc, constraintEngine)
	formationSvc := formation.NewService(transact, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, formationStatusSvc, conf.Features.RuntimeTypeLabelKey, conf.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, conf.SelfRegisterDistinguishLabelKey, ordWebhookMapping)

	ordOpCreator := operationsmanager.NewOperationCreator(operationsmanager.OrdCreatorType, transact, opSvc, webhookSvc, appSvc)
	return operationsmanager.NewOperationService(transact, opSvc, ordOpCreator), nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
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

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeadersTimeout,
	}

	runFn := func() {
		log.C(ctx).Infof("Running %s server on %s...", name, cfg.Address)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.C(ctx).Errorf("%s HTTP server ListenAndServe: %v", name, err)
		}
	}

	shutdownFn := func() {
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()

		log.C(ctx).Infof("Shutting down %s server...", name)
		if err := server.Shutdown(ctx); err != nil {
			log.C(ctx).Errorf("%s HTTP server Shutdown: %v", name, err)
		}
	}

	return runFn, shutdownFn
}
