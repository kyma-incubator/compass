package product

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ProductRepository missing godoc
//
//go:generate mockery --name=ProductRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ProductRepository interface {
	Create(ctx context.Context, tenant string, item *model.Product) error
	CreateGlobal(ctx context.Context, model *model.Product) error
	Update(ctx context.Context, tenant string, item *model.Product) error
	UpdateGlobal(ctx context.Context, model *model.Product) error
	Delete(ctx context.Context, tenant, id string) error
	DeleteGlobal(ctx context.Context, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Product, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Product, error)
	ListByResourceID(ctx context.Context, tenantID, appID string, resourceType resource.Type) ([]*model.Product, error)
	ListGlobal(ctx context.Context) ([]*model.Product, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	productRepo ProductRepository
	uidService  UIDService
}

// NewService creates a new instance of Product Service.
func NewService(productRepo ProductRepository, uidService UIDService) *service {
	return &service{
		productRepo: productRepo,
		uidService:  uidService,
	}
}

// Create creates a new product.
func (s *service) Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.ProductInput) (string, error) {
	id := s.uidService.Generate()
	product := in.ToProduct(id, resourceType, resourceID)

	var (
		err error
		tnt string
	)
	if resourceType.IsTenantIgnorable() {
		err = s.productRepo.CreateGlobal(ctx, product)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return "", err
		}

		err = s.productRepo.Create(ctx, tnt, product)
	}
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Product with id %s and title %s for %v with id %s", id, product.Title, resourceType, resourceID)
	}

	log.C(ctx).Debugf("Successfully created a Product with id %s and title %s for %s with id %s", id, product.Title, resourceType, resourceID)

	return product.OrdID, nil
}

// CreateGlobal creates a new global product (with NULL app_id).
func (s *service) CreateGlobal(ctx context.Context, in model.ProductInput) (string, error) {
	id := s.uidService.Generate()
	product := in.ToProduct(id, "", "")

	if err := s.productRepo.CreateGlobal(ctx, product); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating Global Product with id %s and title %s", id, product.Title)
	}
	log.C(ctx).Debugf("Successfully created a Global Product with id %s and title %s", id, product.Title)

	return product.OrdID, nil
}

// Update updates an existing product.
func (s *service) Update(ctx context.Context, resourceType resource.Type, id string, in model.ProductInput) error {
	var (
		product *model.Product
		tnt     string
		err     error
	)

	if resourceType.IsTenantIgnorable() {
		product, err = s.productRepo.GetByIDGlobal(ctx, id)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return err
		}

		product, err = s.productRepo.GetByID(ctx, tnt, id)
	}
	if err != nil {
		return errors.Wrapf(err, "while getting Product with id %s", id)
	}

	product.SetFromUpdateInput(in)

	if resourceType.IsTenantIgnorable() {
		err = s.productRepo.UpdateGlobal(ctx, product)
	} else {
		err = s.productRepo.Update(ctx, tnt, product)
	}
	if err != nil {
		return errors.Wrapf(err, "while updating Product with id %s", id)
	}

	return nil
}

// UpdateGlobal updates an existing product without tenant isolation.
func (s *service) UpdateGlobal(ctx context.Context, id string, in model.ProductInput) error {
	product, err := s.productRepo.GetByIDGlobal(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Product with id %s", id)
	}

	product.SetFromUpdateInput(in)

	if err = s.productRepo.UpdateGlobal(ctx, product); err != nil {
		return errors.Wrapf(err, "while updating Product with id %s", id)
	}
	return nil
}

// Delete deletes an existing product.
func (s *service) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	var (
		tnt string
		err error
	)

	if resourceType.IsTenantIgnorable() {
		err = s.productRepo.DeleteGlobal(ctx, id)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return errors.Wrap(err, "while loading tenant from context")
		}

		err = s.productRepo.Delete(ctx, tnt, id)
	}
	if err != nil {
		return errors.Wrapf(err, "while deleting Product with id %s", id)
	}

	log.C(ctx).Infof("Successfully deleted Product with id %s", id)

	return nil
}

// DeleteGlobal deletes an existing product without tenant isolation.
func (s *service) DeleteGlobal(ctx context.Context, id string) error {
	if err := s.productRepo.DeleteGlobal(ctx, id); err != nil {
		return errors.Wrapf(err, "while deleting Product with id %s", id)
	}

	return nil
}

// Exist checks if a product exists.
func (s *service) Exist(ctx context.Context, id string) (bool, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading tenant from context")
	}

	exist, err := s.productRepo.Exists(ctx, tnt, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Product with ID: %q", id)
	}

	return exist, nil
}

// Get returns a product by its ID.
func (s *service) Get(ctx context.Context, id string) (*model.Product, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while loading tenant from context")
	}

	product, err := s.productRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Product with ID: %q", id)
	}

	return product, nil
}

// ListByApplicationID returns a list of products for a given application ID.
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.productRepo.ListByResourceID(ctx, tnt, appID, resource.Application)
}

// ListByApplicationTemplateVersionID returns a list of products for a given application ID.
func (s *service) ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Product, error) {
	return s.productRepo.ListByResourceID(ctx, "", appID, resource.ApplicationTemplateVersion)
}

// ListGlobal returns a list of global products (with NULL app_id).
func (s *service) ListGlobal(ctx context.Context) ([]*model.Product, error) {
	return s.productRepo.ListGlobal(ctx)
}
