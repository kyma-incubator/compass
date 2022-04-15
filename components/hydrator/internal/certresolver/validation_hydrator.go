package certresolver

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type ValidationHydrator interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

//go:generate mockery --name=RevokedCertificatesCache --output=automock --outpkg=automock --case=underscore
type RevokedCertificatesCache interface {
	Get() map[string]string
}

type validationHydrator struct {
	certHeaderParsers []CertificateHeaderParser
	revokedCertsCache RevokedCertificatesCache
}

func NewValidationHydrator(cache RevokedCertificatesCache, certHeaderParsers ...CertificateHeaderParser) ValidationHydrator {
	return &validationHydrator{
		certHeaderParsers: certHeaderParsers,
		revokedCertsCache: cache,
	}
}

// ServeHTTP checks the certificate forwarded by the istio mtls gateway against all the configured CertificateHeaderParsers
// First CertificateHeaderParser that matches (successfully parse the subject) is used to extract the clientID, certificate hash and issuer.
// If there is no matching CertificateHeaderParser, an empty oathkeeper session is returned.
func (vh *validationHydrator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		ctx         = r.Context()
		authSession oathkeeper.AuthenticationSession
	)

	if err := json.NewDecoder(r.Body).Decode(&authSession); err != nil {
		log.C(ctx).WithError(err).Errorf("Failed to decode request body: %v", err)
		httputils.RespondWithError(ctx, w, http.StatusBadRequest, errors.Wrap(err, "failed to decode Authentication Session from body"))
		return
	}
	defer httputils.Close(ctx, r.Body)

	log.C(ctx).Info("Trying to validate certificate header...")

	var (
		issuer   string
		certData *CertificateData
	)

	for _, certHeaderParser := range vh.certHeaderParsers {
		certData = certHeaderParser.GetCertificateData(r)
		if certData == nil {
			log.C(ctx).Infof("Certificate header is not valid for issuer %s", certHeaderParser.GetIssuer())
			continue
		}
		issuer = certHeaderParser.GetIssuer()
		break
	}

	if certData == nil {
		log.C(ctx).Info("No valid certificate header found")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	log.C(ctx).Infof("Certificate header is valid for issuer %s", issuer)

	if isCertificateRevoked := vh.Contains(certData.CertificateHash); isCertificateRevoked {
		log.C(ctx).Info("Certificate is revoked.")
		respondWithAuthSession(ctx, w, authSession)
		return
	}

	if authSession.Header == nil {
		authSession.Header = map[string][]string{}
	}

	authSession.Header.Add(oathkeeper.ClientIdFromCertificateHeader, certData.ClientID)
	authSession.Header.Add(oathkeeper.ClientCertificateHashHeader, certData.CertificateHash)
	authSession.Header.Add(oathkeeper.ClientCertificateIssuerHeader, issuer)

	authSession.Extra = appendExtra(authSession.Extra, certData.AuthSessionExtra)

	log.C(ctx).Info("Certificate header validated successfully")
	respondWithAuthSession(ctx, w, authSession)
}

func (vh *validationHydrator) Contains(hash string) bool {
	configMap := vh.revokedCertsCache.Get()

	found := false
	if configMap != nil {
		_, found = configMap[hash]
	}

	return found
}

func respondWithAuthSession(ctx context.Context, w http.ResponseWriter, authSession oathkeeper.AuthenticationSession) {
	httputils.RespondWithBody(ctx, w, http.StatusOK, authSession)
}

func appendExtra(extra, extraFromHeaderParser map[string]interface{}) map[string]interface{} {
	if extra == nil {
		extra = map[string]interface{}{}
	}
	for k, v := range extraFromHeaderParser {
		extra[k] = v
	}
	return extra
}
