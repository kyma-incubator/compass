package authentication

import (
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	tenantTokenKey       = "tenant"
	consumerTypeTokenKey = "consumerType"
)

type authContextMiddleware struct {
}

func NewAuthenticationContextMiddleware() *authContextMiddleware {
	return &authContextMiddleware{}
}

type tokenClaims struct {
	Tenant       string `json:"tenant"`
	ConsumerType string `json:"consumerType"`
}

func (tokenClaims) Valid() error {
	return nil
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			parser := jwt.Parser{}
			var tokenClaims tokenClaims
			_, _, err := parser.ParseUnverified(authorizationHeader[7:], tokenClaims)
			if err != nil {
				log.C(ctx).WithError(err).Error("could not parse token")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			tenant := tokenClaims.Tenant
			if tenant != "" {
				r = r.WithContext(PutIntoContext(r.Context(), TenantKey, tenant))
			}

			consumerType := tokenClaims.ConsumerType
			if consumerType != "" {
				r = r.WithContext(PutIntoContext(r.Context(), ConsumerType, consumerType))
			}
		}

		clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromTokenKey, clientIdFromToken))

		clientIdFromCertificate := r.Header.Get(oathkeeper.ClientIdFromCertificateHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromCertificateKey, clientIdFromCertificate))

		clientCertificateHash := r.Header.Get(oathkeeper.ClientCertificateHashHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientCertificateHashKey, clientCertificateHash))

		handler.ServeHTTP(w, r)
	})
}
