package main

import (
	"context"
	"net/http"
	"time"

	httputil "github.com/kyma-incubator/compass/components/director/pkg/http"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"
	mp_bundle "github.com/kyma-incubator/compass/components/director/internal/domain/bundle"
	"github.com/kyma-incubator/compass/components/director/internal/domain/bundleinstanceauth"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

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

	TenantProvider       string        `envconfig:"APP_TENANT_PROVIDER"`
	MetricsPushEndpoint  string        `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`
	MovedRuntimeLabelKey string        `envconfig:"default=moved_runtime,APP_MOVED_RUNTIME_LABEL_KEY"`
	ClientTimeout        time.Duration `envconfig:"default=60s"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	ctx, err := log.Configure(context.Background(), &cfg.Log)

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

	httpClient := &http.Client{
		Timeout:   cfg.ClientTimeout,
		Transport: httputil.NewCorrelationIDTransport(http.DefaultTransport),
	}

	tenantFetcherSvc := createTenantFetcherSvc(cfg, transact, kubeClient, metricsPusher, httpClient)
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

func createTenantFetcherSvc(cfg config, transact persistence.Transactioner, kubeClient tenantfetcher.KubeClient,
	metricsPusher *metrics.Pusher, httpClient *http.Client) *tenantfetcher.Service {
	uidSvc := uid.NewService()

	tenantStorageConv := tenant.NewConverter()
	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewService(tenantStorageRepo, uidSvc)

	labelDefConverter := labeldef.NewConverter()
	labelDefRepository := labeldef.NewRepository(labelDefConverter)
	scenariosDefService := labeldef.NewScenariosService(labelDefRepository, uidSvc, cfg.Features.DefaultScenarioEnabled)

	labelConverter := label.NewConverter()
	labelRepository := label.NewRepository(labelConverter)
	labelUpsertService := label.NewLabelUpsertService(labelRepository, labelDefRepository, uidSvc)

	authConverter := auth.NewConverter()

	frConverter := fetchrequest.NewConverter(authConverter)
	fetchRequestRepo := fetchrequest.NewRepository(frConverter)
	fetchRequestSvc := fetchrequest.NewService(fetchRequestRepo, httpClient)

	specConverter := spec.NewConverter(frConverter)
	specRepo := spec.NewRepository(specConverter)
	specSvc := spec.NewService(specRepo, fetchRequestRepo, uidSvc, fetchRequestSvc)

	versionConverter := version.NewConverter()

	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	eventAPIRepo := eventdef.NewRepository(eventAPIConverter)

	bundleReferenceConv := bundlereferences.NewConverter()
	bundleReferenceRepo := bundlereferences.NewRepository(bundleReferenceConv)
	bundleReferenceSvc := bundlereferences.NewService(bundleReferenceRepo)

	eventAPISvc := eventdef.NewService(eventAPIRepo, uidSvc, specSvc, bundleReferenceSvc)

	apiConverter := api.NewConverter(versionConverter, specConverter)
	apiRepo := api.NewRepository(apiConverter)
	apiSvc := api.NewService(apiRepo, uidSvc, specSvc, bundleReferenceSvc)

	docConverter := document.NewConverter(frConverter)
	docRepo := document.NewRepository(docConverter)
	documentSvc := document.NewService(docRepo, fetchRequestRepo, uidSvc)

	scenariosSvc := label.NewScenarioService(labelRepository)
	labelDefRepo := labeldef.NewRepository(labelDefConverter)
	labelUpsertSvc := label.NewLabelUpsertService(labelRepository, labelDefRepo, uidSvc)
	bundleSvc := mp_bundle.NewService(bundleRepo(), apiSvc, eventAPISvc, documentSvc, uidSvc)
	bundleInstanceAuthSvc := bundleinstanceauth.NewService(bundleInstanceAuthRepo(), uidSvc, bundleSvc, scenariosSvc, labelUpsertSvc)

	scenarioAssignConv := scenarioassignment.NewConverter()
	scenarioAssignRepo := scenarioassignment.NewRepository(scenarioAssignConv)
	scenarioAssignEngine := scenarioassignment.NewEngine(labelUpsertService, labelRepository, scenarioAssignRepo, bundleInstanceAuthSvc)

	scenarioAssignmentRepo := scenarioassignment.NewRepository(scenarioAssignConv)
	labelDefService := labeldef.NewService(labelDefRepository, labelRepository, scenarioAssignmentRepo, scenariosDefService, uidSvc)

	runtimeConverter := runtime.NewConverter()
	runtimeRepository := runtime.NewRepository(runtimeConverter)
	runtimeService := runtime.NewService(runtimeRepository, labelRepository, scenariosDefService, scenariosSvc, labelUpsertService, uidSvc, scenarioAssignEngine, bundleInstanceAuthSvc, cfg.Features.ProtectedLabelPattern)

	eventAPIClient := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.APIConfig, cfg.ClientTimeout)
	if metricsPusher != nil {
		eventAPIClient.SetMetricsPusher(metricsPusher)
	}

	return tenantfetcher.NewService(cfg.QueryConfig, transact, kubeClient, cfg.TenantFieldMapping, cfg.MovedRuntimeByLabelFieldMapping, cfg.TenantProvider, eventAPIClient, tenantStorageSvc, runtimeService, labelDefService, cfg.MovedRuntimeLabelKey)
}

func bundleRepo() mp_bundle.BundleRepository {
	authConverter := auth.NewConverter()
	frConverter := fetchrequest.NewConverter(authConverter)
	versionConverter := version.NewConverter()
	specConverter := spec.NewConverter(frConverter)
	eventAPIConverter := eventdef.NewConverter(versionConverter, specConverter)
	docConverter := document.NewConverter(frConverter)
	apiConverter := api.NewConverter(versionConverter, specConverter)

	return mp_bundle.NewRepository(mp_bundle.NewConverter(authConverter, apiConverter, eventAPIConverter, docConverter))
}

func bundleInstanceAuthRepo() bundleinstanceauth.Repository {
	authConverter := auth.NewConverter()

	return bundleinstanceauth.NewRepository(bundleinstanceauth.NewConverter(authConverter))
}
