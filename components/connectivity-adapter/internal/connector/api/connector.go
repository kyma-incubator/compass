package api

import (
	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/compass"
	"github.com/pkg/errors"
	"net/http"
)

type Config struct {
	CompassConnectorURL string `envconfig:"default=http://compass-connector.compass-system.svc.cluster.local:3000/graphql"`
	AdapterBaseURL      string `envconfig:"default=https://adapter-gateway.kyma.local"`
}

func RegisterHandler(router *mux.Router, config Config) error {
	client, err := compass.NewClient(config.CompassConnectorURL, true, true)
	if err != nil {
		return errors.Wrap(err, "Failed to initialize compass client")
	}

	signingRequestInfoHandler := NewSigningRequestInfo(client, config.AdapterBaseURL)
	router.HandleFunc("/signingRequests/info", http.HandlerFunc(signingRequestInfoHandler.GetSigningRequestInfo)).Methods(http.MethodGet)

	return nil
}
