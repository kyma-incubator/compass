package processor

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// BundleService is responsible for the service-layer Bundle operations.
//
//go:generate mockery --name=BundleService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleService interface {
	CreateBundle(ctx context.Context, resourceType resource.Type, resourceID string, in model.BundleCreateInput, bndlHash uint64) (string, error)
	UpdateBundle(ctx context.Context, resourceType resource.Type, id string, in model.BundleUpdateInput, bndlHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationIDNoPaging(ctx context.Context, appID string) ([]*model.Bundle, error)
	ListByApplicationTemplateVersionIDNoPaging(ctx context.Context, appTemplateVersionID string) ([]*model.Bundle, error)
}

// TombstonedResourcesDeleter defines tombstoned resources deleter
type TombstonedResourcesDeleter struct {
	transact                 persistence.Transactioner
	packageSvc               PackageService
	apiSvc                   APIService
	eventSvc                 EventService
	entityTypeSvc            EntityTypeService
	capabilitySvc            CapabilityService
	integrationDependencySvc IntegrationDependencyService
	vendorSvc                VendorService
	productSvc               ProductService
	bundleSvc                BundleService
}

// NewTombstonedResourcesDeleter creates new instance of TombstonedResourcesDeleter
func NewTombstonedResourcesDeleter(transact persistence.Transactioner, packageSvc PackageService, apiSvc APIService, eventSvc EventService, entityTypeSvc EntityTypeService, capabilitySvc CapabilityService, integrationDependencySvc IntegrationDependencyService, vendorSvc VendorService, productSvc ProductService, bundleSvc BundleService) *TombstonedResourcesDeleter {
	return &TombstonedResourcesDeleter{
		transact:                 transact,
		packageSvc:               packageSvc,
		apiSvc:                   apiSvc,
		eventSvc:                 eventSvc,
		entityTypeSvc:            entityTypeSvc,
		capabilitySvc:            capabilitySvc,
		integrationDependencySvc: integrationDependencySvc,
		vendorSvc:                vendorSvc,
		productSvc:               productSvc,
		bundleSvc:                bundleSvc,
	}
}

// Delete deletes all tombstoned resources.
func (td *TombstonedResourcesDeleter) Delete(ctx context.Context, resourceType resource.Type, vendorsFromDB []*model.Vendor, productsFromDB []*model.Product, packagesFromDB []*model.Package, bundlesFromDB []*model.Bundle, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition, entityTypesFromDB []*model.EntityType, capabilitiesFromDB []*model.Capability, integrationDependenciesFromDB []*model.IntegrationDependency, tombstonesFromDB []*model.Tombstone, fetchRequests []*OrdFetchRequest) ([]*OrdFetchRequest, error) {
	tx, err := td.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer td.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	frIdxToExclude := make([]int, 0)
	for _, ts := range tombstonesFromDB {
		if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
			return packagesFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := td.packageSvc.Delete(ctx, resourceType, packagesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.apiSvc.Delete(ctx, resourceType, apisFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.eventSvc.Delete(ctx, resourceType, eventsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(entityTypesFromDB), func(i int) bool {
			return equalStrings(&entityTypesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.entityTypeSvc.Delete(ctx, resourceType, entityTypesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(capabilitiesFromDB), func(i int) bool {
			return equalStrings(capabilitiesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.capabilitySvc.Delete(ctx, resourceType, capabilitiesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(integrationDependenciesFromDB), func(i int) bool {
			return equalStrings(integrationDependenciesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.integrationDependencySvc.Delete(ctx, resourceType, integrationDependenciesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
			return equalStrings(bundlesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := td.bundleSvc.Delete(ctx, resourceType, bundlesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
			return vendorsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := td.vendorSvc.Delete(ctx, resourceType, vendorsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
			return productsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := td.productSvc.Delete(ctx, resourceType, productsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		for i := range fetchRequests {
			if equalStrings(&fetchRequests[i].RefObjectOrdID, &ts.OrdID) {
				frIdxToExclude = append(frIdxToExclude, i)
			}
		}
	}

	return excludeUnnecessaryFetchRequests(fetchRequests, frIdxToExclude), tx.Commit()
}

func excludeUnnecessaryFetchRequests(fetchRequests []*OrdFetchRequest, frIdxToExclude []int) []*OrdFetchRequest {
	finalFetchRequests := make([]*OrdFetchRequest, 0)
	for i := range fetchRequests {
		shouldExclude := false
		for _, idxToExclude := range frIdxToExclude {
			if i == idxToExclude {
				shouldExclude = true
				break
			}
		}

		if !shouldExclude {
			finalFetchRequests = append(finalFetchRequests, fetchRequests[i])
		}
	}

	return finalFetchRequests
}
