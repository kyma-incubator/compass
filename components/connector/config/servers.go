package config

import (
	"net/http"

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

func PrepareExternalGraphQLServer(cfg Config, certResolver api.CertificateResolver, authContextMiddleware mux.MiddlewareFunc) *http.Server {
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

	return &http.Server{
		Addr:    cfg.ExternalAddress,
		Handler: externalRouter,
	}
}

func PrepareInternalGraphQLServer(cfg Config, tokenResolver api.TokenResolver) *http.Server {
	internalResolver := api.InternalResolver{TokenResolver: tokenResolver}

	gqlInternalCfg := internalschema.Config{
		Resolvers: &internalResolver,
	}

	internalExecutableSchema := internalschema.NewExecutableSchema(gqlInternalCfg)

	internalRouter := mux.NewRouter()
	internalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	internalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(internalExecutableSchema))

	return &http.Server{
		Addr:    cfg.InternalAddress,
		Handler: internalRouter,
	}
}

func PrepareHydratorServer(cfg Config, tokenService tokens.Service, subjectConsts certificates.CSRSubjectConsts, revokedCertsRepository revocation.RevocationListRepository) *http.Server {
	certHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, subjectConsts)

	validationHydrator := oathkeeper.NewValidationHydrator(tokenService, certHeaderParser, revokedCertsRepository)

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	v1Router := router.PathPrefix("/v1").Subrouter()
	v1Router.HandleFunc("/tokens/resolve", validationHydrator.ResolveConnectorTokenHeader)
	v1Router.HandleFunc("/certificate/data/resolve", validationHydrator.ResolveIstioCertHeader)

	return &http.Server{
		Addr:    cfg.HydratorAddress,
		Handler: router,
	}
}
