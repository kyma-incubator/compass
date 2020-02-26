package authenticator

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/dgrijalva/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

const AuthorizationHeaderKey = "Authorization"

type Authenticator struct {
	jwksEndpoint        string
	allowJWTSigningNone bool
	cachedJWKs          *jwk.Set
	mux                 sync.Mutex
}

func New(jwksEndpoint string, allowJWTSigningNone bool) *Authenticator {
	return &Authenticator{jwksEndpoint: jwksEndpoint, allowJWTSigningNone: allowJWTSigningNone}
}

func (a *Authenticator) SynchronizeJWKS() error {
	log.Info("Synchronizing JWKS...")
	a.mux.Lock()
	defer a.mux.Unlock()
	jwks, err := FetchJWK(a.jwksEndpoint)
	if err != nil {
		return errors.Wrapf(err, "while fetching JWKS from endpoint %s", a.jwksEndpoint)
	}

	a.cachedJWKs = jwks
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

			claims, err := a.parseClaimsWithRetry(bearerToken)
			if err != nil {
				log.Error(err)
				a.writeError(w, err.Error(), http.StatusUnauthorized)
				return
			}

			ctx := a.contextWithClaims(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) parseClaimsWithRetry(bearerToken string) (Claims, error) {
	var claims Claims
	var err error

	claims, err = a.parseClaims(bearerToken)
	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || validationErr.Inner != rsa.ErrVerification {
			return Claims{}, errors.Wrap(err, "while parsing token")
		}

		err := a.SynchronizeJWKS()
		if err != nil {
			return Claims{}, errors.Wrap(err, "while synchronizing JWKs during parsing token")
		}

		claims, err = a.parseClaims(bearerToken)
		if err != nil {
			return Claims{}, errors.Wrap(err, "while parsing token")
		}

		return claims, err
	}

	return claims, nil
}

func (a *Authenticator) parseClaims(bearerToken string) (Claims, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc())
	if err != nil {
		return Claims{}, err
	}

	return claims, nil
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
			a.mux.Lock()
			keys := a.cachedJWKs.Keys
			a.mux.Unlock()
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
	Errors []gqlError `json:"errors"`
}

type gqlError struct {
	Message string `json:"message"`
}

func (a *Authenticator) writeError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	resp := errorResponse{Errors: []gqlError{{Message: message}}}
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Error(err, "while encoding data")
	}
}
