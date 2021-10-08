package authentication

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

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

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			jwtTokenPayload, err := base64.RawStdEncoding.DecodeString(strings.Split(authorizationHeader[7:], ".")[1])
			if err != nil {
				log.C(ctx).WithError(err).Error("could not read token")
				w.WriteHeader(400)
				return
			}

			var payload map[string]interface{}
			err = json.Unmarshal([]byte(jwtTokenPayload), &payload)
			if err != nil {
				log.C(ctx).WithError(err).Error("could not parse token")
				w.WriteHeader(400)
				return
			}

			tenant := payload[tenantTokenKey].(string)
			r = r.WithContext(PutIntoContext(r.Context(), TenantKey, tenant))

			consumerType := payload[consumerTypeTokenKey].(string)
			r = r.WithContext(PutIntoContext(r.Context(), ConsumerType, consumerType))
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
