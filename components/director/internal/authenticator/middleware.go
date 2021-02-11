package authenticator

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/scope"

	"github.com/form3tech-oss/jwt-go"
	"github.com/lestrrat-go/jwx/jwk"
)

const (
	AuthorizationHeaderKey = "Authorization"
	ClientUserHeader       = "client_user"
)

type Authenticator struct {
	jwksEndpoint        string
	allowJWTSigningNone bool
	cachedJWKs          *jwk.Set
	mux                 sync.Mutex
}

func New(jwksEndpoint string, allowJWTSigningNone bool) *Authenticator {
	return &Authenticator{jwksEndpoint: jwksEndpoint, allowJWTSigningNone: allowJWTSigningNone}
}

func (a *Authenticator) SynchronizeJWKS(ctx context.Context) error {
	log.C(ctx).Info("Synchronizing JWKS...")
	a.mux.Lock()
	defer a.mux.Unlock()
	jwks, err := FetchJWK(ctx, a.jwksEndpoint)
	if err != nil {
		return errors.Wrapf(err, "while fetching JWKS from endpoint %s", a.jwksEndpoint)
	}

	a.cachedJWKs = jwks
	return nil
}

func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			bearerToken, err := a.getBearerToken(r)
			if err != nil {
				log.C(ctx).WithError(err).Error("An error has occurred while getting token from header. Error code: ", http.StatusBadRequest)
				apperrors.WriteAppError(ctx, w, err, http.StatusBadRequest)
				return
			}

			claims, err := a.parseClaimsWithRetry(r.Context(), bearerToken)
			if err != nil {
				log.C(ctx).WithError(err).Error("An error has occurred while parsing claims. Error code: ", http.StatusUnauthorized)
				apperrors.WriteAppError(ctx, w, err, http.StatusUnauthorized)
				return
			}

			if claims.Tenant == "" && claims.ExternalTenant != "" {
				err := apperrors.NewTenantNotFoundError(claims.ExternalTenant)
				log.C(ctx).WithError(err).Error("Tenant not found. Error code: ", http.StatusBadRequest)
				apperrors.WriteAppError(ctx, w, err, http.StatusBadRequest)
				return
			}

			ctx = a.contextWithClaims(r.Context(), claims)

			if clientUser := r.Header.Get(ClientUserHeader); clientUser != "" {
				log.C(ctx).Infof("Found %s header in request with value: %s", ClientUserHeader, clientUser)
				ctx = client.SaveToContext(ctx, clientUser)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) parseClaimsWithRetry(ctx context.Context, bearerToken string) (Claims, error) {
	var claims Claims
	var err error

	claims, err = a.parseClaims(bearerToken)
	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || validationErr.Inner != rsa.ErrVerification {
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
		}

		err := a.SynchronizeJWKS(ctx)
		if err != nil {
			return Claims{}, apperrors.InternalErrorFrom(err, "while synchronizing JWKs during parsing token")
		}

		claims, err = a.parseClaims(bearerToken)
		if err != nil {
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
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
		return "", apperrors.NewUnauthorizedError("invalid bearer token")
	}

	reqToken = strings.TrimPrefix(reqToken, "Bearer ")
	return reqToken, nil
}

func (a *Authenticator) contextWithClaims(ctx context.Context, claims Claims) context.Context {
	ctxWithTenants := tenant.SaveToContext(ctx, claims.Tenant, claims.ExternalTenant)
	scopesArray := strings.Split(claims.Scopes, " ")
	ctxWithScopes := scope.SaveToContext(ctxWithTenants, scopesArray)
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
