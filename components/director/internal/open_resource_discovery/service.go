package ord

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	// MultiErrorSeparator represents the separator for splitting multi error into slice of validation errors
	MultiErrorSeparator string = "* "

	tenantMappingCustomTypeIdentifier = "sap.ucl:tenant_mapping"
	customTypeProperty                = "customType"
	callbackURLProperty               = "callbackUrl"
)

// ServiceConfig contains configuration for the ORD aggregator service
type ServiceConfig struct {
	maxParallelWebhookProcessors       int
	maxParallelSpecificationProcessors int
	ordWebhookPartialProcessMaxDays    int
	ordWebhookPartialProcessURL        string
	ordWebhookPartialProcessing        bool
}

type ordFetchRequest struct {
	*model.FetchRequest
	refObjectOrdID string
}

type fetchRequestResult struct {
	fetchRequest *model.FetchRequest
	data         *string
	status       *model.FetchRequestStatus
}

// MetricsConfig is the ord aggregator configuration for pushing metrics to Prometheus
type MetricsConfig struct {
	PushEndpoint  string        `envconfig:"optional,APP_METRICS_PUSH_ENDPOINT"`
	ClientTimeout time.Duration `envconfig:"default=60s"`
	JobName       string        `envconfig:"default=compass-ord-aggregator"`
}

// NewServiceConfig creates new ServiceConfig from the supplied parameters
func NewServiceConfig(maxParallelWebhookProcessors, maxParallelSpecificationProcessors, ordWebhookPartialProcessMaxDays int, ordWebhookPartialProcessURL string, ordWebhookPartialProcessing bool) ServiceConfig {
	return ServiceConfig{
		maxParallelWebhookProcessors:       maxParallelWebhookProcessors,
		maxParallelSpecificationProcessors: maxParallelSpecificationProcessors,
		ordWebhookPartialProcessMaxDays:    ordWebhookPartialProcessMaxDays,
		ordWebhookPartialProcessURL:        ordWebhookPartialProcessURL,
		ordWebhookPartialProcessing:        ordWebhookPartialProcessing,
	}
}

// Service consists of various resource services responsible for service-layer ORD operations.
type Service struct {
	config ServiceConfig

	transact persistence.Transactioner

	appSvc             ApplicationService
	webhookSvc         WebhookService
	bundleSvc          BundleService
	bundleReferenceSvc BundleReferenceService
	apiSvc             APIService
	eventSvc           EventService
	specSvc            SpecService
	fetchReqSvc        FetchRequestService
	packageSvc         PackageService
	productSvc         ProductService
	vendorSvc          VendorService
	tombstoneSvc       TombstoneService
	tenantSvc          TenantService

	webhookConverter WebhookConverter

	globalRegistrySvc GlobalRegistryService
	ordClient         Client
}

// NewAggregatorService returns a new object responsible for service-layer ORD operations.
func NewAggregatorService(config ServiceConfig, transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, eventSvc EventService, specSvc SpecService, fetchReqSvc FetchRequestService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, tenantSvc TenantService, globalRegistrySvc GlobalRegistryService, client Client, webhookConverter WebhookConverter) *Service {
	return &Service{
		config:             config,
		transact:           transact,
		appSvc:             appSvc,
		webhookSvc:         webhookSvc,
		bundleSvc:          bundleSvc,
		bundleReferenceSvc: bundleReferenceSvc,
		apiSvc:             apiSvc,
		eventSvc:           eventSvc,
		specSvc:            specSvc,
		fetchReqSvc:        fetchReqSvc,
		packageSvc:         packageSvc,
		productSvc:         productSvc,
		vendorSvc:          vendorSvc,
		tombstoneSvc:       tombstoneSvc,
		tenantSvc:          tenantSvc,
		globalRegistrySvc:  globalRegistrySvc,
		ordClient:          client,
		webhookConverter:   webhookConverter,
	}
}

// SyncORDDocuments performs resync of ORD information provided via ORD documents for each application
func (s *Service) SyncORDDocuments(ctx context.Context, cfg MetricsConfig) error {
	globalResourcesOrdIDs := s.retrieveGlobalResources(ctx)
	ordWebhooks, err := s.getWebhooksWithOrdType(ctx)
	if err != nil {
		return err
	}

	queue := make(chan *model.Webhook)
	var webhookErrors = int32(0)

	workers := s.config.maxParallelWebhookProcessors
	wg := &sync.WaitGroup{}
	wg.Add(workers)

	log.C(ctx).Infof("Starting %d parallel webhook processor workers...", workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for webhook := range queue {
				entry := log.C(ctx)
				entry = entry.WithField(log.FieldRequestID, uuid.New().String())
				ctx = log.ContextWithLogger(ctx, entry)

				if err := s.processWebhook(ctx, cfg, webhook, globalResourcesOrdIDs); err != nil {
					log.C(ctx).WithError(err).Errorf("error while processing webhook %q", webhook.ID)
					atomic.AddInt32(&webhookErrors, 1)
				}
			}
		}()
	}

	if s.config.ordWebhookPartialProcessing {
		log.C(ctx).Infof("Partial ord webhook processing is enabled for URL [%s] and max days [%d]", s.config.ordWebhookPartialProcessURL, s.config.ordWebhookPartialProcessMaxDays)
	}
	date := time.Now().AddDate(0, 0, -1*s.config.ordWebhookPartialProcessMaxDays)
	for _, webhook := range ordWebhooks {
		webhookURL := str.PtrStrToStr(webhook.URL)
		if s.config.ordWebhookPartialProcessing && strings.Contains(webhookURL, s.config.ordWebhookPartialProcessURL) {
			if webhook.CreatedAt == nil || webhook.CreatedAt.After(date) {
				queue <- webhook
			}
		} else {
			queue <- webhook
		}
	}
	close(queue)
	wg.Wait()

	if webhookErrors != 0 {
		log.C(ctx).Errorf("failed to process %d webhooks", webhookErrors)
	}

	return nil
}

// ProcessApplications performs resync of ORD information provided via ORD documents for list of applications
func (s *Service) ProcessApplications(ctx context.Context, cfg MetricsConfig, appIDs []string) error {
	if len(appIDs) == 0 {
		return nil
	}

	globalResourcesOrdIDs := s.retrieveGlobalResources(ctx)
	for _, appID := range appIDs {
		if err := s.processApplication(ctx, cfg, globalResourcesOrdIDs, appID); err != nil {
			return errors.Wrapf(err, "processing of ORD data for application with id %q failed", appID)
		}
	}
	return nil
}

func (s *Service) processApplication(ctx context.Context, cfg MetricsConfig, globalResourcesOrdIDs map[string]bool, appID string) error {
	webhooks, err := s.getWebhooksForApplication(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "retrieving of webhooks for application with id %q failed", appID)
	}

	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			if err := s.processApplicationWebhook(ctx, cfg, wh, appID, globalResourcesOrdIDs); err != nil {
				return errors.Wrapf(err, "processing of ORD webhook for application with id %q failed", appID)
			}
		}
	}
	return nil
}

// ProcessApplicationTemplates performs resync of ORD information provided via ORD documents for list of application templates
func (s *Service) ProcessApplicationTemplates(ctx context.Context, cfg MetricsConfig, appTemplateIDs []string) error {
	if len(appTemplateIDs) == 0 {
		return nil
	}

	globalResourcesOrdIDs := s.retrieveGlobalResources(ctx)
	for _, appTemplateID := range appTemplateIDs {
		if err := s.processApplicationTemplate(ctx, cfg, globalResourcesOrdIDs, appTemplateID); err != nil {
			return errors.Wrapf(err, "processing of ORD data for application template with id %q failed", appTemplateID)
		}
	}
	return nil
}

func (s *Service) processApplicationTemplate(ctx context.Context, cfg MetricsConfig, globalResourcesOrdIDs map[string]bool, appTemplateID string) error {
	webhooks, err := s.getWebhooksForApplicationTemplate(ctx, appTemplateID)
	if err != nil {
		return errors.Wrapf(err, "retrieving of webhooks for application template with id %q failed", appTemplateID)
	}

	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			apps, err := s.getApplicationsForAppTemplate(ctx, appTemplateID)
			if err != nil {
				return errors.Wrapf(err, "retrieving of applications for application template with id %q failed", appTemplateID)
			}

			for _, app := range apps {
				if err := s.processApplicationWebhook(ctx, cfg, wh, app.ID, globalResourcesOrdIDs); err != nil {
					return errors.Wrapf(err, "processing of ORD webhook for application with id %q failed", app.ID)
				}
			}
		}
	}
	return nil
}

func (s *Service) processWebhook(ctx context.Context, cfg MetricsConfig, webhook *model.Webhook, globalResourcesOrdIDs map[string]bool) error {
	if webhook.ObjectType == model.ApplicationTemplateWebhookReference {
		appTemplateID := webhook.ObjectID
		apps, err := s.getApplicationsForAppTemplate(ctx, appTemplateID)
		if err != nil {
			return err
		}

		for _, app := range apps {
			if err := s.processApplicationWebhook(ctx, cfg, webhook, app.ID, globalResourcesOrdIDs); err != nil {
				return err
			}
		}
	} else if webhook.ObjectType == model.ApplicationWebhookReference {
		appID := webhook.ObjectID
		if err := s.processApplicationWebhook(ctx, cfg, webhook, appID, globalResourcesOrdIDs); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) retrieveGlobalResources(ctx context.Context) map[string]bool {
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
	return globalResourcesOrdIDs
}

func (s *Service) getWebhooksForApplicationTemplate(ctx context.Context, appTemplateID string) ([]*model.Webhook, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhooks, err := s.webhookSvc.ListForApplicationTemplate(ctx, appTemplateID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks for application template with id %s", appTemplateID)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ordWebhooks, nil
}

func (s *Service) getWebhooksForApplication(ctx context.Context, appID string) ([]*model.Webhook, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhooks, err := s.webhookSvc.ListForApplication(ctx, appID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks for application with id %s", appID)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ordWebhooks, nil
}

func (s *Service) processDocuments(ctx context.Context, app *model.Application, baseURL string, documents Documents, globalResourcesOrdIDs map[string]bool, validationErrors *error) error {
	apiDataFromDB, eventDataFromDB, packageDataFromDB, bundleDataFromDB, err := s.fetchResources(ctx, app.ID)
	if err != nil {
		return err
	}

	resourceHashes, err := hashResources(documents)
	if err != nil {
		return err
	}

	validationResult := documents.Validate(baseURL, apiDataFromDB, eventDataFromDB, packageDataFromDB, bundleDataFromDB, resourceHashes, globalResourcesOrdIDs)
	if validationResult != nil {
		validationResult = &ORDDocumentValidationError{errors.Wrap(validationResult, "invalid documents")}
		*validationErrors = validationResult
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

	ordLocalID := s.getUniqueLocalTenantID(documents)
	if ordLocalID != "" && app.LocalTenantID == nil {
		if err := s.appSvc.Update(ctx, app.ID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr(ordLocalID)}); err != nil {
			return err
		}
	}

	vendorsFromDB, err := s.processVendors(ctx, app.ID, vendorsInput)
	if err != nil {
		return err
	}

	productsFromDB, err := s.processProducts(ctx, app.ID, productsInput)
	if err != nil {
		return err
	}

	packagesFromDB, err := s.processPackages(ctx, app.ID, packagesInput, resourceHashes)
	if err != nil {
		return err
	}

	bundlesFromDB, err := s.processBundles(ctx, app, bundlesInput, resourceHashes)
	if err != nil {
		return err
	}

	apisFromDB, apiFetchRequests, err := s.processAPIs(ctx, app.ID, bundlesFromDB, packagesFromDB, apisInput, resourceHashes)
	if err != nil {
		return err
	}

	eventsFromDB, eventFetchRequests, err := s.processEvents(ctx, app.ID, bundlesFromDB, packagesFromDB, eventsInput, resourceHashes)
	if err != nil {
		return err
	}

	tombstonesFromDB, err := s.processTombstones(ctx, app.ID, tombstonesInput)
	if err != nil {
		return err
	}

	fetchRequests := append(apiFetchRequests, eventFetchRequests...)
	fetchRequests, err = s.deleteTombstonedResources(ctx, vendorsFromDB, productsFromDB, packagesFromDB, bundlesFromDB, apisFromDB, eventsFromDB, tombstonesFromDB, fetchRequests)
	if err != nil {
		return err
	}

	return s.processSpecs(ctx, fetchRequests)
}

func (s *Service) processSpecs(ctx context.Context, ordFetchRequests []*ordFetchRequest) error {
	queue := make(chan *model.FetchRequest)

	workers := s.config.maxParallelSpecificationProcessors
	wg := &sync.WaitGroup{}
	wg.Add(workers)

	fetchReqMutex := sync.Mutex{}
	fetchRequestResults := make([]*fetchRequestResult, 0)
	log.C(ctx).Infof("Starting %d parallel specification processor workers to process %d fetch requests...", workers, len(ordFetchRequests))
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()

			for fetchRequest := range queue {
				fr := *fetchRequest
				ctx = addFieldToLogger(ctx, "fetch_request_id", fr.ID)
				log.C(ctx).Infof("Will attempt to execute spec fetch request for spec with id %q and spec entity type %q", fr.ObjectID, fr.ObjectType)
				data, status := s.fetchReqSvc.FetchSpec(ctx, &fr)
				log.C(ctx).Infof("Finished executing spec fetch request for spec with id %q and spec entity type %q with result: %s. Adding to result queue...", fr.ObjectID, fr.ObjectType, status.Condition)
				s.addFetchRequestResult(&fetchRequestResults, &fetchRequestResult{
					fetchRequest: &fr,
					data:         data,
					status:       status,
				}, &fetchReqMutex)
			}
		}()
	}

	for _, ordFetchRequest := range ordFetchRequests {
		queue <- ordFetchRequest.FetchRequest
	}
	close(queue)
	wg.Wait()

	if err := s.processFetchRequestResults(ctx, fetchRequestResults); err != nil {
		errMsg := "error while processing fetch request results"
		log.C(ctx).WithError(err).Error(errMsg)
		return errors.Errorf(errMsg)
	} else {
		log.C(ctx).Info("Successfully processed fetch request results")
	}

	fetchRequestErrors := 0
	for _, fr := range fetchRequestResults {
		if fr.status.Condition == model.FetchRequestStatusConditionFailed {
			fetchRequestErrors += 1
		}
	}

	if fetchRequestErrors != 0 {
		return errors.Errorf("failed to process %d specification fetch requests", fetchRequestErrors)
	}

	return nil
}

func (s *Service) addFetchRequestResult(fetchReqResults *[]*fetchRequestResult, result *fetchRequestResult, mutex *sync.Mutex) {
	mutex.Lock()
	defer mutex.Unlock()
	*fetchReqResults = append(*fetchReqResults, result)
}

func (s *Service) processFetchRequestResults(ctx context.Context, results []*fetchRequestResult) error {
	tx, err := s.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while opening transaction to process fetch request results")
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	for _, result := range results {
		specReferenceType := model.APISpecReference
		if result.fetchRequest.ObjectType == model.EventSpecFetchRequestReference {
			specReferenceType = model.EventSpecReference
		}

		if result.status.Condition == model.FetchRequestStatusConditionSucceeded {
			spec, err := s.specSvc.GetByID(ctx, result.fetchRequest.ObjectID, specReferenceType)
			if err != nil {
				return err
			}

			spec.Data = result.data

			if err = s.specSvc.UpdateSpecOnly(ctx, *spec); err != nil {
				return err
			}
		}

		result.fetchRequest.Status = result.status
		if err = s.fetchReqSvc.Update(ctx, result.fetchRequest); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) deleteTombstonedResources(ctx context.Context, vendorsFromDB []*model.Vendor, productsFromDB []*model.Product, packagesFromDB []*model.Package, bundlesFromDB []*model.Bundle, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition, tombstonesFromDB []*model.Tombstone, fetchRequests []*ordFetchRequest) ([]*ordFetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	frIdxToExclude := make([]int, 0)
	for _, ts := range tombstonesFromDB {
		if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
			return packagesFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.packageSvc.Delete(ctx, packagesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.apiSvc.Delete(ctx, apisFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.eventSvc.Delete(ctx, eventsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
			return equalStrings(bundlesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.bundleSvc.Delete(ctx, bundlesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
			return vendorsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.vendorSvc.Delete(ctx, vendorsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
			return productsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.productSvc.Delete(ctx, productsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		for i := range fetchRequests {
			if equalStrings(&fetchRequests[i].refObjectOrdID, &ts.OrdID) {
				frIdxToExclude = append(frIdxToExclude, i)
			}
		}
	}

	return excludeUnnecessaryFetchRequests(fetchRequests, frIdxToExclude), tx.Commit()
}

func (s *Service) processVendors(ctx context.Context, appID string, vendors []*model.VendorInput) ([]*model.Vendor, error) {
	vendorsFromDB, err := s.listVendorsInTx(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, vendor := range vendors {
		if err := s.resyncVendorInTx(ctx, appID, vendorsFromDB, vendor); err != nil {
			return nil, err
		}
	}

	vendorsFromDB, err = s.listVendorsInTx(ctx, appID)
	if err != nil {
		return nil, err
	}
	return vendorsFromDB, nil
}

func (s *Service) listVendorsInTx(ctx context.Context, appID string) ([]*model.Vendor, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	vendorsFromDB, err := s.vendorSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing vendors for app with id %q", appID)
	}

	return vendorsFromDB, tx.Commit()
}

func (s *Service) resyncVendorInTx(ctx context.Context, appID string, vendorsFromDB []*model.Vendor, vendor *model.VendorInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncVendor(ctx, appID, vendorsFromDB, *vendor); err != nil {
		return errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
	}
	return tx.Commit()
}

func (s *Service) processProducts(ctx context.Context, appID string, products []*model.ProductInput) ([]*model.Product, error) {
	productsFromDB, err := s.listProductsInTx(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		if err := s.resyncProductInTx(ctx, appID, productsFromDB, product); err != nil {
			return nil, err
		}
	}

	productsFromDB, err = s.listProductsInTx(ctx, appID)
	if err != nil {
		return nil, err
	}
	return productsFromDB, nil
}

func (s *Service) listProductsInTx(ctx context.Context, appID string) ([]*model.Product, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	productsFromDB, err := s.productSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing products for app with id %q", appID)
	}

	return productsFromDB, tx.Commit()
}

func (s *Service) resyncProductInTx(ctx context.Context, appID string, productsFromDB []*model.Product, product *model.ProductInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncProduct(ctx, appID, productsFromDB, *product); err != nil {
		return errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
	}
	return tx.Commit()
}

func (s *Service) processPackages(ctx context.Context, appID string, packages []*model.PackageInput, resourceHashes map[string]uint64) ([]*model.Package, error) {
	packagesFromDB, err := s.listPackagesInTx(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		pkgHash := resourceHashes[pkg.OrdID]
		if err := s.resyncPackageInTx(ctx, appID, packagesFromDB, pkg, pkgHash); err != nil {
			return nil, err
		}
	}

	packagesFromDB, err = s.listPackagesInTx(ctx, appID)
	if err != nil {
		return nil, err
	}
	return packagesFromDB, nil
}

func (s *Service) listPackagesInTx(ctx context.Context, appID string) ([]*model.Package, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	packagesFromDB, err := s.packageSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing packages for app with id %q", appID)
	}

	return packagesFromDB, tx.Commit()
}

func (s *Service) resyncPackageInTx(ctx context.Context, appID string, packagesFromDB []*model.Package, pkg *model.PackageInput, pkgHash uint64) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncPackage(ctx, appID, packagesFromDB, *pkg, pkgHash); err != nil {
		return errors.Wrapf(err, "error while resyncing package with ORD ID %q", pkg.OrdID)
	}
	return tx.Commit()
}

func (s *Service) processBundles(ctx context.Context, app *model.Application, bundles []*model.BundleCreateInput, resourceHashes map[string]uint64) ([]*model.Bundle, error) {
	bundlesFromDB, err := s.listBundlesInTx(ctx, app.ID)
	if err != nil {
		return nil, err
	}

	credentialExchangeStrategyHashCurrent := uint64(0)
	var credentialExchangeStrategyJSON gjson.Result
	for _, bndl := range bundles {
		bndlHash := resourceHashes[str.PtrStrToStr(bndl.OrdID)]
		if err := s.resyncBundleInTx(ctx, app.ID, bundlesFromDB, bndl, bndlHash); err != nil {
			return nil, err
		}

		credentialExchangeStrategies, err := bndl.CredentialExchangeStrategies.MarshalJSON()
		if err != nil {
			return nil, err
		}

		for _, credentialExchangeStrategy := range gjson.ParseBytes(credentialExchangeStrategies).Array() {
			customType := credentialExchangeStrategy.Get(customTypeProperty).String()
			isTenantMappingType := strings.Contains(customType, tenantMappingCustomTypeIdentifier)

			if !isTenantMappingType {
				continue
			}

			currentHash, err := HashObject(credentialExchangeStrategy)
			if err != nil {
				return nil, err
			}

			if credentialExchangeStrategyHashCurrent != 0 && currentHash != credentialExchangeStrategyHashCurrent {
				return nil, errors.Errorf("There are differences in the Credential Exchange Strategies for Tenant Mappings for application with ID %s. They should be the same.", app.ID)
			}

			credentialExchangeStrategyHashCurrent = currentHash
			credentialExchangeStrategyJSON = credentialExchangeStrategy
		}
	}

	//if credentialExchangeStrategyJSON.IsObject() {
	//	subParts := strings.Split(credentialExchangeStrategyJSON.Get("customType").String(), ":")
	//	matched, _ := regexp.MatchString(`(sap.ucl:tenant_mapping)(\.\w+){1,2}(:\w+)`, tenantMappingType)
	//	if len(subParts) != 3 || !matched {
	//		return nil, errors.Errorf("Credential Exchange Strategy has invalid customType value: %s", tenantMappingType)
	//	}
	//
	//	tenantMappingSubParts := strings.Split(subParts[1], ".")
	//
	//	var mode, version string
	//	if len(tenantMappingSubParts) == 3 {
	//		version = fmt.Sprintf("%s:%s", tenantMappingSubParts[1], subParts[2])
	//		mode = tenantMappingSubParts[2]
	//	} else {
	//		version = subParts[2]
	//		mode = tenantMappingSubParts[1]
	//	}
	//
	//	inputMode := graphql.WebhookMode(strings.ToUpper(mode))
	//	whInput := &graphql.WebhookInput{
	//		URL: str.Ptr(credentialExchangeStrategyJSON.Get("callbackUrl").String()),
	//		Auth: &graphql.AuthInput{
	//			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
	//		},
	//		Mode:    &inputMode,
	//		Version: str.Ptr(version),
	//	}
	//
	//	enrichedWhs, err := s.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks([]*graphql.WebhookInput{whInput})
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	ctxWithNoTenant := context.Background()
	//	tx, err := s.transact.Begin()
	//	if err != nil {
	//		return nil, err
	//	}
	//	defer s.transact.RollbackUnlessCommitted(ctxWithNoTenant, tx)
	//
	//	ctxWithNoTenant = persistence.SaveToContext(ctxWithNoTenant, tx)
	//	ctxWithNoTenant = tenant.SaveToContext(ctxWithNoTenant, "", "")
	//
	//	appWhs, err := s.webhookSvc.ListForApplicationGlobal(ctxWithNoTenant, app.ID)
	//	if err != nil {
	//		fmt.Printf("List err %+v\n\n", err)
	//		return nil, err
	//	}
	//
	//	tenantMappingRelatedWebhooksFromDB := make([]*model.Webhook, 0)
	//	enrichedWhModels := make([]*model.Webhook, 0)
	//	for _, wh := range enrichedWhs {
	//		convertedIn, err := s.webhookConverter.InputFromGraphQL(wh)
	//		if err != nil {
	//			return nil, errors.Wrap(err, "while converting the WebhookInput")
	//		}
	//
	//		for _, appWh := range appWhs {
	//			if appWh.Type == convertedIn.Type {
	//				whModel := convertedIn.ToWebhook(appWh.ID, appWh.ObjectID, appWh.ObjectType)
	//				whModel.CreatedAt = appWh.CreatedAt
	//				enrichedWhModels = append(enrichedWhModels, whModel)
	//				tenantMappingRelatedWebhooksFromDB = append(tenantMappingRelatedWebhooksFromDB, appWh)
	//				break
	//			}
	//		}
	//
	//	}
	//
	//	fmt.Printf("tenantMappingRelatedWebhooks %+v\n\n", tenantMappingRelatedWebhooksFromDB)
	//	fmt.Printf("enrichedWhModels %+v\n\n", enrichedWhModels)
	//
	//	appWhsFromDBMarshaled, _ := json.Marshal(tenantMappingRelatedWebhooksFromDB)
	//	appWhsFromDBHash, err := HashObject(string(appWhsFromDBMarshaled))
	//
	//	enrichedWhsMarshaled, _ := json.Marshal(enrichedWhModels)
	//	enrichedHash, err := HashObject(string(enrichedWhsMarshaled))
	//
	//	fmt.Printf("\n%s ----- %s \n\n", strconv.FormatUint(appWhsFromDBHash, 10), strconv.FormatUint(enrichedHash, 10))
	//	fmt.Printf("\nAre equal? %v \n\n", strconv.FormatUint(appWhsFromDBHash, 10) == strconv.FormatUint(enrichedHash, 10))
	//
	//	if strconv.FormatUint(appWhsFromDBHash, 10) != strconv.FormatUint(enrichedHash, 10) {
	//		for _, webhook := range tenantMappingRelatedWebhooksFromDB {
	//			fmt.Println("Deleting webhook with type " + webhook.Type)
	//			err := s.webhookSvc.Delete(ctxWithNoTenant, webhook.ID, webhook.ObjectType)
	//			if err != nil {
	//				log.C(ctx).Errorf("err while deleting webhook with ID %s", webhook.ID)
	//				return nil, err
	//			}
	//		}
	//
	//		for _, webhook := range enrichedWhs {
	//			convertedIn, err := s.webhookConverter.InputFromGraphQL(webhook)
	//			if err != nil {
	//				return nil, errors.Wrap(err, "while converting the WebhookInput")
	//			}
	//
	//			fmt.Println("Creating webhook with type " + webhook.Type)
	//			_, err = s.webhookSvc.Create(ctxWithNoTenant, app.ID, *convertedIn, model.ApplicationWebhookReference)
	//			if err != nil {
	//				log.C(ctx).Errorf("err while creating webhook for app %s with type %s", app.ID, convertedIn.Type)
	//				return nil, err
	//			}
	//		}
	//	}
	//
	//	//appWhs, err := s.webhookSvc.ListForApplicationGlobal(ctxWithNoTenant, app.ID)
	//	//if err != nil {
	//	//	fmt.Println("List err")
	//	//	fmt.Println(err)
	//	//	return nil, err
	//	//}
	//	//
	//	//for _, wh := range enrichedWhs {
	//	//	convertedIn, err := s.webhookConverter.InputFromGraphQL(wh)
	//	//	if err != nil {
	//	//		return nil, errors.Wrap(err, "while converting the WebhookInput")
	//	//	}
	//	//
	//	//	ctxWithNoTenant = tenant.SaveToContext(ctxWithNoTenant, "", "")
	//	//	if _, err := s.webhookSvc.Create(ctxWithNoTenant, app.ID, *convertedIn, model.ApplicationWebhookReference); err != nil {
	//	//		fmt.Println("Create err")
	//	//		fmt.Println(err)
	//	//		return nil, err
	//	//	}
	//	//}
	//	//
	//	//appWhs, err = s.webhookSvc.ListForApplicationGlobal(ctxWithNoTenant, app.ID)
	//	//if err != nil {
	//	//	fmt.Println("List err")
	//	//	fmt.Println(err)
	//	//	return nil, err
	//	//}
	//	//
	//	//originalWhModels := make([]*model.Webhook, 0)
	//	//enrichedWhModels := make([]*model.Webhook, 0)
	//	//for _, wh := range enrichedWhs {
	//	//	convertedIn, err := s.webhookConverter.InputFromGraphQL(wh)
	//	//	if err != nil {
	//	//		return nil, errors.Wrap(err, "while converting the WebhookInput")
	//	//	}
	//	//
	//	//	for _, appWh := range appWhs {
	//	//		if appWh.Type == convertedIn.Type {
	//	//			whModel := convertedIn.ToWebhook(appWh.ID, appWh.ObjectID, appWh.ObjectType)
	//	//			whModel.CreatedAt = appWh.CreatedAt
	//	//			enrichedWhModels = append(enrichedWhModels, whModel)
	//	//			originalWhModels = append(originalWhModels, appWh)
	//	//			break
	//	//		}
	//	//	}
	//	//
	//	//}
	//	//
	//	//appWhsMarshaled, _ := json.Marshal(originalWhModels)
	//	//appWhsHash, err := HashObject(string(appWhsMarshaled))
	//	//
	//	//enrichedWhsMashalled, _ := json.Marshal(enrichedWhModels)
	//	//enrichedHash, err := HashObject(string(enrichedWhsMashalled))
	//	//
	//	//fmt.Println("ALEX app " + app.ID)
	//	//fmt.Println(len(originalWhModels))
	//	//fmt.Println(string(appWhsMarshaled))
	//	//fmt.Println()
	//	//fmt.Println(len(enrichedWhModels))
	//	//fmt.Println(string(enrichedWhsMashalled))
	//	//
	//	//fmt.Printf("\n%s ----- %s \n\n", strconv.FormatUint(appWhsHash, 10), strconv.FormatUint(enrichedHash, 10))
	//	tx.Commit()
	//}

	if err := s.resyncTenantMappingWebhooksInTx(ctx, credentialExchangeStrategyJSON, app.ID); err != nil {
		return nil, err
	}

	bundlesFromDB, err = s.listBundlesInTx(ctx, app.ID)
	if err != nil {
		return nil, err
	}

	return bundlesFromDB, nil
}

func (s *Service) listBundlesInTx(ctx context.Context, appID string) ([]*model.Bundle, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	bundlesFromDB, err := s.bundleSvc.ListByApplicationIDNoPaging(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing bundles for app with id %q", appID)
	}

	return bundlesFromDB, tx.Commit()
}

func (s *Service) resyncBundleInTx(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, bundle *model.BundleCreateInput, bndlHash uint64) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncBundle(ctx, appID, bundlesFromDB, *bundle, bndlHash); err != nil {
		return errors.Wrapf(err, "error while resyncing bundle with ORD ID %q", *bundle.OrdID)
	}
	return tx.Commit()
}

func (s *Service) resyncTenantMappingWebhooksInTx(ctx context.Context, credentialExchangeStrategyJSON gjson.Result, appID string) error {
	if !credentialExchangeStrategyJSON.IsObject() {
		return nil
	}

	tenantMappingType := credentialExchangeStrategyJSON.Get(customTypeProperty).String()
	subParts := strings.Split(credentialExchangeStrategyJSON.Get(customTypeProperty).String(), ":")
	matched, _ := regexp.MatchString(fmt.Sprintf(`(%s)(\.\w+){1,2}(:\w+)`, tenantMappingCustomTypeIdentifier), tenantMappingType)
	if len(subParts) != 3 || !matched {
		return errors.Errorf("Credential Exchange Strategy has invalid %s value: %s for application with ID %s", customTypeProperty, tenantMappingType, appID)
	}

	tenantMappingSubParts := strings.Split(subParts[1], ".")

	var mode, version string
	if len(tenantMappingSubParts) == 3 {
		version = fmt.Sprintf("%s:%s", tenantMappingSubParts[1], subParts[2])
		mode = tenantMappingSubParts[2]
	} else {
		version = subParts[2]
		mode = tenantMappingSubParts[1]
	}

	inputMode := graphql.WebhookMode(strings.ToUpper(mode))
	whInput := &graphql.WebhookInput{
		URL: str.Ptr(credentialExchangeStrategyJSON.Get(callbackURLProperty).String()),
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode:    &inputMode,
		Version: str.Ptr(version),
	}

	enrichedWhs, err := s.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks([]*graphql.WebhookInput{whInput})
	if err != nil {
		return err
	}

	ctxWithNoTenant := context.Background()
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctxWithNoTenant, tx)

	ctxWithNoTenant = persistence.SaveToContext(ctxWithNoTenant, tx)
	ctxWithNoTenant = tenant.SaveToContext(ctxWithNoTenant, "", "")

	appWebhooksFromDB, err := s.webhookSvc.ListForApplicationGlobal(ctxWithNoTenant, appID)
	if err != nil {
		return errors.Wrapf(err, "while listing webhooks from application with ID %s", appID)
	}

	tenantMappingRelatedWebhooksFromDB := make([]*model.Webhook, 0)
	enrichedWhModels := make([]*model.Webhook, 0)
	enrichedWhsModelInputs := make([]*model.WebhookInput, 0)

	for _, wh := range enrichedWhs {
		convertedIn, err := s.webhookConverter.InputFromGraphQL(wh)
		if err != nil {
			return errors.Wrap(err, "while converting the WebhookInput")
		}

		enrichedWhsModelInputs = append(enrichedWhsModelInputs, convertedIn)

		whModel := convertedIn.ToWebhook("", "", "")

		for _, appWh := range appWebhooksFromDB {
			if appWh.Type == convertedIn.Type {
				whModel.ID = appWh.ID
				whModel.ObjectType = appWh.ObjectType
				whModel.ObjectID = appWh.ObjectID
				whModel.CreatedAt = appWh.CreatedAt

				tenantMappingRelatedWebhooksFromDB = append(tenantMappingRelatedWebhooksFromDB, appWh)
				break
			}
		}

		enrichedWhModels = append(enrichedWhModels, whModel)
	}

	appWhsFromDBMarshaled, _ := json.Marshal(tenantMappingRelatedWebhooksFromDB)
	appWhsFromDBHash, err := HashObject(string(appWhsFromDBMarshaled))
	if err != nil {
		return errors.Wrapf(err, "while hashing webhooks from DB")
	}

	enrichedWhsMarshaled, _ := json.Marshal(enrichedWhModels)
	enrichedHash, err := HashObject(string(enrichedWhsMarshaled))
	if err != nil {
		return errors.Wrapf(err, "while hashing webhooks from ORD document")
	}

	if strconv.FormatUint(appWhsFromDBHash, 10) == strconv.FormatUint(enrichedHash, 10) {
		log.C(ctxWithNoTenant).Infof("There are no differences in tenant mapping webhooks from the DB and the ORD document")

		return tx.Commit()
	}

	log.C(ctxWithNoTenant).Infof("There are differences in tenant mapping webhooks from the DB and the ORD document. Continuing the sync.")
	for _, webhook := range tenantMappingRelatedWebhooksFromDB {
		log.C(ctxWithNoTenant).Infof("Deleting webhook with ID %s for application %s", webhook.ID, appID)
		if err = s.webhookSvc.Delete(ctxWithNoTenant, webhook.ID, webhook.ObjectType); err != nil {
			log.C(ctx).Errorf("error while deleting webhook with ID %s", webhook.ID)
			return errors.Wrapf(err, "while deleting webhook with ID %s", webhook.ID)
		}
	}

	for _, webhook := range enrichedWhsModelInputs {
		log.C(ctxWithNoTenant).Infof("Creating webhook with type %s for application %s", webhook.Type, appID)
		if _, err = s.webhookSvc.Create(ctxWithNoTenant, appID, *webhook, model.ApplicationWebhookReference); err != nil {
			log.C(ctx).Errorf("error while creating webhook for app %s with type %s", appID, webhook.Type)
			return errors.Wrapf(err, "error while creating webhook for app %s with type %s", appID, webhook.Type)
		}
	}

	return tx.Commit()
}

func (s *Service) processAPIs(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apis []*model.APIDefinitionInput, resourceHashes map[string]uint64) ([]*model.APIDefinition, []*ordFetchRequest, error) {
	apisFromDB, err := s.listAPIsInTx(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*ordFetchRequest, 0)
	for _, api := range apis {
		apiHash := resourceHashes[str.PtrStrToStr(api.OrdID)]
		apiFetchRequests, err := s.resyncAPIInTx(ctx, appID, apisFromDB, bundlesFromDB, packagesFromDB, api, apiHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range apiFetchRequests {
			fetchRequests = append(fetchRequests, &ordFetchRequest{
				FetchRequest:   apiFetchRequests[i],
				refObjectOrdID: *api.OrdID,
			})
		}
	}

	apisFromDB, err = s.listAPIsInTx(ctx, appID)
	if err != nil {
		return nil, nil, err
	}
	return apisFromDB, fetchRequests, nil
}

func (s *Service) listAPIsInTx(ctx context.Context, appID string) ([]*model.APIDefinition, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	apisFromDB, err := s.apiSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing apis for app with id %q", appID)
	}

	return apisFromDB, tx.Commit()
}

func (s *Service) resyncAPIInTx(ctx context.Context, appID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api *model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := s.resyncAPI(ctx, appID, apisFromDB, bundlesFromDB, packagesFromDB, *api, apiHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (s *Service) processEvents(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput, resourceHashes map[string]uint64) ([]*model.EventDefinition, []*ordFetchRequest, error) {
	eventsFromDB, err := s.listEventsInTx(ctx, appID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*ordFetchRequest, 0)
	for _, event := range events {
		eventHash := resourceHashes[str.PtrStrToStr(event.OrdID)]
		eventFetchRequests, err := s.resyncEventInTx(ctx, appID, eventsFromDB, bundlesFromDB, packagesFromDB, event, eventHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range eventFetchRequests {
			fetchRequests = append(fetchRequests, &ordFetchRequest{
				FetchRequest:   eventFetchRequests[i],
				refObjectOrdID: *event.OrdID,
			})
		}
	}

	eventsFromDB, err = s.listEventsInTx(ctx, appID)
	if err != nil {
		return nil, nil, err
	}
	return eventsFromDB, fetchRequests, nil
}

func (s *Service) listEventsInTx(ctx context.Context, appID string) ([]*model.EventDefinition, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	eventsFromDB, err := s.eventSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing events for app with id %q", appID)
	}

	return eventsFromDB, tx.Commit()
}

func (s *Service) resyncEventInTx(ctx context.Context, appID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event *model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := s.resyncEvent(ctx, appID, eventsFromDB, bundlesFromDB, packagesFromDB, *event, eventHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing event with ORD ID %q", *event.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (s *Service) processTombstones(ctx context.Context, appID string, tombstones []*model.TombstoneInput) ([]*model.Tombstone, error) {
	tombstonesFromDB, err := s.listTombstonesInTx(ctx, appID)
	if err != nil {
		return nil, err
	}

	for _, tombstone := range tombstones {
		if err := s.resyncTombstoneInTx(ctx, appID, tombstonesFromDB, tombstone); err != nil {
			return nil, err
		}
	}

	tombstonesFromDB, err = s.listTombstonesInTx(ctx, appID)
	if err != nil {
		return nil, err
	}
	return tombstonesFromDB, nil
}

func (s *Service) listTombstonesInTx(ctx context.Context, appID string) ([]*model.Tombstone, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	tombstonesFromDB, err := s.tombstoneSvc.ListByApplicationID(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing tombstones for app with id %q", appID)
	}

	return tombstonesFromDB, tx.Commit()
}

func (s *Service) resyncTombstoneInTx(ctx context.Context, appID string, tombstonesFromDB []*model.Tombstone, tombstone *model.TombstoneInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncTombstone(ctx, appID, tombstonesFromDB, *tombstone); err != nil {
		return errors.Wrapf(err, "error while resyncing tombstone for resource with ORD ID %q", tombstone.OrdID)
	}
	return tx.Commit()
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

func (s *Service) resyncBundle(ctx context.Context, appID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput, bndlHash uint64) error {
	ctx = addFieldToLogger(ctx, "bundle_ord_id", *bndl.OrdID)
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		return s.bundleSvc.UpdateBundle(ctx, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl), bndlHash)
	}
	_, err := s.bundleSvc.CreateBundle(ctx, appID, bndl, bndlHash)
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

func (s *Service) resyncAPI(ctx context.Context, appID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
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
		apiID, err := s.apiSvc.Create(ctx, appID, nil, packageID, api, nil, defaultTargetURLPerBundle, apiHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}
		return s.createSpecs(ctx, model.APISpecReference, apiID, specs)
	}

	allBundleIDsForAPI, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &apisFromDB[i].ID)
	if err != nil {
		return nil, err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion := extractBundleReferencesForDeletion(allBundleIDsForAPI, defaultTargetURLPerBundle)

	// in case of API update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	defaultTargetURLPerBundleForCreation := extractAllBundleReferencesForCreation(defaultTargetURLPerBundle, allBundleIDsForAPI)

	if err := s.apiSvc.UpdateInManyBundles(ctx, apisFromDB[i].ID, api, nil, defaultTargetURLPerBundle, defaultTargetURLPerBundleForCreation, bundleIDsForDeletion, apiHash, defaultConsumptionBundleID); err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	if api.VersionInput.Value != apisFromDB[i].Version.Value {
		fetchRequests, err = s.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = s.refetchFailedSpecs(ctx, model.APISpecReference, apisFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return fetchRequests, nil
}

func (s *Service) resyncEvent(ctx context.Context, appID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
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
		eventID, err := s.eventSvc.Create(ctx, appID, nil, packageID, event, nil, bundleIDsFromBundleReference, eventHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}
		return s.createSpecs(ctx, model.EventSpecReference, eventID, specs)
	}

	allBundleIDsForEvent, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleEventReference, &eventsFromDB[i].ID)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	if event.VersionInput.Value != eventsFromDB[i].Version.Value {
		fetchRequests, err = s.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = s.refetchFailedSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}

	return fetchRequests, nil
}

func (s *Service) createSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput) ([]*model.FetchRequest, error) {
	fetchRequests := make([]*model.FetchRequest, 0)
	for _, spec := range specs {
		if spec == nil {
			continue
		}
		_, fr, err := s.specSvc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, *spec, objectType, objectID)
		if err != nil {
			return nil, err
		}
		fetchRequests = append(fetchRequests, fr)
	}
	return fetchRequests, nil
}

func (s *Service) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput) ([]*model.FetchRequest, error) {
	if err := s.specSvc.DeleteByReferenceObjectID(ctx, objectType, objectID); err != nil {
		return nil, err
	}
	return s.createSpecs(ctx, objectType, objectID, specs)
}

func (s *Service) refetchFailedSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]*model.FetchRequest, error) {
	specIDsFromDB, err := s.specSvc.ListIDByReferenceObjectID(ctx, objectType, objectID)
	if err != nil {
		return nil, err
	}

	tnt, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return nil, err
	}

	fetchRequestsFromDB, err := s.specSvc.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDsFromDB, objectType)
	if err != nil {
		return nil, err
	}

	fetchRequests := make([]*model.FetchRequest, 0)
	for _, fr := range fetchRequestsFromDB {
		if fr.Status != nil && fr.Status.Condition != model.FetchRequestStatusConditionSucceeded {
			fetchRequests = append(fetchRequests, fr)
		}
	}
	return fetchRequests, nil
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

func (s *Service) fetchBundlesFromDB(ctx context.Context, appID string) (map[string]*model.Bundle, error) {
	bundlesFromDB, err := s.bundleSvc.ListByApplicationIDNoPaging(ctx, appID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing bundles for app with id %s", appID)
	}

	bundleDataFromDB := make(map[string]*model.Bundle)

	for _, bndl := range bundlesFromDB {
		bndlOrdID := str.PtrStrToStr(bndl.OrdID)
		bundleDataFromDB[bndlOrdID] = bndl
	}

	return bundleDataFromDB, nil
}

func (s *Service) fetchResources(ctx context.Context, appID string) (map[string]*model.APIDefinition, map[string]*model.EventDefinition, map[string]*model.Package, map[string]*model.Bundle, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	apiDataFromDB, err := s.fetchAPIDefFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "while fetching apis for app with id %s", appID)
	}

	eventDataFromDB, err := s.fetchEventDefFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "while fetching events for app with id %s", appID)
	}

	packageDataFromDB, err := s.fetchPackagesFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "while fetching packages for app with id %s", appID)
	}

	bundlDataFromDB, err := s.fetchBundlesFromDB(ctx, appID)
	if err != nil {
		return nil, nil, nil, nil, errors.Wrapf(err, "while fetching bundles for app with id %s", appID)
	}

	return apiDataFromDB, eventDataFromDB, packageDataFromDB, bundlDataFromDB, tx.Commit()
}

func (s *Service) processWebhookAndDocuments(ctx context.Context, cfg MetricsConfig, webhook *model.Webhook, app *model.Application, globalResourcesOrdIDs map[string]bool) error {
	var documents Documents
	var baseURL string
	var err error

	metricsCfg := metrics.PusherConfig{
		Enabled:    len(cfg.PushEndpoint) > 0,
		Endpoint:   cfg.PushEndpoint,
		MetricName: strings.ReplaceAll(strings.ToLower(cfg.JobName), "-", "_") + "_job_sync_failure_number",
		Timeout:    cfg.ClientTimeout,
		Subsystem:  metrics.OrdAggregatorSubsystem,
		Labels:     []string{metrics.ErrorMetricLabel, metrics.AppIDMetricLabel, metrics.CorrelationIDMetricLabel},
	}

	if webhook.Type == model.WebhookTypeOpenResourceDiscovery && webhook.URL != nil {
		ctx = addFieldToLogger(ctx, "app_id", app.ID)
		documents, baseURL, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, app, webhook)
		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			return errors.Wrapf(err, "error fetching ORD document for webhook with id %q: %v", webhook.ID, err)
		}
	}

	if len(documents) > 0 {
		log.C(ctx).Info("Processing ORD documents")
		var validationErrors error
		if err = s.processDocuments(ctx, app, baseURL, documents, globalResourcesOrdIDs, &validationErrors); err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			log.C(ctx).WithError(err).Errorf("error processing ORD documents: %v", err)
			return errors.Wrapf(err, "error processing ORD documents")
		}
		if ordValidationError, ok := (validationErrors).(*ORDDocumentValidationError); ok {
			validationErrors := strings.Split(ordValidationError.Error(), MultiErrorSeparator)

			// the first item in the slice is the message 'invalid documents' for the wrapped errors
			validationErrors = validationErrors[1:]

			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)

			for i := range validationErrors {
				validationErrors[i] = strings.TrimSpace(validationErrors[i])
				metricsPusher.ReportAggregationFailureORD(ctx, validationErrors[i])
			}

			log.C(ctx).WithError(ordValidationError.Err).WithField("validation_errors", validationErrors).Error("error processing ORD documents")
			return errors.Wrapf(ordValidationError.Err, "error processing ORD documents")
		}
		log.C(ctx).Info("Successfully processed ORD documents")
	}
	return nil
}

func (s *Service) getWebhooksWithOrdType(ctx context.Context) ([]*model.Webhook, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ordWebhooks, err := s.webhookSvc.ListByWebhookType(ctx, model.WebhookTypeOpenResourceDiscovery)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks with type %s", model.WebhookTypeOpenResourceDiscovery)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return ordWebhooks, nil
}

func (s *Service) getApplicationsForAppTemplate(ctx context.Context, appTemplateID string) ([]*model.Application, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	apps, err := s.appSvc.ListAllByApplicationTemplateID(ctx, appTemplateID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return apps, err
}

func (s *Service) getUniqueLocalTenantID(documents Documents) string {
	var uniqueLocalTenantIds []string
	localTenants := make(map[string]bool, 0)
	var systemInstanceLocalTenantID *string

	for _, doc := range documents {
		if doc != nil && doc.DescribedSystemInstance != nil {
			systemInstanceLocalTenantID = doc.DescribedSystemInstance.LocalTenantID
			if systemInstanceLocalTenantID != nil {
				if _, exists := localTenants[*systemInstanceLocalTenantID]; !exists {
					localTenants[*systemInstanceLocalTenantID] = true
					uniqueLocalTenantIds = append(uniqueLocalTenantIds, *doc.DescribedSystemInstance.LocalTenantID)
				}
			}
		}
	}
	if len(uniqueLocalTenantIds) == 1 {
		return uniqueLocalTenantIds[0]
	}

	return ""
}

func (s *Service) saveLowestOwnerForAppToContext(ctx context.Context, appID string) (context.Context, error) {
	internalTntID, err := s.tenantSvc.GetLowestOwnerForResource(ctx, resource.Application, appID)
	if err != nil {
		return nil, err
	}

	tnt, err := s.tenantSvc.GetTenantByID(ctx, internalTntID)
	if err != nil {
		return nil, err
	}

	ctx = tenant.SaveToContext(ctx, internalTntID, tnt.ExternalTenant)

	return ctx, nil
}

func (s *Service) processApplicationWebhook(ctx context.Context, cfg MetricsConfig, webhook *model.Webhook, appID string, globalResourcesOrdIDs map[string]bool) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	ctx, err = s.saveLowestOwnerForAppToContext(ctx, appID)
	if err != nil {
		return err
	}
	app, err := s.appSvc.Get(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "error while retrieving app with id %q", appID)
	}

	localTenantID := str.PtrStrToStr(app.LocalTenantID)
	ctx = tenant.SaveLocalTenantIDToContext(ctx, localTenantID)

	if err = tx.Commit(); err != nil {
		return err
	}

	if err = s.processWebhookAndDocuments(ctx, cfg, webhook, app, globalResourcesOrdIDs); err != nil {
		return err
	}

	return nil
}

func excludeUnnecessaryFetchRequests(fetchRequests []*ordFetchRequest, frIdxToExclude []int) []*ordFetchRequest {
	finalFetchRequests := make([]*ordFetchRequest, 0)
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

		for _, bundleInput := range doc.ConsumptionBundles {
			normalizedBndl, err := normalizeBundle(bundleInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedBndl)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing bundle with ORD ID: %v", normalizedBndl.OrdID)
			}

			resourceHashes[str.PtrStrToStr(bundleInput.OrdID)] = hash
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
