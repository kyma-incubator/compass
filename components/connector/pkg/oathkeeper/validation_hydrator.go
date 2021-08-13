package oathkeeper

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/internal/revocation"
	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type ValidationHydrator interface {
	ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request)
}

type validationHydrator struct {
	certHeaderParsers      []CertificateHeaderParser
	revokedCertsRepository revocation.RevokedCertificatesRepository
}

func NewValidationHydrator(revokedCertsRepository revocation.RevokedCertificatesRepository, certHeaderParsers ...CertificateHeaderParser) ValidationHydrator {
	return &validationHydrator{
		certHeaderParsers:      certHeaderParsers,
		revokedCertsRepository: revokedCertsRepository,
	}
}

func (tvh *validationHydrator) ResolveIstioCertHeader(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var authSession AuthenticationSession
	err := json.NewDecoder(r.Body).Decode(&authSession)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to decode request body: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(ctx, r.Body)

	log.C(ctx).Info("Trying to validate certificate header...")

	var clientID, hash, issuer string
	var found bool

	for _, certHeaderParser := range tvh.certHeaderParsers {
		clientID, hash, found = certHeaderParser.GetCertificateData(r)
		if !found {
			log.C(ctx).Infof("Certificate header is not valid for issuer %s", certHeaderParser.GetIssuer())
			continue
		}
		issuer = certHeaderParser.GetIssuer()
		break
	}

	if !found {
		log.C(ctx).Info("No valid certificate header found")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Infof("Certificate header is valid for issuer %s", issuer)

	if isCertificateRevoked := tvh.revokedCertsRepository.Contains(hash); isCertificateRevoked {
		log.C(ctx).Info("Certificate is revoked.")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(ClientIdFromCertificateHeader, clientID)
	authSession.Header.Add(ClientCertificateHashHeader, hash)
	authSession.Header.Add(ClientCertificateIssuerHeader, issuer)

	log.C(ctx).Info("Certificate header validated successfully")
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
