package authnmappinghandler

import (
	"sync"
	"time"
)

// Entry represents a single cache Entry
type Entry struct {
	TokenData TokenData
	Taken     time.Time
}

// TokenDataCache is a local short-lived cache
type TokenDataCache struct {
	mutex sync.Mutex
	ttl   time.Duration
	data  map[string]map[string]Entry
}

func NewTokenDataCache(validity time.Duration) TokenDataCache {
	return TokenDataCache{
		ttl:  validity,
		data: make(map[string]map[string]Entry),
	}
}

func (c *TokenDataCache) Cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	validityStart := time.Now().Add((-1) * c.ttl)
	cleanCache := make(map[string]map[string]Entry)
	for token, landscapesMap := range c.data {
		cleanLandscapesMap := make(map[string]Entry)
		for landscape, entry := range landscapesMap {
			if entry.Taken.After(validityStart) {
				cleanLandscapesMap[landscape] = entry
			}
		}
		cleanCache[token] = cleanLandscapesMap
	}
	c.data = cleanCache
}

func (c *TokenDataCache) GetTokenData(token, issuerURL string) (bool, TokenData) {
	landscapesMap, landscapesMapExists := c.data[token]
	if landscapesMapExists {
		entry, entryExists := landscapesMap[issuerURL]
		if entryExists {
			return true, entry.TokenData
		}
	}
	return false, nil
}

func (c *TokenDataCache) PutTokenData(token, issuerURL string, tokenData TokenData) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, landscapesMapExists := c.data[token]
	if !landscapesMapExists {
		landscapesMap := make(map[string]Entry)
		c.data[token] = landscapesMap
	}
	landscapesMap := c.data[token]
	landscapesMap[issuerURL] = Entry{
		TokenData: tokenData,
		Taken:     time.Now(),
	}
}
