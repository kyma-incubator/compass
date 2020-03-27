package memory

import (
	"sync"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
)

type lmsTenants struct {
	mu sync.Mutex

	data map[key]internal.LMSTenant
}

type key struct {
	Name   string
	Region string
}

func NewLMSTenants() *lmsTenants {
	return &lmsTenants{
		data: make(map[key]internal.LMSTenant, 0),
	}
}

func (s *lmsTenants) FindTenantByName(name, region string) (internal.LMSTenant, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := key{Name: name, Region: region}
	tenant, exists := s.data[k]
	if !exists {
		return internal.LMSTenant{}, false, nil
	}

	return tenant, true, nil
}

func (s *lmsTenants) InsertTenant(tenant internal.LMSTenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	k := key{Name: tenant.Name, Region: tenant.Region}
	if _, exists := s.data[k]; exists {
		return dberr.AlreadyExists("tenant already exists")
	}
	s.data[k] = tenant

	return nil
}
