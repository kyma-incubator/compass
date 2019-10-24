package testkit

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalProvisionerURL string `envconfig:"default=http://localhost:3000/graphql"`
	CredentialsNamespace   string `envconfig:"default=compass-system"`

	GCPCredentials string
	GCPProjectName string

	QueryLogging bool `envconfig:"default=false"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("InternalProvisionerURL=%s, CredentialsNamespace=%s, QueryLogging=%v",
		c.InternalProvisionerURL, c.CredentialsNamespace, c.QueryLogging)
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
