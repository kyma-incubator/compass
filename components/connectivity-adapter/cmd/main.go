package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/signal"

	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry"
	connector "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/health"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`

	ServerTimeout time.Duration `envconfig:"default=119s"`

	Log log.Config

	AppRegistry appregistry.Config
	Connector   connector.Config
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	cfg := config{}

	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading app config")

	h, err := initAPIHandler(cfg)
	exitOnError(err, "Failed to init External Connector handler")

	handlerWithTimeout, err := handler.WithTimeout(h, cfg.ServerTimeout)
	exitOnError(err, "Failed configuring timeout on handler")

	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}

	ctx, err = log.Configure(ctx, &cfg.Log)
	exitOnError(err, "Failed to configure Logger")
	logger := log.C(ctx)

	logger.Infof("API listening on %s", cfg.Address)
	exitOnError(server.ListenAndServe(), fmt.Sprintf("while listening on %s", cfg.Address))
}

func initAPIHandler(cfg config) (http.Handler, error) {
	router := mux.NewRouter()
	const healthEndpoint = "/v1/health"
	router.HandleFunc(healthEndpoint, health.HandleFunc).Methods(http.MethodGet)
	router.Use(correlation.AttachCorrelationIDToContext(), log.RequestLogger(healthEndpoint))

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
		log.D().Fatal(wrappedError)
	}
}
