package certloader

import (
	"crypto/tls"
	"sync"
)

// Cache returns a client certificate stored in-memory
type Cache interface {
	Get() map[string]*tls.Certificate
}

type certificateCache struct {
	tlsCerts map[string]*tls.Certificate
	mutex    sync.RWMutex
}

// NewCertificateCache is responsible for in-memory managing of a TLS certificate
func NewCertificateCache() *certificateCache {
	return &certificateCache{
		tlsCerts: make(map[string]*tls.Certificate, 2),
	}
}

// Get returns a map of parsed TLS certificates
func (cc *certificateCache) Get() map[string]*tls.Certificate {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return cc.tlsCerts
}

func (cc *certificateCache) put(secretName string, tlsCert *tls.Certificate) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.tlsCerts[secretName] = tlsCert
}
