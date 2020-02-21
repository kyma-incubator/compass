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

type infoHandler struct {
	connectorClient                connector.Client
	directorClientProvider         director.ClientProvider
	makeResponseFunc               model.InfoProviderFunc
	connectivityAdapterBaseURL     string
	connectivityAdapterMTLSBaseURL string
	logger                         *log.Logger
}

func NewInfoHandler(
	connectorClient connector.Client,
	directorClientProvider director.ClientProvider,
	logger *log.Logger,
	makeResponseFunc model.InfoProviderFunc) infoHandler {

	return infoHandler{
		connectorClient:        connectorClient,
		directorClientProvider: directorClientProvider,
		makeResponseFunc:       makeResponseFunc,
		logger:                 logger,
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
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to get application from Director"), apperrors.CodeInternal)

		return
	}

	configuration, err := ih.connectorClient.Configuration(authorizationHeaders)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to get configuration from Connector"), apperrors.CodeInternal)

		return
	}

	infoResponse, err := ih.makeResponseFunc(
		application.Name,
		application.EventingConfiguration.DefaultURL,
		"",
		configuration)

	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to build info response"), apperrors.CodeInternal)

		return
	}

	respondWithBody(w, http.StatusOK, infoResponse, contextLogger)
}
