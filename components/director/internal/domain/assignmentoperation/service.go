package assignmentOperation

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// AssignmentOperationRepository represents the Assignment Operation repository layer
//
//go:generate mockery --name=AssignmentOperationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type AssignmentOperationRepository interface {
	Create(ctx context.Context, item *model.AssignmentOperation) error
	Finish(ctx context.Context, m *model.AssignmentOperation) error
	GetLatestOperation(ctx context.Context, formationAssignmentID, formationID string, operationType model.AssignmentOperationType) (*model.AssignmentOperation, error)
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
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}

	assignmentOperationID := s.uidSvc.Generate()
	log.C(ctx).Debugf("ID: %q generated for assignment operation for tenant with ID: %q", assignmentOperationID, tenantID)

	log.C(ctx).Infof("Creating assignemnt operation for formation assignment %s in formation %s with type %s, triggered by %s ", in.FormationAssignmentID, in.FormationID, in.Type, in.TriggeredBy)
	if err = s.repo.Create(ctx, in.ToModel(assignmentOperationID)); err != nil {
		return "", errors.Wrapf(err, "while creating formation assignment for formation with ID: %q", in.FormationID)
	}

	return assignmentOperationID, nil
}

// Finish finishes the Assignment Operation
func (s *service) Finish(ctx context.Context, assignmentID, formationID string, operationType model.AssignmentOperationType) error {
	operation, err := s.repo.GetLatestOperation(ctx, assignmentID, formationID, operationType)
	if err != nil {
		return err
	}

	err = s.repo.Finish(ctx, operation)
	if err != nil {
		return err
	}

	return nil
}

// ListByFormationAssignmentIDs fetches the Assignment Operations for the provided Formation Assignment IDs
func (s *service) ListByFormationAssignmentIDs(ctx context.Context, formationAssignmentIDs []string, pageSize int, cursor string) ([]*model.AssignmentOperationPage, error) {
	return s.repo.ListForFormationAssignmentIDs(ctx, formationAssignmentIDs, pageSize, cursor)
}
