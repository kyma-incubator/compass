package testkit

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	SystemBrokerURL      string
	ConnectorURL         string
	InternalConnectorURL string `envconfig:"default=http://compass-connector:3001/graphql"`
}

func ReadConfig() (Config, error) {
	cfg := Config{}

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
