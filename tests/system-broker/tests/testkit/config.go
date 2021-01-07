package testkit

import (
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type Config struct {
	URL string `envconfig:"default=http://compass-connector:3000/graphql"`
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
