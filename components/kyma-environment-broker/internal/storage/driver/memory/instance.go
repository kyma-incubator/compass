package memory

import (
	"database/sql"
	"sync"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/predicate"
)

type Instance struct {
	mu                sync.Mutex
	instances         map[string]internal.Instance
	operationsStorage *operations
}

func NewInstance(operations *operations) *Instance {
	return &Instance{
		instances:         make(map[string]internal.Instance, 0),
		operationsStorage: operations,
	}
}
func (s *Instance) FindAllJoinedWithOperations(prct ...predicate.Predicate) ([]internal.InstanceWithOperation, error) {
	var instances []internal.InstanceWithOperation
	// simulate left join without grouping on column
	for id, v := range s.instances {
		dOp, dErr := s.operationsStorage.GetDeprovisioningOperationByInstanceID(id)
		if dErr != nil && !dberr.IsNotFound(dErr) {
			return nil, dErr
		}
		pOp, pErr := s.operationsStorage.GetProvisioningOperationByInstanceID(id)
		if pErr != nil && !dberr.IsNotFound(pErr) {
			return nil, pErr
		}

		if !dberr.IsNotFound(dErr) {
			instances = append(instances, internal.InstanceWithOperation{
				Instance:    v,
				Type:        sql.NullString{String: string(dbmodel.OperationTypeDeprovision), Valid: true},
				State:       sql.NullString{String: string(dOp.State), Valid: true},
				Description: sql.NullString{String: dOp.Description, Valid: true},
			})
		}
		if !dberr.IsNotFound(pErr) {
			instances = append(instances, internal.InstanceWithOperation{
				Instance:    v,
				Type:        sql.NullString{String: string(dbmodel.OperationTypeProvision), Valid: true},
				State:       sql.NullString{String: string(pOp.State), Valid: true},
				Description: sql.NullString{String: pOp.Description, Valid: true},
			})
		}
		if dberr.IsNotFound(dErr) && dberr.IsNotFound(pErr) {
			instances = append(instances, internal.InstanceWithOperation{Instance: v})
		}
	}

	for _, p := range prct {
		p.ApplyToInMemory(instances)
	}

	return instances, nil
}

func (s *Instance) GetByID(instanceID string) (*internal.Instance, error) {
	inst, ok := s.instances[instanceID]
	if !ok {
		return nil, dberr.NotFound("instance with id %s not exist", instanceID)
	}

	return &inst, nil
}

func (s *Instance) Delete(instanceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.instances, instanceID)
	return nil
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
