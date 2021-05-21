package authenticator

import (
	"context"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/lestrrat-go/iter/arrayiter"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

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
	JwksKeyIDKey           = "kid"
)

type Authenticator struct {
	jwksEndpoint        string
	allowJWTSigningNone bool
	cachedJWKS          jwk.Set
	clientIDHeaderKey   string
	mux                 sync.RWMutex
}

func New(jwksEndpoint string, allowJWTSigningNone bool, clientIDHeaderKey string) *Authenticator {
	return &Authenticator{
		jwksEndpoint:        jwksEndpoint,
		allowJWTSigningNone: allowJWTSigningNone,
		clientIDHeaderKey:   clientIDHeaderKey,
	}
}

func (a *Authenticator) SynchronizeJWKS(ctx context.Context) error {
	log.C(ctx).Info("Synchronizing JWKS...")
	a.mux.Lock()
	defer a.mux.Unlock()
	jwks, err := authenticator.FetchJWK(ctx, a.jwksEndpoint)
	if err != nil {
		return errors.Wrapf(err, "while fetching JWKS from endpoint %s", a.jwksEndpoint)
	}

	a.cachedJWKS = jwks
	log.C(ctx).Info("Successfully synchronized JWKS")

	return nil
}

func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			bearerToken, err := a.getBearerToken(r)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while getting token from header. Error code: %d: %v", http.StatusBadRequest, err)
				apperrors.WriteAppError(ctx, w, err, http.StatusBadRequest)
				return
			}

			claims, err := a.parseClaimsWithRetry(r.Context(), bearerToken)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while parsing claims. Error code: %d: %v", http.StatusUnauthorized, err)
				apperrors.WriteAppError(ctx, w, err, http.StatusUnauthorized)
				return
			}

			if claims.Tenant == "" && claims.ExternalTenant != "" {
				err := apperrors.NewTenantNotFoundError(claims.ExternalTenant)
				log.C(ctx).WithError(err).Errorf("Tenant not found. Error code: %d: %v", http.StatusBadRequest, err)
				apperrors.WriteAppError(ctx, w, err, http.StatusBadRequest)
				return
			}

			ctx = a.contextWithClaims(r.Context(), claims)

			if clientUser := r.Header.Get(a.clientIDHeaderKey); clientUser != "" {
				log.C(ctx).Infof("Found %s header in request with value: %s", a.clientIDHeaderKey, clientUser)
				ctx = client.SaveToContext(ctx, clientUser)
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) parseClaimsWithRetry(ctx context.Context, bearerToken string) (Claims, error) {
	var claims Claims
	var err error

	claims, err = a.parseClaims(ctx, bearerToken)
	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || (validationErr.Inner != rsa.ErrVerification && !apperrors.IsKeyDoesNotExist(validationErr.Inner)) {
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
		}

		err := a.SynchronizeJWKS(ctx)
		if err != nil {
			return Claims{}, apperrors.InternalErrorFrom(err, "while synchronizing JWKS during parsing token")
		}

		claims, err = a.parseClaims(ctx, bearerToken)
		if err != nil {
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
		}

		return claims, err
	}

	return claims, nil
}

func (a *Authenticator) parseClaims(ctx context.Context, bearerToken string) (Claims, error) {
	claims := Claims{}
	_, err := jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc(ctx))
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

func (a *Authenticator) getKeyFunc(ctx context.Context) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		unsupportedErr := fmt.Errorf("unexpected signing method: %v", token.Method.Alg())

		switch token.Method.Alg() {
		case jwt.SigningMethodRS256.Name:
			a.mux.RLock()
			keys := a.cachedJWKS
			a.mux.RUnlock()

			keyID, err := a.getKeyID(*token)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while getting the token signing key ID: %v", err)
				return nil, errors.Wrap(err, "while getting the key ID")
			}

			keyIterator := &authenticator.JWTKeyIterator{
				AlgorithmCriteria: func(alg string) bool {
					return token.Method.Alg() == alg
				},
				IDCriteria: func(id string) bool {
					return id == keyID
				},
			}

			if err := arrayiter.Walk(ctx, keys, keyIterator); err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while walking through the JWKS: %v", err)
				return nil, err
			}

			if keyIterator.ResultingKey == nil {
				log.C(ctx).Debug("Signing key is not found")
				return nil, apperrors.NewKeyDoesNotExistError(keyID)
			}

			return keyIterator.ResultingKey, nil
		case jwt.SigningMethodNone.Alg():
			if !a.allowJWTSigningNone {
				return nil, unsupportedErr
			}
			return jwt.UnsafeAllowNoneSignatureType, nil
		}

		return nil, unsupportedErr
	}
}

func (a *Authenticator) getKeyID(token jwt.Token) (string, error) {
	keyID, ok := token.Header[JwksKeyIDKey]
	if !ok {
		return "", apperrors.NewInternalError("unable to find the key ID in the token")
	}

	keyIDStr, ok := keyID.(string)
	if !ok {
		return "", apperrors.NewInternalError("unable to cast the key ID to a string")
	}

	return keyIDStr, nil
}
