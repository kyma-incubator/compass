package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"time"

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
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
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

	if err := calculateTemplateMappings(ctx, cfg, transact); err != nil {
		log.D().Fatal(err)
	}

	certCache, err := certloader.StartCertLoader(ctx, cfg.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to initialize certificate loader"))
	}

	sf, err := createSystemFetcher(cfg, cfgProvider, transact, &http.Client{Timeout: cfg.ClientTimeout}, certCache)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to initialize System Fetcher"))
	}

	if err = sf.SyncSystems(ctx); err != nil {
		log.D().Fatal(errors.Wrap(err, "failed to sync systems"))
	}
}

func calculateTemplateMappings(ctx context.Context, cfg config, transact persistence.Transactioner) error {
	var systemToTemplateMappings []systemfetcher.TemplateMapping
	if err := json.Unmarshal([]byte(cfg.TemplateConfig.SystemToTemplateMappingsString), &systemToTemplateMappings); err != nil {
		return errors.Wrap(err, "failed to read system template mappings")
	}

	authConverter := auth.NewConverter()
	versionConverter := version.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	webhookConverter := webhook.NewConverter(authConverter)
	specConverter := spec.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	appTemplateConv := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConv)
	webhookRepo := webhook.NewRepository(webhookConverter)
	labelDefConverter := labeldef.NewConverter()
	labelConverter := label.NewConverter()
	labelRepo := label.NewRepository(labelConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)

	uidSvc := uid.NewService()
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo)

	tx, err := transact.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	for index, tm := range systemToTemplateMappings {
		appTemplate, err := appTemplateSvc.GetByName(ctx, tm.Name)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return err
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

func createSystemFetcher(cfg config, cfgProvider *configprovider.Provider, tx persistence.Transactioner, httpClient *http.Client, certCache certloader.Cache) (*systemfetcher.SystemFetcher, error) {
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

	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	assignmentConv := scenarioassignment.NewConverter()
	scenarioAssignmentRepo := scenarioassignment.NewRepository(assignmentConv)
	scenariosSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantRepo, uidSvc, cfg.Features.DefaultScenarioEnabled)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient, accessstrategy.NewDefaultExecutorProvider(certCache))
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo, uidSvc)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)
	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)
	docSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)
	bundleSvc := bundleutil.NewService(bundleRepo, apiSvc, eventAPISvc, docSvc, uidSvc)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, scenariosSvc)
	tntSvc := tenant.NewServiceWithLabels(tenantRepo, uidSvc, labelRepo, labelSvc)
	formationSvc := formation.NewService(labelDefRepo, labelRepo, labelSvc, uidSvc, scenariosSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tntSvc, runtimeRepo, runtimeContextRepo)
	appSvc := application.NewService(&normalizer.DefaultNormalizator{}, cfgProvider, applicationRepo, webhookRepo, runtimeRepo, labelRepo, intSysRepo, labelSvc, scenariosSvc, bundleSvc, uidSvc, formationSvc)
	appTemplateConv := apptemplate.NewConverter(appConverter, webhookConverter)
	appTemplateRepo := apptemplate.NewRepository(appTemplateConv)
	appTemplateSvc := apptemplate.NewService(appTemplateRepo, webhookRepo, uidSvc, labelSvc, labelRepo)

	authProvider := pkgAuth.NewMtlsTokenAuthorizationProvider(cfg.OAuth2Config, certCache, pkgAuth.DefaultMtlsClientCreator)
	client := &http.Client{
		Transport: httputil.NewSecuredTransport(http.DefaultTransport, authProvider),
		Timeout:   cfg.APIConfig.Timeout,
	}
	oauthMtlsClient := systemfetcher.NewOauthMtlsClient(cfg.OAuth2Config, certCache, client)
	systemsAPIClient := systemfetcher.NewClient(cfg.APIConfig, oauthMtlsClient)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: cfg.SystemFetcher.DirectorSkipSSLValidation,
		},
	}

	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewErrorHandlerTransport(tr))

	securedClient := &http.Client{
		Transport: httpTransport,
		Timeout:   cfg.SystemFetcher.DirectorRequestTimeout,
	}

	graphqlClient := gcli.NewClient(cfg.SystemFetcher.DirectorGraphqlURL, gcli.WithHTTPClient(securedClient))
	directorClient := &systemfetcher.DirectorGraphClient{
		Client:        graphqlClient,
		Authenticator: pkgAuth.NewServiceAccountTokenAuthorizationProvider(),
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
