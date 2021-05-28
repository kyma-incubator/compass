package open_resource_discovery

import (
	"context"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type Service struct {
	transact persistence.Transactioner

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

	ordClient Client
}

func NewAggregatorService(transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, eventSvc EventService, specSvc SpecService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, client Client) *Service {
	return &Service{
		transact:           transact,
		appSvc:             appSvc,
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
		ordClient:          client,
	}
}

// SyncORDDocuments performs resync of ORD information provided via ORD documents for each application
func (s *Service) SyncORDDocuments(ctx context.Context) error {
	pageCount := 1
	pageSize := 200

	pageCursor := ""
	hasNextPage := true

	for hasNextPage {
		page, err := s.listAppPage(ctx, pageSize, pageCursor)
		if err != nil {
			return errors.Wrapf(err, "error while fetching application page number %d", pageCount)
		}
		for _, app := range page.Data {
			if err := s.processApp(ctx, app); err != nil {
				return errors.Wrapf(err, "error while processing app %q", app.ID)
			}
		}
		pageCursor = page.PageInfo.EndCursor
		hasNextPage = page.PageInfo.HasNextPage
		pageCount++
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
	return page, tx.Commit()
}

func (s *Service) processApp(ctx context.Context, app *model.Application) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	ctx = tenant.SaveToContext(ctx, app.Tenant, "")

	if err := ValidateSystemInstanceInput(app); err != nil {
		log.C(ctx).WithError(err).Errorf("error validating app %q", app.ID)
	}

	webhooks, err := s.webhookSvc.ListForApplication(ctx, app.ID)
	if err != nil {
		return errors.Wrapf(err, "error fetching webhooks for app with id %q", app.ID)
	}
	var documents Documents
	var baseURL string
	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			ctx = addFieldToLogger(ctx, "app_id", app.ID)
			documents, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, *wh.URL)
			if err != nil {
				log.C(ctx).WithError(err).Errorf("error fetching ORD document for webhook with id %q: %v", wh.ID, err)
			}
			baseURL = *wh.URL
			break
		}
	}
	if len(documents) > 0 {
		log.C(ctx).Info("Processing ORD documents")
		if err := s.processDocuments(ctx, app.ID, baseURL, documents); err != nil {
			log.C(ctx).WithError(err).Errorf("error processing ORD documents: %v", err)
		} else {
			log.C(ctx).Info("Successfully processed ORD documents")
			return tx.Commit()
		}
	}
	return nil
}

func (s *Service) processDocuments(ctx context.Context, appID string, baseURL string, documents Documents) error {
	if err := documents.Validate(baseURL); err != nil {
		return errors.Wrap(err, "invalid documents")
	}

	if err := documents.Sanitize(baseURL); err != nil {
		return errors.Wrap(err, "while sanitizing ORD documents")
	}

	vendorsInput := make([]*model.VendorInput, 0, 0)
	productsInput := make([]*model.ProductInput, 0, 0)
	packagesInput := make([]*model.PackageInput, 0, 0)
	bundlesInput := make([]*model.BundleCreateInput, 0, 0)
	apisInput := make([]*model.APIDefinitionInput, 0, 0)
	eventsInput := make([]*model.EventDefinitionInput, 0, 0)
	tombstonesInput := make([]*model.TombstoneInput, 0, 0)
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

	packagesFromDB, err := s.processPackages(ctx, appID, packagesInput)
	if err != nil {
		return err
	}

	bundlesFromDB, err := s.processBundles(ctx, appID, bundlesInput)
	if err != nil {
		return err
	}

	apisFromDB, err := s.processAPIs(ctx, appID, bundlesFromDB, packagesFromDB, apisInput)
	if err != nil {
		return err
	}

	eventsFromDB, err := s.processEvents(ctx, appID, bundlesFromDB, packagesFromDB, eventsInput)
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
			if err := s.vendorSvc.Delete(ctx, vendorsFromDB[i].OrdID); err != nil {
				return errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
			return productsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.productSvc.Delete(ctx, productsFromDB[i].OrdID); err != nil {
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

func (s *Service) processPackages(ctx context.Context, appID string, packages []*model.PackageInput) ([]*model.Package, error) {
	packagesFromDB, err := s.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing packages for app with id %q", appID)
	}

	for _, pkg := range packages {
		if err := s.resyncPackage(ctx, appID, packagesFromDB, *pkg); err != nil {
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

func (s *Service) processAPIs(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apis []*model.APIDefinitionInput) ([]*model.APIDefinition, error) {
	apisFromDB, err := s.apiSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing apis for app with id %q", appID)
	}

	for _, api := range apis {
		if err := s.resyncAPI(ctx, appID, apisFromDB, bundlesFromDB, packagesFromDB, *api); err != nil {
			return nil, errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
		}
	}

	return s.apiSvc.ListByApplicationID(ctx, appID)
}

func (s *Service) processEvents(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput) ([]*model.EventDefinition, error) {
	eventsFromDB, err := s.eventSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing events for app with id %q", appID)
	}

	for _, event := range events {
		if err := s.resyncEvent(ctx, appID, eventsFromDB, bundlesFromDB, packagesFromDB, *event); err != nil {
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

func (s *Service) resyncPackage(ctx context.Context, appID string, packagesFromDB []*model.Package, pkg model.PackageInput) error {
	ctx = addFieldToLogger(ctx, "package_ord_id", pkg.OrdID)
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return packagesFromDB[i].OrdID == pkg.OrdID
	}); found {
		return s.packageSvc.Update(ctx, packagesFromDB[i].ID, pkg)
	}
	_, err := s.packageSvc.Create(ctx, appID, pkg)
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
		return s.productSvc.Update(ctx, productsFromDB[i].OrdID, product)
	}
	_, err := s.productSvc.Create(ctx, appID, product)
	return err
}

func (s *Service) resyncVendor(ctx context.Context, appID string, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	ctx = addFieldToLogger(ctx, "vendor_ord_id", vendor.OrdID)
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return s.vendorSvc.Update(ctx, vendorsFromDB[i].OrdID, vendor)
	}
	_, err := s.vendorSvc.Create(ctx, appID, vendor)
	return err
}

func (s *Service) resyncAPI(ctx context.Context, appID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput) error {
	ctx = addFieldToLogger(ctx, "api_ord_id", *api.OrdID)
	i, isAPIFound := searchInSlice(len(apisFromDB), func(i int) bool {
		return equalStrings(apisFromDB[i].OrdID, api.OrdID)
	})

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
		_, err := s.apiSvc.Create(ctx, appID, nil, packageID, api, specs, defaultTargetURLPerBundle)
		return err
	}

	allBundleIDsForAPI, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &apisFromDB[i].ID)
	if err != nil {
		return err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion, err := extractBundleReferencesForDeletion(allBundleIDsForAPI, defaultTargetURLPerBundle)
	if err != nil {
		return err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	defaultTargetURLPerBundleForCreation := extractAllBundleReferencesForCreation(defaultTargetURLPerBundle, allBundleIDsForAPI)

	if err := s.apiSvc.UpdateInManyBundles(ctx, apisFromDB[i].ID, api, nil, defaultTargetURLPerBundle, defaultTargetURLPerBundleForCreation, bundleIDsForDeletion); err != nil {
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
	ctx = addFieldToLogger(ctx, "event_ord_id", *event.OrdID)
	i, isEventFound := searchInSlice(len(eventsFromDB), func(i int) bool {
		return equalStrings(eventsFromDB[i].OrdID, event.OrdID)
	})

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
		_, err := s.eventSvc.Create(ctx, appID, nil, packageID, event, specs, bundleIDsFromBundleReference)
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

	if err := s.eventSvc.UpdateInManyBundles(ctx, eventsFromDB[i].ID, event, nil, bundleIDsForCreation, bundleIDsForDeletion); err != nil {
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

func (s *Service) resyncTombstone(ctx context.Context, appID string, tombstonesFromDB []*model.Tombstone, tombstone model.TombstoneInput) error {
	if i, found := searchInSlice(len(tombstonesFromDB), func(i int) bool {
		return tombstonesFromDB[i].OrdID == tombstone.OrdID
	}); found {
		return s.tombstoneSvc.Update(ctx, tombstonesFromDB[i].OrdID, tombstone)
	}
	_, err := s.tombstoneSvc.Create(ctx, appID, tombstone)
	return err
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

func extractBundleReferencesForDeletion(allBundleIDsForAPI []string, defaultTargetURLPerBundle map[string]string) ([]string, error) {
	bundleIDsToBeDeleted := make([]string, 0)

	for _, bndlID := range allBundleIDsForAPI {
		if _, ok := defaultTargetURLPerBundle[bndlID]; !ok {
			bundleIDsToBeDeleted = append(bundleIDsToBeDeleted, bndlID)
		}
	}

	return bundleIDsToBeDeleted, nil
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
