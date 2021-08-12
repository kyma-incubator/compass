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
	certHeaderParser       CertificateHeaderParser
	revokedCertsRepository revocation.RevokedCertificatesRepository
	issuer                 string
}

func NewValidationHydrator(certHeaderParser CertificateHeaderParser, revokedCertsRepository revocation.RevokedCertificatesRepository, issuer string) ValidationHydrator {
	return &validationHydrator{
		certHeaderParser:       certHeaderParser,
		revokedCertsRepository: revokedCertsRepository,
		issuer:                 issuer,
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
	authSession.Header.Add(ClientCertificateIssuerHeader, tvh.issuer)

	log.C(ctx).Info("Certificate header validated successfully")
	respondWithAuthSession(ctx, w, authSession)
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}
