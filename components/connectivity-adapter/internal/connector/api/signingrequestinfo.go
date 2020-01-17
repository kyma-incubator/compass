package api

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/compass"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"log"
	"net/http"
)

const (
	TokenFormat                       = "?token=%s"
	CertsEndpoint                     = "/v1/applications/certificates"
	ManagementInfoEndpoint            = "/v1/applications/management/info"
	ApplicationRegistryEndpointFormat = "/%s/v1/metadata"
	EventsEndpointFormat              = "/%s/v1/events"
)

const (
	HeaderContentType = "Content-Type"
)

const (
	ContentTypeApplicationJson = "application/json;charset=UTF-8"
)

type csrInfoHandler struct {
	getInfoURL string
	baseURL    string
	client     compass.Client
}

func NewSigningRequestInfo(client compass.Client, baseURL string) csrInfoHandler {
	return csrInfoHandler{
		client:     client,
		getInfoURL: "",
		baseURL:    baseURL,
	}
}

func (ih *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")

	for queryParam := range r.URL.Query() {
		log.Println("Header: " + queryParam + ":" + r.URL.Query().Get(queryParam))
	}

	clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
	if clientIdFromToken == "" {
		log.Println("Client Id not provided.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	configuration, err := ih.client.Configuration(map[string]string{
		oathkeeper.ClientIdFromTokenHeader: clientIdFromToken,
	})

	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	csrInfoResponse := ih.makeCSRInfoResponse(configuration, clientIdFromToken)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}

func (ih *csrInfoHandler) makeCSRInfoResponse(configuration schema.Configuration, clientIdFromToken string) connector.CSRInfoResponse {
	return connector.CSRInfoResponse{
		CsrURL:          ih.makeCSRURLs(configuration.Token.Token, clientIdFromToken),
		API:             ih.makeApiURLs(clientIdFromToken),
		CertificateInfo: connector.ToCertInfo(configuration),
	}
}

func (ih *csrInfoHandler) makeCSRURLs(newToken string, clientIdFromToken string) string {
	csrURL := ih.baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func (ih *csrInfoHandler) makeApiURLs(clientIdFromToken string) connector.Api {
	return connector.Api{
		CertificatesURL: ih.baseURL + CertsEndpoint,
		InfoURL:         ih.baseURL + ManagementInfoEndpoint,
		RuntimeURLs: &connector.RuntimeURLs{
			MetadataURL: ih.baseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, clientIdFromToken),
			EventsURL:   ih.baseURL + fmt.Sprintf(EventsEndpointFormat, clientIdFromToken),
		},
	}
}

func Respond(w http.ResponseWriter, statusCode int) {
	w.Header().Set(HeaderContentType, ContentTypeApplicationJson)
	w.WriteHeader(statusCode)
}

func respondWithBody(w http.ResponseWriter, statusCode int, responseBody interface{}) {
	Respond(w, statusCode)
	json.NewEncoder(w).Encode(responseBody)
}
