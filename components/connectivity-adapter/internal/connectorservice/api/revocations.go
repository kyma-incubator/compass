package api

import (
	"net/http"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/res"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/connector"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/connectorservice/api/middlewares"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type revocationsHandler struct {
	connectorClientProvider connector.ClientProvider
}

func NewRevocationsHandler(connectorClientProvider connector.ClientProvider) revocationsHandler {
	return revocationsHandler{
		connectorClientProvider: connectorClientProvider,
	}
}

func (rh *revocationsHandler) RevokeCertificate(w http.ResponseWriter, r *http.Request) {
	authorizationHeaders, err := middlewares.GetAuthHeadersFromContext(r.Context(), middlewares.AuthorizationHeadersKey)
	if err != nil {
		log.C(r.Context()).Errorf("Failed to read authorization context: %s.", err)
		res.WriteErrorMessage(w, "Failed to read authorization context.", apperrors.CodeForbidden)

		return
	}

	contextLogger := contextLogger(r.Context(), authorizationHeaders.GetSystemAuthID())

	contextLogger.Info("Revoke certificate")

	err = rh.connectorClientProvider.Client(r).Revoke(r.Context(), authorizationHeaders)

	if err != nil {
		respondWithError(w, contextLogger, errors.Wrap(err, "Failed to revoke certificate"), apperrors.CodeInternal)

		return
	}

	respond(w, http.StatusCreated)
}
