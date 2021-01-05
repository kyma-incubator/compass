package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/external-services-mock/internal/apispec"

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
	configChangeRouter.Use(oauthMiddleware)
	configurationchange.InitConfigurationChangeHandler(configChangeRouter, configChangeHandler)

	router.HandleFunc("/audit-log/v2/oauth/token", oauthHandler.Generate).Methods(http.MethodPost)

	router.HandleFunc("/external-api/unsecured/spec", apispec.HandleFunc)

	oauthRouter := router.PathPrefix("/external-api/secured/oauth").Subrouter()
	oauthRouter.Use(oauthMiddleware)
	oauthRouter.HandleFunc("/spec", apispec.HandleFunc)
	oauthRouter.HandleFunc("/token", oauthHandler.Generate).Methods(http.MethodPost)

	basicAuthRouter := router.PathPrefix("/external-api/secured/basic").Subrouter()
	basicAuthRouter.Use(basicAuthMiddleware)
	basicAuthRouter.HandleFunc("/spec", apispec.HandleFunc)

	return router
}

func oauthMiddleware(next http.Handler) http.Handler {
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

func basicAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()

		if !ok {
			httphelpers.WriteError(w, errors.New("No Basic credentials"), http.StatusUnauthorized)
			return
		}
		if len(username) == 0 || len(password) == 0 {
			httphelpers.WriteError(w, errors.New("Empty credentials"), http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
