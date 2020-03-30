package memory

import (
	"sync"

	"github.com/pivotal-cf/brokerapi/v7/domain"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dberr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/dbsession/dbmodel"
)

type operations struct {
	mu sync.Mutex

	provisioningOperations   map[string]internal.ProvisioningOperation
	deprovisioningOperations map[string]internal.DeprovisioningOperation
}

// NewOperation creates in-memory storage for OSB operations.
func NewOperation() *operations {
	return &operations{
		provisioningOperations:   make(map[string]internal.ProvisioningOperation, 0),
		deprovisioningOperations: make(map[string]internal.DeprovisioningOperation, 0),
	}
}

func (s *operations) InsertProvisioningOperation(operation internal.ProvisioningOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := operation.ID
	if _, exists := s.provisioningOperations[id]; exists {
		return dberr.AlreadyExists("instance operation with id %s already exist", id)
	}

	s.provisioningOperations[id] = operation
	return nil
}

func (s *operations) GetProvisioningOperationByID(operationID string) (*internal.ProvisioningOperation, error) {
	op, exists := s.provisioningOperations[operationID]
	if !exists {
		return nil, dberr.NotFound("instance provisioning operation with id %s not found", operationID)
	}
	return &op, nil
}

func (s *operations) GetProvisioningOperationByInstanceID(instanceID string) (*internal.ProvisioningOperation, error) {
	for _, op := range s.provisioningOperations {
		if op.InstanceID == instanceID {
			return &op, nil
		}
	}

	return nil, dberr.NotFound("instance provisioning operation with instanceID %s not found", instanceID)
}

func (s *operations) UpdateProvisioningOperation(op internal.ProvisioningOperation) (*internal.ProvisioningOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldOp, exists := s.provisioningOperations[op.ID]
	if !exists {
		return nil, dberr.NotFound("instance operation with id %s not found", op.ID)
	}
	if oldOp.Version != op.Version {
		return nil, dberr.Conflict("unable to update provisioning operation with id %s (for instance id %s) - conflict", op.ID, op.InstanceID)
	}
	op.Version = op.Version + 1
	s.provisioningOperations[op.ID] = op

	return &op, nil
}

func (s *operations) InsertDeprovisioningOperation(operation internal.DeprovisioningOperation) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := operation.ID
	if _, exists := s.deprovisioningOperations[id]; exists {
		return dberr.AlreadyExists("instance operation with id %s already exist", id)
	}

	s.deprovisioningOperations[id] = operation
	return nil
}

func (s *operations) GetDeprovisioningOperationByID(operationID string) (*internal.DeprovisioningOperation, error) {
	op, exists := s.deprovisioningOperations[operationID]
	if !exists {
		return nil, dberr.NotFound("instance deprovisioning operation with id %s not found", operationID)
	}
	return &op, nil
}

func (s *operations) GetDeprovisioningOperationByInstanceID(instanceID string) (*internal.DeprovisioningOperation, error) {
	for _, op := range s.deprovisioningOperations {
		if op.InstanceID == instanceID {
			return &op, nil
		}
	}

	return nil, dberr.NotFound("instance deprovisioning operation with instanceID %s not found", instanceID)
}

func (s *operations) UpdateDeprovisioningOperation(op internal.DeprovisioningOperation) (*internal.DeprovisioningOperation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldOp, exists := s.deprovisioningOperations[op.ID]
	if !exists {
		return nil, dberr.NotFound("instance operation with id %s not found", op.ID)
	}
	if oldOp.Version != op.Version {
		return nil, dberr.Conflict("unable to update deprovisioning operation with id %s (for instance id %s) - conflict", op.ID, op.InstanceID)
	}
	op.Version = op.Version + 1
	s.deprovisioningOperations[op.ID] = op

	return &op, nil
}

func (s *operations) GetOperationByID(operationID string) (*internal.Operation, error) {
	var res *internal.Operation

	provisionOp, exists := s.provisioningOperations[operationID]
	if exists {
		res = &provisionOp.Operation
	}
	deprovisionOp, exists := s.deprovisioningOperations[operationID]
	if exists {
		res = &deprovisionOp.Operation
	}
	if res == nil {
		return nil, dberr.NotFound("instance operation with id %s not found", operationID)
	}

	return res, nil
}

func (s *operations) GetOperationsInProgressByType(opType dbmodel.OperationType) ([]internal.Operation, error) {
	ops := make([]internal.Operation, 0)
	switch opType {
	case dbmodel.OperationTypeProvision:
		for _, op := range s.provisioningOperations {
			if op.State == domain.InProgress {
				ops = append(ops, op.Operation)
			}
		}
	case dbmodel.OperationTypeDeprovision:
		for _, op := range s.deprovisioningOperations {
			if op.State == domain.InProgress {
				ops = append(ops, op.Operation)
			}
		}
	}

	return ops, nil
}
