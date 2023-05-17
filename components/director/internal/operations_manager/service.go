package operationsmanager

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	completedOpStatus = "COMPLETED"
	failedOpStatus    = "FAILED"
	scheduledOpStatus = "SCHEDULED"
)

// Service consists of various resource services responsible for service-layer Operations.
type Service struct {
	transact persistence.Transactioner

	opSvc        OperationService
	ordOpCreator OperationCreator
}

// NewOperationService returns a new object responsible for service-layer Operation operations.
func NewOperationService(transact persistence.Transactioner, opSvc OperationService, ordOpCreator OperationCreator) *Service {
	return &Service{
		transact:     transact,
		opSvc:        opSvc,
		ordOpCreator: ordOpCreator,
	}
}

// CreateORDOperations creates ord operations
func (s *Service) CreateORDOperations(ctx context.Context) error {
	return s.ordOpCreator.Create(ctx)
}

// DeleteOldOperations deletes all operations of type `opType` which are:
// - in status COMPLETED and older than `completedOpDays`
// - in status FAILED and older than `failedOpDays`
func (s *Service) DeleteOldOperations(ctx context.Context, opType string, completedOpDays, failedOpDays int) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.opSvc.DeleteOlderThan(ctx, opType, completedOpStatus, completedOpDays); err != nil {
		return errors.Wrap(err, "while deleting completed operations")
	}

	if err := s.opSvc.DeleteOlderThan(ctx, opType, failedOpStatus, failedOpDays); err != nil {
		return errors.Wrap(err, "while deleting failed operations")
	}

	return tx.Commit()
}
