package runtimemapping

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/form3tech-oss/jwt-go"
	"github.com/pkg/errors"
)

type JwkCacheEntry struct {
	Key      interface{}
	CachedAt time.Time
	ExpireAt time.Time
}

// IsExpired missing godoc
func (e JwkCacheEntry) IsExpired() bool {
	return time.Now().After(e.ExpireAt)
}

type jwksCache struct {
	fetch     KeyGetter
	expPeriod time.Duration
	cache     map[string]JwkCacheEntry
	flag      sync.RWMutex
}

// NewJWKsCache missing godoc
func NewJWKsCache(fetch KeyGetter, expPeriod time.Duration) *jwksCache {
	return &jwksCache{
		fetch:     fetch,
		expPeriod: expPeriod,
		cache:     make(map[string]JwkCacheEntry),
	}
}

func (c *jwksCache) GetSize() int {
	return len(c.cache)
}

// GetKey missing godoc
func (c *jwksCache) GetKey(ctx context.Context, token *jwt.Token) (interface{}, error) {
	if token == nil {
		return nil, apperrors.NewUnauthorizedError("token cannot be nil")
	}

	keyID, err := getKeyID(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the key ID")
	}

	c.flag.RLock()
	cachedKey, exists := c.cache[keyID]
	c.flag.RUnlock()

	if !exists || cachedKey.IsExpired() {
		key, err := c.fetch.GetKey(ctx, token)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting the key with ID [kid=%s]", keyID)
		}

		log.C(ctx).Info(fmt.Sprintf("Adding key %s to cache", keyID))

		c.flag.Lock()
		c.cache[keyID] = JwkCacheEntry{
			Key:      key,
			CachedAt: time.Now(),
			ExpireAt: time.Now().Add(c.expPeriod),
		}
		c.flag.Unlock()

		return key, nil
	}

	log.C(ctx).Info(fmt.Sprintf("Using key %s from cache", keyID))

	return cachedKey.Key, nil
}

// Cleanup missing godoc
func (c *jwksCache) Cleanup(ctx context.Context) {
	expiredKeys := make([]string, 0, len(c.cache))
	c.flag.RLock()
	for keyID := range c.cache {
		if !c.cache[keyID].IsExpired() {
			continue
		}
		expiredKeys = append(expiredKeys, keyID)
	}
	c.flag.RUnlock()

	if len(expiredKeys) == 0 {
		return
	}

	c.flag.Lock()
	for _, keyID := range expiredKeys {
		log.C(ctx).Info(fmt.Sprintf("Removing key %s from cache", keyID))

		delete(c.cache, keyID)
	}
	c.flag.Unlock()
}
