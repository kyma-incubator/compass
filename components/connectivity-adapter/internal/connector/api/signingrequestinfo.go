package api

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
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
	gqlClient  graphql.Client
}

func NewSigningRequestInfoHandler(client graphql.Client) csrInfoHandler {
	return csrInfoHandler{
		gqlClient:  client,
		getInfoURL: "",
	}
}

func (ih *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")

	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		log.Println("Client Id not provided.")
		reqerror.WriteErrorMessage(w, "Client Id not provided.", apperrors.CodeInternal)

		return
	}

	baseURLs, err := middlewares.GetBaseURLsFromContext(r.Context(), middlewares.BaseURLsKey)
	if err != nil {
		reqerror.WriteErrorMessage(w, "Base URLS not provided.", apperrors.CodeInternal)
	}

	configuration, err := ih.gqlClient.Configuration(authorizationHeaders)

	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	certInfo := graphql.ToCertInfo(configuration)

	csrInfoResponse := model.NewCSRInfoResponse(certInfo, authorizationHeaders.GetClientID(), configuration.Token.Token, baseURLs.ConnectivityAdapterBaseURL, baseURLs.EventServiceBaseURL)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}
