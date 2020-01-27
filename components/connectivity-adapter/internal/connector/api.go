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
	CompassConnectorURL         string `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	CompassConnectorInternalURL string `envconfig:"default=http://compass-connector-internal.compass-system.svc.cluster.local:3001/graphql"`
	AdapterBaseURL              string `envconfig:"default=https://adapter-gateway.kyma.local"`
	AdapterMtlsBaseURL          string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
}

const (
	timeout = 30 * time.Second
)

func RegisterExternalHandler(router *mux.Router, config Config) error {
	logger := logrus.New().WithField("component", "connector").Logger
	logger.SetReportCaller(true)

	client, err := graphql.NewClient(config.CompassConnectorURL, config.CompassConnectorInternalURL, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(mux.MiddlewareFunc(authorizationMiddleware.GetAuthorizationHeaders))

	eventBaseURLProvider := eventBaseURLProvider{
		eventBaseUrl: config.AdapterMtlsBaseURL,
	}

	{
		baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterBaseURL, config.AdapterMtlsBaseURL, eventBaseURLProvider)
		signingRequestInfo := api.NewSigningRequestInfoHandler(client, logger)
		signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetSigningRequestInfo)

		router.Handle("/signingRequests/info", baseURLsMiddleware.GetBaseUrls(signingRequestInfoHandler)).Methods(http.MethodGet)
	}

	{
		baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterMtlsBaseURL, config.AdapterMtlsBaseURL, eventBaseURLProvider)
		managementInfo := api.NewManagementInfoHandler(client, logger)
		managementInfoHandler := http.HandlerFunc(managementInfo.GetManagementInfo)

		router.Handle("/management/info", baseURLsMiddleware.GetBaseUrls(managementInfoHandler)).Methods(http.MethodGet)
	}

	certificates := api.NewCertificatesHandler(client, logger)
	router.HandleFunc("/certificates", certificates.SignCSR)
	router.HandleFunc("/certificates/renewals", certificates.SignCSR)

	revocationsHandler := api.NewRevocationsHandler(client, logger)
	router.HandleFunc("/certificates/revocations", revocationsHandler.RevokeCertificate)

	return nil
}

func RegisterInternalHandler(router *mux.Router, config Config) error {
	logger := logrus.New().WithField("component", "connector internal").Logger
	logger.SetReportCaller(true)

	client, err := graphql.NewClient(config.CompassConnectorURL, config.CompassConnectorInternalURL, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	tokenHandler := api.NewTokenHandler(client, config.AdapterBaseURL, logger)
	router.HandleFunc("/tokens", tokenHandler.GetToken).Methods(http.MethodPost)

	return nil
}

// Mock implementation of EventServiceBaseURLProvider
type eventBaseURLProvider struct {
	eventBaseUrl string
}

func (e eventBaseURLProvider) EventServiceBaseURL() (string, error) {
	// TODO: call Director for getting events base url
	return e.eventBaseUrl, nil
}
