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

type jwkCacheEntry struct {
	key      interface{}
	cachedAt time.Time
	expireAt time.Time
}

func (e jwkCacheEntry) IsExpired() bool {
	return time.Now().After(e.expireAt)
}

type jwksCache struct {
	fetch     KeyGetter
	expPeriod time.Duration
	cache     map[string]jwkCacheEntry
	flag      sync.RWMutex
}

func NewJWKsCache(fetch KeyGetter, expPeriod time.Duration) *jwksCache {
	return &jwksCache{
		fetch:     fetch,
		expPeriod: expPeriod,
		cache:     make(map[string]jwkCacheEntry),
	}
}

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
		c.cache[keyID] = jwkCacheEntry{
			key:      key,
			cachedAt: time.Now(),
			expireAt: time.Now().Add(c.expPeriod),
		}
		c.flag.Unlock()

		return key, nil
	}

	log.C(ctx).Info(fmt.Sprintf("Using key %s from cache", keyID))

	return cachedKey.key, nil
}

func (c *jwksCache) Cleanup(ctx context.Context) {
	var expiredKeys []string
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
