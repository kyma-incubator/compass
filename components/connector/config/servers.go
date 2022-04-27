package config

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/99designs/gqlgen/graphql/handler"
	handler2 "github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/healthz"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"
)

func PrepareExternalGraphQLServer(cfg Config, certResolver api.CertificateResolver, middlewares ...mux.MiddlewareFunc) (*http.Server, error) {
	gqlInternalCfg := externalschema.Config{
		Resolvers: &api.ExternalResolver{CertificateResolver: certResolver},
	}

	externalExecutableSchema := externalschema.NewExecutableSchema(gqlInternalCfg)
	gqlServer := handler.NewDefaultServer(externalExecutableSchema)
	gqlServer.Use(log.NewGqlLoggingInterceptor())

	externalRouter := mux.NewRouter()
	externalRouter.HandleFunc("/", handler2.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	externalRouter.HandleFunc(cfg.APIEndpoint, gqlServer.ServeHTTP)
	externalRouter.HandleFunc("/healthz", healthz.NewHTTPHandler())

	externalRouter.Use(middlewares...)

	handlerWithTimeout, err := timeouthandler.WithTimeout(externalRouter, cfg.ServerTimeout)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.ExternalAddress,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}, nil
}
