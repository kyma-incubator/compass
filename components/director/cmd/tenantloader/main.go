package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/domain/labeldef"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"os"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
)

type jobConfig struct {
	Database            persistence.DatabaseConfig
	Log                 log.Config
	TenantLabelsPath    string `envconfig:"APP_TENANT_LABELS_PATH"`
	DefaultTenantRegion string `envconfig:"APP_DEFAULT_TENANT_REGION,default=eu-1"`
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
	tenantSvc := tenant.NewServiceWithLabels(tenantRepo, UIDSvc, labelRepo, labelSvc, tenantConverter)

	tenants, err := externaltenant.MapTenants(tenantsDirectoryPath, cfg.DefaultTenantRegion)
	exitOnError(err, "error while mapping tenants from file")

	tx, err := transact.Begin()
	exitOnError(err, "error while beginning db transaction")
	defer transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	_, err = tenantSvc.CreateManyIfNotExists(ctx, tenants...)
	exitOnError(err, "error while creating tenants")

	tenantLabelsFromEnv, err := os.ReadFile(cfg.TenantLabelsPath)
	exitOnError(err, fmt.Sprintf("while reading tenant labels from file %s", cfg.TenantLabelsPath))

	var tenantLabels []tenantLabel
	if err := json.Unmarshal(tenantLabelsFromEnv, &tenantLabels); err != nil {
		exitOnError(err, fmt.Sprintf("while unmarshalling tenant labels from file %s", cfg.TenantLabelsPath))
	}
	for _, tntLbl := range tenantLabels {
		internalID, err := tenantSvc.GetInternalTenant(ctx, tntLbl.tenantID)
		exitOnError(err, fmt.Sprintf("while getting internal ID for %s", tntLbl.tenantID))
		err = tenantSvc.UpsertLabel(ctx, internalID, tntLbl.key, tntLbl.value)
		exitOnError(err, fmt.Sprintf("while upserting label with key %s and value %s for tenant with extenal ID %s and internal ID %s", tntLbl.key, tntLbl.value, tntLbl.tenantID, internalID))
	}

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

type tenantLabel struct {
	key      string `json:"key"`
	value    string `json:"value"`
	tenantID string `json:"tenantID"`
}
