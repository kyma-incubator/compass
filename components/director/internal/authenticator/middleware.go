package authenticator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

const AuthorizationHeaderKey = "Authorization"

type Authenticator struct {
	jwksEndpoint        string
	allowJWTSigningNone bool
	Jwks                *jwk.Set
}

func New(jwksEndpoint string, allowJWTSigningNone bool) *Authenticator {
	return &Authenticator{jwksEndpoint: jwksEndpoint, allowJWTSigningNone: allowJWTSigningNone}
}

func (a *Authenticator) SynchronizeJWKS() error {
	jwks, err := FetchJWK(a.jwksEndpoint)
	if err != nil {
		return errors.Wrapf(err, "while fetching JWKS from endpoint %s", a.jwksEndpoint)
	}

	a.Jwks = jwks
	return nil
}

func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			bearerToken, err := a.getBearerToken(r)
			if err != nil {
				log.Error(errors.Wrap(err, "while getting token from header"))
				a.writeError(w, err.Error(), http.StatusBadRequest)
				return
			}

			claims := Claims{}
			_, err = jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc())
			if err != nil {
				wrappedErr := errors.Wrap(err, "while parsing token")
				log.Error(wrappedErr)

				vErr, ok := err.(*jwt.ValidationError)
				if !ok || !isInvalidTenantError(vErr.Inner) {
					a.writeError(w, wrappedErr.Error(), http.StatusUnauthorized)
					return
				}

				a.writeError(w, fmt.Sprintf("forbidden: %s", err.Error()), http.StatusForbidden)
				return
			}

			ctx := a.contextWithClaims(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) getBearerToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get(AuthorizationHeaderKey)
	if reqToken == "" {
		return "", errors.New("invalid bearer token")
	}

	reqToken = strings.TrimPrefix(reqToken, "Bearer ")
	return reqToken, nil
}

func (a *Authenticator) contextWithClaims(ctx context.Context, claims Claims) context.Context {
	ctxWithTenant := tenant.SaveToContext(ctx, claims.Tenant)
	scopesArray := strings.Split(claims.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenant, scopesArray)
	apiConsumer := consumer.Consumer{ConsumerID: claims.ConsumerID, ConsumerType: claims.ConsumerType}
	ctxWithConsumerInfo := consumer.SaveToContext(ctxWithScopes, apiConsumer)
	return ctxWithConsumerInfo
}

func (a *Authenticator) getKeyFunc() func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		unsupportedErr := fmt.Errorf("unexpected signing method: %v", token.Method.Alg())

		switch token.Method.Alg() {
		case jwt.SigningMethodRS256.Name:
			keys := a.Jwks.Keys
			for _, key := range keys {
				if key.Algorithm() == token.Method.Alg() {
					return key.Materialize()
				}
			}

			return nil, fmt.Errorf("unable to find key for algorithm %s", token.Method.Alg())
		case jwt.SigningMethodNone.Alg():
			if !a.allowJWTSigningNone {
				return nil, unsupportedErr
			}
			return jwt.UnsafeAllowNoneSignatureType, nil
		}

		return nil, unsupportedErr
	}
}

type errorResponse struct {
	Errors []string `json:"errors"`
}

func (a *Authenticator) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	resp := errorResponse{Errors: []string{message}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error(err, "while encoding data")
	}
}
