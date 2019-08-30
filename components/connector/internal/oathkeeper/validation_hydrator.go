package oathkeeper

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/httputils"

	"github.com/kyma-incubator/compass/components/connector/internal/authentication"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
)

type ValidationHydrator struct {
	tokenService tokens.Service
	log          *logrus.Entry
}

func NewValidationHydrator(tokenService tokens.Service) ValidationHydrator {
	return ValidationHydrator{
		tokenService: tokenService,
		log:          logrus.WithField("Handler", "ValidationHydrator"),
	}
}

func (tvh ValidationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		tvh.log.Error(err)
		httputils.RespondWithError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(r.Body)

	connectorToken := r.Header.Get(authentication.ConnectorTokenHeader)
	if connectorToken == "" {
		tvh.log.Info("Token not provided")
		respondWithAuthSession(w, authSession)
		return
	}

	tvh.log.Info("Trying to resolve token...")

	tokenData, err := tvh.tokenService.Resolve(connectorToken)
	if err != nil {
		tvh.log.Infof("Invalid token provided: %s", err.Error())
		respondWithAuthSession(w, authSession)
		return
	}

	authSession.Header.Add(ClientIdFromTokenHeader, tokenData.ClientId)
	authSession.Header.Add(TokenTypeHeader, string(tokenData.Type))

	respondWithAuthSession(w, authSession)
}

// TODO: implement handler that will validate Istio cert header and set appropriate headers

func respondWithAuthSession(w http.ResponseWriter, authSession AuthenticationSession) {
	httputils.RespondWithBody(w, http.StatusOK, authSession)
}
