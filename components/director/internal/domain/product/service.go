package product

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// ProductRepository missing godoc
//go:generate mockery --name=ProductRepository --output=automock --outpkg=automock --case=underscore
type ProductRepository interface {
	Create(ctx context.Context, tenant string, item *model.Product) error
	Update(ctx context.Context, tenant string, item *model.Product) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Product, error)
	ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Product, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	productRepo ProductRepository
	uidService  UIDService
}

// NewService missing godoc
func NewService(productRepo ProductRepository, uidService UIDService) *service {
	return &service{
		productRepo: productRepo,
		uidService:  uidService,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, applicationID string, in model.ProductInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	id := s.uidService.Generate()
	product := in.ToProduct(id, applicationID)

	if err = s.productRepo.Create(ctx, tnt, product); err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Product with id %s and title %s for Application with id %s", id, product.Title, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Product with id %s and title %s for Application with id %s", id, product.Title, applicationID)

	return product.OrdID, nil
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.ProductInput) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return err
	}

	product, err := s.productRepo.GetByID(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while getting Product with id %s", id)
	}

	product.SetFromUpdateInput(in)

	if err = s.productRepo.Update(ctx, tnt, product); err != nil {
		return errors.Wrapf(err, "while updating Product with id %s", id)
	}
	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading tenant from context")
	}

	err = s.productRepo.Delete(ctx, tnt, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Product with id %s", id)
	}

	return nil
}

// Exist missing godoc
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

// Get missing godoc
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

// ListByApplicationID missing godoc
func (s *service) ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return s.productRepo.ListByApplicationID(ctx, tnt, appID)
}
