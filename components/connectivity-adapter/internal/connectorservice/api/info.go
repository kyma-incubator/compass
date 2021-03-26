package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/director"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	"github.com/pkg/errors"
)

const (
	HeaderContentType          = "Content-Type"
	ContentTypeApplicationJson = "application/json;charset=UTF-8"
	appNameLabel               = "name"
)

type infoHandler struct {
	connectorClientProvider        connector.ClientProvider
	directorClientProvider         director.ClientProvider
	makeResponseFunc               model.InfoProviderFunc
	connectivityAdapterBaseURL     string
	connectivityAdapterMTLSBaseURL string
}

func NewInfoHandler(
	connectorClientProvider connector.ClientProvider,
	directorClientProvider director.ClientProvider,
	makeResponseFunc model.InfoProviderFunc) infoHandler {

	return infoHandler{
		connectorClientProvider: connectorClientProvider,
		directorClientProvider:  directorClientProvider,
		makeResponseFunc:        makeResponseFunc,
	}
}

func (ih *infoHandler) GetInfo(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		res.WriteErrorMessage(w, "Client ID not provided.", apperrors.CodeForbidden)

		return
	}
	systemAuthID := authorizationHeaders.GetSystemAuthID()

	contextLogger := contextLogger(r.Context(), systemAuthID)

	contextLogger.Info("Getting Info")

	application, err := ih.directorClientProvider.Client(r).GetApplication(r.Context(), systemAuthID)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to get application from Director"), apperrors.CodeInternal)

		return
	}

	configuration, err := ih.connectorClientProvider.Client(r).Configuration(r.Context(), authorizationHeaders)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to get configuration from Connector"), apperrors.CodeInternal)
		return
	}

	infoResponse, err := ih.makeResponseFunc(getApplicationName(application), application.EventingConfiguration.DefaultURL, "", configuration)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to build info response"), apperrors.CodeInternal)

		return
	}

	respondWithBody(w, http.StatusOK, infoResponse, contextLogger)
}

func getApplicationName(application graphql.ApplicationExt) string {
	appName := application.Name

	labelsMap := map[string]interface{}(application.Labels)
	if labelsMap == nil {
		return appName
	}

	if label, ok := labelsMap[appNameLabel]; ok {
		name, ok := label.(string)
		if ok && name != "" {
			appName = name
		}
	}

	return appName
}
