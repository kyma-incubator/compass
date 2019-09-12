package testkit

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalConnectorURL  string `envconfig:"default=http://compass-connector:3000/graphql"`
	HydratorURL           string `envconfig:"default=http://compass-connector:8080"`
	ConnectorURL          string
	SecuredConnectorURL   string
	CertificateDataHeader string `envconfig:"default=Certificate-Data"`
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
