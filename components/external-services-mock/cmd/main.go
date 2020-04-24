package main

import (
	"io"
	"log"
	"net/http"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/oauth"

	"github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/security"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configuration"
	"github.com/kyma-incubator/compass/components/external-services-mock/pkg/health"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

type config struct {
	Address string `envconfig:"default=127.0.0.1:8080"`
	OAuthConfig
}
type OAuthConfig struct {
	ClientID     string `envconfig:"APP_AUDITLOG_CLIENT_ID"`
	ClientSecret string `envconfig:"APP_AUDITLOG_CLIENT_SECRET"`
}

func main() {
	cfg := config{}
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")
	handler := initApiHandlers(cfg)
	log.Printf("External Services Mock up and running on address: %s", cfg.Address)
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

	oauthHandler := oauth.NewHandler(cfg.ClientSecret, cfg.ClientID)
	router.Use(authMiddleware)

	router.HandleFunc("/v1/healtz", health.HandleFunc)
	configChangeRouter := router.PathPrefix("/auditlog/v2/configuration-changes").Subrouter()
	configuration.InitConfigurationChangeHandler(configChangeRouter, configHandler)

	router.HandleFunc("/auditlog/v2/security-events", securityHandler.Save).Methods(http.MethodPost)
	router.HandleFunc("/auditlog/v2/security-events", securityHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/auditlog/v2/security-events/{id}", securityHandler.Get).Methods(http.MethodGet)
	router.HandleFunc("/auditlog/v2/security-events/{id}", securityHandler.Delete).Methods(http.MethodDelete)

	router.HandleFunc("/auditlog/v2/oauth/token", oauthHandler.Generate).Methods(http.MethodPost)
	return router
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, err := io.WriteString(w, `{"error":"No auth header"}`)
			exitOnError(err, "while writing auth response")
			return
		}

		next.ServeHTTP(w, r)
	})
}
