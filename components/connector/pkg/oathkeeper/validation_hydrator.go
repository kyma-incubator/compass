package oathkeeper

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/connector/internal/httputils"

	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
)

type ValidationHydrator interface {
	ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request)
	ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request)
}

type validationHydrator struct {
	tokenService     tokens.Service
	certHeaderParser CertificateHeaderParser
	log              *logrus.Entry
}

func NewValidationHydrator(tokenService tokens.Service, certHeaderParser CertificateHeaderParser) ValidationHydrator {
	return &validationHydrator{
		tokenService:     tokenService,
		certHeaderParser: certHeaderParser,
		log:              logrus.WithField("Handler", "ValidationHydrator"),
	}
}

func (tvh *validationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		tvh.log.Error(err)
		httputils.RespondWithError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(r.Body)

	connectorToken := r.Header.Get(ConnectorTokenHeader)
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

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(ClientIdFromTokenHeader, tokenData.ClientId)

	tvh.tokenService.Delete(connectorToken)

	tvh.log.Infof("Token for %s resolved successfully", tokenData.ClientId)
	respondWithAuthSession(w, authSession)
}

func (tvh *validationHydrator) ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request) {
	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		tvh.log.Error(err)
		httputils.RespondWithError(w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(r.Body)

	tvh.log.Info("Trying to validate certificate header...")

	commonName, hash, found := tvh.certHeaderParser.GetCertificateData(r)
	if !found {
		tvh.log.Info("No valid certificate header found")
		respondWithAuthSession(w, authSession)
		return
	}

	// TODO: Check if certificate is revoked

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(ClientIdFromCertificateHeader, commonName)
	authSession.Header.Add(ClientCertificateHashHeader, hash)

	tvh.log.Info("Certificate header validated successfully")
	respondWithAuthSession(w, authSession)
}

func respondWithAuthSession(w http.ResponseWriter, authSession AuthenticationSession) {
	httputils.RespondWithBody(w, http.StatusOK, authSession)
}
