package main

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry"
	connector "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address       string        `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout time.Duration `envconfig:"default=119s"`

	AppRegistry appregistry.Config
	Connector   connector.Config
}

func main() {
	cfg := config{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading app config")

	h, err := initAPIHandler(cfg)
	if err != nil {
		exitOnError(err, "Failed to init External Connector handler")
	}

	handlerWithTimeout, err := handler.WithTimeout(h, cfg.ServerTimeout)
	if err != nil {
		exitOnError(err, "Filed configuring timeout on handler")
	}

	log.Printf("API listening on %s", cfg.Address)
	err = http.ListenAndServe(cfg.Address, handlerWithTimeout)
	exitOnError(err, fmt.Sprintf("while listening on %s", cfg.Address))
}

func initAPIHandler(cfg config) (http.Handler, error) {
	router := mux.NewRouter()
	router.HandleFunc("/v1/health", health.HandleFunc).Methods(http.MethodGet)

	applicationRegistryRouter := router.PathPrefix("/{app-name}/v1").Subrouter()
	connectorRouter := router.PathPrefix("/v1/applications").Subrouter()

	appRegistryRouter := applicationRegistryRouter.PathPrefix("/metadata").Subrouter()
	appregistry.RegisterHandler(appRegistryRouter, cfg.AppRegistry)
	err := connector.RegisterHandler(connectorRouter, cfg.Connector, cfg.AppRegistry.DirectorEndpoint, cfg.AppRegistry.ClientTimeout)
	if err != nil {
		return nil, err
	}

	return router, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
