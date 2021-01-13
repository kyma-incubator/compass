package oathkeeper

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/connector/internal/httputils"
	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/connector/internal/tokens"
	"github.com/pkg/errors"
)

type ValidationHydrator interface {
	ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request)
	ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request)
}

type validationHydrator struct {
	transact               persistence.Transactioner
	tokenService           tokens.Service
	certHeaderParser       CertificateHeaderParser
	revokedCertsRepository revocation.RevokedCertificatesRepository
}

func NewValidationHydrator(tokenService tokens.Service,
	certHeaderParser CertificateHeaderParser,
	revokedCertsRepository revocation.RevokedCertificatesRepository,
	transact persistence.Transactioner) ValidationHydrator {
	return &validationHydrator{
		tokenService:           tokenService,
		certHeaderParser:       certHeaderParser,
		revokedCertsRepository: revokedCertsRepository,
		transact:               transact,
	}
}

func (tvh *validationHydrator) ResolveConnectorTokenHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to decode request body")
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(ctx, r.Body)

	connectorToken := r.Header.Get(ConnectorTokenHeader)
	if connectorToken == "" {
		connectorToken = r.URL.Query().Get(ConnectorTokenQueryParam)
	}

	if connectorToken == "" {
		log.C(ctx).Info("Token not provided")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Info("Trying to resolve token...")

	tx, err := tvh.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to open db transaction")
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not fulfill request"))
		return
	}
	defer tvh.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tokenData, err := tvh.tokenService.Resolve(ctx, connectorToken)
	if err != nil {
		log.C(ctx).Infof("Invalid token provided: %s", err.Error())
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(ClientIdFromTokenHeader, tokenData.ClientId)

	if err := tvh.tokenService.Delete(ctx, connectorToken); err != nil {
		log.C(ctx).WithError(err).Error("Failed to invalidate token")
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not invalidate token"))
		return
	}

	err = tx.Commit()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Could not commit transaction")
		httputils.RespondWithError(ctx, w, http.StatusInternalServerError, errors.New("could not fulfill request"))
		return
	}

	log.C(ctx).Infof("Token for %s resolved successfully", tokenData.ClientId)
	respondWithAuthSession(ctx, w, authSession)
}

func (tvh *validationHydrator) ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to decode request body")
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(ctx, r.Body)

	log.C(ctx).Info("Trying to validate certificate header...")

	commonName, hash, found := tvh.certHeaderParser.GetCertificateData(r)
	if !found {
		log.C(ctx).Info("No valid certificate header found")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if isCertificateRevoked := tvh.revokedCertsRepository.Contains(hash); isCertificateRevoked {
		log.C(ctx).Info("Certificate is revoked.")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(ClientIdFromCertificateHeader, commonName)
	authSession.Header.Add(ClientCertificateHashHeader, hash)

	log.C(ctx).Info("Certificate header validated successfully")
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
