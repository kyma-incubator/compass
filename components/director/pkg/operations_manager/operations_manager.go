package operationsmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

// OperationsManager provides methods for operations management
type OperationsManager struct {
	opType   model.OperationType
	transact persistence.Transactioner
	opSvc    operationsmanager.OperationService
	mutex    sync.Mutex
}

// NewOperationsManager creates new OperationsManager
func NewOperationsManager(transact persistence.Transactioner, opSvc operationsmanager.OperationService, opType model.OperationType) *OperationsManager {
	return &OperationsManager{
		transact: transact,
		opSvc:    opSvc,
		opType:   opType,
	}
}

// GetOperation retrieves one scheduled operation
func (om *OperationsManager) GetOperation(ctx context.Context, opType model.OperationType) (*model.Operation, error) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	operations, err := om.opSvc.ListPriorityQueue(ctx, opType)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching operations from priority queue with type %v ", opType)
	}

	for _, operation := range operations {
		tx, err := om.transact.Begin()
		if err != nil {
			return nil, err
		}
		defer om.transact.RollbackUnlessCommitted(ctx, tx)
		ctx = persistence.SaveToContext(ctx, tx)

		lock, err := om.opSvc.LockOperation(ctx, operation.ID)
		if err != nil {
			return nil, err
		}
		if lock {
			currentOperation, err := om.opSvc.Get(ctx, operation.ID)
			if err != nil {
				return nil, err
			}
			if currentOperation.Status == model.OperationStatusScheduled {
				currentOperation.Status = model.OperationStatusInProgress
				now := time.Now()
				currentOperation.UpdatedAt = &now
				err = om.opSvc.Update(ctx, currentOperation)
				if err != nil {
					return nil, err
				}
				return currentOperation, tx.Commit()
			}
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
	}
	return nil, apperrors.NewNoScheduledOperationsError()
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
