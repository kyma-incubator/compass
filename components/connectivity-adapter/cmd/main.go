package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry"
	connector "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	AppRegistry appregistry.Config
	Connector   connector.Config
}

func main() {
	cfg := config{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading app config")

	router := mux.NewRouter()

	v1Router := router.PathPrefix("/{app-name}/v1").Subrouter()
	v1Router.HandleFunc("/health", health.HandleFunc).Methods(http.MethodGet)

	appRegistryRouter := v1Router.PathPrefix("/metadata").Subrouter()
	appregistry.RegisterHandler(appRegistryRouter, cfg.AppRegistry)

	connectorRouter := v1Router.PathPrefix("/applications").Subrouter()
	err = connector.RegisterHandler(connectorRouter, cfg.Connector)
	if err != nil {
		exitOnError(err, "Failed to init Connector handler")
	}

	log.Printf("Listening on %s", cfg.Address)
	err = http.ListenAndServe(cfg.Address, router)
	exitOnError(err, fmt.Sprintf("while listening on %s", cfg.Address))
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
