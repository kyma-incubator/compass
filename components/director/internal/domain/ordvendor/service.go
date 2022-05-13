package ordvendor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// VendorRepository missing godoc
//go:generate mockery --name=VendorRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type VendorRepository interface {
	Create(ctx context.Context, tenant string, item *model.Vendor) error
	CreateGlobal(ctx context.Context, model *model.Vendor) error
	Update(ctx context.Context, tenant string, item *model.Vendor) error
	UpdateGlobal(ctx context.Context, model *model.Vendor) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Vendor, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Vendor, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Vendor, error)
	ListGlobal(ctx context.Context) ([]*model.Vendor, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	vendorRepo VendorRepository
	uidService UIDService
}

// NewService creates a new instance of Vendor Service.
func NewService(vendorRepo VendorRepository, uidService UIDService) *service {
	return &service{
		vendorRepo: vendorRepo,
		uidService: uidService,
	}
}

// Create creates a new Vendor.
func (s *service) Create(ctx context.Context, applicationID string, in model.VendorInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	vendor := in.ToVendor(id, &applicationID)

	if err = s.vendorRepo.Create(ctx, tnt, vendor); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Vendor with id %s and title %s for Application with id %s", id, vendor.Title, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Vendor with id %s and title %s for Application with id %s", id, vendor.Title, applicationID)

	return vendor.OrdID, nil
}

// CreateGlobal creates a new Global Vendor (with NULL app_id).
func (s *service) CreateGlobal(ctx context.Context, in model.VendorInput) (string, error) {
	id := s.uidService.Generate()
	vendor := in.ToVendor(id, nil)

	if err := s.vendorRepo.CreateGlobal(ctx, vendor); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating Global Vendor with id %s and title %s", id, vendor.Title)
	}
	log.C(ctx).Debugf("Successfully created Global Vendor with id %s and title %s", id, vendor.Title)

	return vendor.OrdID, nil
}

// Update updates an existing Vendor.
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

	if err = s.vendorRepo.Update(ctx, tnt, vendor); err != nil {
		return errors.Wrapf(err, "while updating Vendor with id %s", id)
	}
	return nil
}

// UpdateGlobal updates an existing Vendor without tenant isolation.
func (s *service) UpdateGlobal(ctx context.Context, id string, in model.VendorInput) error {
	vendor, err := s.vendorRepo.GetByIDGlobal(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Vendor with id %s", id)
	}

	vendor.SetFromUpdateInput(in)

	if err = s.vendorRepo.UpdateGlobal(ctx, vendor); err != nil {
		return errors.Wrapf(err, "while updating Vendor with id %s", id)
	}
	return nil
}

// Delete deletes an existing Vendor.
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

// DeleteGlobal deletes an existing Vendor without tenant isolation.
func (s *service) DeleteGlobal(ctx context.Context, id string) error {
	err := s.vendorRepo.DeleteGlobal(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Vendor with id %s", id)
	}

	return nil
}

// Exist checks if a Vendor exists.
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

// Get returns a Vendor by its ID.
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

// ListByApplicationID returns a list of Vendors by Application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.vendorRepo.ListByApplicationID(ctx, tnt, appID)
}

// ListGlobal returns a list of Global Vendors (with NULL app_id).
func (s *service) ListGlobal(ctx context.Context) ([]*model.Vendor, error) {
	return s.vendorRepo.ListGlobal(ctx)
}
