package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	HeaderContentType = "Content-Type"
)

const (
	ContentTypeApplicationJson = "application/json;charset=UTF-8"
)

type csrInfoHandler struct {
	connectorClient                connector.Client
	directorClientProvider         director.ClientProvider
	logger                         *log.Logger
	connectivityAdapterBaseURL     string
	connectivityAdapterMTLSBaseURL string
}

func NewSigningRequestInfoHandler(
	connectorClient connector.Client,
	directorClientProvider director.ClientProvider,
	logger *log.Logger,
	connectivityAdapterBaseURL string,
	connectivityAdapterMTLSBaseURL string) csrInfoHandler {

	return csrInfoHandler{
		connectorClient:                connectorClient,
		directorClientProvider:         directorClientProvider,
		logger:                         logger,
		connectivityAdapterBaseURL:     connectivityAdapterBaseURL,
		connectivityAdapterMTLSBaseURL: connectivityAdapterMTLSBaseURL,
	}
}

func (ci *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		ci.logger.Errorf("Failed to read authorization context: %s.", err)
		reqerror.WriteErrorMessage(w, "Client Id not provided.", apperrors.CodeForbidden)

		return
	}
	systemAuthID := authorizationHeaders.GetSystemAuthID()

	contextLogger := contextLogger(ci.logger, systemAuthID)
	contextLogger.Info("Getting Certificate Signing Request Info")

	application, err := ci.directorClientProvider.Client(r).GetApplication(systemAuthID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get Application from Director")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	configuration, err := ci.connectorClient.Configuration(authorizationHeaders)
	if err != nil {
		err = errors.Wrap(err, "Failed to get Configuration from Connector")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}
	certInfo := connector.ToCertInfo(configuration)

	//TODO: handle case when configuration.Token is nil
	csrInfoResponse := ci.makeCSRInfoResponse(
		application.Name,
		configuration.Token.Token,
		ci.connectivityAdapterBaseURL,
		ci.connectivityAdapterMTLSBaseURL,
		application.EventingConfiguration.DefaultURL,
		certInfo)

	respondWithBody(w, http.StatusOK, csrInfoResponse, contextLogger)
}

func (ci *csrInfoHandler) makeCSRInfoResponse(
	application,
	newToken,
	connectivityAdapterBaseURL,
	connectivityAdapterMTLSBaseURL,
	eventServiceBaseURL string,
	certInfo model.CertInfo) model.CSRInfoResponse {

	return model.CSRInfoResponse{
		CsrURL:          model.MakeCSRURL(newToken, connectivityAdapterBaseURL),
		API:             model.MakeApiURLs(application, connectivityAdapterBaseURL, connectivityAdapterMTLSBaseURL, eventServiceBaseURL),
		CertificateInfo: certInfo,
	}
}
