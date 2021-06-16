package bundlereferences

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

//go:generate mockery --name=BundleReferenceRepository --output=automock --outpkg=automock --case=underscore
type BundleReferenceRepository interface {
	Create(ctx context.Context, item *model.BundleReference) error
	Update(ctx context.Context, item *model.BundleReference) error
	DeleteByReferenceObjectID(ctx context.Context, tenant, bundleID string, objectType model.BundleReferenceObjectType, objectID string) error
	GetByID(ctx context.Context, objectType model.BundleReferenceObjectType, tenantID string, objectID, bundleID *string) (*model.BundleReference, error)
	GetBundleIDsForObject(ctx context.Context, tenantID string, objectType model.BundleReferenceObjectType, objectID *string) (ids []string, err error)
	ListAllForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, tenantID string, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error)
}

type service struct {
	repo BundleReferenceRepository
}

func NewService(repo BundleReferenceRepository) *service {
	return &service{
		repo: repo,
	}
}

func (s *service) GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) (*model.BundleReference, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	bundleRef, err := s.repo.GetByID(ctx, objectType, tnt, objectID, bundleID)
	if err != nil {
		return nil, err
	}

	return bundleRef, nil
}

func (s *service) GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ids, err := s.repo.GetBundleIDsForObject(ctx, tnt, objectType, objectID)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (s *service) CreateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bundleReference, err := in.ToBundleReference(tnt, objectType, bundleID, objectID)
	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, bundleReference)
	if err != nil {
		return errors.Wrapf(err, "while creating record for %s with id %q for Bundle with id %q", objectType, *objectID, *bundleID)
	}

	return nil
}

func (s *service) UpdateByReferenceObjectID(ctx context.Context, in model.BundleReferenceInput, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	bundleReference, err := s.repo.GetByID(ctx, objectType, tnt, objectID, bundleID)
	if err != nil {
		return err
	}

	bundleReference, err = in.ToBundleReference(tnt, objectType, bundleID, objectID)
	if err != nil {
		return err
	}

	err = s.repo.Update(ctx, bundleReference)
	if err != nil {
		return errors.Wrapf(err, "while updating record for %s with id %q", objectType, *objectID)
	}

	return nil
}

func (s *service) DeleteByReferenceObjectID(ctx context.Context, objectType model.BundleReferenceObjectType, objectID, bundleID *string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	return s.repo.DeleteByReferenceObjectID(ctx, tnt, *bundleID, objectType, *objectID)
}

func (s *service) ListAllByBundleIDs(ctx context.Context, objectType model.BundleReferenceObjectType, bundleIDs []string, pageSize int, cursor string) ([]*model.BundleReference, map[string]int, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	if pageSize < 1 || pageSize > 200 {
		return nil, nil, apperrors.NewInvalidDataError("page size must be between 1 and 100")
	}

	return s.repo.ListAllForBundle(ctx, objectType, tnt, bundleIDs, pageSize, cursor)
}
