package connectorservice

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"
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

func RegisterHandler(router *mux.Router, config Config, directorURL string) error {
	logger := logrus.New().WithField("component", "connector").Logger
	logger.SetReportCaller(true)

	client, err := connector.NewClient(config.ConnectorEndpoint, config.ConnectorInternalEndpoint, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(mux.MiddlewareFunc(authorizationMiddleware.GetAuthorizationHeaders))

	signingRequestInfoHandler := newSigningRequestInfoHandler(config, client, logger)
	managementInfoHandler := newManagementInfoHandler(config, client, logger, directorURL)
	certificatesHandler := newCertificateHandler(client, logger)
	revocationsHandler := newRevocationsHandler(client, logger)

	router.Handle("/signingRequests/info", signingRequestInfoHandler).Methods(http.MethodGet)
	router.Handle("/management/info", managementInfoHandler).Methods(http.MethodGet)
	router.Handle("/certificates", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/renewals", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/revocations", revocationsHandler).Methods(http.MethodPost)

	return nil
}

func newSigningRequestInfoHandler(config Config, client connector.Client, logger *logrus.Logger) http.Handler {
	signingRequestInfo := api.NewSigningRequestInfoHandler(client, logger, config.AdapterBaseURL, config.AdapterMtlsBaseURL)
	signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetSigningRequestInfo)

	return signingRequestInfoHandler
}

func newManagementInfoHandler(config Config, client connector.Client, logger *logrus.Logger, directorURL string) http.Handler {
	managementInfo := api.NewManagementInfoHandler(client, logger, config.AdapterMtlsBaseURL, director.NewClientProvider(directorURL))
	managementInfoHandler := http.HandlerFunc(managementInfo.GetManagementInfo)

	return managementInfoHandler
}

func newCertificateHandler(client connector.Client, logger *logrus.Logger) http.Handler {
	handler := api.NewCertificatesHandler(client, logger)

	return http.HandlerFunc(handler.SignCSR)
}

func newRevocationsHandler(client connector.Client, logger *logrus.Logger) http.Handler {
	handler := api.NewRevocationsHandler(client, logger)

	return http.HandlerFunc(handler.RevokeCertificate)
}
