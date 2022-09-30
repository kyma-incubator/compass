package formationassignment

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// FormationAssignmentRepository represents the formation assignment repository layer
//go:generate mockery --name=FormationAssignmentRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationAssignmentRepository interface {
	Create(ctx context.Context, item *model.FormationAssignment) error
	Get(ctx context.Context, id, tenantID string) (*model.FormationAssignment, error)
	GetForFormation(ctx context.Context, tenantID, id, formationID string) (*model.FormationAssignment, error)
	List(ctx context.Context, pageSize int, cursor, tenantID string) (*model.FormationAssignmentPage, error)
	ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error)
	Update(ctx context.Context, model *model.FormationAssignment) error
	Delete(ctx context.Context, id, tenantID string) error
	Exists(ctx context.Context, id, tenantID string) (bool, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo   FormationAssignmentRepository
	uidSvc UIDService
}

// NewService creates a FormationTemplate service
func NewService(repo FormationAssignmentRepository, uidSvc UIDService) *service {
	return &service{
		repo:   repo,
		uidSvc: uidSvc,
	}
}

// Create creates a formation assignment using `in`
func (s *service) Create(ctx context.Context, in *model.FormationAssignmentInput) (string, error) {
	formationAssignmentID := s.uidSvc.Generate()
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", errors.Wrapf(err, "while loading tenant from context")
	}
	log.C(ctx).Debugf("ID: %q generated for formation assignment for tenant with ID: %q", formationAssignmentID, tenantID)

	log.C(ctx).Infof("Creating formation assignment with source: %q and source type: %q, and target: %q with target type: %q", in.Source, in.SourceType, in.Target, in.TargetType)
	if err = s.repo.Create(ctx, in.ToModel(formationAssignmentID, tenantID)); err != nil {
		return "", errors.Wrapf(err, "while creating formation assignment for formation with ID: %q", in.FormationID)
	}

	return formationAssignmentID, nil
}

// Get queries formation assignment matching ID `id`
func (s *service) Get(ctx context.Context, id string) (*model.FormationAssignment, error) {
	log.C(ctx).Infof("Getting formation assignment with ID: %q", id)

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := s.repo.Get(ctx, id, tenantID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q and tenant: %q", id, tenantID)
	}

	return fa, nil
}

// GetForFormation retrieves the formation assignment with the provided `id` associated with formation with id `formationID`
func (s *service) GetForFormation(ctx context.Context, id, formationID string) (*model.FormationAssignment, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	fa, err := s.repo.GetForFormation(ctx, tenantID, id, formationID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting formation assignment with ID: %q for formation with ID: %q", id, formationID)
	}

	return fa, nil
}

// List retrieves formation assignment page based on `pageSize` and `cursor`
func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationAssignmentPage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	return s.repo.List(ctx, pageSize, cursor, tenantID)
}

// ListByFormationIDs retrieves a pages of formation assignment objects for each of the provided formation IDs
func (s *service) ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading tenant from context")
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.ListByFormationIDs(ctx, tnt, formationIDs, pageSize, cursor)
}

// Update updates a formation assignment matching ID `id` using `in`
func (s *service) Update(ctx context.Context, id string, in *model.FormationAssignmentInput) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	exists, err := s.repo.Exists(ctx, id, tenantID)
	if err != nil {
		return errors.Wrapf(err, "while ensuring formation assignment with ID: %q exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationAssignment, id)
	}

	err = s.repo.Update(ctx, in.ToModel(id, tenantID))
	if err != nil {
		return errors.Wrapf(err, "while updating formation assignment with ID: %q", id)
	}

	return nil
}

// Delete deletes a formation assignment matching ID `id`
func (s *service) Delete(ctx context.Context, id string) error {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrapf(err, "while loading tenant from context")
	}

	if err := s.repo.Delete(ctx, id, tenantID); err != nil {
		return errors.Wrapf(err, "while deleting formation assignment with ID: %q", id)
	}
	return nil
}

// Exists check if a formation assignment with given ID exists
func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrapf(err, "while loading tenant from context")
	}

	exists, err := s.repo.Exists(ctx, id, tenantID)
	if err != nil {
		return false, errors.Wrapf(err, "while checking formation assignment existence for ID: %q and tenant: %q", id, tenantID)
	}
	return exists, nil
}
