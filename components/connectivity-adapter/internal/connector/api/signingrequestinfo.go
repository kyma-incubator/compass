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

func (c *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")

	// TODO: make sure only calls with token are accepted

	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		log.Println("Client Id not provided.")
		reqerror.WriteErrorMessage(w, "Client Id not provided.", apperrors.CodeForbidden)

		return
	}

	baseURLs, err := middlewares.GetBaseURLsFromContext(r.Context(), middlewares.BaseURLsKey)
	if err != nil {
		reqerror.WriteErrorMessage(w, "Base URLS not provided.", apperrors.CodeInternal)

		return
	}

	configuration, err := c.gqlClient.Configuration(authorizationHeaders)
	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	certInfo := graphql.ToCertInfo(configuration)
	application := authorizationHeaders.GetClientID()

	//TODO: handle case when configuration.Token is nil
	csrInfoResponse := c.makeCSRInfoResponse(application, configuration.Token.Token, baseURLs.ConnectivityAdapterBaseURL, baseURLs.EventServiceBaseURL, certInfo)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}

func (c *csrInfoHandler) makeCSRInfoResponse(application, newToken, connectivityAdapterBaseURL, eventServiceBaseURL string, certInfo model.CertInfo) model.CSRInfoResponse {
	return model.CSRInfoResponse{
		CsrURL:          model.MakeCSRURL(newToken, connectivityAdapterBaseURL),
		API:             model.MakeApiURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
		CertificateInfo: certInfo,
	}
}
