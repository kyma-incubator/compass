package authenticator

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
)

const (
	SubscriptionCallbacksScope = "Callback"
	AuthorizationHeaderKey     = "Authorization"
)

type Error struct {
	Message string `json:"message"`
}

type Authenticator struct {
	mux                  sync.Mutex
	cachedJWKs           *jwk.Set
	jwksEndpoint         string
	zoneId               string
	trustedClaimPrefixes []string
}

func New(jwksEndpoint, zoneId string, trustedClaimPrefixes []string) *Authenticator {
	return &Authenticator{
		jwksEndpoint:         jwksEndpoint,
		zoneId:               zoneId,
		trustedClaimPrefixes: trustedClaimPrefixes,
	}
}

func (a *Authenticator) Handler() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			token, err := a.getBearerToken(r)
			if err != nil {
				a.writeAppError(ctx, w, err, http.StatusBadRequest)
			}

			claims, err := a.parseClaimsWithRetry(r.Context(), token)
			if err != nil {
				log.C(ctx).WithError(err).Error("An error has occurred while parsing claims. Error code: ", http.StatusUnauthorized)
				a.writeAppError(ctx, w, err, http.StatusUnauthorized)
				return
			}

			if claims.ZID != a.zoneId {
				log.C(ctx).Errorf(`Zone id "%s" from user token does not match the trusted zone %s`, claims.ZID, a.zoneId)
				a.writeAppError(ctx, w, errors.Errorf(`Zone id "%s" from user token is not trusted`, claims.ZID), http.StatusUnauthorized)
				return
			}

			scopes := PrefixScopes(a.trustedClaimPrefixes, SubscriptionCallbacksScope)
			if !stringsAnyEquals(scopes, strings.Join(claims.Scopes, " ")) {
				log.C(ctx).Errorf(`Scope "%s" from user token does not match the trusted scopes`, claims.Scopes)
				a.writeAppError(ctx, w, errors.Errorf(`Scope "%s" is not trusted`, claims.Scopes), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
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
		log.C(ctx).WithError(err).Error("An error occurred while encoding data.")
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

	// TODO: adjust this productive code
	//_, err := jwt.ParseWithClaims(bearerToken, &claims, a.getKeyFunc())
	//if err != nil {
	//	return Claims{}, err
	//}

	// TODO: remove testing version
	p := jwt.Parser{SkipClaimsValidation: true}
	if _, _, err := p.ParseUnverified(bearerToken, &claims); err != nil {
		return Claims{}, err
	}

	return claims, nil
}

func (a *Authenticator) SynchronizeJWKS(ctx context.Context) error {
	log.C(ctx).Info("Synchronizing JWKS...")
	a.mux.Lock()
	defer a.mux.Unlock()
	jwks, err := authenticator.FetchJWK(ctx, a.jwksEndpoint)
	if err != nil {
		return errors.Wrapf(err, "while fetching JWKS from endpoint %s", a.jwksEndpoint)
	}

	a.cachedJWKs = jwks
	return nil
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
