package api

import (
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/compass"
	schema "github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"log"
	"net/http"
)

const (
	TokenFormat   = "?token=%s"
	CertsEndpoint = "/certificates"
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
	//csrSubject               certificates.CSRSubject
}

func NewSigningRequestInfo(client compass.Client) csrInfoHandler {
	return csrInfoHandler{
		client:     client,
		getInfoURL: "",
		baseURL:    "",
	}
}

func (ih *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")
	log.Println("Connector Token: " + r.Header.Get("Connector-Token"))
	log.Println("Connector Token: " + r.Header.Get("ClientIdFromTokenHeader"))

	configuration, err := ih.client.Configuration(map[string]string{
		"Client-Id-From-Token": r.Header.Get("Connector-Token"),
	})

	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	csrInfoResponse := ih.makeCSRInfoResponse(configuration)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}

func (ih *csrInfoHandler) makeCSRInfoResponse(configuration schema.Configuration) connector.CSRInfoResponse {
	return connector.CSRInfoResponse{
		CsrURL:          ih.makeCSRURLs("TOKEN"),
		API:             ih.makeApiURLs(),
		CertificateInfo: connector.ToCertInfo(configuration),
	}
}

func (ih *csrInfoHandler) makeCSRURLs(newToken string) string {
	csrURL := ih.baseURL + CertsEndpoint
	tokenParam := fmt.Sprintf(TokenFormat, newToken)

	return csrURL + tokenParam
}

func (ih *csrInfoHandler) makeApiURLs() connector.Api {
	return connector.Api{
		CertificatesURL: ih.baseURL + CertsEndpoint,
		InfoURL:         ih.getInfoURL,
		//RuntimeURLs:     clientContextService.GetRuntimeUrls(),
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
