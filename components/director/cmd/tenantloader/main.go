package main

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
	log "github.com/sirupsen/logrus"
)

type jobConfig struct {
	Database persistence.DatabaseConfig
}

func main() {
	const tenantsDirectoryPath = "/data/"

	cfg := jobConfig{}
	err := envconfig.Init(&cfg)
	exitOnError(err, "error while loading app config")

	configureLogger()

	connString := persistence.GetConnString(cfg.Database)
	transact, closeFunc, err := persistence.Configure(log.StandardLogger(), connString)
	exitOnError(err, "error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "error while closing the connection to the database")
	}()

	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	UIDSvc := uid.NewService()
	tenantSvc := tenant.NewService(tenantRepo, UIDSvc)

	tenants, err := externaltenant.MapTenants(tenantsDirectoryPath)
	exitOnError(err, "error while mapping tenants from file")

	tx, err := transact.Begin()
	exitOnError(err, "error while beginning db transaction")
	defer transact.RollbackUnlessCommited(tx)

	ctx := persistence.SaveToContext(context.Background(), tx)

	err = tenantSvc.Sync(ctx, tenants)
	exitOnError(err, "error while synchronising tenants with db")

	err = tx.Commit()
	exitOnError(err, "error while committing the transaction")

	log.Println("Tenants were successfully synchronized.")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func configureLogger() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	log.SetReportCaller(true)
}
