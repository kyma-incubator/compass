package operationsmanager

import (
	"context"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
	"sync"
)

// OperationType defines supported operation types
type OperationType string

const (
	// OrdAggregationOpType specifies open resource discovery operation type
	OrdAggregationOpType OperationType = "ORD_AGGREGATION"
)

// OperationsManager provides methods for operations management
type OperationsManager struct {
	opType   OperationType
	transact persistence.Transactioner
	opSvc    operationsmanager.OperationService
	mutex    sync.Mutex
}

// NewOperationsManager creates new OperationsManager
func NewOperationsManager(transact persistence.Transactioner, opSvc operationsmanager.OperationService, opType OperationType) *OperationsManager {
	return &OperationsManager{
		transact: transact,
		opSvc:    opSvc,
		opType:   opType,
	}
}

// GetOperation retrieves one scheduled operation
func (om *OperationsManager) GetOperation() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	//TODO implement me
}

// MarkOperationCompleted marks the operation with the given ID as completed
func (om *OperationsManager) MarkOperationCompleted(ctx context.Context, id string) error {
	tx, err := om.transact.Begin()
	if err != nil {
		return err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := om.opSvc.MarkAsCompleted(ctx, id); err != nil {
		return errors.Wrapf(err, "while marking operation with id %q as completed", id)
	}

	return tx.Commit()
}

// MarkOperationFailed marks the operation with the given ID as failed
func (om *OperationsManager) MarkOperationFailed(ctx context.Context, id, errorMsg string) error {
	tx, err := om.transact.Begin()
	if err != nil {
		return err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := om.opSvc.MarkAsFailed(ctx, id, errorMsg); err != nil {
		return errors.Wrapf(err, "while marking operation with id %q as failed", id)
	}
	return tx.Commit()
}
