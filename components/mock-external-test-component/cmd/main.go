package main

import (
	"log"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/mock-external-test-component/internal/security"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/mock-external-test-component/internal/configuration"
	"github.com/kyma-incubator/compass/components/mock-external-test-component/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
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
	logger := logrus.New()

	router := mux.NewRouter()
	configService := configuration.NewService()
	securityEventService := security.NewService()

	securityHandler := security.NewSecurityEventHandler(securityEventService, logger)
	configHandler := configuration.NewConfigurationHandler(configService, logger)

	router.HandleFunc("/v1/healtz", health.HandleFunc)
	router.HandleFunc("/audit-log/v2/configuration-changes", configHandler.Save).Methods("POST")
	router.HandleFunc("/audit-log/v2/configuration-changes", configHandler.List).Methods("GET")
	router.HandleFunc("/audit-log/v2/configuration-changes/{id}", configHandler.Get).Methods("GET")
	router.HandleFunc("/audit-log/v2/configuration-changes/{id}", configHandler.Delete).Methods("DELETE")

	router.HandleFunc("/audit-log/v2/security-event", securityHandler.Save).Methods("POST")
	router.HandleFunc("/v1/logs/security-event/{id}", securityHandler.Get).Methods("GET")
	router.HandleFunc("/v1/logs/security-event/{id}", securityHandler.Delete).Methods("DELETE")
	return router
}
