package processors

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// PackageService is responsible for the service-layer Package operations.
//
//go:generate mockery --name=PackageService --output=automock --outpkg=automock --case=underscore --disable-version-string
type PackageService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.PackageInput, pkgHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.PackageInput, pkgHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Package, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Package, error)
}

type PackageProcessor struct {
	transact   persistence.Transactioner
	packageSvc PackageService
}

// NewPackageProcessor creates new instance of PackageProcessor
func NewPackageProcessor(transact persistence.Transactioner, packageSvc PackageService) *PackageProcessor {
	return &PackageProcessor{
		transact:   transact,
		packageSvc: packageSvc,
	}
}

func (pp *PackageProcessor) Process(ctx context.Context, resourceType directorresource.Type, resourceID string, packages []*model.PackageInput, resourceHashes map[string]uint64) ([]*model.Package, error) {
	packagesFromDB, err := pp.listPackagesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		pkgHash := resourceHashes[pkg.OrdID]
		if err := pp.resyncPackageInTx(ctx, resourceType, resourceID, packagesFromDB, pkg, pkgHash); err != nil {
			return nil, err
		}
	}

	packagesFromDB, err = pp.listPackagesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return packagesFromDB, nil
}

func (pp *PackageProcessor) listPackagesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Package, error) {
	tx, err := pp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer pp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var packagesFromDB []*model.Package
	switch resourceType {
	case directorresource.Application:
		packagesFromDB, err = pp.packageSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
		packagesFromDB, err = pp.packageSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing packages for %s with id %q", resourceType, resourceID)
	}

	return packagesFromDB, tx.Commit()
}

func (pp *PackageProcessor) resyncPackageInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, packagesFromDB []*model.Package, pkg *model.PackageInput, pkgHash uint64) error {
	tx, err := pp.transact.Begin()
	if err != nil {
		return err
	}
	defer pp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := pp.resyncPackage(ctx, resourceType, resourceID, packagesFromDB, *pkg, pkgHash); err != nil {
		return errors.Wrapf(err, "error while resyncing package with ORD ID %q", pkg.OrdID)
	}
	return tx.Commit()
}

func (pp *PackageProcessor) resyncPackage(ctx context.Context, resourceType directorresource.Type, resourceID string, packagesFromDB []*model.Package, pkg model.PackageInput, pkgHash uint64) error {
	ctx = addFieldToLogger(ctx, "package_ord_id", pkg.OrdID)
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return packagesFromDB[i].OrdID == pkg.OrdID
	}); found {
		return pp.packageSvc.Update(ctx, resourceType, packagesFromDB[i].ID, pkg, pkgHash)
	}

	_, err := pp.packageSvc.Create(ctx, resourceType, resourceID, pkg, pkgHash)
	return err
}
