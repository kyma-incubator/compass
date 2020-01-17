package api

import (
	"encoding/json"
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

	signingRequestInfoHandler := NewSigningRequestInfoHandler(client, config.AdapterBaseURL)
	router.HandleFunc("/signingRequests/info", http.HandlerFunc(signingRequestInfoHandler.GetSigningRequestInfo)).Methods(http.MethodGet)

	certificatesHandler := NewCertificatesHandler(client)
	router.HandleFunc("/certificates", certificatesHandler.SignCSR)

	return nil
}

func respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}
