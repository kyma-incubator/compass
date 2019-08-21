package tokens

import (
	"time"

	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	"github.com/patrickmn/go-cache"
)

const (
	defaultTTLMinutes      = 5 * time.Minute
	defaultCleanupInterval = 1 * time.Minute
)

//go:generate mockery -name=Cache
type Cache interface {
	Put(token string, data TokenData)
	Get(token string) (TokenData, apperrors.AppError)
	Delete(token string)
}

type tokenCache struct {
	tokenCache          *cache.Cache
	applicationTokenTTL time.Duration
	runtimeTokenTTL     time.Duration
}

func NewTokenCache(applicationTokenTTL, runtimeTokenTTL time.Duration) Cache {
	return &tokenCache{
		tokenCache:          cache.New(defaultTTLMinutes, defaultCleanupInterval),
		applicationTokenTTL: applicationTokenTTL,
		runtimeTokenTTL:     runtimeTokenTTL,
	}
}

func (c *tokenCache) Put(token string, data TokenData) {
	var tokenTTL time.Duration

	switch data.Type {
	case RuntimeToken:
		tokenTTL = c.runtimeTokenTTL
	case ApplicationToken:
		tokenTTL = c.applicationTokenTTL
	}

	c.tokenCache.Set(token, data, tokenTTL)
}

func (c *tokenCache) Get(token string) (TokenData, apperrors.AppError) {
	data, found := c.tokenCache.Get(token)
	if !found {
		return TokenData{}, apperrors.NotFound("Token not found in the cache.")
	}

	tokenData, ok := data.(TokenData)
	if !ok {
		return TokenData{}, apperrors.Internal("Failed to get token from cache")
	}

	return tokenData, nil
}

func (c *tokenCache) Delete(token string) {
	c.tokenCache.Delete(token)
}
