package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"

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
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/features"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	pkgAuth "github.com/kyma-incubator/compass/components/director/pkg/auth"
	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	configprovider "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/executor"
	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/normalizer"
	oauth "github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	gcli "github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

const discoverSystemsOpMode = "DISCOVER_SYSTEMS"

type config struct {
	APIConfig      systemfetcher.APIConfig
	OAuth2Config   oauth.Config
	SystemFetcher  systemfetcher.Config
	Database       persistence.DatabaseConfig
	TemplateConfig appTemplateConfig

	Log log.Config

	Features features.Config

	ConfigurationFile string

	ConfigurationFileReload time.Duration `envconfig:"default=1m"`
	ClientTimeout           time.Duration `envconfig:"default=60s"`

	CertLoaderConfig certloader.Config

	SelfRegisterDistinguishLabelKey string `envconfig:"APP_SELF_REGISTER_DISTINGUISH_LABEL_KEY"`

	ORDWebhookMappings string `envconfig:"APP_ORD_WEBHOOK_MAPPINGS"`

	ExternalClientCertSecretName string `envconfig:"APP_EXTERNAL_CLIENT_CERT_SECRET_NAME"`
	ExtSvcClientCertSecretName   string `envconfig:"APP_EXT_SVC_CLIENT_CERT_SECRET_NAME"`
}

type appTemplateConfig struct {
	SystemToTemplateMappingsString string `envconfig:"APP_SYSTEM_INFORMATION_SYSTEM_TO_TEMPLATE_MAPPINGS"`
	OverrideApplicationInput       string `envconfig:"APP_TEMPLATE_OVERRIDE_APPLICATION_INPUT"`
	PlaceholderToSystemKeyMappings string `envconfig:"APP_TEMPLATE_PLACEHOLDER_TO_SYSTEM_KEY_MAPPINGS"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to load config"))
	}

	ctx, err := log.Configure(context.Background(), &cfg.Log)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to configure logger"))
	}

	cfgProvider := createAndRunConfigProvider(ctx, cfg)

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to connect to the database"))
	}
	defer func() {
		err := closeFunc()
		if err != nil {
			log.D().Fatal(errors.Wrap(err, "failed to close database connection"))
		}
	}()

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to initialize certificate loader"))
	}

	httpClient := &http.Client{Timeout: cfg.ClientTimeout}
	securedHTTPClient := pkgAuth.PrepareHTTPClient(cfg.ClientTimeout)
	mtlsClient := pkgAuth.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExternalClientCertSecretName)
	extSvcMtlsClient := pkgAuth.PrepareMTLSClient(cfg.ClientTimeout, certCache, cfg.ExtSvcClientCertSecretName)

	sf, err := createSystemFetcher(ctx, cfg, cfgProvider, transact, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient, certCache)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to initialize System Fetcher"))
	}

	if cfg.SystemFetcher.OperationalMode != discoverSystemsOpMode {
		log.C(ctx).Infof("The operatioal mode is set to %q, skipping systems discovery.", cfg.SystemFetcher.OperationalMode)
		return
	}

	if err = sf.SyncSystems(ctx); err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to sync systems"))
	}
}

func createSystemFetcher(ctx context.Context, cfg config, cfgProvider *configprovider.Provider, tx persistence.Transactioner, httpClient, securedHTTPClient, mtlsClient, extSvcMtlsClient *http.Client, certCache certloader.Cache) (*systemfetcher.SystemFetcher, error) {
	ordWebhookMapping, err := application.UnmarshalMappings(cfg.ORDWebhookMappings)
	if err != nil {
		return nil, errors.Wrap(err, "failed while unmarshalling ord webhook mappings")
	}

	uidSvc := uid.NewService()

	tenantConv := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConv)
	tenantSvc := tenant.NewService(tenantRepo, uidSvc)

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
	bundleReferenceConv := bundlereferences.NewConverter()
	runtimeContextConv := runtimectx.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter()

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
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConv)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	intSysSvc := integrationsystem.NewService(intSysRepo, uidSvc)
	assignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, accessstrategy.NewDefaultExecutorProvider(certCache, cfg.ExternalClientCertSecretName, cfg.ExtSvcClientCertSecretName))
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)
	webhookClient := webhookclient.NewClient(securedHTTPClient, mtlsClient, extSvcMtlsClient)
	appTemplateConv := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConv)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo)
	formationAssignmentConv := formationassignment.NewConverter()
	formationAssignmentRepo := formationassignment.NewRepository(formationAssignmentConv)
	formationAssignmentSvc := formationassignment.NewService(formationAssignmentRepo, uidSvc)
	notificationSvc := formation.NewNotificationService(applicationRepo, appTemplateRepo, runtimeRepo, runtimeContextRepo, labelRepo, webhookRepo, webhookConverter, webhookClient)
	formationSvc := formation.NewService(tx, labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, applicationRepo, runtimeRepo, runtimeContextRepo, formationAssignmentSvc, formationAssignmentConv, notificationSvc, cfg.Features.RuntimeTypeLabelKey, cfg.Features.ApplicationTypeLabelKey)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, bundleSvc, uidSvc, formationSvc, cfg.SelfRegisterDistinguishLabelKey, ordWebhookMapping)

	authProvider := pkgAuth.NewMtlsTokenAuthorizationProvider(cfg.OAuth2Config, cfg.ExternalClientCertSecretName, certCache, pkgAuth.DefaultMtlsClientCreator)
	client := &http.Client{
		Transport: httputil.NewSecuredTransport(httputil.NewHTTPTransportWrapper(http.DefaultTransport.(*http.Transport)), authProvider),
		Timeout:   cfg.APIConfig.Timeout,
	}
	oauthMtlsClient := systemfetcher.NewOauthMtlsClient(cfg.OAuth2Config, certCache, client)
	systemsAPIClient := systemfetcher.NewClient(cfg.APIConfig, oauthMtlsClient)

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

	dataLoader := systemfetcher.NewDataLoader(tx, appTemplateSvc, intSysSvc)
	if err := dataLoader.LoadData(ctx, ioutil.ReadDir, ioutil.ReadFile); err != nil {
		return nil, err
	}

	if err := calculateTemplateMappings(ctx, cfg, tx, appTemplateSvc); err != nil {
		return nil, errors.Wrap(err, "failed while calculating application templates mappings")
	}

	var placeholdersMapping []systemfetcher.PlaceholderMapping
	if err := json.Unmarshal([]byte(cfg.TemplateConfig.PlaceholderToSystemKeyMappings), &placeholdersMapping); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling placeholders mapping")
	}

	templateRenderer, err := systemfetcher.NewTemplateRenderer(appTemplateSvc, appConverter, cfg.TemplateConfig.OverrideApplicationInput, placeholdersMapping)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating template renderer")
	}

	return systemfetcher.NewSystemFetcher(tx, tenantSvc, appSvc, templateRenderer, systemsAPIClient, directorClient, cfg.SystemFetcher), nil
}

func createAndRunConfigProvider(ctx context.Context, cfg config) *configprovider.Provider {
	provider := configprovider.NewProvider(cfg.ConfigurationFile)
	err := provider.Load()
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "error on loading configuration file"))
	}
	executor.NewPeriodic(cfg.ConfigurationFileReload, func(ctx context.Context) {
		if err = provider.Load(); err != nil {
			if err != nil {
				log.D().Fatal(errors.Wrap(err, "error from Reloader watch"))
			}
		}
		log.C(ctx).Infof("Successfully reloaded configuration file.")
	}).Run(ctx)

	return provider
}

func calculateTemplateMappings(ctx context.Context, cfg config, transact persistence.Transactioner, appTemplateSvc apptemplate.ApplicationTemplateService) error {
	var systemToTemplateMappings []systemfetcher.TemplateMapping
	if err := json.Unmarshal([]byte(cfg.TemplateConfig.SystemToTemplateMappingsString), &systemToTemplateMappings); err != nil {
		return errors.Wrap(err, "failed to read system template mappings")
	}

	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	for index, tm := range systemToTemplateMappings {
		var region interface{}
		region = tm.Region
		if region == "" {
			region = nil
		}
		appTemplate, err := appTemplateSvc.GetByNameAndRegion(ctx, tm.Name, region)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to retrieve application template with name %q and region %v", tm.Name, region))
		}
		systemToTemplateMappings[index].ID = appTemplate.ID
	}
	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	systemfetcher.Mappings = systemToTemplateMappings
	return nil
}
