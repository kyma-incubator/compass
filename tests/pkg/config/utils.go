package config

import (
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"log"
)

func ReadConfig(conf interface{}){
	err := envconfig.InitWithPrefix(conf, "APP")
	exitOnError(err, "Error while loading app config")

	log.Printf("Read configuration: %+v", conf)
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
