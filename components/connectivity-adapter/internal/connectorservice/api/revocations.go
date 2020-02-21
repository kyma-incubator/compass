package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/reqerror"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type revocationsHandler struct {
	gqlClient connector.Client
	logger    *log.Logger
}

func NewRevocationsHandler(client connector.Client, logger *log.Logger) revocationsHandler {
	return revocationsHandler{
		gqlClient: client,
		logger:    logger,
	}
}

func (rh *revocationsHandler) RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		rh.logger.Errorf("Failed to read authorization context: %s.", err)
		reqerror.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	contextLogger := contextLogger(rh.logger, authorizationHeaders.GetSystemAuthID())

	contextLogger.Info("Revoke certificate")

	err = rh.gqlClient.Revoke(authorizationHeaders)
	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to revoke certificate"), apperrors.CodeInternal)

		return
	}

	respond(w, http.StatusCreated)
}
