package testkit

import (
	"fmt"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalProvisionerURL string `envconfig:"default=http://localhost:3000/graphql"`

	Gardener GardenerConfig

	// Currently Provisioner do not support standalone GCP
	GCP GCPConfig

	QueryLogging bool `envconfig:"default=false"`
}

type KymaConfig struct {
	Version string `envconfig:"default=1.8.0"`
}

type GardenerConfig struct {
	Providers   []string `envconfig:"default=GCP"` // TODO: make Azure and GCP both as default
	AzureSecret string
	GCPSecret   string
}

// GCPConfig specifies config for test on GCP
type GCPConfig struct {
	// Credentials is base64 encoded service account key
	Credentials string
	ProjectName string
}

func (c TestConfig) String() string {
	return fmt.Sprintf("InternalProvisionerURL=%s, QueryLogging=%v, "+
		"GardenerAzureSecret=%v, GardenerGCPSecret=%v",
		c.InternalProvisionerURL, c.QueryLogging,
		c.Gardener.AzureSecret, c.Gardener.GCPSecret)
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
