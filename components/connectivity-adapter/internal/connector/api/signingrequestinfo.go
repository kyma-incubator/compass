package api

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
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
	gqlClient  graphql.Client
}

func NewSigningRequestInfoHandler(client graphql.Client, baseURL string) csrInfoHandler {
	return csrInfoHandler{
		gqlClient:  client,
		getInfoURL: "",
		baseURL:    baseURL,
	}
}

func (ih *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")

	clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
	if clientIdFromToken == "" {
		log.Println("Client Id not provided.")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	configuration, err := ih.gqlClient.Configuration(clientIdFromToken)

	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	csrInfoResponse := ih.makeCSRInfoResponse(configuration, clientIdFromToken)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}

func (ih *csrInfoHandler) makeCSRInfoResponse(configuration schema.Configuration, clientIdFromToken string) model.CSRInfoResponse {
	return model.CSRInfoResponse{
		CsrURL:          ih.makeCSRURLs(configuration.Token.Token, clientIdFromToken),
		API:             ih.makeApiURLs(clientIdFromToken),
		CertificateInfo: graphql.ToCertInfo(configuration),
	}
}

func (ih *csrInfoHandler) makeCSRURLs(newToken string, clientIdFromToken string) string {
	csrURL := ih.baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func (ih *csrInfoHandler) makeApiURLs(clientIdFromToken string) model.Api {
	return model.Api{
		CertificatesURL: ih.baseURL + CertsEndpoint,
		InfoURL:         ih.baseURL + ManagementInfoEndpoint,
		RuntimeURLs: &model.RuntimeURLs{
			MetadataURL: ih.baseURL + fmt.Sprintf(ApplicationRegistryEndpointFormat, clientIdFromToken),
			EventsURL:   ih.baseURL + fmt.Sprintf(EventsEndpointFormat, clientIdFromToken),
		},
	}
}
