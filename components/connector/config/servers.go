package config

import (
	"net/http"

	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/kyma-incubator/compass/components/connector/internal/healthz"
	log "github.com/sirupsen/logrus"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/internalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
)

func PrepareExternalGraphQLServer(cfg Config, certResolver api.CertificateResolver, authContextMiddleware mux.MiddlewareFunc) (*http.Server, error) {
	externalResolver := api.ExternalResolver{CertificateResolver: certResolver}

	gqlInternalCfg := externalschema.Config{
		Resolvers: &externalResolver,
	}

	externalExecutableSchema := externalschema.NewExecutableSchema(gqlInternalCfg)

	externalRouter := mux.NewRouter()
	externalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	externalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(externalExecutableSchema))
	externalRouter.HandleFunc("/healthz", healthz.NewHTTPHandler(log.StandardLogger()))

	externalRouter.Use(authContextMiddleware)

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

func PrepareInternalGraphQLServer(cfg Config, tokenResolver api.TokenResolver) (*http.Server, error) {
	internalResolver := api.InternalResolver{TokenResolver: tokenResolver}

	gqlInternalCfg := internalschema.Config{
		Resolvers: &internalResolver,
	}

	internalExecutableSchema := internalschema.NewExecutableSchema(gqlInternalCfg)

	internalRouter := mux.NewRouter()
	internalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	internalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(internalExecutableSchema))

	handlerWithTimeout, err := timeouthandler.WithTimeout(internalRouter, cfg.ServerTimeout)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.InternalAddress,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}, nil
}

func PrepareHydratorServer(cfg Config, tokenService tokens.Service, subjectConsts certificates.CSRSubjectConsts, revokedCertsRepository revocation.RevocationListRepository) (*http.Server, error) {
	certHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, subjectConsts)

	validationHydrator := oathkeeper.NewValidationHydrator(tokenService, certHeaderParser, revokedCertsRepository)

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/tokens/resolve", validationHydrator.ResolveConnectorTokenHeader)
	v1Router.HandleFunc("/certificate/data/resolve", validationHydrator.ResolveIstioCertHeader)

	handlerWithTimeout, err := timeouthandler.WithTimeout(router, cfg.ServerTimeout)
	if err != nil {
		return nil, err
	}

	return &http.Server{
		Addr:              cfg.HydratorAddress,
		Handler:           handlerWithTimeout,
		ReadHeaderTimeout: cfg.ServerTimeout,
	}, nil
}
