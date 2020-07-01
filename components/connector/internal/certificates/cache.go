package certificates

import (
	"github.com/kyma-incubator/compass/components/connector/internal/apperrors"
	"github.com/patrickmn/go-cache"
)

type Cache interface {
	Put(certName string, data map[string][]byte)
	Get(certName string) (map[string][]byte, apperrors.AppError)
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
func (cc *certificatesCache) Get(certName string) (map[string][]byte, apperrors.AppError) {
	data, found := cc.cache.Get(certName)
	if !found {
		return nil, apperrors.NotFound("Certificate data not found in the cache.")
	}

	certData, ok := data.(map[string][]byte)
	if !ok {
		return nil, apperrors.Internal("Failed to get certificate data from cache")
	}

	return certData, nil
}
