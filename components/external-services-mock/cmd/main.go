package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/httphelpers"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configurationchange"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/oauth"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
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

	handler := initHTTP(cfg)
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

func initHTTP(cfg config) http.Handler {
	logger := logrus.New()
	router := mux.NewRouter()
	configChangeSvc := configurationchange.NewService()

	configChangeHandler := configurationchange.NewConfigurationHandler(configChangeSvc, logger)
	oauthHandler := oauth.NewHandler(cfg.ClientSecret, cfg.ClientID)

	router.HandleFunc("/v1/healtz", health.HandleFunc)

	configChangeRouter := router.PathPrefix("/audit-log/v2/configuration-changes").Subrouter()
	configChangeRouter.Use(authMiddleware)
	configurationchange.InitConfigurationChangeHandler(configChangeRouter, configChangeHandler)

	router.HandleFunc("/audit-log/v2/oauth/token", oauthHandler.Generate).Methods(http.MethodPost)
	return router
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if len(authHeader) == 0 {
			httphelpers.WriteError(w, errors.New("No Authorization header"), http.StatusUnauthorized)
			return
		}
		if !strings.Contains(authHeader, "Bearer") {
			httphelpers.WriteError(w, errors.New("No Bearer token"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
