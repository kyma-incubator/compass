package api

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"log"
	"net/http"
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

	certInfo := graphql.ToCertInfo(configuration)

	csrInfoResponse := model.NewCSRInfoResponse(certInfo, clientIdFromToken, configuration.Token.Token, ih.baseURL)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}
