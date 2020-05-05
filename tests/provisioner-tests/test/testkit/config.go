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

	DirectorClient DirectorClientConfig

	// Currently Provisioner do not support standalone GCP
	GCP GCPConfig

	Kyma KymaConfig

	QueryLogging bool `envconfig:"default=false"`
}

type KymaConfig struct {
	Version string `envconfig:"default=1.11.0"`
	// PreUpgradeVersion is used in upgrade test
	PreUpgradeVersion string `envconfig:"default=1.10.0"`
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

type DirectorClientConfig struct {
	URL                        string `envconfig:"default=http://compass-director.compass-system.svc.cluster.local:3000/graphql"`
	Namespace                  string `envconfig:"default=compass-system"`
	OauthCredentialsSecretName string `envconfig:"default=compass-provisioner-credentials"`
}

func (c TestConfig) String() string {
	return fmt.Sprintf("InternalProvisionerURL=%s, QueryLogging=%v, "+
		"GardenerProviders=%v GardenerAzureSecret=%v, GardenerGCPSecret=%v, "+
		"DirectorClientURL=%s, DirectorClientNamespace=%s, DirectorClientOauthCredentialsSecretName=%s",
		c.InternalProvisionerURL, c.QueryLogging,
		c.Gardener.Providers, c.Gardener.AzureSecret, c.Gardener.GCPSecret,
		c.DirectorClient.URL, c.DirectorClient.Namespace, c.DirectorClient.OauthCredentialsSecretName)
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
