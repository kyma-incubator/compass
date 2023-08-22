package operationsmanager

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const (
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
