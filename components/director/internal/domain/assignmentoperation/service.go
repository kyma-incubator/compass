package assignmentoperation

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// AssignmentOperationRepository represents the Assignment Operation repository layer
//
//go:generate mockery --name=AssignmentOperationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type AssignmentOperationRepository interface {
	Create(ctx context.Context, item *model.AssignmentOperation) error
	Update(ctx context.Context, m *model.AssignmentOperation) error
	DeleteByIDs(ctx context.Context, ids []string) error
	GetLatestOperation(ctx context.Context, formationAssignmentID, formationID string) (*model.AssignmentOperation, error)
	GetByID(ctx context.Context, id string) (*model.AssignmentOperation, error)
	ListForFormationAssignmentIDs(ctx context.Context, assignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error)
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo   AssignmentOperationRepository
	uidSvc UIDService
}

// NewService creates a service for managing Assignment Operations
func NewService(repo AssignmentOperationRepository, uidSvc UIDService) *service {
	return &service{
		repo:   repo,
		uidSvc: uidSvc,
	}
}

// Create creates a new Assignment Operation
func (s *service) Create(ctx context.Context, in *model.AssignmentOperationInput) (string, error) {
	assignmentOperationID := s.uidSvc.Generate()

	assignmentOp := in.ToModel(assignmentOperationID)
	now := time.Now()
	assignmentOp.StartedAtTimestamp = &now

	log.C(ctx).Infof("Creating assignment operation for formation assignment %s in formation %s with type %s, triggered by %s ", in.FormationAssignmentID, in.FormationID, in.Type, in.TriggeredBy)
	if err := s.repo.Create(ctx, assignmentOp); err != nil {
		return "", errors.Wrapf(err, "while creating assignment operation for formation assignment %s in formation %s with type %s, triggered by %s", in.FormationAssignmentID, in.FormationID, in.Type, in.TriggeredBy)
	}

	return assignmentOperationID, nil
}

// Finish finishes the Assignment Operation by the provided assignment and formation ID
func (s *service) Finish(ctx context.Context, assignmentID, formationID string) error {
	operation, err := s.repo.GetLatestOperation(ctx, assignmentID, formationID)
	if err != nil {
		return errors.Wrapf(err, "while getting the latest operation for assignment with ID: %s, formation with ID: %s", assignmentID, formationID)
	}
	return s.finishOperation(ctx, operation)
}

// FinishByID finishes the Assignment Operation by its provided ID
func (s *service) FinishByID(ctx context.Context, operationID string) error {
	operation, err := s.repo.GetByID(ctx, operationID)
	if err != nil {
		return errors.Wrapf(err, "while getting the operation with ID: %s", operationID)
	}
	return s.finishOperation(ctx, operation)
}

// GetLatestOperation gets the latest Operation for the provided Assignment and Formation ID
func (s *service) GetLatestOperation(ctx context.Context, assignmentID, formationID string) (*model.AssignmentOperation, error) {
	log.C(ctx).Infof("Getting the latest operation for assignment with ID: %s and formation with ID: %s", assignmentID, formationID)

	latestOperation, err := s.repo.GetLatestOperation(ctx, assignmentID, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting the latest operation for assignment with ID: %s, formation with ID: %s", assignmentID, formationID)
	}
	return latestOperation, nil
}

func (s *service) GetByID(ctx context.Context, operationID string) (*model.AssignmentOperation, error) {
	log.C(ctx).Infof("Getting the operation with ID: %s", operationID)

	operation, err := s.repo.GetByID(ctx, operationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting operation with ID: %s", operationID)
	}

	return operation, nil
}

// Update updates the Assignment Operation's triggered_by field with the new trigger and sets the started_at timestamp to the current time
func (s *service) Update(ctx context.Context, assignmentID, formationID string, newTrigger model.OperationTrigger) error {
	operation, err := s.repo.GetLatestOperation(ctx, assignmentID, formationID)
	if err != nil {
		return errors.Wrapf(err, "while getting the latest operation for assignment with ID: %s, formation with ID: %s", assignmentID, formationID)
	}

	log.C(ctx).Debugf("Updating the finished at timestamp for assignment operation with ID: %s", operation.ID)
	now := time.Now()
	operation.StartedAtTimestamp = &now
	operation.TriggeredBy = newTrigger

	err = s.repo.Update(ctx, operation)
	if err != nil {
		return errors.Wrapf(err, "while updating the finished at timestamp for assignment operation with ID: %s", operation.ID)
	}

	return nil
}

// ListByFormationAssignmentIDs fetches the Assignment Operations for the provided Formation Assignment IDs
func (s *service) ListByFormationAssignmentIDs(ctx context.Context, formationAssignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error) {
	log.C(ctx).Infof("Listing assignment operations for formation assignments with IDs: %q", formationAssignmentIDs)

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}
	return s.repo.ListForFormationAssignmentIDs(ctx, formationAssignmentIDs, pageSize, cursor)
}

// DeleteByIDs deletes Assignment Operations by their IDs
func (s *service) DeleteByIDs(ctx context.Context, ids []string) error {
	log.C(ctx).Infof("Deleting assignment operations with IDs: %q", ids)

	return s.repo.DeleteByIDs(ctx, ids)
}

func (s *service) finishOperation(ctx context.Context, operation *model.AssignmentOperation) error {
	log.C(ctx).Debugf("Updating the finished at timestamp for assignment operation with ID: %s", operation.ID)
	now := time.Now()
	operation.FinishedAtTimestamp = &now

	err := s.repo.Update(ctx, operation)
	if err != nil {
		return errors.Wrapf(err, "while updating the finished at timestamp for assignment operation with ID: %s", operation.ID)
	}

	return nil
}
