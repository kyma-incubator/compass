package connectorservice

import (
	"net/http"
	"time"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/sirupsen/logrus"
)

type Config struct {
	ConnectorEndpoint string        `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	ClientTimeout     time.Duration `envconfig:"default=115s"`

	AdapterBaseURL     string `envconfig:"default=https://adapter-gateway.kyma.local"`
	AdapterMtlsBaseURL string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
}

func RegisterHandler(router *mux.Router, config Config, directorURL string, directorTimeout time.Duration) error {
	logger := logrus.New().WithField("component", "connector").Logger
	logger.SetReportCaller(true)

	directorClientProvider := director.NewClientProvider(directorURL, directorTimeout)
	connectorClientProvider := connector.NewClientProvider(config.ConnectorEndpoint, config.ClientTimeout)

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(authorizationMiddleware.GetAuthorizationHeaders)

	signingRequestInfoHandler := newSigningRequestInfoHandler(config, connectorClientProvider, directorClientProvider, logger)
	managementInfoHandler := newManagementInfoHandler(config, connectorClientProvider, directorClientProvider, logger)
	certificatesHandler := newCertificateHandler(connectorClientProvider, logger)
	revocationsHandler := newRevocationsHandler(connectorClientProvider, logger)

	router.Handle("/signingRequests/info", signingRequestInfoHandler).Methods(http.MethodGet)
	router.Handle("/management/info", managementInfoHandler).Methods(http.MethodGet)
	router.Handle("/certificates", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/renewals", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/revocations", revocationsHandler).Methods(http.MethodPost)

	return nil
}

func newSigningRequestInfoHandler(config Config, connectorClientProvider connector.ClientProvider, directorClientProvider director.ClientProvider, logger *logrus.Logger) http.Handler {
	signingRequestInfo := api.NewInfoHandler(connectorClientProvider, directorClientProvider, logger, model.NewCSRInfoResponseProvider(config.AdapterBaseURL, config.AdapterMtlsBaseURL))
	signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetInfo)

	return signingRequestInfoHandler
}

func newManagementInfoHandler(config Config, connectorClientProvider connector.ClientProvider, directorClientProvider director.ClientProvider, logger *logrus.Logger) http.Handler {
	managementInfo := api.NewInfoHandler(connectorClientProvider, directorClientProvider, logger, model.NewManagementInfoResponseProvider(config.AdapterMtlsBaseURL))
	managementInfoHandler := http.HandlerFunc(managementInfo.GetInfo)

	return managementInfoHandler
}

func newCertificateHandler(connectorClientProvider connector.ClientProvider, logger *logrus.Logger) http.Handler {
	handler := api.NewCertificatesHandler(connectorClientProvider, logger)

	return http.HandlerFunc(handler.SignCSR)
}

func newRevocationsHandler(connectorClientProvider connector.ClientProvider, logger *logrus.Logger) http.Handler {
	handler := api.NewRevocationsHandler(connectorClientProvider, logger)

	return http.HandlerFunc(handler.RevokeCertificate)
}
