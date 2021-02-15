package product

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate mockery -name=ProductRepository -output=automock -outpkg=automock -case=underscore
type ProductRepository interface {
	Create(ctx context.Context, item *model.Product) error
	Update(ctx context.Context, item *model.Product) error
	Delete(ctx context.Context, tenant, id string) error
	Exists(ctx context.Context, tenant, id string) (bool, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Product, error)
}

type service struct {
	productRepo ProductRepository
}

func NewService(productRepo ProductRepository) *service {
	return &service{
		productRepo: productRepo,
	}
}

func (s *service) Create(ctx context.Context, applicationID string, in model.ProductInput) (string, error) {
	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	product := in.ToProduct(tnt, applicationID)

	err = s.productRepo.Create(ctx, product)
	if err != nil {
		return "", errors.Wrapf(err, "error occurred while creating a Product with id %s and title %s for Application with id %s", product.OrdID, product.Title, applicationID)
	}
	log.C(ctx).Debugf("Successfully created a Product with id %s and title %s for Application with id %s", product.OrdID, product.Title, applicationID)

	return product.OrdID, nil
}

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

	err = s.productRepo.Update(ctx, product)
	if err != nil {
		return errors.Wrapf(err, "while updating Product with id %s", id)
	}
	return nil
}

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
