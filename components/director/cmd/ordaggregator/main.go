package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processors"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplateversion"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/operation"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	databuilder "github.com/kyma-incubator/compass/components/director/internal/domain/webhook/datainputbuilder"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	directorTime "github.com/kyma-incubator/compass/components/director/pkg/time"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/authenticator/claims"
	authmiddleware "github.com/kyma-incubator/compass/components/director/pkg/auth-middleware"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
	"github.com/kyma-incubator/compass/components/director/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/auth"
	httputilpkg "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/retry"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address           string `envconfig:"default=127.0.0.1:8080"`
	AggregatorRootAPI string `envconfig:"APP_ROOT_API,default=/ord-aggregator"`

	ServerTimeout   time.Duration `envconfig:"default=110s"`
	ShutdownTimeout time.Duration `envconfig:"default=10s"`

	SecurityConfig securityConfig
	Database       persistence.DatabaseConfig
	Log            log.Config
	Features       features.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	ClientTimeout     time.Duration `envconfig:"default=120s"`
	SkipSSLValidation bool          `envconfig:"default=false"`

	RetryConfig                   retry.Config
	CertLoaderConfig              certloader.Config
	GlobalRegistryConfig          ord.GlobalRegistryConfig
	ElectionConfig                cronjob.ElectionConfig
	MaintainOperationsJobInterval time.Duration `envconfig:"APP_MAINTAIN_OPERATIONS_JOB_INTERVAL,default=60m"`

	ParallelOperationProcessors        int           `envconfig:"APP_PARALLEL_OPERATION_PROCESSORS,default=10"`
	OperationProcessorQuietPeriod      time.Duration `envconfig:"APP_OPERATION_PROCESSORS_QUIET_PERIOD,default=5s"`
	MaxParallelDocumentsPerApplication int           `envconfig:"APP_MAX_PARALLEL_DOCUMENTS_PER_APPLICATION"`
	MaxParallelSpecificationProcessors int           `envconfig:"APP_MAX_PARALLEL_SPECIFICATION_PROCESSORS,default=100"`

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`

	TenantMappingConfigPath                  string `envconfig:"APP_TENANT_MAPPING_CONFIG_PATH"`
	TenantMappingCallbackURL                 string `envconfig:"APP_TENANT_MAPPING_CALLBACK_URL"`
	CredentialExchangeStrategyTenantMappings string `envconfig:"APP_CREDENTIAL_EXCHANGE_STRATEGY_TENANT_MAPPINGS"`

	MetricsConfig           ord.MetricsConfig
	OperationsManagerConfig operationsmanager.OperationsManagerConfig
}

type securityConfig struct {
	JwksEndpoint        string        `envconfig:"APP_JWKS_ENDPOINT"`
	JWKSSyncPeriod      time.Duration `envconfig:"default=5m"`
	AllowJWTSigningNone bool          `envconfig:"APP_ALLOW_JWT_SIGNING_NONE,default=false"`
	AggregatorSyncScope string        `envconfig:"APP_ORD_AGGREGATOR_SYNC_SCOPE,default=ord_aggregator:sync"`
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	tenantMappingConfig, err := apptemplate.UnmarshalTenantMappingConfig(cfg.TenantMappingConfigPath)
	exitOnError(err, "Error while loading Tenant mapping config")

	credentialExchangeStrategyTenantMappings, err := unmarshalMappings(cfg.CredentialExchangeStrategyTenantMappings)
	exitOnError(err, "Error while loading Credential Exchange Strategy Tenant Mappings")

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	exitOnError(err, "failed while unmarshalling ord webhook mappings")

	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	exitOnError(err, "Failed to initialize certificate loader")

	httpClient := &http.Client{
		Timeout: cfg.ClientTimeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: cfg.SkipSSLValidation,
			},
		},
	}

	securedHTTPClient := httputil.PrepareHTTPClientWithSSLValidation(cfg.ClientTimeout, cfg.SkipSSLValidation)
	mtlsClient := httputil.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExternalClientCertSecretName)
	extSvcMtlsClient := httputil.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExtSvcClientCertSecretName)

	accessStrategyExecutorProviderWithTenant := accessstrategy.NewExecutorProviderWithTenant(certCache, ctxTenantProvider, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName)
	accessStrategyExecutorProviderWithoutTenant := accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName)
	retryHTTPExecutor := retry.NewHTTPExecutor(&cfg.RetryConfig)

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
	appTemplateConverter := apptemplate.NewConverter(appConverter, webhookConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	bundleReferenceConv := bundlereferences.NewConverter()
	runtimeContextConv := runtimectx.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter(webhookConverter)
	formationConstraintConverter := formationconstraint.NewConverter()
	formationTemplateConstraintReferencesConverter := formationtemplateconstraintreferences.NewConverter()
	assignmentConv := scenarioassignment.NewConverter()
	tenantConverter := tenant.NewConverter()
	appTemplateVersionConv := apptemplateversion.NewConverter()
	formationAssignmentConv := formationassignment.NewConverter()
	bundleInstanceAuthConv := bundleinstanceauth.NewConverter(authConverter)
	certSubjectMappingConv := certsubjectmapping.NewConverter()

	runtimeRepo := runtime.NewRepository(runtimeConverter)
	applicationRepo := application.NewRepository(appConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConverter)
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

	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)
	formationConstraintRepo := formationconstraint.NewRepository(formationConstraintConverter)
	formationTemplateConstraintReferencesRepo := formationtemplateconstraintreferences.NewRepository(formationTemplateConstraintReferencesConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	tenantRepo := tenant.NewRepository(tenantConverter)
	appTemplateVersionRepo := apptemplateversion.NewRepository(appTemplateVersionConv)
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	bundleInstanceAuthRepo := bundleinstanceauth.NewRepository(bundleInstanceAuthConv)
	certSubjectMappingRepo := certsubjectmapping.NewRepository(certSubjectMappingConv)

	timeSvc := directorTime.NewService()
	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewServiceWithRetry(fetchRequestRepo, httpClient, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc, tenantConverter)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc, tenantSvc, tenantMappingConfig, cfg.TenantMappingCallbackURL)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, bundleInstanceAuthSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc, tenantConverter)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)
	webhookLabelBuilder := databuilder.NewWebhookLabelBuilder(labelRepo)
	webhookTenantBuilder := databuilder.NewWebhookTenantBuilder(webhookLabelBuilder, tenantRepo)
	certSubjectInputBuilder := databuilder.NewWebhookCertSubjectBuilder(certSubjectMappingRepo)
	webhookDataInputBuilder := databuilder.NewWebhookDataInputBuilder(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, webhookLabelBuilder, webhookTenantBuilder, certSubjectInputBuilder)
	formationConstraintSvc := formationconstraint.NewService(formationConstraintRepo, formationTemplateConstraintReferencesRepo, uidSvc, formationConstraintConverter)
	constraintEngine := operators.NewConstraintEngine(transact, formationConstraintSvc, tenantSvc, scenarioAssignmentSvc, nil, nil, formationRepo, labelRepo, labelSvc, applicationRepo, runtimeContextRepo, formationTemplateRepo, formationAssignmentRepo, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsBuilder := formation.NewNotificationsBuilder(webhookConverter, constraintEngine, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	notificationsGenerator := formation.NewNotificationsGenerator(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookDataInputBuilder, notificationsBuilder)
	notificationSvc := formation.NewNotificationService(tenantRepo, webhookClient, notificationsGenerator, constraintEngine, webhookConverter, formationTemplateRepo)
	faNotificationSvc := formationassignment.NewFormationAssignmentNotificationService(formationAssignmentRepo, webhookConverter, webhookRepo, tenantRepo, webhookDataInputBuilder, formationRepo, notificationsBuilder, runtimeContextRepo, labelSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationAssignmentStatusSvc := formationassignment.NewFormationAssignmentStatusService(formationAssignmentRepo, constraintEngine, faNotificationSvc)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc, applicationRepo, runtimeRepo, runtimeContextRepo, notificationSvc, faNotificationSvc, labelSvc, formationRepo, formationAssignmentStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	formationStatusSvc := formation.NewFormationStatusService(formationRepo, labelDefRepo, scenariosSvc, notificationSvc, constraintEngine)
	formationSvc := formation.NewService(transact, applicationRepo, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, faNotificationSvc, notificationSvc, constraintEngine, webhookRepo, formationStatusSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, cfg.SelfRegisterDistinguishLabelKey, ordWebhookMapping)

	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo, applicationRepo)
	appTemplateVersionSvc := apptemplateversion.NewService(appTemplateVersionRepo, appTemplateSvc, uidSvc, timeSvc)

	clientConfig := ord.NewClientConfig(cfg.MaxParallelDocumentsPerApplication)

	ordClientWithTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithTenant)
	ordClientWithoutTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithoutTenant)

	vendorSvc := ordvendor.NewDefaultService()
	vendorProcessor := processors.NewVendorProcessor(transact, vendorSvc)

	productSvc := product.NewDefaultService()
	productProcessor := processors.NewProductProcessor(transact, productSvc)

	packageSvc := ordpackage.NewDefaultService()
	packageProcessor := processors.NewPackageProcessor(transact, packageSvc)

	tombstoneSvc := tombstone.NewDefaultService()
	tombstoneProcessor := processors.NewTombstoneProcessor(transact, tombstoneSvc)

	globalRegistrySvc := ord.NewGlobalRegistryService(transact, cfg.GlobalRegistryConfig, vendorSvc, productSvc, ordClientWithoutTenantExecutor, credentialExchangeStrategyTenantMappings)

	opRepo := operation.NewRepository(operation.NewConverter())
	opSvc := operation.NewService(opRepo, uuid.NewService())

	ordConfig := ord.NewServiceConfig(cfg.MaxParallelSpecificationProcessors, credentialExchangeStrategyTenantMappings)

	ordSvc := ord.NewAggregatorService(ordConfig, cfg.MetricsConfig, transact, appSvc, webhookSvc, bundleSvc, bundleReferenceSvc, apiSvc, eventAPISvc, specSvc, fetchRequestSvc, packageSvc, *packageProcessor, productSvc, *productProcessor, vendorSvc, *vendorProcessor, *tombstoneProcessor, tenantSvc, globalRegistrySvc, ordClientWithTenantExecutor, webhookConverter, appTemplateVersionSvc, appTemplateSvc, labelSvc, ordWebhookMapping, opSvc)
	operationsManager := operationsmanager.NewOperationsManager(transact, opSvc, model.OperationTypeOrdAggregation, cfg.OperationsManagerConfig)

	jwtHTTPClient := &http.Client{
		Transport: httputilpkg.NewCorrelationIDTransport(httputilpkg.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport))),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	onDemandChannel := make(chan string, 100)

	handler := initHandler(ctx, jwtHTTPClient, operationsManager, appSvc, webhookSvc, cfg, transact, onDemandChannel)
	runMainSrv, shutdownMainSrv := createServer(ctx, cfg, handler, "main")

	go func() {
		<-ctx.Done()
		// Interrupt signal received - shut down the servers
		shutdownMainSrv()
	}()

	ordOpProcessor := &ord.OperationsProcessor{
		OrdSvc: ordSvc,
	}
	ordOperationMaintainer := ord.NewOperationMaintainer(model.OperationTypeOrdAggregation, transact, opSvc, webhookSvc, appSvc)

	for i := 0; i < cfg.ParallelOperationProcessors; i++ {
		go func(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *ord.OperationsProcessor, executorIndex int) {
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
						log.C(ctx).Infof("Opeartion %q send for processing through OnDemand channel to executor %d", operationID, executorIndex)
					case <-time.After(cfg.OperationProcessorQuietPeriod):
						log.C(ctx).Infof("Quiet period finished for executor %d", executorIndex)
					}
				}
			}
		}(ctx, operationsManager, ordOpProcessor, i)
	}

	go func() {
		if err := startSyncORDOperationsJob(ctx, ordOperationMaintainer, cfg); err != nil {
			log.C(ctx).WithError(err).Error("Failed to start sync ORD documents cronjob. Stopping app...")
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

	runMainSrv()
}

func claimAndProcessOperation(ctx context.Context, opManager *operationsmanager.OperationsManager, opProcessor *ord.OperationsProcessor) (string, error) {
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
		if errMarkAsFailed := opManager.MarkOperationFailed(ctx, op.ID, errProcess.Error()); errMarkAsFailed != nil {
			log.C(ctx).Errorf("Error while marking operation with id %q as failed. Err: %v", op.ID, errMarkAsFailed)
			return op.ID, errMarkAsFailed
		}
		return op.ID, errProcess
	}
	if errMarkAsCompleted := opManager.MarkOperationCompleted(ctx, op.ID); errMarkAsCompleted != nil {
		log.C(ctx).Errorf("Error while marking operation with id %q as completed. Err: %v", op.ID, errMarkAsCompleted)
		return op.ID, errMarkAsCompleted
	}
	return op.ID, nil
}

func startSyncORDOperationsJob(ctx context.Context, ordOperationMaintainer ord.OperationMaintainer, cfg config) error {
	resyncJob := cronjob.CronJob{
		Name: "SyncORDOperations",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Infof("Start syncing ORD operations...")

			if err := ordOperationMaintainer.Maintain(ctx); err != nil {
				log.C(jobCtx).Errorf("Cannot sync ord operations. Err: %v", err)
			}

			log.C(jobCtx).Infof("ORD documents aggregation finished.")
		},
		SchedulePeriod: cfg.MaintainOperationsJobInterval,
	}
	return cronjob.RunCronJob(ctx, cfg.ElectionConfig, resyncJob)
}

func ctxTenantProvider(ctx context.Context) (string, error) {
	localTenantID, err := tenant.LoadLocalTenantIDFromContext(ctx)
	if err != nil {
		return "", err
	}

	return localTenantID, nil
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

func createServer(ctx context.Context, cfg config, handler http.Handler, name string) (func(), func()) {
	handlerWithTimeout, err := timeouthandler.WithTimeout(handler, cfg.ServerTimeout)
	exitOnError(err, "Error while configuring ord aggregator handler")

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

func initHandler(ctx context.Context, httpClient *http.Client, opMgr *operationsmanager.OperationsManager, appSvc ord.ApplicationService, webhookSvc webhook.WebhookService, cfg config, transact persistence.Transactioner, onDemandChannel chan string) http.Handler {
	const (
		healthzEndpoint   = "/healthz"
		readyzEndpoint    = "/readyz"
		aggregateEndpoint = "/aggregate"
	)
	logger := log.C(ctx)

	mainRouter := mux.NewRouter()
	mainRouter.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(
		cfg.AggregatorRootAPI+healthzEndpoint, cfg.AggregatorRootAPI+readyzEndpoint))

	handler := ord.NewORDAggregatorHTTPHandler(opMgr, appSvc, webhookSvc, transact, onDemandChannel)
	apiRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	configureAuthMiddleware(ctx, httpClient, apiRouter, cfg, cfg.SecurityConfig.AggregatorSyncScope)
	apiRouter.HandleFunc(aggregateEndpoint, handler.ScheduleAggregationForORDData).Methods(http.MethodPost)

	healthCheckRouter := mainRouter.PathPrefix(cfg.AggregatorRootAPI).Subrouter()
	logger.Infof("Registering readiness endpoint...")
	healthCheckRouter.HandleFunc(readyzEndpoint, newReadinessHandler())
	logger.Infof("Registering liveness endpoint...")
	healthCheckRouter.HandleFunc(healthzEndpoint, newReadinessHandler())

	return mainRouter
}

func newReadinessHandler() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}
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

func unmarshalMappings(tenantMappings string) (map[string]ord.CredentialExchangeStrategyTenantMapping, error) {
	var mappingsFromEnv map[string]ord.CredentialExchangeStrategyTenantMapping
	if err := json.Unmarshal([]byte(tenantMappings), &mappingsFromEnv); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling tenant mappings")
	}

	return mappingsFromEnv, nil
}
