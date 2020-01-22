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

type managementInfoHandler struct {
	gqlClient graphql.Client
}

func NewManagementInfoHandler(client graphql.Client) managementInfoHandler {
	return managementInfoHandler{
		gqlClient: client,
	}
}

func (mh *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {
	log.Println("Starting GetSigningRequestInfo")

	// TODO: make sure only calls with certificate are accepted

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

	configuration, err := mh.gqlClient.Configuration(authorizationHeaders)
	if err != nil {
		log.Println("Error getting configuration " + err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	certInfo := graphql.ToCertInfo(configuration)
	application := authorizationHeaders.GetClientID()

	//TODO: handle case when configuration.Token is nil
	csrInfoResponse := mh.makeManagementInfoResponse(application, configuration.Token.Token, baseURLs.ConnectivityAdapterBaseURL, baseURLs.EventServiceBaseURL, certInfo)
	respondWithBody(w, http.StatusOK, csrInfoResponse)
}

func (m *managementInfoHandler) makeManagementInfoResponse(application, newToken, connectivityAdapterBaseURL, eventServiceBaseURL string, certInfo model.CertInfo) model.MgmtInfoReponse {
	return model.MgmtInfoReponse{
		ClientIdentity:  model.MakeClientIdentity(application, "", ""), // TODO: how to get tenant? Is it vital?
		URLs:            model.MakeManagementURLs(application, connectivityAdapterBaseURL, eventServiceBaseURL),
		CertificateInfo: certInfo,
	}
}
