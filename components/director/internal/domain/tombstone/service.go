package tombstone

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery -name=TombstoneRepository -output=automock -outpkg=automock -case=underscore
type TombstoneRepository interface {
	Create(ctx context.Context, item *model.Tombstone) error
	Update(ctx context.Context, item *model.Tombstone) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Tombstone, error)
}

type service struct {
	tombstoneRepo TombstoneRepository
}

func NewService(tombstoneRepo TombstoneRepository) *service {
	return &service{
		tombstoneRepo: tombstoneRepo,
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.TombstoneInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	tombstone := in.ToTombstone(tnt, applicationID)

	err = s.tombstoneRepo.Create(ctx, tombstone)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Tombstone with id %s for Application with id %s", tombstone.OrdID, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Tombstone with id %s for Application with id %s", tombstone.OrdID, applicationID)

	return tombstone.OrdID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.TombstoneInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	tombstone, err := s.tombstoneRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Tombstone with id %s", id)
	}

	tombstone.SetFromUpdateInput(in)

	err = s.tombstoneRepo.Update(ctx, tombstone)
	if err != nil {
		return errors.Wrapf(err, "while updating Tombstone with id %s", id)
	}
	return nil
}

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
