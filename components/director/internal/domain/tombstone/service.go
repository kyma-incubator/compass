package tombstone

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// TombstoneRepository missing godoc
//
//go:generate mockery --name=TombstoneRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type TombstoneRepository interface {
	Create(ctx context.Context, tenant string, item *model.Tombstone) error
	CreateGlobal(ctx context.Context, model *model.Tombstone) error
	Update(ctx context.Context, tenant string, item *model.Tombstone) error
	UpdateGlobal(ctx context.Context, model *model.Tombstone) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Tombstone, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Tombstone, error)
	ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.Tombstone, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	tombstoneRepo TombstoneRepository
	uidService    UIDService
}

// NewService missing godoc
func NewService(tombstoneRepo TombstoneRepository, uidService UIDService) *service {
	return &service{
		tombstoneRepo: tombstoneRepo,
		uidService:    uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.TombstoneInput) (string, error) {
	id := s.uidService.Generate()
	tombstone := in.ToTombstone(id, resourceType, resourceID)

	var (
		err error
		tnt string
	)
	if resourceType == resource.ApplicationTemplateVersion {
		err = s.tombstoneRepo.CreateGlobal(ctx, tombstone)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return "", err
		}

		err = s.tombstoneRepo.Create(ctx, tnt, tombstone)
	}
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Tombstone with id %s for %s with id %s", id, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created a Tombstone with id %s for %s with id %s", id, resourceType, resourceID)

	return tombstone.OrdID, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.TombstoneInput) error {
	var (
		tombstone *model.Tombstone
		tnt       string
		err       error
	)

	if resourceType == resource.ApplicationTemplateVersion {
		tombstone, err = s.tombstoneRepo.GetByIDGlobal(ctx, id)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return err
		}

		tombstone, err = s.tombstoneRepo.GetByID(ctx, tnt, id)
	}
	if err != nil {
		return errors.Wrapf(err, "while getting Tombstone with id %s", id)
	}

	tombstone.SetFromUpdateInput(in)

	if resourceType == resource.ApplicationTemplateVersion {
		err = s.tombstoneRepo.UpdateGlobal(ctx, tombstone)
	} else {
		err = s.tombstoneRepo.Update(ctx, tnt, tombstone)
	}
	if err != nil {
		return errors.Wrapf(err, "while updating Tombstone with id %s", id)
	}

	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.tombstoneRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Tombstone with id %s", id)
	}

	return nil
}

// Exist missing godoc
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.tombstoneRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Tombstone with ID: %q", id)
	}

	return exist, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.Tombstone, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	tombstone, err := s.tombstoneRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Tombstone with ID: %q", id)
	}

	return tombstone, nil
}

// ListByApplicationID missing godoc
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Tombstone, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.tombstoneRepo.ListByResourceID(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionID missing godoc
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Tombstone, error) {
	return s.tombstoneRepo.ListByResourceID(ctx, "", appID, resource.ApplicationTemplateVersion)
}
