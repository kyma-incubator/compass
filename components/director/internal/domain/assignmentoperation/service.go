package formationassignment

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

// AssignmentOperationRepository represents the Assignment Operation repository layer
//
//go:generate mockery --name=AssignmentOperationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type AssignmentOperationRepository interface {
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

// NewService creates
func NewService(repo AssignmentOperationRepository, uidSvc UIDService) *service {
	return &service{
		repo:   repo,
		uidSvc: uidSvc,
	}
}

func (s *service) Create(ctx context.Context, assignmentID string, operationType model.AssignmentOperationType, formationID string, triggered_by model.OperationTrigger) error {
	return nil
}

func (s *service) Finish(ctx context.Context, assignmentID, formationID string, operationType model.AssignmentOperationType) error {
	return nil
}

func (s *service) List(ctx context.Context, assignmentID string) (*model.AssignmentOperationPage, error) {
	return nil, nil
}
