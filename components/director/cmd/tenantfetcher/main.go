package main

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Database     persistence.DatabaseConfig
	OAuthConfig  tenantfetcher.OAuth2Config
	APIConfig    tenantfetcher.APIConfig
	FieldMapping tenantfetcher.TenantFieldMapping

	TenantProvider string `envconfig:"APP_TENANT_PROVIDER"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	configureLogger()

	connString := persistence.GetConnString(cfg.Database)
	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), connString)
	exitOnError(err, "Error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "Error while closing the connection to the database")
	}()

	tenantFetcherSvc := createTenantFetcherSvc(cfg, transact)
	err = tenantFetcherSvc.SyncTenants()
	exitOnError(err, "Error while synchronizing tenants in database with tenant changes from events")

	log.Info("Successfully synchronized tenants")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func createTenantFetcherSvc(cfg config, transact persistence.Transactioner) *tenantfetcher.Service {
	uidSvc := uid.NewService()

	tenantStorageConv := tenant.NewConverter()
	tenantStorageRepo := tenant.NewRepository(tenantStorageConv)
	tenantStorageSvc := tenant.NewService(tenantStorageRepo, uidSvc)

	eventAPIClient := tenantfetcher.NewClient(cfg.OAuthConfig, cfg.APIConfig)

	tenantFetcherConverter := tenantfetcher.NewConverter(cfg.TenantProvider, cfg.FieldMapping)
	return tenantfetcher.NewService(transact, tenantFetcherConverter, eventAPIClient, tenantStorageSvc)
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}
