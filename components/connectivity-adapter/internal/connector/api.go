package connector

import (
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api"
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

	signingRequestInfoHandler := api.NewSigningRequestInfoHandler(client, config.AdapterBaseURL)
	router.HandleFunc("/signingRequests/info", http.HandlerFunc(signingRequestInfoHandler.GetSigningRequestInfo)).Methods(http.MethodGet)

	certificatesHandler := api.NewCertificatesHandler(client)
	router.HandleFunc("/certificates", certificatesHandler.SignCSR)

	return nil
}
