package operationsmanager

import (
	"context"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/cronjob"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type operationPriority int

var now = time.Now

const (
	// LowOperationPriority represents low priority for operations
	LowOperationPriority operationPriority = 1
	// HighOperationPriority represents high priority for operations
	HighOperationPriority operationPriority = 100
)

// OperationsManager provides methods for operations management
type OperationsManager struct {
	opType        model.OperationType
	transact      persistence.Transactioner
	opSvc         operationsmanager.OperationService
	mutex         sync.Mutex
	areJobsStared bool
	cfg           OperationsManagerConfig
}

// NewOperationsManager creates new OperationsManager
func NewOperationsManager(transact persistence.Transactioner, opSvc operationsmanager.OperationService, opType model.OperationType, cfg OperationsManagerConfig) *OperationsManager {
	return &OperationsManager{
		transact: transact,
		opSvc:    opSvc,
		opType:   opType,
		cfg:      cfg,
	}
}

// GetOperation retrieves one scheduled operation
func (om *OperationsManager) GetOperation(ctx context.Context) (*model.Operation, error) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	tx, err := om.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	operations, err := om.opSvc.ListPriorityQueue(ctx, om.cfg.PriorityQueueLimit, om.opType)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching operations from priority queue with type %v ", om.opType)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	for _, operation := range operations {
		op, err := om.tryToGetOperation(ctx, operation.ID)
		if err != nil {
			return nil, err
		}

		if op != nil {
			return op, nil
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

// RescheduleOperation reschedules operation with high priority
func (om *OperationsManager) RescheduleOperation(ctx context.Context, operationID string) error {
	return om.rescheduleOperation(ctx, operationID, HighOperationPriority)
}

// RunMaintenanceJobs runs the maintenance jobs. Should be mandatory during startup of corresponding module.
func (om *OperationsManager) RunMaintenanceJobs(ctx context.Context) error {
	if om.areJobsStared {
		log.C(ctx).Info("Maintenance jobs are already started")
		return nil
	}
	log.C(ctx).Info("Maintenance jobs starting")

	err := om.startRescheduleOperationsJob(ctx)
	if err != nil {
		return err
	}

	err = om.startRescheduleHangedOperationsJob(ctx)
	if err != nil {
		return err
	}
	om.areJobsStared = true
	log.C(ctx).Info("Maintenance jobs now started")
	return nil
}

func (om *OperationsManager) startRescheduleOperationsJob(ctx context.Context) error {
	resyncJob := cronjob.CronJob{
		Name: "RescheduleOperationsJob",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Info("Starting RescheduleOperationsJob...")

			tx, err := om.transact.Begin()
			if err != nil {
				log.C(jobCtx).Errorf("Error during opening transaction in RescheduleOperationsJob %v", err)
			}
			defer om.transact.RollbackUnlessCommitted(ctx, tx)
			ctx = persistence.SaveToContext(ctx, tx)

			if err := om.opSvc.RescheduleOperations(ctx, om.opType, om.cfg.OperationReschedulePeriod); err != nil {
				log.C(jobCtx).Errorf("Error during execution of RescheduleOperationsJob %v", err)
			}
			err = tx.Commit()
			if err != nil {
				log.C(jobCtx).Errorf("Error during committing transaction at RescheduleOperationsJob %v", err)
			}

			log.C(jobCtx).Infof("RescheduleOperationsJob finished.")
		},
		SchedulePeriod: om.cfg.RescheduleOperationsJobInterval,
	}
	return cronjob.RunCronJob(ctx, om.cfg.ElectionConfig, resyncJob)
}

func (om *OperationsManager) startRescheduleHangedOperationsJob(ctx context.Context) error {
	resyncJob := cronjob.CronJob{
		Name: "RescheduleHangedOperationsJob",
		Fn: func(jobCtx context.Context) {
			log.C(jobCtx).Info("Starting RescheduleHangedOperationsJob...")

			tx, err := om.transact.Begin()
			if err != nil {
				log.C(jobCtx).Errorf("Error during opening transaction in RescheduleHangedOperationsJob %v", err)
			}
			defer om.transact.RollbackUnlessCommitted(ctx, tx)
			ctx = persistence.SaveToContext(ctx, tx)

			if err := om.opSvc.RescheduleHangedOperations(ctx, om.opType, om.cfg.OperationHangPeriod); err != nil {
				log.C(jobCtx).Errorf("Error during execution of RescheduleHangedOperationsJob %v", err)
			}
			err = tx.Commit()
			if err != nil {
				log.C(jobCtx).Errorf("Error during commititng transaction at RescheduleHangedOperationsJob %v", err)
			}

			log.C(jobCtx).Infof("RescheduleHangedOperationsJob finished.")
		},
		SchedulePeriod: om.cfg.RescheduleHangedOperationsJobInterval,
	}
	return cronjob.RunCronJob(ctx, om.cfg.ElectionConfig, resyncJob)
}

func (om *OperationsManager) rescheduleOperation(ctx context.Context, operationID string, priority operationPriority) error {
	tx, err := om.transact.Begin()
	if err != nil {
		return err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := om.opSvc.RescheduleOperation(ctx, operationID, int(priority)); err != nil {
		return err
	}
	return tx.Commit()
}

func (om *OperationsManager) tryToGetOperation(ctx context.Context, operationID string) (*model.Operation, error) {
	tx, err := om.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer om.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	lock, err := om.opSvc.LockOperation(ctx, operationID)
	if err != nil {
		return nil, err
	}
	if lock {
		currentOperation, err := om.opSvc.Get(ctx, operationID)
		if err != nil {
			return nil, err
		}
		if currentOperation.Status == model.OperationStatusScheduled {
			currentOperation.Status = model.OperationStatusInProgress
			currentTime := now()
			currentOperation.UpdatedAt = &currentTime
			err = om.opSvc.Update(ctx, currentOperation)
			if err != nil {
				return nil, err
			}
			return currentOperation, tx.Commit()
		}
	}
	return nil, tx.Commit()
}
