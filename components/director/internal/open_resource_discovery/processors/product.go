package processors

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// ProductService is responsible for the service-layer Product operations.
//
//go:generate mockery --name=ProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ProductService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.ProductInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.ProductInput) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Product, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.Product, error)
}

type ProductProcessor struct {
	transact   persistence.Transactioner
	productSvc ProductService
}

// NewProductProcessor creates new instance of ProductProcessor
func NewProductProcessor(transact persistence.Transactioner, productSvc ProductService) *ProductProcessor {
	return &ProductProcessor{
		transact:   transact,
		productSvc: productSvc,
	}
}

func (pp *ProductProcessor) Process(ctx context.Context, resourceType directorresource.Type, resourceID string, products []*model.ProductInput) ([]*model.Product, error) {
	productsFromDB, err := pp.listProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		if err := pp.resyncProductInTx(ctx, resourceType, resourceID, productsFromDB, product); err != nil {
			return nil, err
		}
	}

	productsFromDB, err = pp.listProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return productsFromDB, nil
}

func (pp *ProductProcessor) listProductsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Product, error) {
	tx, err := pp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer pp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var productsFromDB []*model.Product
	switch resourceType {
	case directorresource.Application:
		productsFromDB, err = pp.productSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
		productsFromDB, err = pp.productSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing products for %s with id %q", resourceType, resourceID)
	}

	return productsFromDB, tx.Commit()
}

func (pp *ProductProcessor) resyncProductInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, productsFromDB []*model.Product, product *model.ProductInput) error {
	tx, err := pp.transact.Begin()
	if err != nil {
		return err
	}
	defer pp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := pp.resyncProduct(ctx, resourceType, resourceID, productsFromDB, *product); err != nil {
		return errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
	}
	return tx.Commit()
}

func (pp *ProductProcessor) resyncProduct(ctx context.Context, resourceType directorresource.Type, resourceID string, productsFromDB []*model.Product, product model.ProductInput) error {
	ctx = addFieldToLogger(ctx, "product_ord_id", product.OrdID)
	if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
		return productsFromDB[i].OrdID == product.OrdID
	}); found {
		return pp.productSvc.Update(ctx, resourceType, productsFromDB[i].ID, product)
	}

	_, err := pp.productSvc.Create(ctx, resourceType, resourceID, product)
	return err
}
