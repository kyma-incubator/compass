package main

import (
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/fake-external-test-component/internal/configuration"
	"github.com/kyma-incubator/compass/components/fake-external-test-component/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"log"
	"net/http"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")
	handler := initApiHandlers(cfg)
	err = http.ListenAndServe(cfg.Address, handler)
	exitOnError(err, "while running up http server")
}

func exitOnError(err error, context string) {
	if err != nil {
		wrappedError := errors.Wrap(err, context)
		log.Fatal(wrappedError)
	}
}

func initApiHandlers(cfg config) http.Handler {
	router := mux.NewRouter()
	configService := configuration.NewService()
	configHandler := configuration.NewConfigurationHandler(configService)

	router.HandleFunc("/v1/healtz", health.HandleFunc)
	router.HandleFunc("/v1/logs/configuration-change", configHandler.Save).Methods("POST")
	router.HandleFunc("/v1/logs/configuration-change/{id}", configHandler.Get).Methods("GET")
	router.HandleFunc("/v1/logs/configuration-change/{id}", configHandler.Delete).Methods("DELETE")

	//TODO: Add security-event log

	return router
}
