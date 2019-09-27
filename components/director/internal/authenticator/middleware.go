package authenticator

import (
	"context"
	"fmt"
	"net/http"
	"strings"

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
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			claims := Claims{}

			token, err := jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc())
			if err != nil {
				wrappedErr := errors.Wrap(err, "while parsing token")
				log.Error(wrappedErr)
				http.Error(w, wrappedErr.Error(), http.StatusUnauthorized)
				return
			}

			if !token.Valid {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// TODO: Remove getting tenant from header after implementing https://github.com/kyma-incubator/compass/issues/288
			tenantHeader := r.Header.Get(tenant.TenantHeaderName)
			if tenantHeader != "" {
				claims.Tenant = tenantHeader
			}

			ctx := a.contextWithClaims(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) getBearerToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get(AuthorizationHeaderKey)
	if reqToken == "" {
		return "", errors.New("Invalid bearer token")
	}

	reqToken = strings.TrimPrefix(reqToken, "Bearer ")
	return reqToken, nil
}

func (a *Authenticator) contextWithClaims(ctx context.Context, claims Claims) context.Context {
	ctxWithTenant := tenant.SaveToContext(ctx, claims.Tenant)
	scopesArray := strings.Split(claims.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenant, scopesArray)
	return ctxWithScopes
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
