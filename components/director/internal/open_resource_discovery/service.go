package ord

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/imdario/mergo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/metrics"
	directorresource "github.com/kyma-incubator/compass/components/director/pkg/resource"

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
	// TenantMappingCustomTypeIdentifier represents an identifier for tenant mapping webhooks in Credential exchange strategies
	TenantMappingCustomTypeIdentifier = "sap.ucl:tenant-mapping"

	customTypeProperty  = "customType"
	callbackURLProperty = "callbackUrl"
)

// ServiceConfig contains configuration for the ORD aggregator service
type ServiceConfig struct {
	maxParallelWebhookProcessors       int
	maxParallelSpecificationProcessors int
	ordWebhookPartialProcessMaxDays    int
	ordWebhookPartialProcessURL        string
	ordWebhookPartialProcessing        bool

	credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping
}

// CredentialExchangeStrategyTenantMapping contains tenant mappings configuration
type CredentialExchangeStrategyTenantMapping struct {
	Mode    model.WebhookMode
	Version string
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
func NewServiceConfig(maxParallelWebhookProcessors, maxParallelSpecificationProcessors, ordWebhookPartialProcessMaxDays int, ordWebhookPartialProcessURL string, ordWebhookPartialProcessing bool, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) ServiceConfig {
	return ServiceConfig{
		maxParallelWebhookProcessors:             maxParallelWebhookProcessors,
		maxParallelSpecificationProcessors:       maxParallelSpecificationProcessors,
		ordWebhookPartialProcessMaxDays:          ordWebhookPartialProcessMaxDays,
		ordWebhookPartialProcessURL:              ordWebhookPartialProcessURL,
		ordWebhookPartialProcessing:              ordWebhookPartialProcessing,
		credentialExchangeStrategyTenantMappings: credentialExchangeStrategyTenantMappings,
	}
}

// Service consists of various resource services responsible for service-layer ORD operations.
type Service struct {
	config ServiceConfig

	transact persistence.Transactioner

	appSvc                ApplicationService
	webhookSvc            WebhookService
	bundleSvc             BundleService
	bundleReferenceSvc    BundleReferenceService
	apiSvc                APIService
	eventSvc              EventService
	specSvc               SpecService
	fetchReqSvc           FetchRequestService
	packageSvc            PackageService
	productSvc            ProductService
	vendorSvc             VendorService
	tombstoneSvc          TombstoneService
	tenantSvc             TenantService
	appTemplateVersionSvc ApplicationTemplateVersionService
	appTemplateSvc        ApplicationTemplateService

	webhookConverter WebhookConverter

	globalRegistrySvc GlobalRegistryService
	ordClient         Client
}

// NewAggregatorService returns a new object responsible for service-layer ORD operations.
func NewAggregatorService(config ServiceConfig, transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, eventSvc EventService, specSvc SpecService, fetchReqSvc FetchRequestService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneSvc TombstoneService, tenantSvc TenantService, globalRegistrySvc GlobalRegistryService, client Client, webhookConverter WebhookConverter, appTemplateVersionSvc ApplicationTemplateVersionService, appTemplateSvc ApplicationTemplateService) *Service {
	return &Service{
		config:                config,
		transact:              transact,
		appSvc:                appSvc,
		webhookSvc:            webhookSvc,
		bundleSvc:             bundleSvc,
		bundleReferenceSvc:    bundleReferenceSvc,
		apiSvc:                apiSvc,
		eventSvc:              eventSvc,
		specSvc:               specSvc,
		fetchReqSvc:           fetchReqSvc,
		packageSvc:            packageSvc,
		productSvc:            productSvc,
		vendorSvc:             vendorSvc,
		tombstoneSvc:          tombstoneSvc,
		tenantSvc:             tenantSvc,
		globalRegistrySvc:     globalRegistrySvc,
		ordClient:             client,
		webhookConverter:      webhookConverter,
		appTemplateVersionSvc: appTemplateVersionSvc,
		appTemplateSvc:        appTemplateSvc,
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
			if err := s.processApplicationTemplateWebhook(ctx, cfg, wh, appTemplateID, globalResourcesOrdIDs); err != nil {
				return err
			}

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

		if err := s.processApplicationTemplateWebhook(ctx, cfg, webhook, appTemplateID, globalResourcesOrdIDs); err != nil {
			return err
		}

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

func (s *Service) processDocuments(ctx context.Context, resource Resource, baseURL string, documents Documents, globalResourcesOrdIDs map[string]bool, validationErrors *error) error {
	if _, err := s.processDescribedSystemVersions(ctx, resource, documents); err != nil {
		return err
	}

	resourcesFromDB, err := s.fetchResources(ctx, resource, documents)
	if err != nil {
		return err
	}

	resourceHashes, err := hashResources(documents)
	if err != nil {
		return err
	}

	validationResult := documents.Validate(baseURL, resourcesFromDB, resourceHashes, globalResourcesOrdIDs, s.config.credentialExchangeStrategyTenantMappings)
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
	if ordLocalID != "" && resource.LocalTenantID == nil {
		if err := s.appSvc.Update(ctx, resource.ID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr(ordLocalID)}); err != nil {
			return err
		}
	}
	for _, doc := range documents {
		if resource.Type == directorresource.ApplicationTemplate && doc.DescribedSystemVersion == nil {
			continue
		}

		resourceToAggregate := Resource{
			ID:   resource.ID,
			Type: directorresource.Application,
		}

		if doc.DescribedSystemVersion != nil {
			applicationTemplateID := resource.ID
			if resource.Type == directorresource.Application && resource.ParentID != nil {
				applicationTemplateID = *resource.ParentID
			}

			appTemplateVersion, err := s.getApplicationTemplateVersionByAppTemplateIDAndVersionInTx(ctx, applicationTemplateID, doc.DescribedSystemVersion.Version)
			if err != nil {
				return err
			}

			resourceToAggregate = Resource{
				ID:   appTemplateVersion.ID,
				Type: directorresource.ApplicationTemplateVersion,
			}
		}

		vendorsFromDB, err := s.processVendors(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Vendors)
		if err != nil {
			return err
		}

		productsFromDB, err := s.processProducts(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Products)
		if err != nil {
			return err
		}

		packagesFromDB, err := s.processPackages(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Packages, resourceHashes)
		if err != nil {
			return err
		}

		bundlesFromDB, err := s.processBundles(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.ConsumptionBundles, resourceHashes)
		if err != nil {
			return err
		}

		apisFromDB, apiFetchRequests, err := s.processAPIs(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.APIResources, resourceHashes)
		if err != nil {
			return err
		}

		eventsFromDB, eventFetchRequests, err := s.processEvents(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.EventResources, resourceHashes)
		if err != nil {
			return err
		}

		tombstonesFromDB, err := s.processTombstones(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Tombstones)
		if err != nil {
			return err
		}

		fetchRequests := append(apiFetchRequests, eventFetchRequests...)
		fetchRequests, err = s.deleteTombstonedResources(ctx, resourceToAggregate.Type, vendorsFromDB, productsFromDB, packagesFromDB, bundlesFromDB, apisFromDB, eventsFromDB, tombstonesFromDB, fetchRequests)
		if err != nil {
			return err
		}

		if err := s.processSpecs(ctx, resourceToAggregate.Type, fetchRequests); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) processSpecs(ctx context.Context, resourceType directorresource.Type, ordFetchRequests []*ordFetchRequest) error {
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

	if err := s.processFetchRequestResults(ctx, resourceType, fetchRequestResults); err != nil {
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

func (s *Service) processFetchRequestResults(ctx context.Context, resourceType directorresource.Type, results []*fetchRequestResult) error {
	tx, err := s.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while opening transaction to process fetch request results")
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	for _, result := range results {
		if resourceType.IsTenantIgnorable() {
			if err = s.processFetchRequestResultGlobal(ctx, result); err != nil {
				return err
			}
		} else {
			if err = s.processFetchRequestResult(ctx, result); err != nil {
				return err
			}
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) processFetchRequestResult(ctx context.Context, result *fetchRequestResult) error {
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
	return s.fetchReqSvc.Update(ctx, result.fetchRequest)
}

func (s *Service) processFetchRequestResultGlobal(ctx context.Context, result *fetchRequestResult) error {
	if result.status.Condition == model.FetchRequestStatusConditionSucceeded {
		spec, err := s.specSvc.GetByIDGlobal(ctx, result.fetchRequest.ObjectID)
		if err != nil {
			return err
		}

		spec.Data = result.data

		if err = s.specSvc.UpdateSpecOnlyGlobal(ctx, *spec); err != nil {
			return err
		}
	}

	result.fetchRequest.Status = result.status
	return s.fetchReqSvc.UpdateGlobal(ctx, result.fetchRequest)
}

func (s *Service) deleteTombstonedResources(ctx context.Context, resourceType directorresource.Type, vendorsFromDB []*model.Vendor, productsFromDB []*model.Product, packagesFromDB []*model.Package, bundlesFromDB []*model.Bundle, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition, tombstonesFromDB []*model.Tombstone, fetchRequests []*ordFetchRequest) ([]*ordFetchRequest, error) {
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
			if err := s.packageSvc.Delete(ctx, resourceType, packagesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(apisFromDB), func(i int) bool {
			return equalStrings(apisFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.apiSvc.Delete(ctx, resourceType, apisFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(eventsFromDB), func(i int) bool {
			return equalStrings(eventsFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.eventSvc.Delete(ctx, resourceType, eventsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
			return equalStrings(bundlesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.bundleSvc.Delete(ctx, resourceType, bundlesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
			return vendorsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.vendorSvc.Delete(ctx, resourceType, vendorsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
			return productsFromDB[i].OrdID == ts.OrdID
		}); found {
			if err := s.productSvc.Delete(ctx, resourceType, productsFromDB[i].ID); err != nil {
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

func (s *Service) processDescribedSystemVersions(ctx context.Context, resource Resource, documents Documents) ([]*model.ApplicationTemplateVersion, error) {
	appTemplateID := resource.ID
	if resource.Type == directorresource.Application && resource.ParentID != nil {
		appTemplateID = *resource.ParentID
	}

	appTemplateVersions, err := s.listApplicationTemplateVersionByAppTemplateIDInTx(ctx, appTemplateID)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return nil, err
	}

	for _, document := range documents {
		if document.DescribedSystemVersion == nil {
			continue
		}

		if err := s.resyncApplicationTemplateVersionInTx(ctx, appTemplateID, appTemplateVersions, document.DescribedSystemVersion); err != nil {
			return nil, err
		}
	}

	return s.listApplicationTemplateVersionByAppTemplateIDInTx(ctx, appTemplateID)
}

func (s *Service) listApplicationTemplateVersionByAppTemplateIDInTx(ctx context.Context, applicationTemplateID string) ([]*model.ApplicationTemplateVersion, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	appTemplateVersions, err := s.appTemplateVersionSvc.ListByAppTemplateID(ctx, applicationTemplateID)
	if err != nil {
		return nil, err
	}

	return appTemplateVersions, tx.Commit()
}

func (s *Service) getApplicationTemplateVersionByAppTemplateIDAndVersionInTx(ctx context.Context, applicationTemplateID, version string) (*model.ApplicationTemplateVersion, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	systemVersion, err := s.appTemplateVersionSvc.GetByAppTemplateIDAndVersion(ctx, applicationTemplateID, version)
	if err != nil {
		return nil, err
	}

	return systemVersion, tx.Commit()
}

func (s *Service) processVendors(ctx context.Context, resourceType directorresource.Type, resourceID string, vendors []*model.VendorInput) ([]*model.Vendor, error) {
	vendorsFromDB, err := s.listVendorsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, vendor := range vendors {
		if err := s.resyncVendorInTx(ctx, resourceType, resourceID, vendorsFromDB, vendor); err != nil {
			return nil, err
		}
	}

	vendorsFromDB, err = s.listVendorsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return vendorsFromDB, nil
}

func (s *Service) listVendorsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Vendor, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var vendorsFromDB []*model.Vendor
	if resourceType == directorresource.Application {
		vendorsFromDB, err = s.vendorSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		vendorsFromDB, err = s.vendorSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing vendors for %s with id %q", resourceType, resourceID)
	}

	return vendorsFromDB, tx.Commit()
}

func (s *Service) resyncVendorInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, vendorsFromDB []*model.Vendor, vendor *model.VendorInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncVendor(ctx, resourceType, resourceID, vendorsFromDB, *vendor); err != nil {
		return errors.Wrapf(err, "error while resyncing vendor with ORD ID %q", vendor.OrdID)
	}
	return tx.Commit()
}

func (s *Service) resyncApplicationTemplateVersionInTx(ctx context.Context, appTemplateID string, appTemplateVersionsFromDB []*model.ApplicationTemplateVersion, appTemplateVersion *model.ApplicationTemplateVersionInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncAppTemplateVersion(ctx, appTemplateID, appTemplateVersionsFromDB, appTemplateVersion); err != nil {
		return errors.Wrapf(err, "error while resyncing App Template Version for App template %q", appTemplateID)
	}
	return tx.Commit()
}

func (s *Service) processProducts(ctx context.Context, resourceType directorresource.Type, resourceID string, products []*model.ProductInput) ([]*model.Product, error) {
	productsFromDB, err := s.listProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, product := range products {
		if err := s.resyncProductInTx(ctx, resourceType, resourceID, productsFromDB, product); err != nil {
			return nil, err
		}
	}

	productsFromDB, err = s.listProductsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return productsFromDB, nil
}

func (s *Service) listProductsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Product, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var productsFromDB []*model.Product
	if resourceType == directorresource.Application {
		productsFromDB, err = s.productSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		productsFromDB, err = s.productSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing products for %s with id %q", resourceType, resourceID)
	}

	return productsFromDB, tx.Commit()
}

func (s *Service) resyncProductInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, productsFromDB []*model.Product, product *model.ProductInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncProduct(ctx, resourceType, resourceID, productsFromDB, *product); err != nil {
		return errors.Wrapf(err, "error while resyncing product with ORD ID %q", product.OrdID)
	}
	return tx.Commit()
}

func (s *Service) processPackages(ctx context.Context, resourceType directorresource.Type, resourceID string, packages []*model.PackageInput, resourceHashes map[string]uint64) ([]*model.Package, error) {
	packagesFromDB, err := s.listPackagesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, pkg := range packages {
		pkgHash := resourceHashes[pkg.OrdID]
		if err := s.resyncPackageInTx(ctx, resourceType, resourceID, packagesFromDB, pkg, pkgHash); err != nil {
			return nil, err
		}
	}

	packagesFromDB, err = s.listPackagesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return packagesFromDB, nil
}

func (s *Service) listPackagesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Package, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var packagesFromDB []*model.Package
	if resourceType == directorresource.Application {
		packagesFromDB, err = s.packageSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		packagesFromDB, err = s.packageSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing packages for %s with id %q", resourceType, resourceID)
	}

	return packagesFromDB, tx.Commit()
}

func (s *Service) resyncPackageInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, packagesFromDB []*model.Package, pkg *model.PackageInput, pkgHash uint64) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncPackage(ctx, resourceType, resourceID, packagesFromDB, *pkg, pkgHash); err != nil {
		return errors.Wrapf(err, "error while resyncing package with ORD ID %q", pkg.OrdID)
	}
	return tx.Commit()
}

func (s *Service) processBundles(ctx context.Context, resourceType directorresource.Type, resourceID string, bundles []*model.BundleCreateInput, resourceHashes map[string]uint64) ([]*model.Bundle, error) {
	bundlesFromDB, err := s.listBundlesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	credentialExchangeStrategyHashCurrent := uint64(0)
	var credentialExchangeStrategyJSON gjson.Result
	for _, bndl := range bundles {
		bndlHash := resourceHashes[str.PtrStrToStr(bndl.OrdID)]
		if err := s.resyncBundleInTx(ctx, resourceType, resourceID, bundlesFromDB, bndl, bndlHash); err != nil {
			return nil, err
		}

		credentialExchangeStrategies, err := bndl.CredentialExchangeStrategies.MarshalJSON()
		if err != nil {
			return nil, errors.Wrapf(err, "while marshalling credential exchange strategies for %s with ID %s", resourceType, resourceID)
		}

		for _, credentialExchangeStrategy := range gjson.ParseBytes(credentialExchangeStrategies).Array() {
			customType := credentialExchangeStrategy.Get(customTypeProperty).String()
			isTenantMappingType := strings.Contains(customType, TenantMappingCustomTypeIdentifier)

			if !isTenantMappingType {
				continue
			}

			currentHash, err := HashObject(credentialExchangeStrategy)
			if err != nil {
				return nil, errors.Wrapf(err, "while hasing credential exchange strategy for application with ID %s", resourceID)
			}

			if credentialExchangeStrategyHashCurrent != 0 && currentHash != credentialExchangeStrategyHashCurrent {
				return nil, errors.Errorf("There are differences in the Credential Exchange Strategies for Tenant Mappings for application with ID %s. They should be the same.", resourceID)
			}

			credentialExchangeStrategyHashCurrent = currentHash
			credentialExchangeStrategyJSON = credentialExchangeStrategy
		}
	}

	if err := s.resyncTenantMappingWebhooksInTx(ctx, credentialExchangeStrategyJSON, resourceID); err != nil {
		return nil, err
	}

	bundlesFromDB, err = s.listBundlesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	return bundlesFromDB, nil
}

func (s *Service) listBundlesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Bundle, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var bundlesFromDB []*model.Bundle
	if resourceType == directorresource.Application {
		bundlesFromDB, err = s.bundleSvc.ListByApplicationIDNoPaging(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		bundlesFromDB, err = s.bundleSvc.ListByApplicationTemplateVersionIDNoPaging(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing bundles for %s with id %q", resourceType, resourceID)
	}

	return bundlesFromDB, tx.Commit()
}

func (s *Service) resyncBundleInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, bundle *model.BundleCreateInput, bndlHash uint64) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncBundle(ctx, resourceType, resourceID, bundlesFromDB, *bundle, bndlHash); err != nil {
		return errors.Wrapf(err, "error while resyncing bundle with ORD ID %q", *bundle.OrdID)
	}
	return tx.Commit()
}

func (s *Service) resyncTenantMappingWebhooksInTx(ctx context.Context, credentialExchangeStrategyJSON gjson.Result, appID string) error {
	if !credentialExchangeStrategyJSON.IsObject() {
		log.C(ctx).Debugf("There are no tenant mappings to resync")
		return nil
	}

	tenantMappingData, err := s.getTenantMappingData(credentialExchangeStrategyJSON, appID)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("Enriching tenant mapping webhooks for application with ID %s", appID)

	enrichedWebhooks, err := s.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks([]*graphql.WebhookInput{createWebhookInput(credentialExchangeStrategyJSON, tenantMappingData)})
	if err != nil {
		return errors.Wrapf(err, "while enriching webhooks with tenant mapping webhooks for application with ID %s", appID)
	}

	ctxWithoutTenant := context.Background()
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctxWithoutTenant, tx)

	ctxWithoutTenant = persistence.SaveToContext(ctxWithoutTenant, tx)
	ctxWithoutTenant = tenant.SaveToContext(ctxWithoutTenant, "", "")

	appWebhooksFromDB, err := s.webhookSvc.ListForApplicationGlobal(ctxWithoutTenant, appID)
	if err != nil {
		return errors.Wrapf(err, "while listing webhooks from application with ID %s", appID)
	}

	tenantMappingRelatedWebhooksFromDB, enrichedWhModels, enrichedWhModelInputs, err := s.processEnrichedWebhooks(enrichedWebhooks, appWebhooksFromDB)
	if err != nil {
		return err
	}

	isEqual, err := isWebhookDataEqual(tenantMappingRelatedWebhooksFromDB, enrichedWhModels)
	if err != nil {
		return err
	}

	if isEqual {
		log.C(ctxWithoutTenant).Infof("There are no differences in tenant mapping webhooks from the DB and the ORD document")
		return tx.Commit()
	}

	log.C(ctxWithoutTenant).Infof("There are differences in tenant mapping webhooks from the DB and the ORD document. Continuing the sync.")

	if err := s.deleteWebhooks(ctxWithoutTenant, tenantMappingRelatedWebhooksFromDB, appID); err != nil {
		return err
	}

	if err := s.createWebhooks(ctxWithoutTenant, enrichedWhModelInputs, appID); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *Service) deleteWebhooks(ctx context.Context, webhooks []*model.Webhook, appID string) error {
	for _, webhook := range webhooks {
		log.C(ctx).Infof("Deleting webhook with ID %s for application %s", webhook.ID, appID)
		if err := s.webhookSvc.Delete(ctx, webhook.ID, webhook.ObjectType); err != nil {
			log.C(ctx).Errorf("error while deleting webhook with ID %s", webhook.ID)
			return errors.Wrapf(err, "while deleting webhook with ID %s", webhook.ID)
		}
	}

	return nil
}

func (s *Service) createWebhooks(ctx context.Context, webhooks []*model.WebhookInput, appID string) error {
	for _, webhook := range webhooks {
		log.C(ctx).Infof("Creating webhook with type %s for application %s", webhook.Type, appID)
		if _, err := s.webhookSvc.Create(ctx, appID, *webhook, model.ApplicationWebhookReference); err != nil {
			log.C(ctx).Errorf("error while creating webhook for app %s with type %s", appID, webhook.Type)
			return errors.Wrapf(err, "error while creating webhook for app %s with type %s", appID, webhook.Type)
		}
	}

	return nil
}

func (s *Service) getTenantMappingData(credentialExchangeStrategyJSON gjson.Result, appID string) (CredentialExchangeStrategyTenantMapping, error) {
	tenantMappingType := credentialExchangeStrategyJSON.Get(customTypeProperty).String()
	tenantMappingData, ok := s.config.credentialExchangeStrategyTenantMappings[tenantMappingType]
	if !ok {
		return CredentialExchangeStrategyTenantMapping{}, errors.Errorf("Credential Exchange Strategy has invalid %s value: %s for application with ID %s", customTypeProperty, tenantMappingType, appID)
	}
	return tenantMappingData, nil
}

func (s *Service) processEnrichedWebhooks(enrichedWebhooks []*graphql.WebhookInput, webhooksFromDB []*model.Webhook) ([]*model.Webhook, []*model.Webhook, []*model.WebhookInput, error) {
	tenantMappingRelatedWebhooksFromDB := make([]*model.Webhook, 0)
	enrichedWebhookModels := make([]*model.Webhook, 0)
	enrichedWebhookModelInputs := make([]*model.WebhookInput, 0)

	for _, wh := range enrichedWebhooks {
		convertedIn, err := s.webhookConverter.InputFromGraphQL(wh)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "while converting the WebhookInput")
		}

		enrichedWebhookModelInputs = append(enrichedWebhookModelInputs, convertedIn)

		webhookModel := convertedIn.ToWebhook("", "", "")

		for _, webhookFromDB := range webhooksFromDB {
			if webhookFromDB.Type == convertedIn.Type {
				webhookModel.ID = webhookFromDB.ID
				webhookModel.ObjectType = webhookFromDB.ObjectType
				webhookModel.ObjectID = webhookFromDB.ObjectID
				webhookModel.CreatedAt = webhookFromDB.CreatedAt

				tenantMappingRelatedWebhooksFromDB = append(tenantMappingRelatedWebhooksFromDB, webhookFromDB)
				break
			}
		}

		enrichedWebhookModels = append(enrichedWebhookModels, webhookModel)
	}

	return tenantMappingRelatedWebhooksFromDB, enrichedWebhookModels, enrichedWebhookModelInputs, nil
}

func (s *Service) processAPIs(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apis []*model.APIDefinitionInput, resourceHashes map[string]uint64) ([]*model.APIDefinition, []*ordFetchRequest, error) {
	apisFromDB, err := s.listAPIsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*ordFetchRequest, 0)
	for _, api := range apis {
		apiHash := resourceHashes[str.PtrStrToStr(api.OrdID)]
		apiFetchRequests, err := s.resyncAPIInTx(ctx, resourceType, resourceID, apisFromDB, bundlesFromDB, packagesFromDB, api, apiHash)
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

	apisFromDB, err = s.listAPIsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}
	return apisFromDB, fetchRequests, nil
}

func (s *Service) listAPIsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.APIDefinition, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var apisFromDB []*model.APIDefinition

	if resourceType == directorresource.Application {
		apisFromDB, err = s.apiSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		apisFromDB, err = s.apiSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing apis for %s with id %q", resourceType, resourceID)
	}

	return apisFromDB, tx.Commit()
}

func (s *Service) resyncAPIInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api *model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := s.resyncAPI(ctx, resourceType, resourceID, apisFromDB, bundlesFromDB, packagesFromDB, *api, apiHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (s *Service) processEvents(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput, resourceHashes map[string]uint64) ([]*model.EventDefinition, []*ordFetchRequest, error) {
	eventsFromDB, err := s.listEventsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*ordFetchRequest, 0)
	for _, event := range events {
		eventHash := resourceHashes[str.PtrStrToStr(event.OrdID)]
		eventFetchRequests, err := s.resyncEventInTx(ctx, resourceType, resourceID, eventsFromDB, bundlesFromDB, packagesFromDB, event, eventHash)
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

	eventsFromDB, err = s.listEventsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}
	return eventsFromDB, fetchRequests, nil
}

func (s *Service) listEventsInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.EventDefinition, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var eventsFromDB []*model.EventDefinition
	if resourceType == directorresource.Application {
		eventsFromDB, err = s.eventSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		eventsFromDB, err = s.eventSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing events for %s with id %q", resourceType, resourceID)
	}

	return eventsFromDB, tx.Commit()
}

func (s *Service) resyncEventInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event *model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := s.resyncEvent(ctx, resourceType, resourceID, eventsFromDB, bundlesFromDB, packagesFromDB, *event, eventHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing event with ORD ID %q", *event.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (s *Service) processTombstones(ctx context.Context, resourceType directorresource.Type, resourceID string, tombstones []*model.TombstoneInput) ([]*model.Tombstone, error) {
	tombstonesFromDB, err := s.listTombstonesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, tombstone := range tombstones {
		if err := s.resyncTombstoneInTx(ctx, resourceType, resourceID, tombstonesFromDB, tombstone); err != nil {
			return nil, err
		}
	}

	tombstonesFromDB, err = s.listTombstonesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return tombstonesFromDB, nil
}

func (s *Service) listTombstonesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Tombstone, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var tombstonesFromDB []*model.Tombstone
	if resourceType == directorresource.Application {
		tombstonesFromDB, err = s.tombstoneSvc.ListByApplicationID(ctx, resourceID)
	} else if resourceType == directorresource.ApplicationTemplateVersion {
		tombstonesFromDB, err = s.tombstoneSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing tombstones for %s with id %q", resourceType, resourceID)
	}

	return tombstonesFromDB, tx.Commit()
}

func (s *Service) resyncTombstoneInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, tombstonesFromDB []*model.Tombstone, tombstone *model.TombstoneInput) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := s.resyncTombstone(ctx, resourceType, resourceID, tombstonesFromDB, *tombstone); err != nil {
		return errors.Wrapf(err, "error while resyncing tombstone for resource with ORD ID %q", tombstone.OrdID)
	}
	return tx.Commit()
}

func (s *Service) resyncPackage(ctx context.Context, resourceType directorresource.Type, resourceID string, packagesFromDB []*model.Package, pkg model.PackageInput, pkgHash uint64) error {
	ctx = addFieldToLogger(ctx, "package_ord_id", pkg.OrdID)
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return packagesFromDB[i].OrdID == pkg.OrdID
	}); found {
		return s.packageSvc.Update(ctx, resourceType, packagesFromDB[i].ID, pkg, pkgHash)
	}

	_, err := s.packageSvc.Create(ctx, resourceType, resourceID, pkg, pkgHash)
	return err
}

func (s *Service) resyncBundle(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput, bndlHash uint64) error {
	ctx = addFieldToLogger(ctx, "bundle_ord_id", *bndl.OrdID)
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		return s.bundleSvc.UpdateBundle(ctx, resourceType, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl), bndlHash)
	}

	_, err := s.bundleSvc.CreateBundle(ctx, resourceType, resourceID, bndl, bndlHash)
	return err
}

func (s *Service) resyncProduct(ctx context.Context, resourceType directorresource.Type, resourceID string, productsFromDB []*model.Product, product model.ProductInput) error {
	ctx = addFieldToLogger(ctx, "product_ord_id", product.OrdID)
	if i, found := searchInSlice(len(productsFromDB), func(i int) bool {
		return productsFromDB[i].OrdID == product.OrdID
	}); found {
		return s.productSvc.Update(ctx, resourceType, productsFromDB[i].ID, product)
	}

	_, err := s.productSvc.Create(ctx, resourceType, resourceID, product)
	return err
}

func (s *Service) resyncVendor(ctx context.Context, resourceType directorresource.Type, resourceID string, vendorsFromDB []*model.Vendor, vendor model.VendorInput) error {
	ctx = addFieldToLogger(ctx, "vendor_ord_id", vendor.OrdID)
	if i, found := searchInSlice(len(vendorsFromDB), func(i int) bool {
		return vendorsFromDB[i].OrdID == vendor.OrdID
	}); found {
		return s.vendorSvc.Update(ctx, resourceType, vendorsFromDB[i].ID, vendor)
	}

	_, err := s.vendorSvc.Create(ctx, resourceType, resourceID, vendor)
	return err
}

func (s *Service) resyncAppTemplateVersion(ctx context.Context, appTemplateID string, appTemplateVersionsFromDB []*model.ApplicationTemplateVersion, appTemplateVersion *model.ApplicationTemplateVersionInput) error {
	ctx = addFieldToLogger(ctx, "app_template_version_id", appTemplateID)
	if i, found := searchInSlice(len(appTemplateVersionsFromDB), func(i int) bool {
		return appTemplateVersionsFromDB[i].Version == appTemplateVersion.Version
	}); found {
		return s.appTemplateVersionSvc.Update(ctx, appTemplateVersionsFromDB[i].ID, appTemplateID, *appTemplateVersion)
	}

	_, err := s.appTemplateVersionSvc.Create(ctx, appTemplateID, appTemplateVersion)
	return err
}

func (s *Service) resyncAPI(ctx context.Context, resourceType directorresource.Type, resourceID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
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
		apiID, err := s.apiSvc.Create(ctx, resourceType, resourceID, nil, packageID, api, nil, defaultTargetURLPerBundle, apiHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}

		fr, err := s.createSpecs(ctx, model.APISpecReference, apiID, specs, resourceType)
		if err != nil {
			return nil, err
		}

		return fr, nil
	}

	allBundleIDsForAPI, err := s.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &apisFromDB[i].ID)
	if err != nil {
		return nil, err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion := extractBundleReferencesForDeletion(allBundleIDsForAPI, defaultTargetURLPerBundle)

	// in case of API update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	defaultTargetURLPerBundleForCreation := extractAllBundleReferencesForCreation(defaultTargetURLPerBundle, allBundleIDsForAPI)

	if err = s.apiSvc.UpdateInManyBundles(ctx, resourceType, apisFromDB[i].ID, api, nil, defaultTargetURLPerBundle, defaultTargetURLPerBundleForCreation, bundleIDsForDeletion, apiHash, defaultConsumptionBundleID); err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	if api.VersionInput.Value != apisFromDB[i].Version.Value {
		fetchRequests, err = s.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs, resourceType)
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

func (s *Service) resyncEvent(ctx context.Context, resourceType directorresource.Type, resourceID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
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
		eventID, err := s.eventSvc.Create(ctx, resourceType, resourceID, nil, packageID, event, nil, bundleIDsFromBundleReference, eventHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}
		return s.createSpecs(ctx, model.EventSpecReference, eventID, specs, resourceType)
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

	if err = s.eventSvc.UpdateInManyBundles(ctx, resourceType, eventsFromDB[i].ID, event, nil, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion, eventHash, defaultConsumptionBundleID); err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	if event.VersionInput.Value != eventsFromDB[i].Version.Value {
		fetchRequests, err = s.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs, resourceType)
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

func (s *Service) createSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType directorresource.Type) ([]*model.FetchRequest, error) {
	fetchRequests := make([]*model.FetchRequest, 0)
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		_, fr, err := s.specSvc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, *spec, resourceType, objectType, objectID)
		if err != nil {
			return nil, err
		}
		fetchRequests = append(fetchRequests, fr)
	}
	return fetchRequests, nil
}

func (s *Service) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType directorresource.Type) ([]*model.FetchRequest, error) {
	if err := s.specSvc.DeleteByReferenceObjectID(ctx, resourceType, objectType, objectID); err != nil {
		return nil, err
	}
	return s.createSpecs(ctx, objectType, objectID, specs, resourceType)
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

func (s *Service) resyncTombstone(ctx context.Context, resourceType directorresource.Type, resourceID string, tombstonesFromDB []*model.Tombstone, tombstone model.TombstoneInput) error {
	if i, found := searchInSlice(len(tombstonesFromDB), func(i int) bool {
		return tombstonesFromDB[i].OrdID == tombstone.OrdID
	}); found {
		return s.tombstoneSvc.Update(ctx, resourceType, tombstonesFromDB[i].ID, tombstone)
	}

	_, err := s.tombstoneSvc.Create(ctx, resourceType, resourceID, tombstone)
	return err
}

func (s *Service) fetchAPIDefFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.APIDefinition, error) {
	var (
		apisFromDB []*model.APIDefinition
		err        error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		apisFromDB, err = s.apiSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		apisFromDB, err = s.apiSvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	apiDataFromDB := make(map[string]*model.APIDefinition, len(apisFromDB))

	for _, api := range apisFromDB {
		apiOrdID := str.PtrStrToStr(api.OrdID)
		apiDataFromDB[apiOrdID] = api
	}

	return apiDataFromDB, nil
}

func (s *Service) fetchPackagesFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.Package, error) {
	var (
		packagesFromDB []*model.Package
		err            error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		packagesFromDB, err = s.packageSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		packagesFromDB, err = s.packageSvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	packageDataFromDB := make(map[string]*model.Package)

	for _, pkg := range packagesFromDB {
		packageDataFromDB[pkg.OrdID] = pkg
	}

	return packageDataFromDB, nil
}

func (s *Service) fetchEventDefFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.EventDefinition, error) {
	var (
		eventsFromDB []*model.EventDefinition
		err          error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		eventsFromDB, err = s.eventSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		eventsFromDB, err = s.eventSvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	eventDataFromDB := make(map[string]*model.EventDefinition)

	for _, event := range eventsFromDB {
		eventOrdID := str.PtrStrToStr(event.OrdID)
		eventDataFromDB[eventOrdID] = event
	}

	return eventDataFromDB, nil
}

func (s *Service) fetchBundlesFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.Bundle, error) {
	var (
		bundlesFromDB []*model.Bundle
		err           error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		bundlesFromDB, err = s.bundleSvc.ListByApplicationTemplateVersionIDNoPaging(ctx, resourceID)
	} else {
		bundlesFromDB, err = s.bundleSvc.ListByApplicationIDNoPaging(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	bundleDataFromDB := make(map[string]*model.Bundle)

	for _, bndl := range bundlesFromDB {
		bndlOrdID := str.PtrStrToStr(bndl.OrdID)
		bundleDataFromDB[bndlOrdID] = bndl
	}

	return bundleDataFromDB, nil
}

func (s *Service) fetchResources(ctx context.Context, resource Resource, documents Documents) (ResourcesFromDB, error) {
	resourceIDs := make(map[string]directorresource.Type, 0)

	if resource.Type == directorresource.Application {
		resourceIDs[resource.ID] = directorresource.Application
	}

	for _, doc := range documents {
		if doc.DescribedSystemVersion != nil {
			appTemplateID := resource.ID
			if resource.Type == directorresource.Application && resource.ParentID != nil {
				appTemplateID = *resource.ParentID
			}

			appTemplateVersion, err := s.getApplicationTemplateVersionByAppTemplateIDAndVersionInTx(ctx, appTemplateID, doc.DescribedSystemVersion.Version)
			if err != nil {
				return ResourcesFromDB{}, err
			}
			resourceIDs[appTemplateVersion.ID] = directorresource.ApplicationTemplateVersion
		}
	}

	tx, err := s.transact.Begin()
	if err != nil {
		return ResourcesFromDB{}, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	apiDataFromDB := make(map[string]*model.APIDefinition)
	eventDataFromDB := make(map[string]*model.EventDefinition)
	packageDataFromDB := make(map[string]*model.Package)
	bundleDataFromDB := make(map[string]*model.Bundle)

	for resourceID, resourceType := range resourceIDs {
		apiData, err := s.fetchAPIDefFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching apis for %s with id %s", resourceType, resourceID)
		}

		eventData, err := s.fetchEventDefFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching events for %s with id %s", resourceType, resourceID)
		}

		packageData, err := s.fetchPackagesFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching packages for %s with id %s", resourceType, resourceID)
		}

		bundleData, err := s.fetchBundlesFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching bundles for %s with id %s", resourceType, resourceID)
		}

		if err = mergo.Merge(&apiDataFromDB, apiData); err != nil {
			return ResourcesFromDB{}, err
		}
		if err = mergo.Merge(&eventDataFromDB, eventData); err != nil {
			return ResourcesFromDB{}, err
		}
		if err = mergo.Merge(&packageDataFromDB, packageData); err != nil {
			return ResourcesFromDB{}, err
		}
		if err = mergo.Merge(&bundleDataFromDB, bundleData); err != nil {
			return ResourcesFromDB{}, err
		}
	}

	return ResourcesFromDB{
		APIs:     apiDataFromDB,
		Events:   eventDataFromDB,
		Packages: packageDataFromDB,
		Bundles:  bundleDataFromDB,
	}, tx.Commit()
}

func (s *Service) processWebhookAndDocuments(ctx context.Context, cfg MetricsConfig, webhook *model.Webhook, resource Resource, globalResourcesOrdIDs map[string]bool) error {
	var (
		documents Documents
		baseURL   string
		err       error
	)

	metricsCfg := metrics.PusherConfig{
		Enabled:    len(cfg.PushEndpoint) > 0,
		Endpoint:   cfg.PushEndpoint,
		MetricName: strings.ReplaceAll(strings.ToLower(cfg.JobName), "-", "_") + "_job_sync_failure_number",
		Timeout:    cfg.ClientTimeout,
		Subsystem:  metrics.OrdAggregatorSubsystem,
		Labels:     []string{metrics.ErrorMetricLabel, metrics.ResourceIDMetricLabel, metrics.ResourceTypeMetricLabel, metrics.CorrelationIDMetricLabel},
	}

	ctx = addFieldToLogger(ctx, "resource_id", resource.ID)
	ctx = addFieldToLogger(ctx, "resource_type", string(resource.Type))

	if webhook.Type == model.WebhookTypeOpenResourceDiscovery && webhook.URL != nil {
		documents, baseURL, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, resource, webhook)
		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			return errors.Wrapf(err, "error fetching ORD document for webhook with id %q: %v", webhook.ID, err)
		}
	}

	if len(documents) > 0 {
		log.C(ctx).Info("Processing ORD documents")
		var validationErrors error

		err = s.processDocuments(ctx, resource, baseURL, documents, globalResourcesOrdIDs, &validationErrors)
		if err != nil {
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
	internalTntID, err := s.tenantSvc.GetLowestOwnerForResource(ctx, directorresource.Application, appID)
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

	resource := Resource{
		Type:          directorresource.Application,
		ID:            app.ID,
		ParentID:      app.ApplicationTemplateID,
		Name:          app.Name,
		LocalTenantID: app.LocalTenantID,
	}
	if err = s.processWebhookAndDocuments(ctx, cfg, webhook, resource, globalResourcesOrdIDs); err != nil {
		return err
	}

	return nil
}

func (s *Service) processApplicationTemplateWebhook(ctx context.Context, cfg MetricsConfig, webhook *model.Webhook, appTemplateID string, globalResourcesOrdIDs map[string]bool) error {
	tx, err := s.transact.Begin()
	if err != nil {
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplate, err := s.appTemplateSvc.Get(ctx, appTemplateID)
	if err != nil {
		return errors.Wrapf(err, "error while retrieving app template with id %q", appTemplateID)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	resource := Resource{
		Type: directorresource.ApplicationTemplate,
		ID:   appTemplate.ID,
		Name: appTemplate.Name,
	}

	if err := s.processWebhookAndDocuments(ctx, cfg, webhook, resource, globalResourcesOrdIDs); err != nil {
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

func createWebhookInput(credentialExchangeStrategyJSON gjson.Result, tenantMappingData CredentialExchangeStrategyTenantMapping) *graphql.WebhookInput {
	inputMode := graphql.WebhookMode(tenantMappingData.Mode)
	return &graphql.WebhookInput{
		URL: str.Ptr(credentialExchangeStrategyJSON.Get(callbackURLProperty).String()),
		Auth: &graphql.AuthInput{
			AccessStrategy: str.Ptr(string(accessstrategy.CMPmTLSAccessStrategy)),
		},
		Mode:    &inputMode,
		Version: str.Ptr(tenantMappingData.Version),
	}
}

func isWebhookDataEqual(tenantMappingRelatedWebhooksFromDB, enrichedWhModels []*model.Webhook) (bool, error) {
	appWhsFromDBMarshaled, err := json.Marshal(tenantMappingRelatedWebhooksFromDB)
	if err != nil {
		return false, errors.Wrapf(err, "while marshalling webhooks from DB")
	}

	appWhsFromDBHash, err := HashObject(string(appWhsFromDBMarshaled))
	if err != nil {
		return false, errors.Wrapf(err, "while hashing webhooks from DB")
	}

	enrichedWhsMarshaled, err := json.Marshal(enrichedWhModels)
	if err != nil {
		return false, errors.Wrapf(err, "while marshalling webhooks from DB")
	}

	enrichedHash, err := HashObject(string(enrichedWhsMarshaled))
	if err != nil {
		return false, errors.Wrapf(err, "while hashing webhooks from ORD document")
	}

	if strconv.FormatUint(appWhsFromDBHash, 10) == strconv.FormatUint(enrichedHash, 10) {
		return true, nil
	}

	return false, nil
}
