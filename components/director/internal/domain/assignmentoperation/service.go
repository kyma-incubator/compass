package formationassignment

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

func (s *service) Finish(ctx context.Context, assignmentID, formationID string, operationType model.AssignmentOperationType) error {
	return nil
}

func (s *service) List(ctx context.Context, assignmentID string) (*model.AssignmentOperationPage, error) {
	return nil, nil
}
