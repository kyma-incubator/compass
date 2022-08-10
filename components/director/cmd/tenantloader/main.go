package main

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
)

type jobConfig struct {
	Database            persistence.DatabaseConfig
	Log                 log.Config
	DefaultTenantRegion string
}

func main() {
	const tenantsDirectoryPath = "/data/"

	cfg := jobConfig{}
	err := envconfig.Init(&cfg)
	exitOnError(err, "error while loading app config")

	ctx, err := log.Configure(context.Background(), &cfg.Log)
	exitOnError(err, "Error while configuring logger")

	transact, closeFunc, err := persistence.Configure(ctx, cfg.Database)
	exitOnError(err, "error while establishing the connection to the database")

	defer func() {
		err := closeFunc()
		exitOnError(err, "error while closing the connection to the database")
	}()

	UIDSvc := uid.NewService()

	labelConv := label.NewConverter()
	labelRepo := label.NewRepository(labelConv)
	labelDefConv := labeldef.NewConverter()
	labelDefRepo := labeldef.NewRepository(labelDefConv)
	labelSvc := label.NewLabelService(labelRepo, labelDefRepo, UIDSvc)

	tenantConverter := tenant.NewConverter()
	tenantRepo := tenant.NewRepository(tenantConverter)
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, UIDSvc, labelRepo, labelSvc)

	tenants, err := externaltenant.MapTenants(tenantsDirectoryPath, cfg.DefaultTenantRegion)
	exitOnError(err, "error while mapping tenants from file")

	tx, err := transact.Begin()
	exitOnError(err, "error while beginning db transaction")
	defer transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	err = tenantSvc.CreateManyIfNotExists(ctx, tenants...)
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
