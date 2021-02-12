package ordvendor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery -name=VendorRepository -output=automock -outpkg=automock -case=underscore
type VendorRepository interface {
	Create(ctx context.Context, item *model.Vendor) error
	Update(ctx context.Context, item *model.Vendor) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Vendor, error)
}

type service struct {
	vendorRepo VendorRepository
}

func NewService(vendorRepo VendorRepository) *service {
	return &service{
		vendorRepo: vendorRepo,
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.VendorInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	vendor := in.ToVendor(tnt, applicationID)

	err = s.vendorRepo.Create(ctx, vendor)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Vendor with id %s and title %s for Application with id %s", vendor.OrdID, vendor.Title, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Vendor with id %s and title %s for Application with id %s", vendor.OrdID, vendor.Title, applicationID)

	return vendor.OrdID, nil
}

func (s *service) Update(ctx context.Context, id string, in model.VendorInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	vendor, err := s.vendorRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Vendor with id %s", id)
	}

	vendor.SetFromUpdateInput(in)

	err = s.vendorRepo.Update(ctx, vendor)
	if err != nil {
		return errors.Wrapf(err, "while updating Vendor with id %s", id)
	}
	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.vendorRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Vendor with id %s", id)
	}

	return nil
}

func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.vendorRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Vendor with ID: %q", id)
	}

	return exist, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.Vendor, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	vendor, err := s.vendorRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Vendor with ID: %q", id)
	}

	return vendor, nil
}
