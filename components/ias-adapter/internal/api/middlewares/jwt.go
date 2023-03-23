package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/jwk"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

const (
	authorizationHeader = "Authorization"
	tenantCtxKey        = "tenant"
	keyIDHeader         = "kid"
)

type JWTMiddleware struct {
	cache jwk.Cache
}

func NewJWTMiddleware(cache jwk.Cache) JWTMiddleware {
	return JWTMiddleware{
		cache: cache,
	}
}

func (m JWTMiddleware) JWT(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	bearerToken, err := getBearerToken(ctx.Request)
	if err != nil {
		log.Err(err).Msg("Failed to get token from Authorization header")
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	jwtClaims, err := m.verifyJWT(ctx, bearerToken)
	if err != nil {
		log.Err(err).Msg("Failed to verify Authorization header token")
		internal.RespondWithError(ctx, http.StatusUnauthorized, err)
		return
	}

	tenant, err := jwtClaims.extractTenant()
	if err != nil {
		log.Err(err).Msg("Failed to extract tenant from claims")
		internal.RespondWithError(ctx, http.StatusUnauthorized, err)
		return
	}
	ctx.Set(tenantCtxKey, tenant)

	ctx.Next()
}

func getBearerToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get(authorizationHeader)
	if reqToken == "" {
		return "", errors.Newf("Authorization header is empty: %s", errors.InvalidAccessToken)
	}
	return strings.TrimPrefix(reqToken, "Bearer "), nil
}

type jwtClaims struct {
	Tenants string `json:"tenant"`
	jwt.Claims
}

type tenant struct {
	ProviderExternalTenant string
}

func (c jwtClaims) extractTenant() (string, error) {
	tenant := tenant{}
	if err := json.Unmarshal([]byte(c.Tenants), &tenant); err != nil {
		return "", errors.Newf("failed to unmarshal tenant claim: %w", err)
	}
	return tenant.ProviderExternalTenant, nil
}

func (m JWTMiddleware) verifyJWT(ctx context.Context, jwtToken string) (jwtClaims, error) {
	claims := jwtClaims{}
	// if _, _, err := jwt.NewParser().ParseUnverified(jwtToken, &claims); err != nil {
	// 	return jwtClaims{}, errors.Newf("failed to parse token: %s: %w", err, errors.InvalidAccessToken)
	// }
	token, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (any, error) {
		keyID, ok := token.Header[keyIDHeader]
		if !ok {
			return []byte{}, errors.Newf("jwt header %s not found in token", keyIDHeader)
		}

		kid, ok := keyID.(string)
		if !ok {
			return []byte{}, errors.Newf("failed to cast jwt header %s of type %T to string", keyIDHeader, keyID)
		}

		return m.cache.Get(ctx, kid)
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return jwtClaims{}, errors.Newf("invalid token signature %s: %w", err, errors.InvalidAccessToken)
		}
		return jwtClaims{}, errors.Newf("failed to validate token signature %s: %w", err, errors.InvalidAccessToken)
	}
	if token.Method.Alg() != jwt.SigningMethodRS256.Alg() {
		return jwtClaims{}, errors.Newf("invalid signing method: %w", errors.InvalidAccessToken)
	}
	if !token.Valid {
		return jwtClaims{}, errors.Newf("%s: %w", err, errors.InvalidAccessToken)
	}

	return claims, nil
}
