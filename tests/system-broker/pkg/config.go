package pkg

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	Tenant          string
	SystemBrokerURL string
	DirectorURL     string
	ConnectorURL    string
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
