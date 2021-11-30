package config

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/pkg/errors"

	"github.com/99designs/gqlgen/handler"
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connector/internal/api"
	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/internal/healthz"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
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
	connectorCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ConnectorIssuer,
		oathkeeper.ConnectorCertificateSubjectMatcher(CSRSubjectConsts), cert.GetCommonName, nil)

	mappings, err := unmarshalMappings(cfg.SubjectConsumerMappingConfig)
	if err != nil {
		return nil, err
	}

	authSessionFunc := authSessionExtraFromSubject(mappings)
	authIDFromMappingFunc := authIDFromSubjectToConsumerMapping(mappings)
	authIDFromOUsFunc := cert.GetRemainingOrganizationalUnit(externalSubjectConsts.OrganizationalUnitPattern)
	authIDFromSubjectFunc := func(subject string) string {
		if authIDFromMapping := authIDFromMappingFunc(subject); authIDFromMapping != "" {
			return authIDFromMapping
		}
		return authIDFromOUsFunc(subject)
	}

	externalCertHeaderParser := oathkeeper.NewHeaderParser(cfg.CertificateDataHeader, oathkeeper.ExternalIssuer,
		oathkeeper.ExternalCertIssuerSubjectMatcher(externalSubjectConsts), authIDFromSubjectFunc, authSessionFunc)

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

func authSessionExtraFromSubject(mappings []subjectConsumerTypeMapping) func(subject string) map[string]interface{} {
	return func(subject string) map[string]interface{} {
		for _, m := range mappings {
			r, err := regexp.Compile(m.SubjectPattern)
			if err != nil { // already validated during bootstrap
				continue
			}
			if r.MatchString(subject) {
				return cert.GetExtra(m.TenantAccessLevel, m.ConsumerType, m.InternalConsumerID)
			}
		}
		return nil
	}
}

func authIDFromSubjectToConsumerMapping(mappings []subjectConsumerTypeMapping) func(subject string) string {
	return func(subject string) string {
		for _, m := range mappings {
			r, err := regexp.Compile(m.SubjectPattern)
			if err != nil { // already validated during bootstrap
				continue
			}
			if r.MatchString(subject) {
				return m.InternalConsumerID
			}
		}
		return ""
	}
}

// TODO add configuration validation
func unmarshalMappings(mappingsConfig string) ([]subjectConsumerTypeMapping, error) {
	var mappings []subjectConsumerTypeMapping
	if err := json.Unmarshal([]byte(mappingsConfig), &mappings); err != nil {
		return nil, errors.Wrap(err, "while unmarshalling mappings")
	}
	for _, m := range mappings {
		if _, err := regexp.Compile(m.SubjectPattern); err != nil {
			return nil, errors.Wrapf(err, "Failed to compile regex %s", m.SubjectPattern)
		}
	}
	return mappings, nil
}
