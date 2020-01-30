package api

import (
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
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
	gqlClient connector.Client
	logger    *log.Logger
}

func NewSigningRequestInfoHandler(client connector.Client, logger *log.Logger) csrInfoHandler {
	return csrInfoHandler{
		gqlClient: client,
		logger:    logger,
	}
}

func (ci *csrInfoHandler) GetSigningRequestInfo(w http.ResponseWriter, r *http.Request) {
	// TODO: make sure only calls with token are accepted

	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		reqerror.WriteErrorMessage(w, "Client Id not provided.", apperrors.CodeForbidden)

		return
	}

	application := authorizationHeaders.GetClientID()
	contextLogger := contextLogger(ci.logger, authorizationHeaders.GetClientID())

	baseURLs, err := middlewares.GetBaseURLsFromContext(r.Context(), middlewares.BaseURLsKey)
	if err != nil {
		contextLogger.Errorf("Failed to read Base URL context: %s.", err)
		reqerror.WriteErrorMessage(w, "Base URLs not provided.", apperrors.CodeInternal)

		return
	}

	contextLogger.Info("Getting Certificate Signing Request Info")

	configuration, err := ci.gqlClient.Configuration(authorizationHeaders)
	if err != nil {
		err = errors.Wrap(err, "Failed to get configuration")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	certInfo := connector.ToCertInfo(configuration)

	//TODO: handle case when configuration.Token is nil
	csrInfoResponse := ci.makeCSRInfoResponse(
		application,
		configuration.Token.Token,
		baseURLs.ConnectivityAdapterBaseURL,
		baseURLs.ConnectivityAdapterMTLSBaseURL,
		baseURLs.EventServiceBaseURL,
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
