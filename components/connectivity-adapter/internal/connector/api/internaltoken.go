package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/graphql"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connector/model"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	log "github.com/sirupsen/logrus"
)

const ApplicationHeader = "Application"

type tokenHandler struct {
	gqlClient                  graphql.Client
	logger                     *log.Logger
	connectivityAdapterBaseURL string
}

func NewTokenHandler(client graphql.Client, connectivityAdapterBaseURL string, logger *log.Logger) tokenHandler {
	return tokenHandler{
		gqlClient:                  client,
		connectivityAdapterBaseURL: connectivityAdapterBaseURL,
		logger:                     logger,
	}
}

func (th *tokenHandler) GetToken(w http.ResponseWriter, r *http.Request) {
	application := r.Header.Get(ApplicationHeader)
	contextLogger := th.logger.WithField("application", application)

	token, err := th.gqlClient.Token(application)
	if err != nil {
		th.logger.Errorf("Failed to get token: %s.", err)
		reqerror.WriteErrorMessage(w, "Failed to get token.", apperrors.CodeForbidden)
	}

	res := model.MakeTokenResponse(application, th.connectivityAdapterBaseURL, token)

	respondWithBody(w, http.StatusCreated, res, contextLogger)
}
