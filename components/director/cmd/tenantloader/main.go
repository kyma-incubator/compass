package main

import (
	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

//type TenantMappingService interface { //TODO REMOVE?
//	Sync(ctx context.Context, tenants []externaltenant.TenantMappingInput) error
//}

type jobConfig struct {
	//Database             persistence.DatabaseConfig
	TenantsSrc           string `envconfig:"default=/data/tenants.json"`            //`envconfig:"default=/Users/i500678/go/src/github.com/kyma-incubator/compass/chart/compass/charts/director/tenants.json"`
	TenantKeyMappingFile string `envconfig:"default=/data/tenant-key-mapping.json"` //`envconfig:"default=/Users/i500678/go/src/github.com/kyma-incubator/compass/chart/compass/charts/director/tenant-key-mapping.json"`
	TenantProvider       string `envconfig:"default=compass"`
}

func main() {
	cfg := jobConfig{}
	err := envconfig.Init(&cfg)
	//exitOnError(err, "error while loading app config")
	if err != nil {
		panic(err)
	}

	//configureLogger()

	//connString := persistence.GetConnString(cfg.Database)
	//_, closeFunc, err := persistence.Configure(log.StandardLogger(), connString) //TODO transact 1st var
	//exitOnError(err, "Error while establishing the connection to the database")
	//if err != nil {
	//	panic(err)
	//}
	//defer func() {
	//	err := closeFunc()
	//	//exitOnError(err, "Error while closing the connection to the database")
	//	if err != nil {
	//		panic(err)
	//	}
	//}()

	//tms := setupTenantMappingService() TODO sync

	externalTenantsSrc := cfg.TenantsSrc
	tenantKeyMappingFileSrc := cfg.TenantKeyMappingFile

	mappingOverrides, err := externaltenant.ParseMappingOverrides(tenantKeyMappingFileSrc)
	if err != nil {
		panic(err)
	}

	tenants, err := externaltenant.MapTenants(externalTenantsSrc, cfg.TenantProvider, *mappingOverrides)
	if err != nil {
		panic(err)
	}

	log.Println("tenanty:", tenants)

	//tx, err := transact.Begin()  TODO sync
	//if err != nil {
	//	//TODO
	//	return
	//}
	//defer transact.RollbackUnlessCommited(tx)
	//
	//ctx := persistence.SaveToContext(context.Background(), tx)
	//err = tms.Sync(ctx, tenants)
	//if err != nil {
	//	panic(err)
	//}

}

//func setupTenantMappingService() TenantMappingService {  TODO sync
//	tmc := tenant.NewConverter()
//	tmr := tenant.NewRepository(conv)
//	svc := tenant.NewService(tmr, uidSvc)
//	return svc
//}
