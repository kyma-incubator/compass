package connector

import (
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

type Config struct {
	CompassConnectorURL    string `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	AdapterBaseURL         string `envconfig:"default=https://adapter-gateway.kyma.local"`
	AdapterBaseURLMTLS     string `envconfig:"default=https://adapter-gateway-mtls.kyma.local"`
	EventBaseURL           string `envconfig:"default=https://gateway.kyma.local"` // TODO: remove once Event Base URL is read from Director
	InsecureConnectorCalls bool   `envconfig:"default=true"`
}

const (
	timeout = 30 * time.Second
)

func RegisterHandler(router *mux.Router, config Config) error {
	client, err := graphql.NewClient(config.CompassConnectorURL, config.InsecureConnectorCalls, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	authorizationMiddleware := middlewares.NewAuthorizationMiddleware()
	router.Use(mux.MiddlewareFunc(authorizationMiddleware.GetAuthorizationHeaders))

	eventBaseURLProvider := eventBaseURLProvider{
		eventBaseUrl: config.EventBaseURL,
	}

	{
		baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterBaseURL, eventBaseURLProvider)
		signingRequestInfo := api.NewSigningRequestInfoHandler(client)
		signingRequestInfoHandler := http.HandlerFunc(signingRequestInfo.GetSigningRequestInfo)

		router.Handle("/signingRequests/info", baseURLsMiddleware.GetBaseUrls(signingRequestInfoHandler)).Methods(http.MethodGet)
	}

	{
		baseURLsMiddleware := middlewares.NewBaseURLsMiddleware(config.AdapterBaseURLMTLS, eventBaseURLProvider)
		managementInfo := api.NewManagementInfoHandler(client)
		managementInfoHandler := http.HandlerFunc(managementInfo.GetManagementInfo)

		router.Handle("/management/info", baseURLsMiddleware.GetBaseUrls(managementInfoHandler))
	}

	certificatesHandler := api.NewCertificatesHandler(client).SignCSR
	router.HandleFunc("/certificates", certificatesHandler)

	return nil
}

type eventBaseURLProvider struct {
	eventBaseUrl string
}

func (e eventBaseURLProvider) EventServiceBaseURL() (string, error) {
	// TODO: call Director for getting events base url
	return e.eventBaseUrl, nil
}
