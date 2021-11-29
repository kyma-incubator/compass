package certloader

import (
	"crypto/tls"
	"sync"
)

// Cache returns a client certificate stored in-memory
type Cache interface {
	Get() *tls.Certificate
}

type certificatesCache struct {
	tlsCert *tls.Certificate
	mutex   sync.Mutex
}

// NewCertificateCache is responsible for in-memory managing of a TLS certificate
func NewCertificateCache() *certificatesCache {
	return &certificatesCache{}
}

// Get returns a parsed TLS certificate
func (cc *certificatesCache) Get() *tls.Certificate {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	return cc.tlsCert
}

func (cc *certificatesCache) put(tlsCert *tls.Certificate) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.tlsCert = tlsCert
}
