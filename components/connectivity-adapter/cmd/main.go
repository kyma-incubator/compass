package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry"
	connector "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	ExternalAPIAddress string `envconfig:"default=127.0.0.1:8080"`
	InternalAPIAddress string `envconfig:"default=127.0.0.1:8090"`

	AppRegistry appregistry.Config
	Connector   connector.Config
}

func main() {
	cfg := config{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading app config")

	externalAPIHandler, err := initExternalAPIHandler(cfg)
	if err != nil {
		exitOnError(err, "Failed to init External Connector handler")
	}

	internalAPIHandler, err := initInternalAPIHandler(cfg)
	if err != nil {
		exitOnError(err, "Failed to init Internal Connector handler")
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		log.Printf("External API listening on %s", cfg.ExternalAPIAddress)
		err = http.ListenAndServe(cfg.ExternalAPIAddress, externalAPIHandler)
		exitOnError(err, fmt.Sprintf("while listening on %s", cfg.ExternalAPIAddress))
	}()

	go func() {
		log.Printf("Internal API listening on %s", cfg.InternalAPIAddress)
		err = http.ListenAndServe(cfg.InternalAPIAddress, internalAPIHandler)
		exitOnError(err, fmt.Sprintf("while listening on %s", cfg.InternalAPIAddress))
	}()

	wg.Wait()
}

func initInternalAPIHandler(cfg config) (http.Handler, error) {
	router := mux.NewRouter().PathPrefix("/v1/applications").Subrouter()

	err := connector.RegisterInternalHandler(router, cfg.Connector)
	if err != nil {
		return nil, err
	}

	return router, nil
}

func initExternalAPIHandler(cfg config) (http.Handler, error) {
	router := mux.NewRouter()

	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/health", health.HandleFunc).Methods(http.MethodGet)

	appRegistryRouter := v1Router.PathPrefix("/metadata").Subrouter()
	appregistry.RegisterHandler(appRegistryRouter, cfg.AppRegistry)

	connectorRouter := v1Router.PathPrefix("/applications").Subrouter()
	err := connector.RegisterExternalHandler(connectorRouter, cfg.Connector)
	if err != nil {
		return nil, err
	}

	return connectorRouter, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}
