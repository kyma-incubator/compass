package revocation

import (
	"github.com/kyma-incubator/compass/components/director/pkg/log"
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

	revokedCertsData, ok := data.(map[string]string)
	if !ok {
		log.D().Error("revocation cache did not have the expected config map type")
		return make(map[string]string)
	}

	return revokedCertsData
}
