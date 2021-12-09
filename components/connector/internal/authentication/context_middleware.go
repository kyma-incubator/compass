package authentication

import (
	"encoding/json"
	"net/http"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	logKeyConsumerType  string = "consumer-type"
	logKeyCertClientId         = "cert-client-id"
	logKeyTokenClientId        = "token-client-id"
)

type authContextMiddleware struct {
}

func NewAuthenticationContextMiddleware() *authContextMiddleware {
	return &authContextMiddleware{}
}

type tokenClaims struct {
	Tenant       map[string]string `json:"tenant"`
	ConsumerType string            `json:"consumerType"`
}

func (tokenClaims) Valid() error {
	return nil
}

// UnmarshalJSON implements Unmarshaler interface. The method unmarshal the data from b into Claims structure.
func (c *tokenClaims) UnmarshalJSON(b []byte) error {
	tokenClaims := struct {
		TenantString string `json:"tenant"`
		ConsumerType string `json:"consumerType"`
	}{}

	err := json.Unmarshal(b, &tokenClaims)
	if err != nil {
		return errors.Wrap(err, "while unmarshaling token claims:")
	}

	c.ConsumerType = tokenClaims.ConsumerType

	if err := json.Unmarshal([]byte(tokenClaims.TenantString), &c.Tenant); err != nil {
		return errors.Wrap(err, "while unmarshaling tenants")
	}

	return nil
}

func (acm *authContextMiddleware) PropagateAuthentication(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authorizationHeader := r.Header.Get("Authorization")
		if authorizationHeader != "" {
			parser := jwt.Parser{}
			var tokenClaims tokenClaims
			_, _, err := parser.ParseUnverified(authorizationHeader[7:], &tokenClaims)
			if err != nil {
				log.C(ctx).WithError(err).Error("could not parse token")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			tenant, found := tokenClaims.Tenant["consumerTenant"]
			if found && tenant != "" {
				r = r.WithContext(PutIntoContext(r.Context(), TenantKey, tenant))
			}

			consumerType := tokenClaims.ConsumerType
			if consumerType != "" {
				r = r.WithContext(PutIntoContext(r.Context(), ConsumerType, consumerType))
			}

			if mdc := log.MdcFromContext(ctx); nil != mdc {
				if consumerType != "" {
					mdc.Set(logKeyConsumerType, tokenClaims.ConsumerType)
				}
			}
		}

		clientIdFromToken := r.Header.Get(oathkeeper.ClientIdFromTokenHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromTokenKey, clientIdFromToken))

		clientIdFromCertificate := r.Header.Get(oathkeeper.ClientIdFromCertificateHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientIdFromCertificateKey, clientIdFromCertificate))

		clientCertificateHash := r.Header.Get(oathkeeper.ClientCertificateHashHeader)
		r = r.WithContext(PutIntoContext(r.Context(), ClientCertificateHashKey, clientCertificateHash))

		if mdc := log.MdcFromContext(ctx); nil != mdc {
			if clientIdFromToken != "" {
				mdc.Set(logKeyTokenClientId, clientIdFromToken)
			}
			if clientIdFromCertificate != "" {
				mdc.Set(logKeyCertClientId, clientIdFromCertificate)
			}
		}
		handler.ServeHTTP(w, r)
	})
}
