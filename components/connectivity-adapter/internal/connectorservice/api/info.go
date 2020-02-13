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
	HeaderContentType          = "Content-Type"
	ContentTypeApplicationJson = "application/json;charset=UTF-8"
)

type newResponseFunction func(string, string, string, string, string, model.CertInfo) interface{}

type infoHandler struct {
	connectorClient                connector.Client
	directorClientProvider         director.ClientProvider
	newResponse                    newResponseFunction
	connectivityAdapterBaseURL     string
	connectivityAdapterMTLSBaseURL string
	logger                         *log.Logger
}

func NewInfoHandler(
	connectorClient connector.Client,
	directorClientProvider director.ClientProvider,
	logger *log.Logger,
	connectivityAdapterBaseURL string,
	connectivityAdapterMTLSBaseURL string,
	newResponse newResponseFunction) infoHandler {

	return infoHandler{
		connectorClient:                connectorClient,
		directorClientProvider:         directorClientProvider,
		newResponse:                    newResponse,
		connectivityAdapterBaseURL:     connectivityAdapterBaseURL,
		connectivityAdapterMTLSBaseURL: connectivityAdapterMTLSBaseURL,
		logger:                         logger,
	}
}

func (ih *infoHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		ih.logger.Errorf("Failed to read authorization context: %s.", err)
		reqerror.WriteErrorMessage(w, "Client ID not provided.", apperrors.CodeForbidden)

		return
	}
	systemAuthID := authorizationHeaders.GetSystemAuthID()

	contextLogger := contextLogger(ih.logger, systemAuthID)
	contextLogger.Info("Getting Info")

	application, err := ih.directorClientProvider.Client(r).GetApplication(systemAuthID)
	if err != nil {
		err = errors.Wrap(err, "Failed to get Application from Director")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}

	configuration, err := ih.connectorClient.Configuration(authorizationHeaders)
	if err != nil {
		err = errors.Wrap(err, "Failed to get Configuration from Connector")
		contextLogger.Error(err.Error())
		reqerror.WriteError(w, err, apperrors.CodeInternal)

		return
	}
	certInfo := connector.ToCertInfo(configuration)

	//TODO: handle case when configuration.Token is nil
	infoResponse := ih.newResponse(
		application.Name,
		configuration.Token.Token,
		ih.connectivityAdapterBaseURL,
		ih.connectivityAdapterMTLSBaseURL,
		application.EventingConfiguration.DefaultURL,
		certInfo)

	respondWithBody(w, http.StatusOK, infoResponse, contextLogger)
}
