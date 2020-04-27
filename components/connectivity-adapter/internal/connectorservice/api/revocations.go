package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type revocationsHandler struct {
	connectorClientProvider connector.ClientProvider
	logger                  *log.Logger
}

func NewRevocationsHandler(connectorClientProvider connector.ClientProvider, logger *log.Logger) revocationsHandler {
	return revocationsHandler{
		connectorClientProvider: connectorClientProvider,
		logger:                  logger,
	}
}

func (rh *revocationsHandler) RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		rh.logger.Errorf("Failed to read authorization context: %s.", err)
		res.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	contextLogger := contextLogger(rh.logger, authorizationHeaders.GetSystemAuthID())

	contextLogger.Info("Revoke certificate")

	err = rh.connectorClientProvider.Client(r).Revoke(authorizationHeaders)

	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to revoke certificate"), apperrors.CodeInternal)

		return
	}

	respond(w, http.StatusCreated)
}
