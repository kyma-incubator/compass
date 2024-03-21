package certsubjectmapping

import (
	"sync"

	csm "github.com/kyma-incubator/compass/components/hydrator/pkg/certsubjmapping"
)

// Cache is responsible for the read and write operation of the cert subject mapping in-memory storage
//
//go:generate mockery --name=Cache --output=automock --outpkg=automock --case=underscore --disable-version-string
type Cache interface {
	Get() []csm.SubjectConsumerTypeMapping
	Put(certSubjectMappings []csm.SubjectConsumerTypeMapping)
}

type certSubjectMappingCache struct {
	mutex    sync.RWMutex
	mappings []csm.SubjectConsumerTypeMapping
}

// NewCertSubjectMappingCache creates a new certificate subject mapping cache
func NewCertSubjectMappingCache() *certSubjectMappingCache {
	return &certSubjectMappingCache{mappings: []csm.SubjectConsumerTypeMapping{}}
}

// Get returns a slice of SubjectConsumerTypeMapping
func (cc *certSubjectMappingCache) Get() []csm.SubjectConsumerTypeMapping {
	cc.mutex.RLock()
	defer cc.mutex.RUnlock()

	return cc.mappings
}

// Put updates the cache with the given slice of SubjectConsumerTypeMapping
func (cc *certSubjectMappingCache) Put(mappings []csm.SubjectConsumerTypeMapping) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()

	cc.mappings = mappings
}
