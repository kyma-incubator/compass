package certloader

import (
	"crypto/rsa"
	"sync"
)

// KeysCache missing godoc
//
//go:generate mockery --name=CertificateCache --output=automock --outpkg=automock --case=underscore --disable-version-string
type KeysCache interface {
	Get() map[string]*KeyStore
}

type secretData struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

type KeyStore struct {
	PublicKey  *rsa.PublicKey
	PrivateKey interface{}
}

type KeyCache struct {
	keys  map[string]*KeyStore
	mutex sync.RWMutex
}

// NewKeyCache is responsible for in-memory managing of a TLS certificate
func NewKeyCache() *KeyCache {
	return &KeyCache{
		keys: make(map[string]*KeyStore, 2),
	}
}

// Get returns a map of parsed TLS certificates
func (cc *KeyCache) Get() map[string]*KeyStore {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return cc.keys
}

func (cc *KeyCache) put(secretName string, keys *KeyStore) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.keys[secretName] = keys
}
