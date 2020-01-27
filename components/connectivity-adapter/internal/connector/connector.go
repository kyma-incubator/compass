package connector

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ConnectorEndpoint         string `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	ConnectorInternalEndpoint string `envconfig:"default=http://compass-connector-internal.compass-system.svc.cluster.local:3001/graphql"`
	AdapterBaseURL            string `envconfig:"default=https://adapter-gateway.kyma.local"`
	AdapterMtlsBaseURL        string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
}

const (
	timeout = 30 * time.Second
)

func RegisterExternalHandler(router *mux.Router, config Config) error {
	logger := logrus.New().WithField("component", "connector").Logger
	logger.SetReportCaller(true)

	client, err := graphql.NewClient(config.ConnectorEndpoint, config.ConnectorInternalEndpoint, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(mux.MiddlewareFunc(authorizationMiddleware.GetAuthorizationHeaders))

	signingRequestInfoHandler := newSigningRequestInfoHandler(config, client, logger)
	managementInfoHandler := newManagementInfoHandler(config, client, logger)
	certificatesHandler := newCertificateHandler(client, logger)
	revocationsHandler := newRevocationsHandler(client, logger)

	router.Handle("/signingRequests/info", signingRequestInfoHandler).Methods(http.MethodGet)
	router.Handle("/management/info", managementInfoHandler).Methods(http.MethodGet)
	router.Handle("/certificates", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/renewals", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/revocations", revocationsHandler).Methods(http.MethodPost)

	return nil
}

func newSigningRequestInfoHandler(config Config, client graphql.Client, logger *logrus.Logger) http.Handler {
	eventBaseURLProvider := eventBaseURLProvider{
		eventBaseUrl: config.AdapterMtlsBaseURL,
	}

	baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterBaseURL, config.AdapterMtlsBaseURL, eventBaseURLProvider)
	signingRequestInfo := api.NewSigningRequestInfoHandler(client, logger)
	signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetSigningRequestInfo)

	return baseURLsMiddleware.GetBaseUrls(signingRequestInfoHandler)
}

func newManagementInfoHandler(config Config, client graphql.Client, logger *logrus.Logger) http.Handler {
	eventBaseURLProvider := eventBaseURLProvider{
		eventBaseUrl: config.AdapterMtlsBaseURL,
	}
	baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterMtlsBaseURL, config.AdapterMtlsBaseURL, eventBaseURLProvider)
	managementInfo := api.NewManagementInfoHandler(client, logger)
	managementInfoHandler := http.HandlerFunc(managementInfo.GetManagementInfo)

	return baseURLsMiddleware.GetBaseUrls(managementInfoHandler)
}

func newCertificateHandler(client graphql.Client, logger *logrus.Logger) http.Handler {
	handler := api.NewCertificatesHandler(client, logger)

	return http.HandlerFunc(handler.SignCSR)
}

func newRevocationsHandler(client graphql.Client, logger *logrus.Logger) http.Handler {
	handler := api.NewRevocationsHandler(client, logger)

	return http.HandlerFunc(handler.RevokeCertificate)
}

func RegisterInternalHandler(router *mux.Router, config Config) error {
	logger := logrus.New().WithField("component", "connector internal").Logger
	logger.SetReportCaller(true)

	client, err := graphql.NewClient(config.ConnectorEndpoint, config.ConnectorInternalEndpoint, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	tokenHandler := api.NewTokenHandler(client, config.AdapterBaseURL, logger)
	router.HandleFunc("/tokens", tokenHandler.GetToken).Methods(http.MethodPost)

	return nil
}
