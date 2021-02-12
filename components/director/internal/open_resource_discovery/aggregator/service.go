package aggregator

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	packageORDType = "package"
	bundleORDType  = "consumptionBundle"
	apiORDType     = "apiResource"
	eventORDType   = "eventResource"
	vendorORDType  = "vendor"
	productORDType = "product"
)

type Service struct {
	transact persistence.Transactioner

	appSvc       ApplicationService
	webhookSvc   WebhookService
	bundleSvc    BundleService
	apiSvc       APIService
	eventSvc     EventService
	specSvc      SpecService
	packageSvc   PackageService
	productSvc   ProductService
	vendorSvc    VendorService
	tombstoneSvc TombstoneService

	ordClient open_resource_discovery.Client
}

func NewService(transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, apiSvc APIService, eventSvc EventService, specSvc SpecService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, client open_resource_discovery.Client) *Service {
	return &Service{
		transact:     transact,
		appSvc:       appSvc,
		webhookSvc:   webhookSvc,
		bundleSvc:    bundleSvc,
		apiSvc:       apiSvc,
		eventSvc:     eventSvc,
		specSvc:      specSvc,
		packageSvc:   packageSvc,
		productSvc:   productSvc,
		vendorSvc:    vendorSvc,
		tombstoneSvc: tombstoneSvc,
		ordClient:    client,
	}
}

func (s *Service) SyncORDDocuments(ctx context.Context) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	pageCount := 1
	pageSize := 200
	page, err := s.appSvc.ListGlobal(ctx, pageSize, "")
	if err != nil {
		return errors.Wrapf(err, "error while fetching application page number %d", pageCount)
	}
	if err := s.processAppPage(ctx, page.Data); err != nil {
		return errors.Wrapf(err, "error while processing application page number %d", pageCount)
	}
	pageCount++

	for page.PageInfo.HasNextPage {
		page, err = s.appSvc.ListGlobal(ctx, pageSize, page.PageInfo.EndCursor)
		if err != nil {
			return errors.Wrapf(err, "error while fetching page number %d", pageCount)
		}
		if err := s.processAppPage(ctx, page.Data); err != nil {
			return errors.Wrapf(err, "error while processing page number %d", pageCount)
		}
		pageCount++
	}

	return tx.Commit()
}

func (s *Service) processAppPage(ctx context.Context, page []*model.Application) error {
	for _, app := range page {
		ctx = tenant.SaveToContext(ctx, app.Tenant, "")
		webhooks, err := s.webhookSvc.List(ctx, app.ID)
		if err != nil {
			return errors.Wrapf(err, "error fetching webhooks for app with id %q", app.ID)
		}
		documents := make([]*open_resource_discovery.Document, 0, 0)
		for _, wh := range webhooks {
			if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
				docs, err := s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, *wh.URL)
				if err != nil {
					return errors.Wrapf(err, "error fetching ORD document for webhook with id %q for app with id %q", wh.ID, app.ID)
				}
				documents = append(documents, docs...)
			}
		}
		if err := s.processDocuments(ctx, app.ID, documents); err != nil {
			return errors.Wrapf(err, "error processing ORD documents for app with id %q", app.ID)
		}
	}
	return nil
}

func (s *Service) processDocuments(ctx context.Context, appID string, documents open_resource_discovery.Documents) error {
	if err := documents.Validate(); err != nil {
		return errors.Wrap(err, "invalid documents")
	}

	if err := documents.Sanitize(); err != nil {
		return errors.Wrap(err, "while sanitizing ORD documents")
	}

	packagesFromDB, err := s.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing packages for app with id %q", appID)
	}

	bundlesFromDB, err := s.bundleSvc.ListByApplicationIDNoPaging(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing bundles for app with id %q", appID)
	}

	productsFromDB, err := s.productSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing products for app with id %q", appID)
	}

	vendorsFromDB, err := s.vendorSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing vendors for app with id %q", appID)
	}

	apisFromDB, err := s.apiSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing apis for app with id %q", appID)
	}

	eventsFromDB, err := s.eventSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing events for app with id %q", appID)
	}

	tombstonesFromDB, err := s.tombstoneSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while listing tombstones for app with id %q", appID)
	}

	for _, doc := range documents {
		for _, vendor := range doc.Vendors {
			if err := s.resyncVendor(ctx, appID, vendorsFromDB, vendor); err != nil {
				return errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
			}
		}
		for _, product := range doc.Products {
			if err := s.resyncProduct(ctx, appID, productsFromDB, product); err != nil {
				return errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
			}
		}
		for _, pkg := range doc.Packages {
			if err := s.resyncPackage(ctx, appID, packagesFromDB, pkg); err != nil {
				return errors.Wrapf(err, "error while resyncing package with ORD ID %q", pkg.OrdID)
			}
		}
		for _, bndl := range doc.ConsumptionBundles {
			if err := s.resyncBundle(ctx, appID, bundlesFromDB, bndl); err != nil {
				return errors.Wrapf(err, "error while resyncing bundle with ORD ID %q", *bndl.OrdID)
			}
		}
		for _, api := range doc.APIResources {
			if err := s.resyncAPI(ctx, appID, apisFromDB, bundlesFromDB, packagesFromDB, api); err != nil {
				return errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
			}
		}
		for _, event := range doc.EventResources {
			if err := s.resyncEvent(ctx, appID, eventsFromDB, bundlesFromDB, packagesFromDB, event); err != nil {
				return errors.Wrapf(err, "error while resyncing event with ORD ID %q", *event.OrdID)
			}
		}
		for _, tombstone := range doc.Tombstones {
			if err := s.resyncTombstones(ctx, appID, tombstonesFromDB, bundlesFromDB, packagesFromDB, apisFromDB, eventsFromDB, tombstone); err != nil {
				return errors.Wrapf(err, "error while resyncing tombstone for resource with ORD ID %q", tombstone.OrdID)
			}
		}
	}
	return nil
}

func (s *Service) resyncPackage(ctx context.Context, appID string, packagesFromDB []*model.Package, pkg model.PackageInput) error {
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return packagesFromDB[i].OrdID == pkg.OrdID
	}); found {
		return s.packageSvc.Update(ctx, packagesFromDB[i].ID, pkg)
	}
	_, err := s.packageSvc.Create(ctx, appID, pkg)
	return err
}

func (s *Service) resyncBundle(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput) error {
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		return s.bundleSvc.Update(ctx, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl))
	}
	_, err := s.bundleSvc.Create(ctx, appID, bndl)
	return err
}

func (s *Service) resyncProduct(ctx context.Context, appID string, productsFromDB []*model.Product, product model.ProductInput) error {
	if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
		return productsFromDB[i].OrdID == product.OrdID
	}); found {
		return s.productSvc.Update(ctx, productsFromDB[i].OrdID, product)
	}
	_, err := s.productSvc.Create(ctx, appID, product)
	return err
}

func (s *Service) resyncVendor(ctx context.Context, appID string, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return s.vendorSvc.Update(ctx, vendorsFromDB[i].OrdID, vendor)
	}
	_, err := s.vendorSvc.Create(ctx, appID, vendor)
	return err
}

func (s *Service) resyncAPI(ctx context.Context, appID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput) error {
	i, found := searchInSlice(len(apisFromDB), func(i int) bool {
		return equalStrings(apisFromDB[i].OrdID, api.OrdID)
	})

	var bundleID *string
	var packageID *string

	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, api.OrdBundleID)
	}); found {
		bundleID = &bundlesFromDB[i].ID
	}

	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, api.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	specs := make([]*model.SpecInput, 0, len(api.ResourceDefinitions))
	for _, resourceDef := range api.ResourceDefinitions {
		specs = append(specs, resourceDef.ToSpec())
	}

	if !found {
		_, err := s.apiSvc.Create(ctx, appID, bundleID, packageID, api, specs)
		return err
	}

	if err := s.apiSvc.Update(ctx, apisFromDB[i].ID, api, nil); err != nil {
		return err
	}
	if api.VersionInput.Value != apisFromDB[i].Version.Value {
		if err := s.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resyncEvent(ctx context.Context, appID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event model.EventDefinitionInput) error {
	i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
		return equalStrings(eventsFromDB[i].OrdID, event.OrdID)
	})

	var bundleID *string
	var packageID *string

	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, event.OrdBundleID)
	}); found {
		bundleID = &bundlesFromDB[i].ID
	}

	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, event.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	specs := make([]*model.SpecInput, 0, len(event.ResourceDefinitions))
	for _, resourceDef := range event.ResourceDefinitions {
		specs = append(specs, resourceDef.ToSpec())
	}

	if !found {
		_, err := s.eventSvc.Create(ctx, appID, bundleID, packageID, event, specs)
		return err
	}

	if err := s.eventSvc.Update(ctx, eventsFromDB[i].ID, event, nil); err != nil {
		return err
	}
	if event.VersionInput.Value != eventsFromDB[i].Version.Value {
		if err := s.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput) error {
	if err := s.specSvc.DeleteByReferenceObjectID(ctx, objectType, objectID); err != nil {
		return err
	}
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		if _, err := s.specSvc.CreateByReferenceObjectID(ctx, *spec, objectType, objectID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resyncTombstones(ctx context.Context, appID string, tombstonesFromDB []*model.Tombstone, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition, tombstone model.TombstoneInput) error {
	if i, found := searchInSlice(len(tombstonesFromDB), func(i int) bool {
		return tombstonesFromDB[i].OrdID == tombstone.OrdID
	}); found {
		return s.tombstoneSvc.Update(ctx, tombstonesFromDB[i].OrdID, tombstone)
	}

	if _, err := s.tombstoneSvc.Create(ctx, appID, tombstone); err != nil {
		return err
	}

	resourceType := strings.Split(tombstone.OrdID, ":")[1]
	switch resourceType {
	case packageORDType:
		if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
			return packagesFromDB[i].OrdID == tombstone.OrdID
		}); found {
			return s.packageSvc.Delete(ctx, packagesFromDB[i].ID)
		}
	case apiORDType:
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &tombstone.OrdID)
		}); found {
			return s.apiSvc.Delete(ctx, apisFromDB[i].ID)
		}
	case eventORDType:
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &tombstone.OrdID)
		}); found {
			return s.eventSvc.Delete(ctx, eventsFromDB[i].ID)
		}
	case vendorORDType:
		if err := s.vendorSvc.Delete(ctx, tombstone.OrdID); err != nil && !apperrors.IsNotFoundError(err) {
			return err
		}
	case productORDType:
		if err := s.productSvc.Delete(ctx, tombstone.OrdID); err != nil && !apperrors.IsNotFoundError(err) {
			return err
		}
	case bundleORDType:
		if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
			return equalStrings(bundlesFromDB[i].OrdID, &tombstone.OrdID)
		}); found {
			return s.bundleSvc.Delete(ctx, bundlesFromDB[i].ID)
		}
	}
	return nil
}

func bundleUpdateInputFromCreateInput(in model.BundleCreateInput) model.BundleUpdateInput {
	return model.BundleUpdateInput{
		Name:                           in.Name,
		Description:                    in.Description,
		InstanceAuthRequestInputSchema: in.InstanceAuthRequestInputSchema,
		DefaultInstanceAuth:            in.DefaultInstanceAuth,
		OrdID:                          in.OrdID,
		ShortDescription:               in.ShortDescription,
		Links:                          in.Links,
		Labels:                         in.Labels,
		CredentialExchangeStrategies:   in.CredentialExchangeStrategies,
	}
}

func equalStrings(first, second *string) bool {
	return first != nil && second != nil && *first == *second
}

func searchInSlice(length int, f func(i int) bool) (int, bool) {
	for i := 0; i < length; i++ {
		if f(i) {
			return i, true
		}
	}
	return -1, false
}
