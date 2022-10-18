package bundlereferences

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

// BundleReferenceRepository is responsible for the repo-layer BundleReference operations.
//go:generate mockery --name=BundleReferenceRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceRepository interface {
	Create(ctx context.Context, item *model.BundleReference) error
	Update(ctx context.Context, item *model.BundleReference) error
	DeleteByReferenceObjectID(ctx context.Context, bundleID string, objectType model.BundleReferenceObjectType, objectID string) error
	GetByID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error)
	GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) (ids []string, err error)
	ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error)
}

// UIDService is responsible for generating GUIDs, which will be used as internal bundleReference IDs when they are created.
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo       BundleReferenceRepository
	uidService UIDService
}

// NewService returns a new object responsible for service-layer BundleReference operations.
func NewService(repo BundleReferenceRepository, uidService UIDService) *service {
	return &service{
		repo:       repo,
		uidService: uidService,
	}
}

// GetForBundle returns the BundleReference that is related to a specific Bundle.
func (s *service) GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error) {
	bundleRef, err := s.repo.GetByID(ctx, objectType, objectID, bundleID)
	if err != nil {
		return nil, err
	}

	return bundleRef, nil
}

// GetBundleIDsForObject returns all bundle IDs that are related for a specific object(APIDefinition/EventDefinition).
func (s *service) GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error) {
	ids, err := s.repo.GetBundleIDsForObject(ctx, objectType, objectID)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// CreateByReferenceObjectID creates a BundleReference between a Bundle and object(APIDefinition/EventDefinition).
func (s *service) CreateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	id := s.uidService.Generate()
	bundleReference, err := in.ToBundleReference(id, objectType, bundleID, objectID)
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, bundleReference)
	if err != nil {
		return errors.Wrapf(err, "while creating record for %s with id %q for Bundle with id %q", objectType, *objectID, *bundleID)
	}

	return nil
}

// UpdateByReferenceObjectID updates a BundleReference for a specific object(APIDefinition/EventDefinition).
func (s *service) UpdateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	bundleReference, err := s.repo.GetByID(ctx, objectType, objectID, bundleID)
	if err != nil {
		return err
	}

	bundleReference, err = in.ToBundleReference(bundleReference.ID, objectType, bundleID, objectID)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, bundleReference)
	if err != nil {
		return errors.Wrapf(err, "while updating record for %s with id %q", objectType, *objectID)
	}

	return nil
}

// DeleteByReferenceObjectID deletes a BundleReference for a specific object(APIDefinition/EventDefinition).
func (s *service) DeleteByReferenceObjectID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	return s.repo.DeleteByReferenceObjectID(ctx, *bundleID, objectType, *objectID)
}

// ListByBundleIDs lists all BundleReferences for given array of bundle IDs. In addition, the number of records for each BundleReference is returned.
func (s *service) ListByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error) {
	if pageSize < 1 || pageSize > 600 {
		return nil, nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.ListByBundleIDs(ctx, objectType, bundleIDs, pageSize, cursor)
}
