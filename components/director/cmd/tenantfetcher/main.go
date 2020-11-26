package main

import (
	"context"
	"time"

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
	Database     persistence.DatabaseConfig
	OAuthConfig  tenantfetcher.OAuth2Config
	APIConfig    tenantfetcher.APIConfig
	QueryConfig  tenantfetcher.QueryConfig
	FieldMapping tenantfetcher.TenantFieldMapping

	Log log.Config

	TenantProvider      string `envconfig:"APP_TENANT_PROVIDER"`
	MetricsPushEndpoint string `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`

	ClientTimeout time.Duration `envconfig:"default=60s"`
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

	transact, closeFunc, err := persistence.Configure(ctx,cfg.Database)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	tenantFetcherSvc := createTenantFetcherSvc(ctx, cfg, transact, metricsPusher)
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

func createTenantFetcherSvc(ctx context.Context, cfg config, transact persistence.Transactioner, metricsPusher *metrics.Pusher) *tenantfetcher.Service {
	uidSvc := uid.NewService()

	tenantStorageConv := tenant.NewConverter()
	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewService(tenantStorageRepo, uidSvc)

	eventAPIClient := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.APIConfig, cfg.ClientTimeout)
	if metricsPusher != nil {
		eventAPIClient.SetMetricsPusher(metricsPusher)
	}

	// tenantFetcherConverter := tenantfetcher.NewConverter(cfg.TenantProvider, cfg.FieldMapping)
	return tenantfetcher.NewService(cfg.QueryConfig, transact, cfg.FieldMapping, cfg.TenantProvider, eventAPIClient, tenantStorageSvc)
}

//func configureLogger() {
//	log.SetFormatter(&log.TextFormatter{
//		FullTimestamp: true,
//	})
//	log.SetReportCaller(true)
//}
