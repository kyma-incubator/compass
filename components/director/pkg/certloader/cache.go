package certloader

import (
	"crypto/tls"
	"sync"
)

// Cache returns a client certificate stored in-memory
type Cache interface {
	Get() []*tls.Certificate
}

type certificateCache struct {
	tlsCerts []*tls.Certificate
	mutex    sync.RWMutex
}

// NewCertificateCache is responsible for in-memory managing of a TLS certificate
func NewCertificateCache() *certificateCache {
	return &certificateCache{
		tlsCerts: make([]*tls.Certificate, 2),
	}
}

// Get returns an array of parsed TLS certificates
func (cc *certificateCache) Get() []*tls.Certificate {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return cc.tlsCerts
}

func (cc *certificateCache) put(idx int, tlsCert *tls.Certificate) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.tlsCerts[idx] = tlsCert
}
