package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/jwk"

	"github.com/kyma-incubator/compass/components/ias-adapter/internal/api/internal"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/config"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/errors"
	"github.com/kyma-incubator/compass/components/ias-adapter/internal/logger"
)

const (
	authorizationHeader = "Authorization"
	tenantCtxKey        = "tenant"
)

func JWT(ctx *gin.Context) {
	log := logger.FromContext(ctx)

	bearerToken, err := getBearerToken(ctx.Request)
	if err != nil {
		log.Err(err).Msg("Failed to get token from Authorization header")
		internal.RespondWithError(ctx, http.StatusBadRequest, err)
		return
	}

	// get key

	jwtClaims, err := verifyJWT(bearerToken, "???")
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

func verifyJWT(jwtToken, key string) (jwtClaims, error) {
	claims := jwtClaims{}
	// if _, _, err := jwt.NewParser().ParseUnverified(jwtToken, &claims); err != nil {
	// 	return jwtClaims{}, errors.Newf("failed to parse token: %s: %w", err, errors.InvalidAccessToken)
	// }
	token, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (any, error) {
		return []byte(key), nil
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

type jwkCache struct {
	jwkEndpoint  string
	syncInterval time.Duration
	lastSyncTime time.Time
	keys         map[string]jwk.Key
	mutex        sync.RWMutex
	client       *http.Client
}

func newJWKCache(ctx context.Context, cfg config.JWKCache) *jwkCache {
	cache := &jwkCache{
		jwkEndpoint:  cfg.Endpoint,
		keys:         make(map[string]jwk.Key),
		syncInterval: cfg.SyncInterval,
	}
	ticker := time.NewTicker(cache.syncInterval)
	go func() {
		for {
			select {
			case <-ticker.C:
				cache.sync(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	return cache
}

func (c *jwkCache) sync(ctx context.Context) {
	if time.Since(c.lastSyncTime) > c.syncInterval {

	}
}

func (c *jwkCache) getKey(ctx context.Context, id string) (jwk.Key, error) {
	key, exists := c.keys[id]
	if !exists {
		return key, errors.New("jwk not found")
	}
	if key.Algorithm() != jwt.SigningMethodRS256.Alg() {
		return key, errors.Newf("jwk doesn't match expected algorithm '%s'", jwt.SigningMethodRS256.Alg())
	}
	return key, nil
}
