package config

import (
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	handler2 "github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/healthz"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/subject"
	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
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

func PrepareHydratorServer(cfg Config, CSRSubjectConsts certificates.CSRSubjectConsts, externalSubjectConsts certificates.ExternalIssuerSubjectConsts, revokedCertsRepository revocation.RevokedCertificatesRepository, middlewares ...mux.MiddlewareFunc) (*http.Server, error) {
	subjectProcessor, err := subject.NewProcessor(cfg.SubjectConsumerMappingConfig, externalSubjectConsts.OrganizationalUnitPattern)
	if err != nil {
		return nil, err
	}

	externalCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ExternalIssuer,
		oathkeeper.ExternalCertIssuerSubjectMatcher(externalSubjectConsts), subjectProcessor.AuthIDFromSubjectFunc(), subjectProcessor.AuthSessionExtraFromSubjectFunc())
	connectorCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ConnectorIssuer,
		oathkeeper.ConnectorCertificateSubjectMatcher(CSRSubjectConsts), cert.GetCommonName, subjectProcessor.EmptyAuthSessionExtraFunc())

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
