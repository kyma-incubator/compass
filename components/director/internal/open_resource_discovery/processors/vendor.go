package processors

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// VendorService is responsible for the service-layer Vendor operations.
//
//go:generate mockery --name=VendorService --output=automock --outpkg=automock --case=underscore --disable-version-string
type VendorService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.VendorInput) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.VendorInput) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Vendor, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Vendor, error)
}

type VendorProcessor struct {
	transact  persistence.Transactioner
	vendorSvc VendorService
}

// NewVendorProcessor creates new instance of VendorProcessor
func NewVendorProcessor(transact persistence.Transactioner, vendorSvc VendorService) *VendorProcessor {
	return &VendorProcessor{
		transact:  transact,
		vendorSvc: vendorSvc,
	}
}

func (vp *VendorProcessor) Process(ctx context.Context, resourceType directorresource.Type, resourceID string, vendors []*model.VendorInput) ([]*model.Vendor, error) {
	vendorsFromDB, err := vp.listVendorsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, vendor := range vendors {
		if err := vp.resyncVendorInTx(ctx, resourceType, resourceID, vendorsFromDB, vendor); err != nil {
			return nil, err
		}
	}

	vendorsFromDB, err = vp.listVendorsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return vendorsFromDB, nil
}

func (vp *VendorProcessor) listVendorsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Vendor, error) {
	tx, err := vp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer vp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var vendorsFromDB []*model.Vendor
	switch resourceType {
	case directorresource.Application:
		vendorsFromDB, err = vp.vendorSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
		vendorsFromDB, err = vp.vendorSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing vendors for %s with id %q", resourceType, resourceID)
	}

	return vendorsFromDB, tx.Commit()
}

func (vp *VendorProcessor) resyncVendorInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, vendorsFromDB []*model.Vendor, vendor *model.VendorInput) error {
	tx, err := vp.transact.Begin()
	if err != nil {
		return err
	}
	defer vp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := vp.resyncVendor(ctx, resourceType, resourceID, vendorsFromDB, *vendor); err != nil {
		return errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
	}
	return tx.Commit()
}

func (vp *VendorProcessor) resyncVendor(ctx context.Context, resourceType directorresource.Type, resourceID string, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	ctx = addFieldToLogger(ctx, "vendor_ord_id", vendor.OrdID)
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return vp.vendorSvc.Update(ctx, resourceType, vendorsFromDB[i].ID, vendor)
	}

	_, err := vp.vendorSvc.Create(ctx, resourceType, resourceID, vendor)
	return err
}
