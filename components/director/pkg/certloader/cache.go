package certloader

import (
    "errors"
    "github.com/kyma-incubator/compass/components/director/pkg/log"
    "github.com/patrickmn/go-cache"
)

type Cache interface {
    Put(certName string, data map[string][]byte)
    Get(certName string) (map[string][]byte, error)
}

type certificatesCache struct {
    cache *cache.Cache
}

func NewCertificateCache() Cache {
    return &certificatesCache{
        cache: cache.New(0, 0),
    }
}

func (cc *certificatesCache) Put(certName string, data map[string][]byte) {
    cc.cache.Set(certName, data, 0)
}

func (cc *certificatesCache) Get(certName string) (map[string][]byte, error) {
    log.D().Infof("Getting certificate data from secret with name: %s", certName)
    data, found := cc.cache.Get(certName)
    if !found {
        log.D().Errorf("Certificate data not found in the cache for secret with name: %s", certName)
        return nil, errors.New("certificate data not found in the cache")
    }

    certData, ok := data.(map[string][]byte)
    if !ok {
        log.D().Errorf("Failed to get certificate data from cache for secret with name: %s", certName)
        return nil, errors.New("failed to get certificate data from cache")
    }

    return certData, nil
}
