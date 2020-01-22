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
	CompassConnectorURL string `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	AdapterBaseURL      string `envconfig:"default=https://adapter-gateway.kyma.local"`
}

const (
	timeout = 30 * time.Second
)

func RegisterHandler(router *mux.Router, config Config) error {
	client, err := graphql.NewClient(config.CompassConnectorURL, true, timeout)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	authorizationMiddleware := middlewares.NewClientFromTokenMiddleware().GetAuthorizationHeaders
	router.Use(mux.MiddlewareFunc(authorizationMiddleware))

	signingRequestInfoHandler := api.NewSigningRequestInfoHandler(client)
	router.HandleFunc("/signingRequests/info", signingRequestInfoHandler.GetSigningRequestInfo).Methods(http.MethodGet)

	certificatesHandler := api.NewCertificatesHandler(client)
	router.HandleFunc("/certificates", certificatesHandler.SignCSR)

	managementInfoHandle := api.NewManagementInfoHandler(client)
	router.HandleFunc("/management/info", managementInfoHandle.GetManagementInfo)

	return nil
}
