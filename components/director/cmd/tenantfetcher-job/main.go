package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formation"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	bundleutil "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	graphqlclient "github.com/kyma-incubator/compass/components/director/pkg/graphql_client"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"
	gcli "github.com/machinebox/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/features"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Database                    persistence.DatabaseConfig
	KubernetesConfig            tenantfetcher.KubeConfig
	OAuthConfig                 tenantfetcher.OAuth2Config
	APIConfig                   tenantfetcher.APIConfig
	QueryConfig                 tenantfetcher.QueryConfig
	TenantFieldMapping          tenantfetcher.TenantFieldMapping
	MovedSubaccountFieldMapping tenantfetcher.MovedSubaccountsFieldMapping

	Log      log.Config
	Features features.Config

	AuthMode                    oauth.AuthMode `envconfig:"APP_OAUTH_AUTH_MODE,default=standard"`
	TenantProvider              string         `envconfig:"APP_TENANT_PROVIDER"`
	AccountsRegion              string         `envconfig:"default=central,APP_ACCOUNT_REGION"`
	SubaccountRegions           []string       `envconfig:"default=central,APP_SUBACCOUNT_REGIONS"`
	MetricsPushEndpoint         string         `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`
	TenantInsertChunkSize       int            `envconfig:"default=500,APP_TENANT_INSERT_CHUNK_SIZE"`
	ClientTimeout               time.Duration  `envconfig:"default=60s"`
	FullResyncInterval          time.Duration  `envconfig:"default=12h"`
	DirectorGraphQLEndpoint     string         `envconfig:"APP_DIRECTOR_GRAPHQL_ENDPOINT"`
	HTTPClientSkipSslValidation bool           `envconfig:"default=false"`

	ShouldSyncSubaccounts bool `envconfig:"default=false,APP_SYNC_SUBACCOUNTS"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	ctx, err := log.Configure(context.Background(), &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	var metricsPusher *metrics.Pusher
	if cfg.MetricsPushEndpoint != "" {
		metricsPusher = metrics.NewPusher(cfg.MetricsPushEndpoint, cfg.ClientTimeout)
	}

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	kubeClient, err := tenantfetcher.NewKubernetesClient(ctx, cfg.KubernetesConfig)
	exitOnError(err, "Failed to initialize Kubernetes client")

	tenantFetcherSvc, err := createTenantFetcherSvc(cfg, transact, kubeClient, metricsPusher)
	exitOnError(err, "failed to create tenant fetcher service")

	err = tenantFetcherSvc.SyncTenants()

	if metricsPusher != nil {
		metricsPusher.Push()
	}

	exitOnError(err, "Error while synchronizing tenants in database with tenant changes from events")

	log.C(ctx).Info("Successfully synchronized tenants")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}

func createTenantFetcherSvc(cfg config, transact persistence.Transactioner, kubeClient tenantfetcher.KubeClient, metricsPusher *metrics.Pusher) (tenantfetcher.TenantSyncService, error) {
	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	tenantStorageConverter := tenant.NewConverter()
	labelConverter := label.NewConverter()
	authConverter := auth.NewConverter()
	webhookConverter := webhook.NewConverter(authConverter)
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	bundleConverter := bundleutil.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter)
	appConverter := application.NewConverter(webhookConverter, bundleConverter)
	runtimeConverter := runtime.NewConverter(webhookConverter)
	runtimeContextConverter := runtimectx.NewConverter()
	scenarioAssignConverter := scenarioassignment.NewConverter()
	formationConv := formation.NewConverter()
	formationTemplateConverter := formationtemplate.NewConverter()

	webhookRepo := webhook.NewRepository(webhookConverter)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelRepo := label.NewRepository(labelConverter)
	tenantStorageRepo := tenant.NewRepository(tenantStorageConverter)
	applicationRepo := application.NewRepository(appConverter)
	runtimeRepo := runtime.NewRepository(runtimeConverter)
	runtimeContextRepo := runtimectx.NewRepository(runtimeContextConverter)
	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignConverter)
	formationRepo := formation.NewRepository(formationConv)
	formationTemplateRepo := formationtemplate.NewRepository(formationTemplateConverter)

	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, uidSvc)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepo, labelSvc)
	webhookSvc := webhook.NewService(webhookRepo, applicationRepo, uidSvc)
	labelDefSvc := labeldef.NewService(labelDefRepo, labelRepo, scenarioAssignmentRepo, tenantStorageRepo, uidSvc, cfg.Features.DefaultScenarioEnabled)
	scenarioAssignmentSvc := scenarioassignment.NewService(scenarioAssignmentRepo, labelDefSvc)

	formationSvc := formation.NewService(labelDefRepo, labelRepo, formationRepo, formationTemplateRepo, labelSvc, uidSvc, labelDefSvc, scenarioAssignmentRepo, scenarioAssignmentSvc, tenantStorageSvc, runtimeRepo, runtimeContextRepo)
	runtimeContextSvc := runtimectx.NewService(runtimeContextRepo, labelRepo, labelSvc, formationSvc, tenantStorageSvc, uidSvc)
	runtimeSvc := runtime.NewService(runtimeRepo, labelRepo, labelDefSvc, labelSvc, uidSvc, formationSvc, tenantStorageSvc, webhookSvc, runtimeContextSvc, cfg.Features.ProtectedLabelPattern, cfg.Features.ImmutableLabelPattern)

	eventAPIClient, err := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.AuthMode, cfg.APIConfig, cfg.ClientTimeout)
	if nil != err {
		return nil, err
	}

	if metricsPusher != nil {
		eventAPIClient.SetMetricsPusher(metricsPusher)
	}

	gqlClient := newInternalGraphQLClient(cfg.DirectorGraphQLEndpoint, cfg.ClientTimeout, cfg.HTTPClientSkipSslValidation)
	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}
	directorClient := graphqlclient.NewDirector(gqlClient)

	if cfg.ShouldSyncSubaccounts {
		return tenantfetcher.NewSubaccountService(cfg.QueryConfig, transact, kubeClient, cfg.TenantFieldMapping, cfg.MovedSubaccountFieldMapping, cfg.TenantProvider, cfg.SubaccountRegions, eventAPIClient, tenantStorageSvc, runtimeSvc, labelRepo, cfg.FullResyncInterval, directorClient, cfg.TenantInsertChunkSize, tenantStorageConverter), nil
	}
	return tenantfetcher.NewGlobalAccountService(cfg.QueryConfig, transact, kubeClient, cfg.TenantFieldMapping, cfg.TenantProvider, cfg.AccountsRegion, eventAPIClient, tenantStorageSvc, cfg.FullResyncInterval, directorClient, cfg.TenantInsertChunkSize, tenantStorageConverter), nil
}

func newInternalGraphQLClient(url string, timeout time.Duration, skipSSLValidation bool) *gcli.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipSSLValidation,
		},
	}

	client := &http.Client{
		Transport: httputil.NewCorrelationIDTransport(httputil.NewServiceAccountTokenTransportWithHeader(tr, "Authorization")),
		Timeout:   timeout,
	}

	return gcli.NewClient(url, gcli.WithHTTPClient(client))
}
