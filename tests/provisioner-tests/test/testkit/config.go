package testkit

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type TestConfig struct {
	InternalProvisionerUrl string `envconfig:"default=http://localhost:3000/graphql"`

	QueryLogging bool `envconfig:"default=false"`
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
