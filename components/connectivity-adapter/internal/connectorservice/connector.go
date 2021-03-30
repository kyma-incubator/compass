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
)

type Config struct {
	ConnectorEndpoint string        `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	ClientTimeout     time.Duration `envconfig:"default=115s"`

	AdapterBaseURL     string `envconfig:"default=https://adapter-gateway.kyma.local"`
	AdapterMtlsBaseURL string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
}

func RegisterHandler(router *mux.Router, config Config, directorURL string, directorTimeout time.Duration) error {
	directorClientProvider := director.NewClientProvider(directorURL, directorTimeout)
	connectorClientProvider := connector.NewClientProvider(config.ConnectorEndpoint, config.ClientTimeout)

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(authorizationMiddleware.GetAuthorizationHeaders)

	signingRequestInfoHandler := newSigningRequestInfoHandler(config, connectorClientProvider, directorClientProvider)
	managementInfoHandler := newManagementInfoHandler(config, connectorClientProvider, directorClientProvider)
	certificatesHandler := newCertificateHandler(connectorClientProvider)
	revocationsHandler := newRevocationsHandler(connectorClientProvider)

	router.Handle("/signingRequests/info", signingRequestInfoHandler).Methods(http.MethodGet)
	router.Handle("/management/info", managementInfoHandler).Methods(http.MethodGet)
	router.Handle("/certificates", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/renewals", certificatesHandler).Methods(http.MethodPost)
	router.Handle("/certificates/revocations", revocationsHandler).Methods(http.MethodPost)

	return nil
}

func newSigningRequestInfoHandler(config Config, connectorClientProvider connector.ClientProvider, directorClientProvider director.ClientProvider) http.Handler {
	signingRequestInfo := api.NewInfoHandler(connectorClientProvider, directorClientProvider, model.NewCSRInfoResponseProvider(config.AdapterBaseURL, config.AdapterMtlsBaseURL))
	signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetInfo)

	return signingRequestInfoHandler
}

func newManagementInfoHandler(config Config, connectorClientProvider connector.ClientProvider, directorClientProvider director.ClientProvider) http.Handler {
	managementInfo := api.NewInfoHandler(connectorClientProvider, directorClientProvider, model.NewManagementInfoResponseProvider(config.AdapterMtlsBaseURL))
	managementInfoHandler := http.HandlerFunc(managementInfo.GetInfo)

	return managementInfoHandler
}

func newCertificateHandler(connectorClientProvider connector.ClientProvider) http.Handler {
	handler := api.NewCertificatesHandler(connectorClientProvider)

	return http.HandlerFunc(handler.SignCSR)
}

func newRevocationsHandler(connectorClientProvider connector.ClientProvider) http.Handler {
	handler := api.NewRevocationsHandler(connectorClientProvider)

	return http.HandlerFunc(handler.RevokeCertificate)
}
