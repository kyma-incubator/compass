package operation

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

// OperationRepository is responsible for repository-layer operation operations
//
//go:generate mockery --name=OperationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type OperationRepository interface {
	Create(ctx context.Context, model *model.Operation) error
	Get(ctx context.Context, id string) (*model.Operation, error)
	Update(ctx context.Context, model *model.Operation) error
	PriorityQueueListByType(ctx context.Context, queueLimit int, opType model.OperationType) ([]*model.Operation, error)
	LockOperation(ctx context.Context, operationID string) (bool, error)
	ResheduleOperations(ctx context.Context, reschedulePeriod time.Duration) error
	RescheduleHangedOperations(ctx context.Context, hangPeriod time.Duration) error
}

// UIDService is responsible for service-layer uid operations
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	opRepo     OperationRepository
	uidService UIDService
}

// NewService creates operations service
func NewService(opRepo OperationRepository, uidService UIDService) *service {
	return &service{
		opRepo:     opRepo,
		uidService: uidService,
	}
}

// Create creates new operation entity
func (s *service) Create(ctx context.Context, in *model.OperationInput) error {
	id := s.uidService.Generate()
	op := in.ToOperation(id)

	if err := s.opRepo.Create(ctx, op); err != nil {
		return errors.Wrapf(err, "error occurred while creating an Operation with id %s and type %s", op.ID, op.OpType)
	}

	log.C(ctx).Infof("Successfully created an Operation with id %s and type %s", op.ID, op.OpType)
	return nil
}

// CreateMultiple creates multiple operations
func (s *service) CreateMultiple(ctx context.Context, in []*model.OperationInput) error {
	if in == nil {
		return nil
	}

	for _, op := range in {
		if op == nil {
			continue
		}

		if err := s.Create(ctx, op); err != nil {
			return err
		}
	}

	return nil
}

// MarkAsCompleted marks an operation as completed
func (s *service) MarkAsCompleted(ctx context.Context, id string) error {
	op, err := s.opRepo.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting opreration with id %q", id)
	}

	op.Status = model.OperationStatusCompleted
	currentTime := time.Now()
	op.UpdatedAt = &currentTime

	if err := s.opRepo.Update(ctx, op); err != nil {
		return errors.Wrapf(err, "while updating operation with id %q", id)
	}
	return nil
}

// Update updates an operation in repository
func (s *service) Update(ctx context.Context, input *model.Operation) error {
	return s.opRepo.Update(ctx, input)
}

// MarkAsFailed marks an operation as failed
func (s *service) MarkAsFailed(ctx context.Context, id, errorMsg string) error {
	op, err := s.opRepo.Get(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting opreration with id %q", id)
	}

	currentTime := time.Now()
	opError := NewOperationError(errorMsg)
	rawMessage, err := opError.ToJSONRawMessage()
	if err != nil {
		return errors.Wrap(err, "while marshaling operation error")
	}

	op.Status = model.OperationStatusFailed
	op.UpdatedAt = &currentTime
	op.Error = rawMessage

	if err := s.opRepo.Update(ctx, op); err != nil {
		return errors.Wrapf(err, "while updating operation with id %q", id)
	}
	return nil
}

// ListPriorityQueue returns top 10 operations of specified type ordered by priority
func (s *service) ListPriorityQueue(ctx context.Context, queueLimit int, opType model.OperationType) ([]*model.Operation, error) {
	return s.opRepo.PriorityQueueListByType(ctx, queueLimit, opType)
}

// LockOperation try to acquire advisory lock on operation with provided ID
func (s *service) LockOperation(ctx context.Context, operationID string) (bool, error) {
	return s.opRepo.LockOperation(ctx, operationID)
}

// ResheduleOperations reschedules all old operations
func (s *service) ResheduleOperations(ctx context.Context, reschedulePeriod time.Duration) error {
	return s.opRepo.ResheduleOperations(ctx, reschedulePeriod)
}

// RescheduleHangedOperations reschedules all hanged operations
func (s *service) RescheduleHangedOperations(ctx context.Context, hangPeriod time.Duration) error {
	return s.opRepo.RescheduleHangedOperations(ctx, hangPeriod)
}

// Get loads operation with specified ID
func (s *service) Get(ctx context.Context, operationID string) (*model.Operation, error) {
	return s.opRepo.Get(ctx, operationID)
}

// OperationError represents an error from operation processing.
type OperationError struct {
	ErrorMsg string `json:"error"`
}

// NewOperationError creates OperationError instance.
func NewOperationError(errorMsg string) *OperationError {
	return &OperationError{ErrorMsg: errorMsg}
}

// ToJSONRawMessage converts the operation error ro JSON
func (or *OperationError) ToJSONRawMessage() (json.RawMessage, error) {
	jsonBytes, err := json.Marshal(or)
	if err != nil {
		return nil, err
	}

	return jsonBytes, nil
}
