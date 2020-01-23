package main

import (
	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"

	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

type jobConfig struct {
	TenantsSrc     string `envconfig:"default=/data/tenants.json"`
	TenantProvider string `envconfig:"default=dummy"`
}

func main() {
	cfg := jobConfig{}
	err := envconfig.Init(&cfg)
	if err != nil {
		panic(err)
	}

	tenants, err := externaltenant.MapTenants(cfg.TenantsSrc, cfg.TenantProvider)
	if err != nil {
		panic(err)
	}

	log.Println("tenants:", tenants)
}
