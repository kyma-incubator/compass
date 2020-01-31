package api

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"net/http"
)

type managementInfoHandler struct {
	gqlClient                      connector.Client
	logger                         *log.Logger
	connectivityAdapterMTLSBaseURL string
	directorClientProvider         director.ClientProvider
}

func NewManagementInfoHandler(client connector.Client, logger *log.Logger, connectivityAdapterMTLSBaseURL string, directorClientProvider director.ClientProvider) managementInfoHandler {
	return managementInfoHandler{
		gqlClient:                      client,
		logger:                         logger,
		connectivityAdapterMTLSBaseURL: connectivityAdapterMTLSBaseURL,
		directorClientProvider:         directorClientProvider,
	}
}

func (mh *managementInfoHandler) GetManagementInfo(w http.ResponseWriter, r *http.Request) {

	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		mh.logger.Errorf("Failed to read authorization context: %s.", err)
		reqerror.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	systemAuthID := authorizationHeaders.GetSystemAuthID()
	contextLogger := contextLogger(mh.logger, systemAuthID)

	directorClient := mh.directorClientProvider.Client(r)

	application, err := directorClient.GetApplication(systemAuthID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get application")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

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

	certInfo := connector.ToCertInfo(configuration)

	//TODO: handle case when configuration.Token is nil
	managementInfoResponse := mh.makeManagementInfoResponse(
		application.Name,
		configuration.Token.Token,
		mh.connectivityAdapterMTLSBaseURL,
		application.EventingConfiguration.DefaultURL,
		certInfo)

	respondWithBody(w, http.StatusOK, managementInfoResponse, contextLogger)
}

func (mh *managementInfoHandler) makeManagementInfoResponse(
	application,
	newToken,
	connectivityAdapterMTLSBaseURL,
	eventServiceBaseURL string,
	certInfo model.CertInfo) model.MgmtInfoReponse {

	return model.MgmtInfoReponse{
		ClientIdentity:  model.MakeClientIdentity(application, "", ""),
		URLs:            model.MakeManagementURLs(application, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
		CertificateInfo: certInfo,
	}
}
