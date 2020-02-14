package memory

import (
	"sync"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/pkg/errors"
)

type Instance struct {
	mu        sync.Mutex
	instances map[string]internal.Instance
}

func NewInstance() *Instance {
	return &Instance{
		instances: make(map[string]internal.Instance, 0),
	}
}

func (s *Instance) GetByID(instanceID string) (*internal.Instance, error) {
	inst, ok := s.instances[instanceID]
	if !ok {
		return nil, errors.Errorf("instance with id %s not exist", instanceID)
	}

	return &inst, nil
}

func (s *Instance) Insert(instance internal.Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[instance.InstanceID] = instance

	return nil
}

func (s *Instance) Update(instance internal.Instance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.instances[instance.InstanceID] = instance

	return nil
}
