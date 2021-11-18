package config

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	timeouthandler "github.com/kyma-incubator/compass/components/director/pkg/handler"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/healthz"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
)

func PrepareExternalGraphQLServer(cfg Config, certResolver api.CertificateResolver, middlewares ...mux.MiddlewareFunc) (*http.Server, error) {
	gqlInternalCfg := externalschema.Config{
		Resolvers: &api.ExternalResolver{CertificateResolver: certResolver},
	}

	externalExecutableSchema := externalschema.NewExecutableSchema(gqlInternalCfg)

	externalRouter := mux.NewRouter()
	externalRouter.HandleFunc("/", handler.Playground("Dataloader", cfg.PlaygroundAPIEndpoint))
	externalRouter.HandleFunc(cfg.APIEndpoint, handler.GraphQL(externalExecutableSchema))
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

func PrepareHydratorServer(cfg Config, CSRSubjectConsts certificates.CSRSubjectConsts, externalSubjectConsts certificates.ExternalIssuerSubjectConsts, revokedCertsRepository revocation.RevokedCertificatesRepository, middlewares ...mux.MiddlewareFunc) (*http.Server, error) {
	connectorCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ConnectorIssuer, oathkeeper.ConnectorCertificateSubjectMatcher(CSRSubjectConsts), cert.GetCommonName)
	externalCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ExternalIssuer, oathkeeper.ExternalCertIssuerSubjectMatcher(externalSubjectConsts), cert.GetRemainingOrganizationalUnit(externalSubjectConsts.OrganizationalUnitPattern))

	validationHydrator := oathkeeper.NewValidationHydrator(revokedCertsRepository, connectorCertHeaderParser, externalCertHeaderParser)

	router := mux.NewRouter()
	router.Path("/health").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	router.Use(middlewares...)

	v1Router := router.PathPrefix("/v1").Subrouter()
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
