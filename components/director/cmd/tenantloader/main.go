package main

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/config/tenantloader"

	"github.com/kyma-incubator/compass/components/director/pkg/env"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
)

func main() {
	const tenantsDirectoryPath = "/data/"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	environment, err := env.Default(ctx, tenantloader.AddPFlags)
	exitOnError(err, "Error while creating environment")

	environment.SetEnvPrefix("APP")

	cfg, err := tenantloader.New(environment)
	exitOnError(err, "Error while creating config")

	err = cfg.Validate()
	exitOnError(err, "Error while validating config")

	ctx, err = log.Configure(ctx, cfg.Log)
	exitOnError(err, "Error while crating context with logger")

	transact, closeFunc, err := persistence.Configure(ctx, *cfg.DB)
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
	defer transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = tenantSvc.CreateManyIfNotExists(ctx, tenants)
	exitOnError(err, "error while creating tenants")

	err = tx.Commit()
	exitOnError(err, "error while committing the transaction")

	log.C(ctx).Println("Tenants were successfully synchronized.")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.D().Fatal(wrappedError)
	}
}
