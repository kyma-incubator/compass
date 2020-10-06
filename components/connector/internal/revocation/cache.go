package revocation

import (
	"github.com/patrickmn/go-cache"
)

const key string = "revocation-list"

type Cache interface {
	Put(data map[string]string)
	Get() map[string]string
}

type revocationCache struct {
	cache *cache.Cache
}

func NewCache() Cache {
	return &revocationCache{
		cache: cache.New(0, 0),
	}
}

func (cc *revocationCache) Put(data map[string]string) {
	cc.cache.Set(key, data, 0)
}
func (cc *revocationCache) Get() map[string]string {
	data, found := cc.cache.Get(key)
	if !found {
		return make(map[string]string)
	}

	revocationListData, ok := data.(map[string]string)
	if !ok {
		return make(map[string]string)
	}

	return revocationListData
}
