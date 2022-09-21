package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/apptemplate"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/auth"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/retry"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationsystem"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Database persistence.DatabaseConfig

	Log log.Config

	Features features.Config

	ConfigurationFile       string
	ConfigurationFileReload time.Duration `envconfig:"default=1m"`

	ClientTimeout     time.Duration `envconfig:"default=60s"`
	SkipSSLValidation bool          `envconfig:"default=false"`

	RetryConfig retry.Config

	CertLoaderConfig certloader.Config

	GlobalRegistryConfig ord.GlobalRegistryConfig

	MaxParallelWebhookProcessors       int `envconfig:"APP_MAX_PARALLEL_WEBHOOK_PROCESSORS,default=1"`
	MaxParallelDocumentsPerApplication int `envconfig:"APP_MAX_PARALLEL_DOCUMENTS_PER_APPLICATION"`

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	ctx, err := log.Configure(context.Background(), &cfg.Log)
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
	mtlsClient := httputil.PrepareMTLSClient(cfg.ClientTimeout, certCache)
	extSvcMtlsClient := httputil.PrepareExtSvcMTLSClient(cfg.ClientTimeout, certCache)

	accessStrategyExecutorProviderWithTenant := accessstrategy.NewExecutorProviderWithTenant(certCache, ctxTenantProvider)
	accessStrategyExecutorProviderWithoutTenant := accessstrategy.NewDefaultExecutorProvider(certCache)
	retryHTTPExecutor := retry.NewHTTPExecutor(&cfg.RetryConfig)

	ordAggregator := createORDAggregatorSvc(cfgProvider, cfg, transact, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient, accessStrategyExecutorProviderWithTenant, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor, ordWebhookMapping)
	err = ordAggregator.SyncORDDocuments(ctx)
	exitOnError(err, "Error while synchronizing Open Resource Discovery Documents")

	log.C(ctx).Info("Successfully synchronized Open Resource Discovery Documents")
}

func ctxTenantProvider(ctx context.Context) (string, error) {
	tenantPair, err := tenant.LoadTenantPairFromContext(ctx)
	if err != nil {
		return "", err
	}

	return tenantPair.ExternalID, nil
}

func createORDAggregatorSvc(cfgProvider *configprovider.Provider, config config, transact persistence.Transactioner, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient *http.Client, accessStrategyExecutorProviderWithTenant *accessstrategy.Provider, accessStrategyExecutorProviderWithoutTenant *accessstrategy.Provider, retryHTTPExecutor *retry.HTTPExecutor, ordWebhookMapping []application.ORDWebhookMapping) *ord.Service {
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
	pkgConverter := ordpackage.NewConverter()
	productConverter := product.NewConverter()
	vendorConverter := ordvendor.NewConverter()
	tombstoneConverter := tombstone.NewConverter()
	runtimeConverter := runtime.NewConverter(webhookConverter)
	bundleReferenceConv := bundlereferences.NewConverter()
	runtimeContextConv := runtimectx.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter()

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
	pkgRepo := ordpackage.NewRepository(pkgConverter)
	productRepo := product.NewRepository(productConverter)
	vendorRepo := ordvendor.NewRepository(vendorConverter)
	tombstoneRepo := tombstone.NewRepository(tombstoneConverter)
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	assignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	tenantRepo := tenant.NewRepository(tenant.NewConverter())
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, config.Features.DefaultScenarioEnabled)
	fetchRequestSvc := fetchrequest.NewServiceWithRetry(fetchRequestRepo, httpClient, accessStrategyExecutorProviderWithoutTenant, retryHTTPExecutor)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)
	formationSvc := formation.NewService(labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo, webhookRepo, webhookClient, applicationRepo, appTemplateRepo, webhookConverter, config.Features.RuntimeTypeLabelKey, config.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, scenariosSvc, bundleSvc, uidSvc, formationSvc, config.SelfRegisterDistinguishLabelKey, ordWebhookMapping)
	packageSvc := ordpackage.NewService(pkgRepo, uidSvc)
	productSvc := product.NewService(productRepo, uidSvc)
	vendorSvc := ordvendor.NewService(vendorRepo, uidSvc)
	tombstoneSvc := tombstone.NewService(tombstoneRepo, uidSvc)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

	clientConfig := ord.NewClientConfig(config.MaxParallelDocumentsPerApplication)

	ordClientWithTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithTenant)
	ordClientWithoutTenantExecutor := ord.NewClient(clientConfig, httpClient, accessStrategyExecutorProviderWithoutTenant)

	globalRegistrySvc := ord.NewGlobalRegistryService(transact, config.GlobalRegistryConfig, vendorSvc, productSvc, ordClientWithoutTenantExecutor)

	ordConfig := ord.NewServiceConfig(config.MaxParallelWebhookProcessors)
	return ord.NewAggregatorService(ordConfig, transact, appSvc, webhookSvc, bundleSvc, bundleReferenceSvc, apiSvc, eventAPISvc, specSvc, packageSvc, productSvc, vendorSvc, tombstoneSvc, tenantSvc, globalRegistrySvc, ordClientWithTenantExecutor)
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
