package main

import (
	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type jobConfig struct {
	TenantsSrc           string `envconfig:"default=/data/tenants.json"`
	TenantIDKeyMapping   string `envconfig:"default=id"`
	TenantNameKeyMapping string `envconfig:"default=name"`
	TenantProvider       string `envconfig:"default=dummy"`
}

func main() {
	cfg := jobConfig{}
	err := envconfig.Init(&cfg)
	if err != nil {
		panic(err)
	}

	mappingOverrides := externaltenant.MappingOverrides{
		Name: cfg.TenantNameKeyMapping,
		ID:   cfg.TenantIDKeyMapping,
	}

	tenants, err := externaltenant.MapTenants(cfg.TenantsSrc, cfg.TenantProvider, mappingOverrides)
	if err != nil {
		panic(err)
	}

	log.Println("tenants:", tenants)
}
