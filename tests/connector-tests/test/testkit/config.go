package testkit

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	Tenant               string `envconfig:"default=3e64ebae-38b5-46a0-b1ed-9ccee153a0ae"`
	ConnectorURL         string `envconfig:"default=http://compass-connector:3000/graphql"`
	DirectorURL          string `envconfig:"default=http://compass-director:3000/graphql"`
	ConnectorHydratorURL string `envconfig:"default=http://compass-connector:8080"`
	DirectorHydratorURL  string `envconfig:"default=http://compass-director:8080"`

	CertificateDataHeader        string `envconfig:"default=Certificate-Data"`
	RevocationConfigMapName      string `envconfig:"default=revocations-config"`
	RevocationConfigMapNamespace string `envconfig:"default=compass-system"`
}

func ReadConfig() (TestConfig, error) {
	cfg := TestConfig{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "Error while loading app config")

	log.Printf("Read configuration: %+v", cfg)
	return cfg, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
