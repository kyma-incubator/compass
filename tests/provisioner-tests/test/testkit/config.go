package testkit

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalProvisionerURL string `envconfig:"default=http://localhost:3000/graphql"`
	Tenant                 string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`

	Gardener GardenerConfig

	// Currently Provisioner do not support standalone GCP
	GCP GCPConfig

	Kyma KymaConfig

	QueryLogging bool `envconfig:"default=false"`
}

type KymaConfig struct {
	Version string `envconfig:"default=1.10.0"`
}

type GardenerConfig struct {
	Providers   []string `envconfig:"default=Azure"`
	AzureSecret string   `envconfig:"default=''"`
	GCPSecret   string   `envconfig:"default=''"`
}

// GCPConfig specifies config for test on GCP
type GCPConfig struct {
	// Credentials is base64 encoded service account key
	Credentials string `envconfig:"default=''"`
	ProjectName string `envconfig:"default=''"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("InternalProvisionerURL=%s, QueryLogging=%v, "+
		"GardenerProviders=%v GardenerAzureSecret=%v, GardenerGCPSecret=%v",
		c.InternalProvisionerURL, c.QueryLogging,
		c.Gardener.Providers, c.Gardener.AzureSecret, c.Gardener.GCPSecret)
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
