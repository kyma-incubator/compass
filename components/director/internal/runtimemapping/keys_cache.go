package runtimemapping

import (
	"fmt"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	logger    *logrus.Logger
	fetch     KeyGetter
	expPeriod time.Duration
	cache     map[string]jwkCacheEntry
	flag      sync.RWMutex
}

func NewJWKsCache(logger *logrus.Logger, fetch KeyGetter, expPeriod time.Duration) *jwksCache {
	return &jwksCache{
		logger:    logger,
		fetch:     fetch,
		expPeriod: expPeriod,
		cache:     make(map[string]jwkCacheEntry),
	}
}

func (c *jwksCache) GetKey(token *jwt.Token) (interface{}, error) {
	if token == nil {
		return nil, errors.New("token cannot be nil")
	}

	keyID, err := getKeyID(*token)
	if err != nil {
		return nil, errors.Wrap(err, "while getting the key ID")
	}

	c.flag.RLock()
	cachedKey, exists := c.cache[keyID]
	c.flag.RUnlock()

	if !exists || cachedKey.IsExpired() {
		key, err := c.fetch.GetKey(token)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting the key with ID [kid=%s]", keyID)
		}

		c.logger.Info(fmt.Sprintf("adding key %s to cache", keyID))

		c.flag.Lock()
		c.cache[keyID] = jwkCacheEntry{
			key:      key,
			cachedAt: time.Now(),
			expireAt: time.Now().Add(c.expPeriod),
		}
		c.flag.Unlock()

		return key, nil
	}

	c.logger.Info(fmt.Sprintf("using key %s from cache", keyID))

	return cachedKey.key, nil
}

func (c *jwksCache) Cleanup() {
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
		c.logger.Info(fmt.Sprintf("removing key %s from cache", keyID))

		delete(c.cache, keyID)
	}
	c.flag.Unlock()
}
