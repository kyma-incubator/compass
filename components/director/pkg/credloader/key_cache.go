package credloader

import (
	"crypto/rsa"
	"sync"
)

// KeysCache missing godoc
//
//go:generate mockery --name=KeysCache --output=automock --outpkg=automock --case=underscore --disable-version-string
type KeysCache interface {
	Get() map[string]*KeyStore
}

type secretData struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// KeyStore is an object that keeps track of a public/private key
type KeyStore struct {
	PublicKey  *rsa.PublicKey
	PrivateKey interface{}
}

// KeyCache is a mutex secured KeyStore
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

// NewKeyCacheWithKeys is responsible for in-memory managing of a TLS certificate
func NewKeyCacheWithKeys(keys map[string]*KeyStore) *KeyCache {
	return &KeyCache{
		keys: keys,
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
