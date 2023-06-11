package ordvendor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// VendorRepository missing godoc
//
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
	ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.Vendor, error)
	ListGlobal(ctx context.Context) ([]*model.Vendor, error)
}

// UIDService missing godoc
//
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
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.VendorInput) (string, error) {
	id := s.uidService.Generate()
	vendor := in.ToVendor(id, resourceType, resourceID)

	var (
		err error
		tnt string
	)
	if resourceType == resource.ApplicationTemplateVersion {
		err = s.vendorRepo.CreateGlobal(ctx, vendor)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return "", err
		}

		err = s.vendorRepo.Create(ctx, tnt, vendor)
	}
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Vendor with id %s and title %s", id, vendor.Title)
	}

	log.C(ctx).Debugf("Successfully created a Vendor with id %s and title %s", id, vendor.Title)

	return vendor.OrdID, nil
}

// CreateGlobal creates a new Global Vendor (with NULL app_id).
func (s *service) CreateGlobal(ctx context.Context, in model.VendorInput) (string, error) {
	id := s.uidService.Generate()
	vendor := in.ToVendor(id, "", "")

	if err := s.vendorRepo.CreateGlobal(ctx, vendor); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating Global Vendor with id %s and title %s", id, vendor.Title)
	}
	log.C(ctx).Debugf("Successfully created Global Vendor with id %s and title %s", id, vendor.Title)

	return vendor.OrdID, nil
}

// Update updates an existing Vendor.
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.VendorInput) error {
	var (
		vendor *model.Vendor
		err    error
		tnt    string
	)

	if resourceType == resource.ApplicationTemplateVersion {
		vendor, err = s.vendorRepo.GetByIDGlobal(ctx, id)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return err
		}

		vendor, err = s.vendorRepo.GetByID(ctx, tnt, id)
	}
	if err != nil {
		return errors.Wrapf(err, "while getting Vendor with id %s", id)
	}

	vendor.SetFromUpdateInput(in)

	if resourceType == resource.ApplicationTemplateVersion {
		err = s.vendorRepo.UpdateGlobal(ctx, vendor)
	} else {
		err = s.vendorRepo.Update(ctx, tnt, vendor)
	}
	if err != nil {
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

	return s.vendorRepo.ListByResourceID(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionID returns a list of Vendors by Application Template Version ID without tenant isolation.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Vendor, error) {
	return s.vendorRepo.ListByResourceID(ctx, "", appTemplateVersionID, resource.ApplicationTemplateVersion)
}

// ListGlobal returns a list of Global Vendors (with NULL app_id).
func (s *service) ListGlobal(ctx context.Context) ([]*model.Vendor, error) {
	return s.vendorRepo.ListGlobal(ctx)
}
