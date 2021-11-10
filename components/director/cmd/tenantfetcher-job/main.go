package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql_client"
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
	Database                        persistence.DatabaseConfig
	KubernetesConfig                tenantfetcher.KubeConfig
	OAuthConfig                     tenantfetcher.OAuth2Config
	APIConfig                       tenantfetcher.APIConfig
	QueryConfig                     tenantfetcher.QueryConfig
	TenantFieldMapping              tenantfetcher.TenantFieldMapping
	MovedRuntimeByLabelFieldMapping tenantfetcher.MovedRuntimeByLabelFieldMapping

	Log      log.Config
	Features features.Config

	TenantProvider              string        `envconfig:"APP_TENANT_PROVIDER"`
	AccountsRegion              string        `envconfig:"default=central,APP_ACCOUNT_REGION"`
	SubaccountRegions           []string      `envconfig:"default=central,APP_SUBACCOUNT_REGIONS"`
	MetricsPushEndpoint         string        `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`
	MovedRuntimeLabelKey        string        `envconfig:"default=moved_runtime,APP_MOVED_RUNTIME_LABEL_KEY"`
	TenantInsertChunkSize       int           `envconfig:"default=500,APP_TENANT_INSERT_CHUNK_SIZE"`
	ClientTimeout               time.Duration `envconfig:"default=60s"`
	FullResyncInterval          time.Duration `envconfig:"default=12h"`
	DirectorGraphQLEndpoint     string        `envconfig:"APP_DIRECTOR_GRAPHQL_ENDPOINT"`
	HTTPClientSkipSslValidation bool          `envconfig:"default=false"`

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

	tenantFetcherSvc := createTenantFetcherSvc(cfg, transact, kubeClient, metricsPusher)
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

func createTenantFetcherSvc(cfg config, transact persistence.Transactioner, kubeClient tenantfetcher.KubeClient, metricsPusher *metrics.Pusher) tenantfetcher.TenantSyncService {
	uidSvc := uid.NewService()

	labelDefConverter := labeldef.NewConverter()
	labelDefRepository := labeldef.NewRepository(labelDefConverter)

	labelConverter := label.NewConverter()
	labelRepository := label.NewRepository(labelConverter)
	labelService := label.NewLabelService(labelRepository, labelDefRepository, uidSvc)

	tenantStorageConv := tenant.NewConverter()
	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewServiceWithLabels(tenantStorageRepo, uidSvc, labelRepository, labelService)

	scenarioAssignConv := scenarioassignment.NewConverter()
	scenarioAssignRepo := scenarioassignment.NewRepository(scenarioAssignConv)
	scenarioAssignEngine := scenarioassignment.NewEngine(labelService, labelRepository, scenarioAssignRepo)

	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignConv)
	labelDefService := labeldef.NewService(labelDefRepository, labelRepository, scenarioAssignmentRepo, tenantStorageRepo, uidSvc, cfg.Features.DefaultScenarioEnabled)

	runtimeConverter := runtime.NewConverter()
	runtimeRepository := runtime.NewRepository(runtimeConverter)
	runtimeService := runtime.NewService(runtimeRepository, labelRepository, labelDefService, labelService, uidSvc, scenarioAssignEngine, cfg.Features.ProtectedLabelPattern)

	eventAPIClient := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.APIConfig, cfg.ClientTimeout)
	if metricsPusher != nil {
		eventAPIClient.SetMetricsPusher(metricsPusher)
	}

	gqlClient := newInternalGraphQLClient(cfg.DirectorGraphQLEndpoint, cfg.ClientTimeout, cfg.HTTPClientSkipSslValidation)
	gqlClient.Log = func(s string) {
		log.D().Debug(s)
	}
	directorClient := graphqlclient.NewDirector(gqlClient)

	if cfg.ShouldSyncSubaccounts {
		return tenantfetcher.NewSubaccountService(cfg.QueryConfig, transact, kubeClient, cfg.TenantFieldMapping, cfg.MovedRuntimeByLabelFieldMapping, cfg.TenantProvider, cfg.SubaccountRegions, eventAPIClient, tenantStorageSvc, runtimeService, labelDefService, labelService, cfg.MovedRuntimeLabelKey, cfg.FullResyncInterval, directorClient, cfg.TenantInsertChunkSize)
	}
	return tenantfetcher.NewGlobalAccountService(cfg.QueryConfig, transact, kubeClient, cfg.TenantFieldMapping, cfg.TenantProvider, cfg.AccountsRegion, eventAPIClient, tenantStorageSvc, cfg.FullResyncInterval, directorClient, cfg.TenantInsertChunkSize)
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
