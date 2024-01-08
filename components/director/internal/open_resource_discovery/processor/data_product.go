package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// DataProductService is responsible for the service-layer Data Product operations.
//
//go:generate mockery --name=DataProductService --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataProductService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.DataProduct, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.DataProduct, error)
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.DataProductInput, dataProductHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, packageID *string, in model.DataProductInput, dataProductHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// DataProductProcessor defines Data Product processor
type DataProductProcessor struct {
	transact       persistence.Transactioner
	dataProductSvc DataProductService
}

// NewDataProductProcessor creates new instance of DataProductProcessor
func NewDataProductProcessor(transact persistence.Transactioner, dataProductSvc DataProductService) *DataProductProcessor {
	return &DataProductProcessor{
		transact:       transact,
		dataProductSvc: dataProductSvc,
	}
}

// Process re-syncs the data products passed as an argument.
func (id *DataProductProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, packagesFromDB []*model.Package, dataProducts []*model.DataProductInput, resourceHashes map[string]uint64) ([]*model.DataProduct, error) {
	dataProductsFromDB, err := id.listDataProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, dataProduct := range dataProducts {
		dataProductHash := resourceHashes[str.PtrStrToStr(dataProduct.OrdID)]
		if err := id.resyncDataProductInTx(ctx, resourceType, resourceID, dataProductsFromDB, packagesFromDB, dataProduct, dataProductHash); err != nil {
			return nil, err
		}
	}

	dataProductsFromDB, err = id.listDataProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return dataProductsFromDB, nil
}

func (id *DataProductProcessor) listDataProductsInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.DataProduct, error) {
	tx, err := id.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer id.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var dataProductsFromDB []*model.DataProduct
	switch resourceType {
	case resource.Application:
		dataProductsFromDB, err = id.dataProductSvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		dataProductsFromDB, err = id.dataProductSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing data products for %s with id %q", resourceType, resourceID)
	}

	return dataProductsFromDB, tx.Commit()
}

func (id *DataProductProcessor) resyncDataProductInTx(ctx context.Context, resourceType resource.Type, resourceID string, dataProductsFromDB []*model.DataProduct, packagesFromDB []*model.Package, dataProduct *model.DataProductInput, dataProductHash uint64) error {
	tx, err := id.transact.Begin()
	if err != nil {
		return err
	}
	defer id.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := id.resyncDataProduct(ctx, resourceType, resourceID, dataProductsFromDB, packagesFromDB, *dataProduct, dataProductHash); err != nil {
		return errors.Wrapf(err, "error while resyncing data product for resource with ORD ID %q", *dataProduct.OrdID)
	}
	return tx.Commit()
}

func (id *DataProductProcessor) resyncDataProduct(ctx context.Context, resourceType resource.Type, resourceID string, dataProductsFromDB []*model.DataProduct, packagesFromDB []*model.Package, dataProduct model.DataProductInput, dataProductHash uint64) error {
	ctx = addFieldToLogger(ctx, "data_product_ord_id", *dataProduct.OrdID)
	i, isDataProductFound := searchInSlice(len(dataProductsFromDB), func(i int) bool {
		return equalStrings(dataProductsFromDB[i].OrdID, dataProduct.OrdID)
	})

	var packageID *string
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, dataProduct.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	if !isDataProductFound {
		_, err := id.dataProductSvc.Create(ctx, resourceType, resourceID, packageID, dataProduct, dataProductHash)
		if err != nil {
			return err
		}

		return nil
	}

	err := id.dataProductSvc.Update(ctx, resourceType, resourceID, dataProductsFromDB[i].ID, packageID, dataProduct, dataProductHash)
	if err != nil {
		return err
	}

	return nil
}
