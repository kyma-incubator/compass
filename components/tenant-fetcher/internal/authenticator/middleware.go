package authenticator

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/lestrrat-go/iter/arrayiter"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	AuthorizationHeaderKey = "Authorization"
	JwksKeyIDKey           = "kid"
)

type Error struct {
	Message string `json:"message"`
}

type Authenticator struct {
	mux                        sync.RWMutex
	cachedJWKS                 jwk.Set
	jwksEndpoints              []string
	zoneId                     string
	trustedClaimPrefixes       []string
	subscriptionCallbacksScope string
	allowJWTSigningNone        bool
}

func New(jwksEndpoints []string, zoneId, subscriptionCallbacksScope string, trustedClaimPrefixes []string, allowJWTSigningNone bool) *Authenticator {
	return &Authenticator{
		jwksEndpoints:              jwksEndpoints,
		zoneId:                     zoneId,
		trustedClaimPrefixes:       trustedClaimPrefixes,
		subscriptionCallbacksScope: subscriptionCallbacksScope,
		allowJWTSigningNone:        allowJWTSigningNone,
	}
}

func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := a.getBearerToken(r)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while extracting the JWT token. Error code: %d: %v", http.StatusUnauthorized, err)
				a.writeAppError(ctx, w, err, http.StatusBadRequest)
				return
			}

			claims, err := a.parseClaimsWithRetry(r.Context(), token)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error has occurred while parsing claims. Error code: %d: %v", http.StatusUnauthorized, err)
				a.writeAppError(ctx, w, err, http.StatusUnauthorized)
				return
			}

			if claims.ZID != a.zoneId {
				log.C(ctx).Errorf(`Zone id "%s" from user token does not match the trusted zone %s`, claims.ZID, a.zoneId)
				a.writeAppError(ctx, w, errors.Errorf(`Zone id "%s" from user token is not trusted`, claims.ZID), http.StatusUnauthorized)
				return
			}

			scopes := PrefixScopes(a.trustedClaimPrefixes, a.subscriptionCallbacksScope)
			if !stringsAnyEquals(scopes, strings.Join(claims.Scopes, " ")) {
				log.C(ctx).Errorf(`Scope "%s" from user token does not match the trusted scopes`, claims.Scopes)
				a.writeAppError(ctx, w, errors.Errorf(`Scope "%s" is not trusted`, claims.Scopes), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Authenticator) SetJWKSEndpoints(urls []string) {
	a.jwksEndpoints = urls
}

func (a *Authenticator) getBearerToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get(AuthorizationHeaderKey)
	if reqToken == "" {
		return "", apperrors.NewUnauthorizedError("invalid bearer token")
	}

	reqToken = strings.TrimPrefix(reqToken, "Bearer ")
	return reqToken, nil
}

func (a *Authenticator) writeAppError(ctx context.Context, w http.ResponseWriter, appErr error, statusCode int) {
	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(Error{Message: appErr.Error()})
	if err != nil {
		log.C(ctx).WithError(err).Errorf("An error occurred while encoding data: %v", err)
	}
}

func (a *Authenticator) parseClaimsWithRetry(ctx context.Context, bearerToken string) (Claims, error) {
	var claims Claims
	var err error

	claims, err = a.parseClaims(ctx, bearerToken)
	if err != nil {
		validationErr, ok := err.(*jwt.ValidationError)
		if !ok || (validationErr.Inner != rsa.ErrVerification && !apperrors.IsKeyDoesNotExist(validationErr.Inner)) {
			log.C(ctx).WithError(err).Errorf("An error occurred while parsing claims: %v", err)
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
		}

		if err := a.SynchronizeJWKS(ctx); err != nil {
			return Claims{}, apperrors.InternalErrorFrom(err, "while synchronizing JWKS during parsing token")
		}

		claims, err = a.parseClaims(ctx, bearerToken)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("An error occurred while parsing claims: %v", err)
			return Claims{}, apperrors.NewUnauthorizedError(err.Error())
		}

		return claims, err
	}

	return claims, nil
}

func (a *Authenticator) getKeyFunc(ctx context.Context) func(token *jwt.Token) (interface{}, error) {
	return func(token *jwt.Token) (interface{}, error) {
		a.mux.RLock()
		defer a.mux.RUnlock()

		unsupportedErr := apperrors.NewInternalError("unexpected signing method: %s", token.Method.Alg())

		switch token.Method.Alg() {
		case jwt.SigningMethodRS256.Name:
			keyID, err := a.getKeyID(*token)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while getting the token signing key ID: %v", err)
				return nil, errors.Wrap(err, "while getting the key ID")
			}

			if a.cachedJWKS == nil {
				log.C(ctx).Debugf("Empty JWKS cache... Signing key %s is not found", keyID)
				return nil, apperrors.NewKeyDoesNotExistError(keyID)

			}

			keyIterator := &authenticator.JWTKeyIterator{
				AlgorithmCriteria: func(alg string) bool {
					return token.Method.Alg() == alg
				},
				IDCriteria: func(id string) bool {
					return id == keyID
				},
			}

			if err := arrayiter.Walk(ctx, a.cachedJWKS, keyIterator); err != nil {
				log.C(ctx).WithError(err).Errorf("An error occurred while walking through the jwks: %v", err)
				return nil, err
			}

			if keyIterator.ResultingKey == nil {
				log.C(ctx).Debugf("Signing key %s is not found", keyID)
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

func (a *Authenticator) parseClaims(ctx context.Context, bearerToken string) (Claims, error) {
	claims := Claims{}

	_, err := jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc(ctx))
	if err != nil {
		return Claims{}, err
	}

	return claims, nil
}

func (a *Authenticator) SynchronizeJWKS(ctx context.Context) error {
	log.C(ctx).Info("Synchronizing JWKS...")
	a.mux.Lock()
	defer a.mux.Unlock()

	a.cachedJWKS = jwk.NewSet()

	for _, jwksEndpoint := range a.jwksEndpoints {
		log.C(ctx).Debugf("Fetching from endpoint: %s", jwksEndpoint)
		jwks, err := authenticator.FetchJWK(ctx, jwksEndpoint)
		if err != nil {
			return errors.Wrapf(err, "while fetching JWKS from endpoint %s", jwksEndpoint)
		}

		keyIterator := &authenticator.JWTKeyIterator{
			AlgorithmCriteria: func(alg string) bool {
				return true
			},
			IDCriteria: func(id string) bool {
				return true
			},
		}

		if err := arrayiter.Walk(ctx, jwks, keyIterator); err != nil {
			return errors.Wrapf(err, "while walking through JWKS")
		}

		for _, key := range keyIterator.AllKeys {
			key, ok := key.(jwk.Key)
			if !ok {
				return apperrors.NewInternalError("unable to parse jwk key")
			}

			a.cachedJWKS.Add(key)
		}
	}
	log.C(ctx).Info("Successfully synchronized JWKS")
	return nil
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

func stringsAnyEquals(stringSlice []string, str string) bool {
	for _, v := range stringSlice {
		if strings.Contains(str, v) {
			return true
		}
	}
	return false
}

func PrefixScopes(prefixes []string, callbackScope string) []string {
	prefixedScopes := make([]string, 0, len(prefixes))
	for _, scope := range prefixes {
		prefixedScopes = append(prefixedScopes, fmt.Sprintf("%s%s", scope, callbackScope))
	}
	return prefixedScopes
}
