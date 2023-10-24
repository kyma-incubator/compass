package ord

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	requestobject "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	"dario.cat/mergo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/accessstrategy"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

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
	maxParallelSpecificationProcessors int

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
func NewServiceConfig(maxParallelSpecificationProcessors int, credentialExchangeStrategyTenantMappings map[string]CredentialExchangeStrategyTenantMapping) ServiceConfig {
	return ServiceConfig{
		maxParallelSpecificationProcessors:       maxParallelSpecificationProcessors,
		credentialExchangeStrategyTenantMappings: credentialExchangeStrategyTenantMappings,
	}
}

// Service consists of various resource services responsible for service-layer ORD operations.
type Service struct {
	config     ServiceConfig
	metricsCfg MetricsConfig

	transact persistence.Transactioner

	appSvc                         ApplicationService
	webhookSvc                     WebhookService
	bundleSvc                      BundleService
	bundleReferenceSvc             BundleReferenceService
	apiSvc                         APIService
	eventSvc                       EventService
	entityTypeSvc         		   EntityTypeService
	capabilitySvc                  CapabilityService
	integrationDependencySvc       IntegrationDependencyService
	specSvc                        SpecService
	fetchReqSvc                    FetchRequestService
	packageSvc                     PackageService
	productSvc                     ProductService
	vendorSvc                      VendorService
	tombstoneProcessor             TombstoneProcessor
	entityTypeProcessor   	       EntityTypeProcessor
	integrationDependencyProcessor IntegrationDependencyProcessor
	tenantSvc                      TenantService
	appTemplateVersionSvc          ApplicationTemplateVersionService
	appTemplateSvc                 ApplicationTemplateService
	labelSvc                       LabelService
	opSvc                          operationsmanager.OperationService

	ordWebhookMapping []application.ORDWebhookMapping

	webhookConverter WebhookConverter

	globalRegistrySvc GlobalRegistryService
	ordClient         Client
}

// NewAggregatorService returns a new object responsible for service-layer ORD operations.
func NewAggregatorService(config ServiceConfig, metricsCfg MetricsConfig, transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, eventSvc EventService, entityTypeSvc EntityTypeService, entityTypeProcessor EntityTypeProcessor, capabilitySvc CapabilityService, integrationDependencySvc IntegrationDependencyService, integrationDependencyProcessor IntegrationDependencyProcessor, specSvc SpecService, fetchReqSvc FetchRequestService, packageSvc PackageService, productSvc ProductService, vendorSvc VendorService, tombstoneProcessor TombstoneProcessor, tenantSvc TenantService, globalRegistrySvc GlobalRegistryService, client Client, webhookConverter WebhookConverter, appTemplateVersionSvc ApplicationTemplateVersionService, appTemplateSvc ApplicationTemplateService, labelService LabelService, ordWebhookMapping []application.ORDWebhookMapping, opSvc operationsmanager.OperationService) *Service {
	return &Service{
		config:                         config,
		metricsCfg:                     metricsCfg,
		transact:                       transact,
		appSvc:                         appSvc,
		webhookSvc:                     webhookSvc,
		bundleSvc:                      bundleSvc,
		bundleReferenceSvc:             bundleReferenceSvc,
		apiSvc:                         apiSvc,
		eventSvc:                       eventSvc,
		entityTypeSvc:         			entityTypeSvc,
		entityTypeProcessor:   			entityTypeProcessor,
		capabilitySvc:                  capabilitySvc,
		integrationDependencySvc:       integrationDependencySvc,
		integrationDependencyProcessor: integrationDependencyProcessor,
		specSvc:                        specSvc,
		fetchReqSvc:                    fetchReqSvc,
		packageSvc:                     packageSvc,
		productSvc:                     productSvc,
		vendorSvc:                      vendorSvc,
		tombstoneProcessor:             tombstoneProcessor,
		tenantSvc:                      tenantSvc,
		globalRegistrySvc:              globalRegistrySvc,
		ordClient:                      client,
		webhookConverter:               webhookConverter,
		appTemplateVersionSvc:          appTemplateVersionSvc,
		appTemplateSvc:                 appTemplateSvc,
		labelSvc:                       labelService,
		ordWebhookMapping:              ordWebhookMapping,
		opSvc:                          opSvc,
	}
}

// ProcessApplication performs resync of ORD information provided via ORD documents for an applications
func (s *Service) ProcessApplication(ctx context.Context, appID string) error {
	ctx, err := s.saveLowestOwnerForAppToContextInTx(ctx, appID)
	if err != nil {
		return err
	}

	webhooks, err := s.getWebhooksForApplication(ctx, appID)
	if err != nil {
		return errors.Wrapf(err, "retrieving of webhooks for application with id %q failed", appID)
	}

	var globalResourcesOrdIDs map[string]bool
	globalResourcesLoaded := false

	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			// lazy loading of global ORD resources on first need
			if !globalResourcesLoaded {
				log.C(ctx).Infof("Retrieving global ORD resources")
				globalResourcesOrdIDs = s.retrieveGlobalResources(ctx)
				globalResourcesLoaded = true
			}
			log.C(ctx).Infof("Process Webhook ID %s for Application with ID %s", wh.ID, appID)
			if err = s.processApplicationWebhook(ctx, wh, appID, globalResourcesOrdIDs); err != nil {
				return errors.Wrapf(err, "processing of ORD webhook for application with id %q failed", appID)
			}
		}
	}
	return nil
}

// ProcessAppInAppTemplateContext performs resync of ORD information provided via ORD documents for an applications in context of application template
func (s *Service) ProcessAppInAppTemplateContext(ctx context.Context, appTemplateID, appID string) error {
	var globalResourcesOrdIDs map[string]bool
	globalResourcesLoaded := false

	webhooks, err := s.getWebhooksForApplicationTemplate(ctx, appTemplateID)
	if err != nil {
		return errors.Wrapf(err, "while retrieving all webhooks for application template with id %q", appTemplateID)
	}

	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscovery && wh.URL != nil {
			// lazy loading of global ORD resources on first need
			if !globalResourcesLoaded {
				log.C(ctx).Infof("Retrieving global ORD resources")
				globalResourcesOrdIDs = s.retrieveGlobalResources(ctx)
				globalResourcesLoaded = true
			}

			apps, err := s.getApplicationsForAppTemplate(ctx, appTemplateID)
			if err != nil {
				return errors.Wrapf(err, "retrieving of applications for application template with id %q failed", appTemplateID)
			}

			found := false
			for _, app := range apps {
				if app.ID == appID {
					found = true
					break
				}
			}
			if !found {
				return errors.Errorf("cannot find application with id %q for app template with id %q", appID, appTemplateID)
			}

			if err = s.processApplicationWebhook(ctx, wh, appID, globalResourcesOrdIDs); err != nil {
				return errors.Wrapf(err, "processing of ORD webhook for application with id %q failed", appID)
			}
		}
	}
	return nil
}

// ProcessApplicationTemplate performs resync of static ORD information for an application template
func (s *Service) ProcessApplicationTemplate(ctx context.Context, appTemplateID string) error {
	var globalResourcesOrdIDs map[string]bool
	globalResourcesLoaded := false

	webhooks, err := s.getWebhooksForApplicationTemplate(ctx, appTemplateID)
	if err != nil {
		return errors.Wrapf(err, "retrieving of webhooks for application template with id %q failed", appTemplateID)
	}

	for _, wh := range webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscoveryStatic && wh.URL != nil {
			// lazy loading of global ORD resources on first need
			if !globalResourcesLoaded {
				log.C(ctx).Infof("Retrieving global ORD resources")
				globalResourcesOrdIDs = s.retrieveGlobalResources(ctx)
				globalResourcesLoaded = true
			}

			log.C(ctx).Infof("Processing Webhook ID %s for Application Tempalate with ID %s", wh.ID, appTemplateID)
			if err = s.processApplicationTemplateWebhook(ctx, wh, appTemplateID, globalResourcesOrdIDs); err != nil {
				return err
			}
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
	webhooks, err := s.webhookSvc.ListForApplicationTemplate(ctx, appTemplateID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks for application template with id %s", appTemplateID)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (s *Service) getWebhooksForApplication(ctx context.Context, appID string) ([]*model.Webhook, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	webhooks, err := s.webhookSvc.ListForApplication(ctx, appID)
	if err != nil {
		log.C(ctx).WithError(err).Errorf("error while fetching webhooks for application with id %s", appID)
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return webhooks, nil
}

func (s *Service) processDocuments(ctx context.Context, resource Resource, webhookBaseURL, webhookBaseProxyURL string, ordRequestObject requestobject.OpenResourceDiscoveryWebhookRequestObject, documents Documents, globalResourcesOrdIDs map[string]bool, validationErrors *error) error {
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

	validationResult := documents.Validate(webhookBaseURL, resourcesFromDB, resourceHashes, globalResourcesOrdIDs, s.config.credentialExchangeStrategyTenantMappings)
	if validationResult != nil {
		validationResult = &ORDDocumentValidationError{errors.Wrap(validationResult, "invalid documents")}
		*validationErrors = validationResult
	}

	if err := documents.Sanitize(webhookBaseURL, webhookBaseProxyURL); err != nil {
		return errors.Wrap(err, "while sanitizing ORD documents")
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

		log.C(ctx).Infof("Starting processing vendors for %s with id: %q", resource.Type, resource.ID)
		vendorsFromDB, err := s.processVendors(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Vendors)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing vendors for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing products for %s with id: %q", resource.Type, resource.ID)
		productsFromDB, err := s.processProducts(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Products)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing products for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing packages for %s with id: %q", resource.Type, resource.ID)
		packagesFromDB, err := s.processPackages(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Packages, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing packages for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing bundles for %s with id: %q", resource.Type, resource.ID)
		bundlesFromDB, err := s.processBundles(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.ConsumptionBundles, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing bundles for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing apis for %s with id: %q", resource.Type, resource.ID)
		apisFromDB, apiFetchRequests, err := s.processAPIs(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.APIResources, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing apis for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing events for %s with id: %q", resource.Type, resource.ID)
		eventsFromDB, eventFetchRequests, err := s.processEvents(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.EventResources, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing events for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing entity types for %s with id: %q", resource.Type, resource.ID)
		entityTypesFromDB, err := s.entityTypeProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.EntityTypes, packagesFromDB, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing entity types for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing capabilities for %s with id: %q", resource.Type, resource.ID)
		capabilitiesFromDB, capabilitiesFetchRequests, err := s.processCapabilities(ctx, resourceToAggregate.Type, resourceToAggregate.ID, packagesFromDB, doc.Capabilities, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing capabilities for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing integration dependencies for %s with id: %q", resource.Type, resource.ID)
		integrationDependenciesFromDB, err := s.integrationDependencyProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, packagesFromDB, doc.IntegrationDependencies, resourceHashes)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing integration dependencies for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing tombstones for %s with id: %q", resource.Type, resource.ID)
		tombstonesFromDB, err := s.tombstoneProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Tombstones)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing tombstones for %s with id: %q", resource.Type, resource.ID)

		fetchRequests := appendFetchRequests(apiFetchRequests, eventFetchRequests, capabilitiesFetchRequests)
		log.C(ctx).Infof("Starting deleting tombstoned resources for %s with id: %q", resource.Type, resource.ID)
		fetchRequests, err = s.deleteTombstonedResources(ctx, resourceToAggregate.Type, vendorsFromDB, productsFromDB, packagesFromDB, bundlesFromDB, apisFromDB, eventsFromDB, entityTypesFromDB, capabilitiesFromDB, integrationDependenciesFromDB, tombstonesFromDB, fetchRequests)
		if err != nil {
			return err
		}
		log.C(ctx).Infof("Finished deleting tombstoned resources for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing specs for %s with id: %q", resource.Type, resource.ID)
		if err := s.processSpecs(ctx, resourceToAggregate.Type, fetchRequests, ordRequestObject); err != nil {
			return err
		}
		log.C(ctx).Infof("Finished processing specs for %s with id: %q", resource.Type, resource.ID)
	}

	return nil
}

func (s *Service) processSpecs(ctx context.Context, resourceType directorresource.Type, ordFetchRequests []*ordFetchRequest, ordRequestObject requestobject.OpenResourceDiscoveryWebhookRequestObject) error {
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
				data, status := s.fetchReqSvc.FetchSpec(ctx, &fr, ordRequestObject.Headers)
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
			err = s.processFetchRequestResultGlobal(ctx, result)
		} else {
			err = s.processFetchRequestResult(ctx, result)
		}
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Service) processFetchRequestResult(ctx context.Context, result *fetchRequestResult) error {
	specReferenceType := model.APISpecReference
	switch result.fetchRequest.ObjectType {
	case model.EventSpecFetchRequestReference:
		specReferenceType = model.EventSpecReference
	case model.CapabilitySpecFetchRequestReference:
		specReferenceType = model.CapabilitySpecReference
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

func (s *Service) deleteTombstonedResources(ctx context.Context, resourceType directorresource.Type, vendorsFromDB []*model.Vendor, productsFromDB []*model.Product, packagesFromDB []*model.Package, bundlesFromDB []*model.Bundle, apisFromDB []*model.APIDefinition, eventsFromDB []*model.EventDefinition, entityTypesFromDB []*model.EntityType, capabilitiesFromDB []*model.Capability, integrationDependenciesFromDB []*model.IntegrationDependency, tombstonesFromDB []*model.Tombstone, fetchRequests []*ordFetchRequest) ([]*ordFetchRequest, error) {
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
		if i, found := searchInSlice(len(entityTypesFromDB), func(i int) bool {
			return equalStrings(&entityTypesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.entityTypeSvc.Delete(ctx, resourceType, eventsFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(capabilitiesFromDB), func(i int) bool {
			return equalStrings(capabilitiesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.capabilitySvc.Delete(ctx, resourceType, capabilitiesFromDB[i].ID); err != nil {
				return nil, errors.Wrapf(err, "error while deleting resource with ORD ID %q based on its tombstone", ts.OrdID)
			}
		}
		if i, found := searchInSlice(len(integrationDependenciesFromDB), func(i int) bool {
			return equalStrings(integrationDependenciesFromDB[i].OrdID, &ts.OrdID)
		}); found {
			if err := s.integrationDependencySvc.Delete(ctx, resourceType, integrationDependenciesFromDB[i].ID); err != nil {
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

		if err = s.resyncApplicationTemplateVersionInTx(ctx, appTemplateID, appTemplateVersions, document.DescribedSystemVersion); err != nil {
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
	switch resourceType {
	case directorresource.Application:
		vendorsFromDB, err = s.vendorSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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

	if err = s.resyncAppTemplateVersion(ctx, appTemplateID, appTemplateVersionsFromDB, appTemplateVersion); err != nil {
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
	switch resourceType {
	case directorresource.Application:
		productsFromDB, err = s.productSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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
	switch resourceType {
	case directorresource.Application:
		packagesFromDB, err = s.packageSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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

	if err = s.resyncTenantMappingWebhooksInTx(ctx, credentialExchangeStrategyJSON, resourceID); err != nil {
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
	switch resourceType {
	case directorresource.Application:
		bundlesFromDB, err = s.bundleSvc.ListByApplicationIDNoPaging(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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

	switch resourceType {
	case directorresource.Application:
		apisFromDB, err = s.apiSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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
	switch resourceType {
	case directorresource.Application:
		eventsFromDB, err = s.eventSvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
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

func (s *Service) processCapabilities(ctx context.Context, resourceType directorresource.Type, resourceID string, packagesFromDB []*model.Package, capabilities []*model.CapabilityInput, resourceHashes map[string]uint64) ([]*model.Capability, []*ordFetchRequest, error) {
	capabilitiesFromDB, err := s.listCapabilitiesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*ordFetchRequest, 0)
	for _, capability := range capabilities {
		capabilityHash := resourceHashes[str.PtrStrToStr(capability.OrdID)]
		capabilityFetchRequests, err := s.resyncCapabilitiesInTx(ctx, resourceType, resourceID, capabilitiesFromDB, packagesFromDB, capability, capabilityHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range capabilityFetchRequests {
			fetchRequests = append(fetchRequests, &ordFetchRequest{
				FetchRequest:   capabilityFetchRequests[i],
				refObjectOrdID: *capability.OrdID,
			})
		}
	}

	capabilitiesFromDB, err = s.listCapabilitiesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	return capabilitiesFromDB, fetchRequests, nil
}

func (s *Service) listCapabilitiesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string) ([]*model.Capability, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var capabilitiesFromDB []*model.Capability

	switch resourceType {
	case directorresource.Application:
		capabilitiesFromDB, err = s.capabilitySvc.ListByApplicationID(ctx, resourceID)
	case directorresource.ApplicationTemplateVersion:
		capabilitiesFromDB, err = s.capabilitySvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing capabilities for %s with id %q", resourceType, resourceID)
	}

	return capabilitiesFromDB, nil
}

func (s *Service) resyncCapabilitiesInTx(ctx context.Context, resourceType directorresource.Type, resourceID string, capabilitiesFromDB []*model.Capability, packagesFromDB []*model.Package, capability *model.CapabilityInput, capabilityHash uint64) ([]*model.FetchRequest, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := s.resyncCapability(ctx, resourceType, resourceID, capabilitiesFromDB, packagesFromDB, *capability, capabilityHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing capability with ORD ID %q", *capability.OrdID)
	}
	return fetchRequests, tx.Commit()
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
	ctx = addFieldToLogger(ctx, "app_template_id", appTemplateID)
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

	shouldFetchSpecs, err := checkIfShouldFetchSpecs(api.LastUpdate, apisFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = s.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = s.refetchFailedSpecs(ctx, resourceType, model.APISpecReference, apisFromDB[i].ID)
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
	shouldFetchSpecs, err := checkIfShouldFetchSpecs(event.LastUpdate, eventsFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = s.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = s.refetchFailedSpecs(ctx, resourceType, model.EventSpecReference, eventsFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}

	return fetchRequests, nil
}

func (s *Service) resyncCapability(ctx context.Context, resourceType directorresource.Type, resourceID string, capabilitiesFromDB []*model.Capability, packagesFromDB []*model.Package, capability model.CapabilityInput, capabilityHash uint64) ([]*model.FetchRequest, error) {
	ctx = addFieldToLogger(ctx, "capability_ord_id", *capability.OrdID)
	i, isCapabilityFound := searchInSlice(len(capabilitiesFromDB), func(i int) bool {
		return equalStrings(capabilitiesFromDB[i].OrdID, capability.OrdID)
	})

	var packageID *string
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, capability.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	specs := make([]*model.SpecInput, 0, len(capability.CapabilityDefinitions))
	for _, resourceDef := range capability.CapabilityDefinitions {
		specs = append(specs, resourceDef.ToSpec())
	}

	if !isCapabilityFound {
		capabilityID, err := s.capabilitySvc.Create(ctx, resourceType, resourceID, packageID, capability, nil, capabilityHash)
		if err != nil {
			return nil, err
		}

		fetchRequests, err := s.createSpecs(ctx, model.CapabilitySpecReference, capabilityID, specs, resourceType)
		if err != nil {
			return nil, err
		}

		return fetchRequests, nil
	}

	err := s.capabilitySvc.Update(ctx, resourceType, capabilitiesFromDB[i].ID, capability, capabilityHash)
	if err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	shouldFetchSpecs, err := checkIfShouldFetchSpecs(capability.LastUpdate, capabilitiesFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = s.resyncSpecs(ctx, model.CapabilitySpecReference, capabilitiesFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = s.refetchFailedSpecs(ctx, resourceType, model.CapabilitySpecReference, capabilitiesFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return fetchRequests, nil
}

func checkIfShouldFetchSpecs(lastUpdateValueFromDoc, lastUpdateValueFromDB *string) (bool, error) {
	if lastUpdateValueFromDoc == nil || lastUpdateValueFromDB == nil {
		return true, nil
	}

	lastUpdateTimeFromDoc, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDoc))
	if err != nil {
		return false, err
	}

	lastUpdateTimeFromDB, err := time.Parse(time.RFC3339, str.PtrStrToStr(lastUpdateValueFromDB))
	if err != nil {
		return false, err
	}

	return lastUpdateTimeFromDoc.After(lastUpdateTimeFromDB), nil
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

func (s *Service) refetchFailedSpecs(ctx context.Context, resourceType directorresource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]*model.FetchRequest, error) {
	specIDsFromDB, err := s.specSvc.ListIDByReferenceObjectID(ctx, resourceType, objectType, objectID)
	if err != nil {
		return nil, err
	}

	var (
		fetchRequestsFromDB []*model.FetchRequest
		tnt                 string
	)
	if resourceType.IsTenantIgnorable() {
		fetchRequestsFromDB, err = s.specSvc.ListFetchRequestsByReferenceObjectIDsGlobal(ctx, specIDsFromDB, objectType)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}

		fetchRequestsFromDB, err = s.specSvc.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDsFromDB, objectType)
	}
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

func (s *Service) fetchCapabilitiesFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.Capability, error) {
	var (
		capabilitiesFromDB []*model.Capability
		err                error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		capabilitiesFromDB, err = s.capabilitySvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		capabilitiesFromDB, err = s.capabilitySvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	capabilitiesDataFromDB := make(map[string]*model.Capability, len(capabilitiesFromDB))

	for _, capability := range capabilitiesFromDB {
		capabilityOrdID := str.PtrStrToStr(capability.OrdID)
		capabilitiesDataFromDB[capabilityOrdID] = capability
	}

	return capabilitiesDataFromDB, nil
}

func (s *Service) fetchIntegrationDependenciesFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.IntegrationDependency, error) {
	var (
		integrationDependenciesFromDB []*model.IntegrationDependency
		err                           error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		integrationDependenciesFromDB, err = s.integrationDependencySvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		integrationDependenciesFromDB, err = s.integrationDependencySvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	integrationDependenciesDataFromDB := make(map[string]*model.IntegrationDependency, len(integrationDependenciesFromDB))

	for _, integrationDependency := range integrationDependenciesFromDB {
		integrationDependencyOrdID := str.PtrStrToStr(integrationDependency.OrdID)
		integrationDependenciesDataFromDB[integrationDependencyOrdID] = integrationDependency
	}

	return integrationDependenciesDataFromDB, nil
}

func (s *Service) fetchEntityTypesFromDB(ctx context.Context, resourceType directorresource.Type, resourceID string) (map[string]*model.EntityType, error) {
	var (
		entityTypesFromDB []*model.EntityType
		err               error
	)

	if resourceType == directorresource.ApplicationTemplateVersion {
		entityTypesFromDB, err = s.entityTypeSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	} else {
		entityTypesFromDB, err = s.entityTypeSvc.ListByApplicationID(ctx, resourceID)
	}
	if err != nil {
		return nil, err
	}

	entityTypesDataFromDB := make(map[string]*model.EntityType, len(entityTypesFromDB))

	for _, entityType := range entityTypesFromDB {
		entityTypesDataFromDB[entityType.OrdID] = entityType
	}

	return entityTypesDataFromDB, nil
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
	capabilitiesDataFromDB := make(map[string]*model.Capability)
	integrationDependenciesFromDB := make(map[string]*model.IntegrationDependency)
	entityTypesDataFromDB := make(map[string]*model.EntityType)

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

		capabilityData, err := s.fetchCapabilitiesFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching capabilities for %s with id %s", resourceType, resourceID)
		}

		integrationDependencyData, err := s.fetchIntegrationDependenciesFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching integration dependencies for %s with id %s", resourceType, resourceID)
		}

		entityTypeData, err := s.fetchEntityTypesFromDB(ctx, resourceType, resourceID)
		if err != nil {
			return ResourcesFromDB{}, errors.Wrapf(err, "while fetching entity types for %s with id %s", resourceType, resourceID)
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
		if err = mergo.Merge(&capabilitiesDataFromDB, capabilityData); err != nil {
			return ResourcesFromDB{}, err
		}
		if err = mergo.Merge(&integrationDependenciesFromDB, integrationDependencyData); err != nil {
			return ResourcesFromDB{}, err
		}
		if err = mergo.Merge(&entityTypesDataFromDB, entityTypeData); err != nil {
			return ResourcesFromDB{}, err
		}
	}

	return ResourcesFromDB{
		APIs:                    apiDataFromDB,
		Events:                  eventDataFromDB,
		Packages:                packageDataFromDB,
		Bundles:                 bundleDataFromDB,
		Capabilities:            capabilitiesDataFromDB,
		IntegrationDependencies: integrationDependenciesFromDB,
		EntityTypes:  			 entityTypesDataFromDB,
	}, tx.Commit()
}

func (s *Service) processWebhookAndDocuments(ctx context.Context, webhook *model.Webhook, resource Resource, globalResourcesOrdIDs map[string]bool, ordWebhookMapping application.ORDWebhookMapping) error {
	var (
		documents      Documents
		webhookBaseURL string
		err            error
	)

	metricsCfg := metrics.PusherConfig{
		Enabled:    len(s.metricsCfg.PushEndpoint) > 0,
		Endpoint:   s.metricsCfg.PushEndpoint,
		MetricName: strings.ReplaceAll(strings.ToLower(s.metricsCfg.JobName), "-", "_") + "_job_sync_failure_number",
		Timeout:    s.metricsCfg.ClientTimeout,
		Subsystem:  metrics.OrdAggregatorSubsystem,
		Labels:     []string{metrics.ErrorMetricLabel, metrics.ResourceIDMetricLabel, metrics.ResourceTypeMetricLabel, metrics.CorrelationIDMetricLabel},
	}

	ctx = addFieldToLogger(ctx, "resource_id", resource.ID)
	ctx = addFieldToLogger(ctx, "resource_type", string(resource.Type))

	var appBaseURL *string
	if resource.Type == directorresource.Application {
		tx, err := s.transact.Begin()
		if err != nil {
			return err
		}
		defer s.transact.RollbackUnlessCommitted(ctx, tx)

		ctx = persistence.SaveToContext(ctx, tx)

		ctx, err = s.saveLowestOwnerForAppToContext(ctx, resource.ID)
		if err != nil {
			return err
		}

		app, err := s.appSvc.Get(ctx, resource.ID)
		if err != nil {
			return errors.Wrapf(err, "error while retrieving app with id %q", resource.ID)
		}

		if err = tx.Commit(); err != nil {
			return err
		}

		appBaseURL = app.BaseURL
	}

	ordRequestObject := requestobject.OpenResourceDiscoveryWebhookRequestObject{
		Application: requestobject.Application{BaseURL: str.PtrStrToStr(appBaseURL)},
		Headers:     &sync.Map{},
	}

	if webhook.HeaderTemplate != nil {
		log.C(ctx).Infof("Header template found for webhook with ID %s. Will parse the template.", webhook.ID)
		headers, err := ordRequestObject.ParseHeadersTemplate(webhook.HeaderTemplate)
		if err != nil {
			return err
		}

		for key, value := range headers {
			ordRequestObject.Headers.Store(key, value[0])
		}
	}

	if (webhook.Type == model.WebhookTypeOpenResourceDiscovery || webhook.Type == model.WebhookTypeOpenResourceDiscoveryStatic) && webhook.URL != nil {
		documents, webhookBaseURL, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, resource, webhook, ordWebhookMapping, ordRequestObject)
		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			return errors.Wrapf(err, "error fetching ORD document for webhook with id %q: %v", webhook.ID, err)
		}
	}

	if len(documents) > 0 {
		log.C(ctx).Infof("Processing ORD documents for resource %s with ID %s", resource.Type, resource.ID)
		var validationError error

		err = s.processDocuments(ctx, resource, webhookBaseURL, ordWebhookMapping.ProxyURL, ordRequestObject, documents, globalResourcesOrdIDs, &validationError)
		if ordValidationError, ok := (validationError).(*ORDDocumentValidationError); ok {
			validationErrors := strings.Split(ordValidationError.Error(), MultiErrorSeparator)

			// the first item in the slice is the message 'invalid documents' for the wrapped errors
			validationErrors = validationErrors[1:]

			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)

			for i := range validationErrors {
				validationErrors[i] = strings.TrimSpace(validationErrors[i])
				metricsPusher.ReportAggregationFailureORD(ctx, validationErrors[i])
			}

			log.C(ctx).WithError(ordValidationError.Err).WithField("validation_errors", validationErrors).Error("error validating ORD documents")
		}
		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			log.C(ctx).WithError(err).Errorf("error processing ORD documents: %v", err)
			return errors.Wrap(err, "error processing ORD documents")
		}

		if validationError != nil {
			return errors.Wrap(err, "error validating ORD documents")
		}

		log.C(ctx).Info("Successfully processed ORD documents")
	}
	return nil
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

func (s *Service) processApplicationWebhook(ctx context.Context, webhook *model.Webhook, appID string, globalResourcesOrdIDs map[string]bool) error {
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

	ordWebhookMapping, err := s.getORDConfigForApplication(ctx, app.ID)
	if err != nil {
		return err
	}

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
	if err = s.processWebhookAndDocuments(ctx, webhook, resource, globalResourcesOrdIDs, ordWebhookMapping); err != nil {
		return errors.Wrapf(err, "while processing webhook %s for application %s", webhook.ID, appID)
	}

	return nil
}

func (s *Service) processApplicationTemplateWebhook(ctx context.Context, webhook *model.Webhook, appTemplateID string, globalResourcesOrdIDs map[string]bool) error {
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

	ordWebhookMapping := s.getMappingORDConfiguration(appTemplate.Name)

	resource := Resource{
		Type: directorresource.ApplicationTemplate,
		ID:   appTemplate.ID,
		Name: appTemplate.Name,
	}
	if err = s.processWebhookAndDocuments(ctx, webhook, resource, globalResourcesOrdIDs, ordWebhookMapping); err != nil {
		return err
	}

	return nil
}

func (s *Service) getMappingORDConfiguration(applicationType string) application.ORDWebhookMapping {
	for _, wm := range s.ordWebhookMapping {
		if wm.Type == applicationType {
			return wm
		}
	}
	return application.ORDWebhookMapping{}
}

func (s *Service) getORDConfigForApplication(ctx context.Context, appID string) (application.ORDWebhookMapping, error) {
	var ordWebhookMapping application.ORDWebhookMapping

	appTenant, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return application.ORDWebhookMapping{}, errors.Wrapf(err, "while loading tenant from context")
	}

	appTypeLbl, err := s.labelSvc.GetByKey(ctx, appTenant, model.ApplicationLabelableObject, appID, application.ApplicationTypeLabelKey)
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return application.ORDWebhookMapping{}, errors.Wrapf(err, "while getting label %q for %s with id %q", application.ApplicationTypeLabelKey, model.ApplicationLabelableObject, appID)
		}

		return application.ORDWebhookMapping{}, nil
	}

	if appTypeLbl != nil {
		ordWebhookMapping = s.getMappingORDConfiguration(appTypeLbl.Value.(string))
	}

	return ordWebhookMapping, nil
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

		for _, entityTypeInput := range doc.EntityTypes {
			normalizedEntityType, err := normalizeEntityType(entityTypeInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedEntityType)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing entity type with ORD ID: %s", normalizedEntityType.OrdID)
			}

			resourceHashes[entityTypeInput.OrdID] = hash
		}

		for _, capabilityInput := range doc.Capabilities {
			normalizedCapabilities, err := normalizeCapability(capabilityInput)
			if err != nil {
				return nil, err
			}
			hash, err := HashObject(normalizedCapabilities)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing capability with ORD ID: %s", str.PtrStrToStr(normalizedCapabilities.OrdID))
			}

			resourceHashes[str.PtrStrToStr(capabilityInput.OrdID)] = hash
		}

		for _, integrationDependencyInput := range doc.IntegrationDependencies {
			normalizedIntegrationDependencies, err := normalizeIntegrationDependency(integrationDependencyInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedIntegrationDependencies)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing integration dependency with ORD ID: %s", str.PtrStrToStr(normalizedIntegrationDependencies.OrdID))
			}

			resourceHashes[str.PtrStrToStr(integrationDependencyInput.OrdID)] = hash
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

func appendFetchRequests(fetchRequestsSlices ...[]*ordFetchRequest) []*ordFetchRequest {
	result := make([]*ordFetchRequest, 0)
	for _, frSlice := range fetchRequestsSlices {
		result = append(result, frSlice...)
	}

	return result
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

func (s *Service) saveLowestOwnerForAppToContextInTx(ctx context.Context, appID string) (context.Context, error) {
	tx, err := s.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)
	ctx, err = s.saveLowestOwnerForAppToContext(ctx, appID)
	if err != nil {
		return nil, err
	}

	return ctx, tx.Commit()
}
