package testkit

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalProvisionerURL string `envconfig:"default=http://localhost:3000/graphql"`
	DirectorURL            string `envconfig:"default=https://gateway.kyma.local/director/graphql"`
	CredentialsNamespace   string `envconfig:"default=compass-system"`

	GCPCredentials string
	GCPProjectName string
	Tenant         string

	HydraPublicURL string `envconfig:"default=http://ory-hydra-public.kyma-system:4444"`
	HydraAdminURL  string `envconfig:"default=http://ory-hydra-admin.kyma-system:4445"`

	QueryLogging bool `envconfig:"default=false"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("InternalProvisionerURL=%s, DirectorURL=%s, CredentialsNamespace=%s, "+
		"Tenant=%s, HydraPublicURL=%s, HydraAdminURL=%s, QueryLogging=%v",
		c.InternalProvisionerURL, c.DirectorURL, c.CredentialsNamespace,
		c.Tenant, c.HydraPublicURL, c.HydraAdminURL, c.QueryLogging)
}

func ReadConfig() (TestConfig, error) {
	cfg := TestConfig{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	if err != nil {
		return TestConfig{}, errors.Wrap(err, "Error while loading app config")
	}

	log.Printf("Read configuration: %s", cfg.String())
	return cfg, nil
}
