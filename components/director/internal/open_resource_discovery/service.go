package ord

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const applicationTypeLabel = "applicationType"

// ServiceConfig contains configuration for the ORD aggregator service
type ServiceConfig struct {
	maxParallelApplicationProcessors int
}

// NewServiceConfig creates new ServiceConfig from the supplied parameters
func NewServiceConfig(maxParallelApplicationProcessors int) ServiceConfig {
	return ServiceConfig{
		maxParallelApplicationProcessors: maxParallelApplicationProcessors,
	}
}

// Service consists of various resource services responsible for service-layer ORD operations.
type Service struct {
	config ServiceConfig

	transact persistence.Transactioner

	labelRepo labelRepository

	appSvc             ApplicationService
	webhookSvc         WebhookService
	bundleSvc          BundleService
	bundleReferenceSvc BundleReferenceService
	apiSvc             APIService
	eventSvc           EventService
	specSvc            SpecService
	packageSvc         PackageService
	productSvc         ProductService
	vendorSvc          VendorService
	tombstoneSvc       TombstoneService
	tenantSvc          TenantService

	globalRegistrySvc GlobalRegistryService
	ordClient         Client
}

// NewAggregatorService returns a new object responsible for service-layer ORD operations.
func NewAggregatorService(config ServiceConfig, transact persistence.Transactioner, labelRepo labelRepository, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, eventSvc EventService, specSvc SpecService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, tenantSvc TenantService, globalRegistrySvc GlobalRegistryService, client Client) *Service {
	return &Service{
		config:             config,
		transact:           transact,
		appSvc:             appSvc,
		labelRepo:          labelRepo,
		webhookSvc:         webhookSvc,
		bundleSvc:          bundleSvc,
		bundleReferenceSvc: bundleReferenceSvc,
		apiSvc:             apiSvc,
		eventSvc:           eventSvc,
		specSvc:            specSvc,
		packageSvc:         packageSvc,
		productSvc:         productSvc,
		vendorSvc:          vendorSvc,
		tombstoneSvc:       tombstoneSvc,
		tenantSvc:          tenantSvc,
		globalRegistrySvc:  globalRegistrySvc,
		ordClient:          client,
	}
}

// SyncORDDocuments performs resync of ORD information provided via ORD documents for each application
func (s *Service) SyncORDDocuments(ctx context.Context) error {
	globalResourcesOrdIDs, err := s.globalRegistrySvc.SyncGlobalResources(ctx)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("Error while synchronizing global resources: %s. Proceeding with already existing global resources...", err)
		globalResourcesOrdIDs, err = s.globalRegistrySvc.ListGlobalResources(ctx)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("Error while listing existing global resource: %s. Proceeding with empty globalResourceOrdIDs... Validation of Documents relying on global resources might fail.", err)
		}
	}

	if globalResourcesOrdIDs == nil {
		globalResourcesOrdIDs = make(map[string]bool)
	}

	pageCount := 1
	pageSize := 200

	pageCursor := ""
	hasNextPage := true

	queue := make(chan *model.Application)
	appErrors := int32(0)

	wg := &sync.WaitGroup{}
	wg.Add(s.config.maxParallelApplicationProcessors)

	log.C(ctx).Infof("Starting %d workers...", s.config.maxParallelApplicationProcessors)
	for i := 0; i < s.config.maxParallelApplicationProcessors; i++ {
		go func() {
			defer wg.Done()

			for app := range queue {
				if err := s.processApp(ctx, app, globalResourcesOrdIDs); err != nil {
					log.C(ctx).WithError(err).Errorf("error while processing app %q", app.ID)
					atomic.AddInt32(&appErrors, 1)
				}
			}
		}()
	}

	var fetchErr error = nil
	for hasNextPage {
		page, err := s.listAppPage(ctx, pageSize, pageCursor)
		if err != nil {
			log.C(ctx).WithError(err).Errorf("error while fetching application page number %d", pageCount)
			fetchErr = err
			break
		}

		for _, app := range page.Data {
			queue <- app
		}

		pageCursor = page.PageInfo.EndCursor
		hasNextPage = page.PageInfo.HasNextPage
		pageCount++
	}

	close(queue)
	wg.Wait()

	if fetchErr != nil && appErrors != 0 {
		return errors.Wrapf(fetchErr, "failed to process %d applications and failed to fetch the next application page", appErrors)
	} else if appErrors != 0 {
		return errors.Errorf("failed to process %d applications", appErrors)
	} else if fetchErr != nil {
		return errors.Wrapf(fetchErr, "failed to fetch the next application page")
	}
	return nil
}

func (s *Service) listAppPage(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)
	page, err := s.appSvc.ListGlobal(ctx, pageSize, cursor)
	if err != nil {
		return nil, err
	}

	if len(page.Data) > 0 {
		applicationIDs := make([]string, 0, page.TotalCount)
		for _, app := range page.Data {
			applicationIDs = append(applicationIDs, app.ID)
		}

		labels, err := s.labelRepo.ListGlobalByKeyAndObjects(ctx, model.ApplicationLabelableObject, applicationIDs, applicationTypeLabel)
		if err != nil {
			return nil, err
		}

		appLabelsMap := make(map[string]interface{}, page.TotalCount)
		for i := range labels {
			appLabelsMap[labels[i].ObjectID] = labels[i].Value
		}

		for i := range page.Data {
			appType, ok := appLabelsMap[page.Data[i].ID].(string)
			if ok {
				page.Data[i].Type = appType
			}
		}
	}

	return page, tx.Commit()
}

func (s *Service) processApp(ctx context.Context, app *model.Application, globalResourcesOrdIDs map[string]bool) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	tnt, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Application, app.ID)
	if err != nil {
		return err
	}

	ctx = tenant.SaveToContext(ctx, tnt, "")

	if _, err := s.appSvc.GetForUpdate(ctx, app.ID); err != nil {
		return errors.Wrapf(err, "error while locking app with id %q for update", app.ID)
	}

	webhooks, err := s.webhookSvc.ListForApplicationWithSelectForUpdate(ctx, app.ID)
	if err != nil {
		return errors.Wrapf(err, "error fetching webhooks for app with id %q", app.ID)
	}
	var documents Documents
	var baseURL string
	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			ctx = addFieldToLogger(ctx, "app_id", app.ID)
			documents, baseURL, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, app, wh)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("error fetching ORD document for webhook with id %q: %v", wh.ID, err)
			}
			break
		}
	}
	if len(documents) > 0 {
		log.C(ctx).Info("Processing ORD documents")
		if err := s.processDocuments(ctx, app.ID, baseURL, documents, globalResourcesOrdIDs); err != nil {
			log.C(ctx).WithError(err).Errorf("error processing ORD documents: %v", err)
		} else {
			log.C(ctx).Info("Successfully processed ORD documents")
			return tx.Commit()
		}
	}
	return nil
}

func (s *Service) processDocuments(ctx context.Context, appID string, baseURL string, documents Documents, globalResourcesOrdIDs map[string]bool) error {
	apiDataFromDB, eventDataFromDB, packageDataFromDB, err := s.fetchResources(ctx, appID)
	if err != nil {
		return err
	}

	resourceHashes, err := hashResources(documents)
	if err != nil {
		return err
	}

	if err := documents.Validate(baseURL, apiDataFromDB, eventDataFromDB, packageDataFromDB, resourceHashes, globalResourcesOrdIDs); err != nil {
		return errors.Wrap(err, "invalid documents")
	}

	if err := documents.Sanitize(baseURL); err != nil {
		return errors.Wrap(err, "while sanitizing ORD documents")
	}

	vendorsInput := make([]*model.VendorInput, 0)
	productsInput := make([]*model.ProductInput, 0)
	packagesInput := make([]*model.PackageInput, 0)
	bundlesInput := make([]*model.BundleCreateInput, 0)
	apisInput := make([]*model.APIDefinitionInput, 0)
	eventsInput := make([]*model.EventDefinitionInput, 0)
	tombstonesInput := make([]*model.TombstoneInput, 0)
	for _, doc := range documents {
		vendorsInput = append(vendorsInput, doc.Vendors...)
		productsInput = append(productsInput, doc.Products...)
		packagesInput = append(packagesInput, doc.Packages...)
		bundlesInput = append(bundlesInput, doc.ConsumptionBundles...)
		apisInput = append(apisInput, doc.APIResources...)
		eventsInput = append(eventsInput, doc.EventResources...)
		tombstonesInput = append(tombstonesInput, doc.Tombstones...)
	}

	vendorsFromDB, err := s.processVendors(ctx, appID, vendorsInput)
	if err != nil {
		return err
	}

	productsFromDB, err := s.processProducts(ctx, appID, productsInput)
	if err != nil {
		return err
	}

	packagesFromDB, err := s.processPackages(ctx, appID, packagesInput, resourceHashes)
	if err != nil {
		return err
	}

	bundlesFromDB, err := s.processBundles(ctx, appID, bundlesInput)
	if err != nil {
		return err
	}

	apisFromDB, err := s.processAPIs(ctx, appID, bundlesFromDB, packagesFromDB, apisInput, resourceHashes)
	if err != nil {
		return err
	}

	eventsFromDB, err := s.processEvents(ctx, appID, bundlesFromDB, packagesFromDB, eventsInput, resourceHashes)
	if err != nil {
		return err
	}

	tombstonesFromDB, err := s.processTombstones(ctx, appID, tombstonesInput)
	if err != nil {
		return err
	}

	for _, ts := range tombstonesFromDB {
		if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
			return packagesFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.packageSvc.Delete(ctx, packagesFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.apiSvc.Delete(ctx, apisFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.eventSvc.Delete(ctx, eventsFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
			return equalStrings(bundlesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.bundleSvc.Delete(ctx, bundlesFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
			return vendorsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.vendorSvc.Delete(ctx, vendorsFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
			return productsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.productSvc.Delete(ctx, productsFromDB[i].ID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
	}

	return nil
}

func (s *Service) processVendors(ctx context.Context, appID string, vendors []*model.VendorInput) ([]*model.Vendor, error) {
	vendorsFromDB, err := s.vendorSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing vendors for app with id %q", appID)
	}

	for _, vendor := range vendors {
		if err := s.resyncVendor(ctx, appID, vendorsFromDB, *vendor); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
		}
	}

	return s.vendorSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processProducts(ctx context.Context, appID string, products []*model.ProductInput) ([]*model.Product, error) {
	productsFromDB, err := s.productSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing products for app with id %q", appID)
	}

	for _, product := range products {
		if err := s.resyncProduct(ctx, appID, productsFromDB, *product); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
		}
	}
	return s.productSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processPackages(ctx context.Context, appID string, packages []*model.PackageInput, resourceHashes map[string]uint64) ([]*model.Package, error) {
	packagesFromDB, err := s.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing packages for app with id %q", appID)
	}

	for _, pkg := range packages {
		pkgHash := resourceHashes[pkg.OrdID]
		if err := s.resyncPackage(ctx, appID, packagesFromDB, *pkg, pkgHash); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing package with ORD ID %q", pkg.OrdID)
		}
	}

	return s.packageSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processBundles(ctx context.Context, appID string, bundles []*model.BundleCreateInput) ([]*model.Bundle, error) {
	bundlesFromDB, err := s.bundleSvc.ListByApplicationIDNoPaging(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing bundles for app with id %q", appID)
	}

	for _, bndl := range bundles {
		if err := s.resyncBundle(ctx, appID, bundlesFromDB, *bndl); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing bundle with ORD ID %q", *bndl.OrdID)
		}
	}

	return s.bundleSvc.ListByApplicationIDNoPaging(ctx, appID)
}

func (s *Service) processAPIs(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apis []*model.APIDefinitionInput, resourceHashes map[string]uint64) ([]*model.APIDefinition, error) {
	apisFromDB, err := s.apiSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing apis for app with id %q", appID)
	}

	for _, api := range apis {
		apiHash := resourceHashes[str.PtrStrToStr(api.OrdID)]
		if err := s.resyncAPI(ctx, appID, apisFromDB, bundlesFromDB, packagesFromDB, *api, apiHash); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
		}
	}

	return s.apiSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processEvents(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput, resourceHashes map[string]uint64) ([]*model.EventDefinition, error) {
	eventsFromDB, err := s.eventSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing events for app with id %q", appID)
	}

	for _, event := range events {
		eventHash := resourceHashes[str.PtrStrToStr(event.OrdID)]
		if err := s.resyncEvent(ctx, appID, eventsFromDB, bundlesFromDB, packagesFromDB, *event, eventHash); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing event with ORD ID %q", *event.OrdID)
		}
	}

	return s.eventSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processTombstones(ctx context.Context, appID string, tombstones []*model.TombstoneInput) ([]*model.Tombstone, error) {
	tombstonesFromDB, err := s.tombstoneSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing tombstones for app with id %q", appID)
	}

	for _, tombstone := range tombstones {
		if err := s.resyncTombstone(ctx, appID, tombstonesFromDB, *tombstone); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing tombstone for resource with ORD ID %q", tombstone.OrdID)
		}
	}

	return s.tombstoneSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) resyncPackage(ctx context.Context, appID string, packagesFromDB []*model.Package, pkg model.PackageInput, pkgHash uint64) error {
	ctx = addFieldToLogger(ctx, "package_ord_id", pkg.OrdID)
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return packagesFromDB[i].OrdID == pkg.OrdID
	}); found {
		return s.packageSvc.Update(ctx, packagesFromDB[i].ID, pkg, pkgHash)
	}
	_, err := s.packageSvc.Create(ctx, appID, pkg, pkgHash)
	return err
}

func (s *Service) resyncBundle(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput) error {
	ctx = addFieldToLogger(ctx, "bundle_ord_id", *bndl.OrdID)
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		return s.bundleSvc.Update(ctx, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl))
	}
	_, err := s.bundleSvc.Create(ctx, appID, bndl)
	return err
}

func (s *Service) resyncProduct(ctx context.Context, appID string, productsFromDB []*model.Product, product model.ProductInput) error {
	ctx = addFieldToLogger(ctx, "product_ord_id", product.OrdID)
	if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
		return productsFromDB[i].OrdID == product.OrdID
	}); found {
		return s.productSvc.Update(ctx, productsFromDB[i].ID, product)
	}
	_, err := s.productSvc.Create(ctx, appID, product)
	return err
}

func (s *Service) resyncVendor(ctx context.Context, appID string, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	ctx = addFieldToLogger(ctx, "vendor_ord_id", vendor.OrdID)
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return s.vendorSvc.Update(ctx, vendorsFromDB[i].ID, vendor)
	}
	_, err := s.vendorSvc.Create(ctx, appID, vendor)
	return err
}

func (s *Service) resyncAPI(ctx context.Context, appID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput, apiHash uint64) error {
	ctx = addFieldToLogger(ctx, "api_ord_id", *api.OrdID)
	i, isAPIFound := searchInSlice(len(apisFromDB), func(i int) bool {
		return equalStrings(apisFromDB[i].OrdID, api.OrdID)
	})

	defaultConsumptionBundleID := extractDefaultConsumptionBundle(bundlesFromDB, api.DefaultConsumptionBundle)
	defaultTargetURLPerBundle := extractAllBundleReferencesForAPI(bundlesFromDB, api)

	var packageID *string
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, api.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	specs := make([]*model.SpecInput, 0, len(api.ResourceDefinitions))
	for _, resourceDef := range api.ResourceDefinitions {
		specs = append(specs, resourceDef.ToSpec())
	}

	if !isAPIFound {
		_, err := s.apiSvc.Create(ctx, appID, nil, packageID, api, specs, defaultTargetURLPerBundle, apiHash, defaultConsumptionBundleID)
		return err
	}

	allBundleIDsForAPI, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &apisFromDB[i].ID)
	if err != nil {
		return err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion := extractBundleReferencesForDeletion(allBundleIDsForAPI, defaultTargetURLPerBundle)

	// in case of API update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	defaultTargetURLPerBundleForCreation := extractAllBundleReferencesForCreation(defaultTargetURLPerBundle, allBundleIDsForAPI)

	if err := s.apiSvc.UpdateInManyBundles(ctx, apisFromDB[i].ID, api, nil, defaultTargetURLPerBundle, defaultTargetURLPerBundleForCreation, bundleIDsForDeletion, apiHash, defaultConsumptionBundleID); err != nil {
		return err
	}
	if api.VersionInput.Value != apisFromDB[i].Version.Value {
		if err := s.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs); err != nil {
			return err
		}
	} else {
		if err := s.refetchFailedSpecs(ctx, model.APISpecReference, apisFromDB[i].ID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) resyncEvent(ctx context.Context, appID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event model.EventDefinitionInput, eventHash uint64) error {
	ctx = addFieldToLogger(ctx, "event_ord_id", *event.OrdID)
	i, isEventFound := searchInSlice(len(eventsFromDB), func(i int) bool {
		return equalStrings(eventsFromDB[i].OrdID, event.OrdID)
	})

	defaultConsumptionBundleID := extractDefaultConsumptionBundle(bundlesFromDB, event.DefaultConsumptionBundle)

	bundleIDsFromBundleReference := make([]string, 0)
	for _, br := range event.PartOfConsumptionBundles {
		for _, bndl := range bundlesFromDB {
			if equalStrings(bndl.OrdID, &br.BundleOrdID) {
				bundleIDsFromBundleReference = append(bundleIDsFromBundleReference, bndl.ID)
			}
		}
	}

	var packageID *string
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, event.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	specs := make([]*model.SpecInput, 0, len(event.ResourceDefinitions))
	for _, resourceDef := range event.ResourceDefinitions {
		specs = append(specs, resourceDef.ToSpec())
	}

	if !isEventFound {
		_, err := s.eventSvc.Create(ctx, appID, nil, packageID, event, specs, bundleIDsFromBundleReference, eventHash, defaultConsumptionBundleID)
		return err
	}

	allBundleIDsForEvent, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleEventReference, &eventsFromDB[i].ID)
	if err != nil {
		return err
	}

	// in case of Event update, we need to filter which ConsumptionBundleReferences(bundle IDs) should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion := make([]string, 0)
	for _, id := range allBundleIDsForEvent {
		if _, found := searchInSlice(len(bundleIDsFromBundleReference), func(i int) bool {
			return equalStrings(&bundleIDsFromBundleReference[i], &id)
		}); !found {
			bundleIDsForDeletion = append(bundleIDsForDeletion, id)
		}
	}

	// in case of Event update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	bundleIDsForCreation := make([]string, 0)
	for _, id := range bundleIDsFromBundleReference {
		if _, found := searchInSlice(len(allBundleIDsForEvent), func(i int) bool {
			return equalStrings(&allBundleIDsForEvent[i], &id)
		}); !found {
			bundleIDsForCreation = append(bundleIDsForCreation, id)
		}
	}

	if err := s.eventSvc.UpdateInManyBundles(ctx, eventsFromDB[i].ID, event, nil, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion, eventHash, defaultConsumptionBundleID); err != nil {
		return err
	}
	if event.VersionInput.Value != eventsFromDB[i].Version.Value {
		if err := s.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs); err != nil {
			return err
		}
	} else {
		if err := s.refetchFailedSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID); err != nil {
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

func (s *Service) resyncTombstone(ctx context.Context, appID string, tombstonesFromDB []*model.Tombstone, tombstone model.TombstoneInput) error {
	if i, found := searchInSlice(len(tombstonesFromDB), func(i int) bool {
		return tombstonesFromDB[i].OrdID == tombstone.OrdID
	}); found {
		return s.tombstoneSvc.Update(ctx, tombstonesFromDB[i].ID, tombstone)
	}
	_, err := s.tombstoneSvc.Create(ctx, appID, tombstone)
	return err
}

func (s *Service) refetchFailedSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error {
	specsFromDB, err := s.specSvc.ListByReferenceObjectID(ctx, objectType, objectID)
	if err != nil {
		return err
	}
	for _, spec := range specsFromDB {
		fr, err := s.specSvc.GetFetchRequest(ctx, spec.ID, objectType)
		if err != nil {
			return err
		}
		if fr.Status != nil && fr.Status.Condition == model.FetchRequestStatusConditionFailed {
			if _, err := s.specSvc.RefetchSpec(ctx, spec.ID, objectType); err != nil {
				return errors.Wrapf(err, "while refetching spec %s", spec.ID)
			}
		}
	}
	return nil
}

func (s *Service) fetchAPIDefFromDB(ctx context.Context, appID string) (map[string]*model.APIDefinition, error) {
	apisFromDB, err := s.apiSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing apis for app with id %s", appID)
	}

	apiDataFromDB := make(map[string]*model.APIDefinition, len(apisFromDB))

	for _, api := range apisFromDB {
		apiOrdID := str.PtrStrToStr(api.OrdID)
		apiDataFromDB[apiOrdID] = api
	}

	return apiDataFromDB, nil
}

func (s *Service) fetchPackagesFromDB(ctx context.Context, appID string) (map[string]*model.Package, error) {
	packagesFromDB, err := s.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing packages for app with id %s", appID)
	}

	packageDataFromDB := make(map[string]*model.Package)

	for _, pkg := range packagesFromDB {
		packageDataFromDB[pkg.OrdID] = pkg
	}

	return packageDataFromDB, nil
}

func (s *Service) fetchEventDefFromDB(ctx context.Context, appID string) (map[string]*model.EventDefinition, error) {
	eventsFromDB, err := s.eventSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing events for app with id %s", appID)
	}

	eventDataFromDB := make(map[string]*model.EventDefinition)

	for _, event := range eventsFromDB {
		eventOrdID := str.PtrStrToStr(event.OrdID)
		eventDataFromDB[eventOrdID] = event
	}

	return eventDataFromDB, nil
}

func (s *Service) fetchResources(ctx context.Context, appID string) (map[string]*model.APIDefinition, map[string]*model.EventDefinition, map[string]*model.Package, error) {
	apiDataFromDB, err := s.fetchAPIDefFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "while fetching apis for app with id %s", appID)
	}

	eventDataFromDB, err := s.fetchEventDefFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "while fetching events for app with id %s", appID)
	}

	packageDataFromDB, err := s.fetchPackagesFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, errors.Wrapf(err, "while fetching packages for app with id %s", appID)
	}

	return apiDataFromDB, eventDataFromDB, packageDataFromDB, nil
}

func hashResources(docs Documents) (map[string]uint64, error) {
	resourceHashes := make(map[string]uint64)

	for _, doc := range docs {
		for _, apiInput := range doc.APIResources {
			normalizedAPIDef, err := normalizeAPIDefinition(apiInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedAPIDef)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing api definition with ORD ID: %s", str.PtrStrToStr(normalizedAPIDef.OrdID))
			}

			resourceHashes[str.PtrStrToStr(apiInput.OrdID)] = hash
		}

		for _, eventInput := range doc.EventResources {
			normalizedEventDef, err := normalizeEventDefinition(eventInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedEventDef)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing event definition with ORD ID: %s", str.PtrStrToStr(normalizedEventDef.OrdID))
			}

			resourceHashes[str.PtrStrToStr(eventInput.OrdID)] = hash
		}

		for _, packageInput := range doc.Packages {
			normalizedPkg, err := normalizePackage(packageInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedPkg)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing package with ORD ID: %s", normalizedPkg.OrdID)
			}

			resourceHashes[packageInput.OrdID] = hash
		}
	}

	return resourceHashes, nil
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
		DocumentationLabels:            in.DocumentationLabels,
		CredentialExchangeStrategies:   in.CredentialExchangeStrategies,
		CorrelationIDs:                 in.CorrelationIDs,
	}
}

// extractDefaultConsumptionBundle converts the defaultConsumptionBundle which is bundle ORD_ID into internal bundle_id
func extractDefaultConsumptionBundle(bundlesFromDB []*model.Bundle, defaultConsumptionBundle *string) string {
	var bundleID string
	if defaultConsumptionBundle == nil {
		return bundleID
	}

	for _, bndl := range bundlesFromDB {
		if equalStrings(bndl.OrdID, defaultConsumptionBundle) {
			bundleID = bndl.ID
			break
		}
	}
	return bundleID
}

func extractAllBundleReferencesForAPI(bundlesFromDB []*model.Bundle, api model.APIDefinitionInput) map[string]string {
	defaultTargetURLPerBundle := make(map[string]string)
	lenTargetURLs := len(gjson.ParseBytes(api.TargetURLs).Array())
	for _, br := range api.PartOfConsumptionBundles {
		for _, bndl := range bundlesFromDB {
			if equalStrings(bndl.OrdID, &br.BundleOrdID) {
				if br.DefaultTargetURL == "" && lenTargetURLs == 1 {
					defaultTargetURLPerBundle[bndl.ID] = gjson.ParseBytes(api.TargetURLs).Array()[0].String()
				} else {
					defaultTargetURLPerBundle[bndl.ID] = br.DefaultTargetURL
				}
			}
		}
	}
	return defaultTargetURLPerBundle
}

func extractAllBundleReferencesForCreation(defaultTargetURLPerBundle map[string]string, allBundleIDsForAPI []string) map[string]string {
	defaultTargetURLPerBundleForCreation := make(map[string]string)
	for bndlID, defaultEntryPoint := range defaultTargetURLPerBundle {
		if _, found := searchInSlice(len(allBundleIDsForAPI), func(i int) bool {
			return equalStrings(&allBundleIDsForAPI[i], &bndlID)
		}); !found {
			defaultTargetURLPerBundleForCreation[bndlID] = defaultEntryPoint
			delete(defaultTargetURLPerBundle, bndlID)
		}
	}
	return defaultTargetURLPerBundleForCreation
}

func extractBundleReferencesForDeletion(allBundleIDsForAPI []string, defaultTargetURLPerBundle map[string]string) []string {
	bundleIDsToBeDeleted := make([]string, 0)

	for _, bndlID := range allBundleIDsForAPI {
		if _, ok := defaultTargetURLPerBundle[bndlID]; !ok {
			bundleIDsToBeDeleted = append(bundleIDsToBeDeleted, bndlID)
		}
	}

	return bundleIDsToBeDeleted
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

func addFieldToLogger(ctx context.Context, fieldName, fieldValue string) context.Context {
	logger := log.LoggerFromContext(ctx)
	logger = logger.WithField(fieldName, fieldValue)
	return log.ContextWithLogger(ctx, logger)
}
