package api

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"net/http"
)

type managementInfoHandler struct {
	gqlClient graphql.Client
	logger    *log.Logger
}

func NewManagementInfoHandler(client graphql.Client, logger *log.Logger) managementInfoHandler {
	return managementInfoHandler{
		gqlClient: client,
		logger:    logger,
	}
}

func (mh *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {

	// TODO: make sure only calls with certificate are accepted
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		mh.logger.Errorf("Failed to read authorization context: %s.", err)
		reqerror.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	contextLogger := contextLogger(mh.logger, authorizationHeaders.GetClientID())

	baseURLs, err := middlewares.GetBaseURLsFromContext(r.Context(), middlewares.BaseURLsKey)
	if err != nil {
		contextLogger.Errorf("Failed to read Base URL context: %s.", err)
		reqerror.WriteErrorMessage(w, "Failed to read Base URL context.", apperrors.CodeInternal)

		return
	}

	contextLogger.Info("Getting Management Info")

	configuration, err := mh.gqlClient.Configuration(authorizationHeaders)
	if err != nil {
		err = errors.Wrap(err, "Failed to get configuration")
		contextLogger.Error(err.Error())
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
