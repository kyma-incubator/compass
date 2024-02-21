package ord

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mitchellh/hashstructure/v2"

	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/processor"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	operationsmanager "github.com/kyma-incubator/compass/components/director/internal/operations_manager"
	requestobject "github.com/kyma-incubator/compass/components/director/pkg/webhook"

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
	// TenantMappingCustomTypeIdentifier represents an identifier for tenant mapping webhooks in Credential exchange strategies
	TenantMappingCustomTypeIdentifier = "sap.ucl:tenant-mapping"

	customTypeProperty  = "customType"
	callbackURLProperty = "callbackUrl"

	// ValidationErrorMsg is the error message for validation error in ORD Documents
	ValidationErrorMsg = "error validating ORD documents"
	// ProcessingErrorMsg is the error message for processing error in ORD Documents
	ProcessingErrorMsg = "error processing ORD documents"
)

// RuntimeError represents the message of the runtime errors
type RuntimeError struct {
	Message string `json:"message"`
}

// ProcessingError represents the error containing the validation and runtime errors from aggregating the ORD documents
type ProcessingError struct {
	ValidationErrors []*ValidationError `json:"validation_errors"`
	RuntimeError     *RuntimeError      `json:"runtime_error"`
}

func (p *ProcessingError) Error() string {
	return p.toJSON()
}

func (p *ProcessingError) toJSON() string {
	bytes, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal error: %s"}`, err)
	}
	return string(bytes)
}

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
	apiProcessor                   APIProcessor
	eventSvc                       EventService
	eventProcessor                 EventProcessor
	entityTypeSvc                  EntityTypeService
	capabilitySvc                  CapabilityService
	capabilityProcessor            CapabilityProcessor
	integrationDependencySvc       IntegrationDependencyService
	dataProductSvc                 DataProductService
	specSvc                        SpecService
	fetchReqSvc                    FetchRequestService
	packageSvc                     PackageService
	packageProcessor               PackageProcessor
	productProcessor               ProductProcessor
	vendorProcessor                VendorProcessor
	tombstoneProcessor             TombstoneProcessor
	entityTypeProcessor            EntityTypeProcessor
	integrationDependencyProcessor IntegrationDependencyProcessor
	dataProductProcessor           DataProductProcessor
	tenantSvc                      TenantService
	appTemplateVersionSvc          ApplicationTemplateVersionService
	appTemplateSvc                 ApplicationTemplateService
	tombstonedResourcesDeleter     TombstonedResourcesDeleter
	labelSvc                       LabelService
	opSvc                          operationsmanager.OperationService

	ordWebhookMapping []application.ORDWebhookMapping

	webhookConverter WebhookConverter

	globalRegistrySvc GlobalRegistryService
	ordClient         Client
	documentValidator Validator
}

// TODO remove redundant services

// NewAggregatorService returns a new object responsible for service-layer ORD operations.
func NewAggregatorService(config ServiceConfig, metricsCfg MetricsConfig, transact persistence.Transactioner, appSvc ApplicationService, webhookSvc WebhookService, bundleSvc BundleService, bundleReferenceSvc BundleReferenceService, apiSvc APIService, apiProcessor APIProcessor, eventSvc EventService, eventProcessor EventProcessor, entityTypeSvc EntityTypeService, entityTypeProcessor EntityTypeProcessor, capabilitySvc CapabilityService, capabilityProcessor CapabilityProcessor, integrationDependencySvc IntegrationDependencyService, integrationDependencyProcessor IntegrationDependencyProcessor, dataProductSvc DataProductService, dataProductProcessor DataProductProcessor, specSvc SpecService, fetchReqSvc FetchRequestService, packageSvc PackageService, packageProcessor PackageProcessor, productProcessor ProductProcessor, vendorProcessor VendorProcessor, tombstoneProcessor TombstoneProcessor, tenantSvc TenantService, globalRegistrySvc GlobalRegistryService, client Client, webhookConverter WebhookConverter, appTemplateVersionSvc ApplicationTemplateVersionService, appTemplateSvc ApplicationTemplateService, tombstonedResourcesDeleter TombstonedResourcesDeleter, labelService LabelService, ordWebhookMapping []application.ORDWebhookMapping, opSvc operationsmanager.OperationService, documentValidator Validator) *Service {
	return &Service{
		config:                         config,
		metricsCfg:                     metricsCfg,
		transact:                       transact,
		appSvc:                         appSvc,
		webhookSvc:                     webhookSvc,
		bundleSvc:                      bundleSvc,
		bundleReferenceSvc:             bundleReferenceSvc,
		apiSvc:                         apiSvc,
		apiProcessor:                   apiProcessor,
		eventSvc:                       eventSvc,
		eventProcessor:                 eventProcessor,
		entityTypeSvc:                  entityTypeSvc,
		entityTypeProcessor:            entityTypeProcessor,
		capabilitySvc:                  capabilitySvc,
		capabilityProcessor:            capabilityProcessor,
		integrationDependencySvc:       integrationDependencySvc,
		integrationDependencyProcessor: integrationDependencyProcessor,
		dataProductSvc:                 dataProductSvc,
		dataProductProcessor:           dataProductProcessor,
		specSvc:                        specSvc,
		fetchReqSvc:                    fetchReqSvc,
		packageSvc:                     packageSvc,
		packageProcessor:               packageProcessor,
		productProcessor:               productProcessor,
		vendorProcessor:                vendorProcessor,
		tombstoneProcessor:             tombstoneProcessor,
		tenantSvc:                      tenantSvc,
		globalRegistrySvc:              globalRegistrySvc,
		ordClient:                      client,
		webhookConverter:               webhookConverter,
		appTemplateVersionSvc:          appTemplateVersionSvc,
		appTemplateSvc:                 appTemplateSvc,
		tombstonedResourcesDeleter:     tombstonedResourcesDeleter,
		labelSvc:                       labelService,
		ordWebhookMapping:              ordWebhookMapping,
		opSvc:                          opSvc,
		documentValidator:              documentValidator,
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

			return s.processApplicationWebhook(ctx, wh, appID, globalResourcesOrdIDs)
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

			return s.processApplicationWebhook(ctx, wh, appID, globalResourcesOrdIDs)
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

			return s.processApplicationTemplateWebhook(ctx, wh, appTemplateID, globalResourcesOrdIDs)
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

func (s *Service) processDocuments(ctx context.Context, resource Resource, webhookBaseURL, webhookBaseProxyURL string, ordRequestObject requestobject.OpenResourceDiscoveryWebhookRequestObject, documents Documents, globalResourcesOrdIDs map[string]bool, docsString []string) ([]*ValidationError, error) {
	if _, err := s.processDescribedSystemVersions(ctx, resource, documents); err != nil {
		return nil, err
	}

	resourceHashes, err := hashResources(documents)
	if err != nil {
		return nil, err
	}

	validationErrors, err := s.documentValidator.Validate(ctx, documents, webhookBaseURL, globalResourcesOrdIDs, docsString)
	if err != nil {
		return validationErrors, err
	}

	documentSanitizer := NewDocumentSanitizer()
	validationErrorsFromSanitize, err := documentSanitizer.Sanitize(documents, webhookBaseURL, webhookBaseProxyURL)
	if err != nil {
		return validationErrors, errors.Wrap(err, "while sanitizing ORD documents")
	}

	if len(validationErrorsFromSanitize) > 0 {
		log.C(ctx).Errorf("Stopping aggregation of resource with id %s", resource.ID)
		return append(validationErrors, validationErrorsFromSanitize...), nil
	}

	//for _, e := range validationErrors {
	//	if e.Severity == ErrorSeverity {
	//		return validationErrors, nil
	//	}
	//}

	ordLocalTenantID := s.getUniqueLocalTenantID(documents)
	if ordLocalTenantID != "" && resource.LocalTenantID == nil {
		if err := s.appSvc.Update(ctx, resource.ID, model.ApplicationUpdateInput{LocalTenantID: str.Ptr(ordLocalTenantID)}); err != nil {
			return validationErrors, err
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
				return validationErrors, err
			}

			resourceToAggregate = Resource{
				ID:   appTemplateVersion.ID,
				Type: directorresource.ApplicationTemplateVersion,
			}
		}

		log.C(ctx).Infof("Starting processing vendors for %s with id: %q", resource.Type, resource.ID)
		vendorsFromDB, err := s.vendorProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Vendors)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing vendors for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing products for %s with id: %q", resource.Type, resource.ID)
		productsFromDB, err := s.productProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Products)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing products for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing packages for %s with id: %q", resource.Type, resource.ID)
		packagesFromDB, err := s.packageProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Packages, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing packages for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing bundles for %s with id: %q", resource.Type, resource.ID)
		bundlesFromDB, err := s.processBundles(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.ConsumptionBundles, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing bundles for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing apis for %s with id: %q", resource.Type, resource.ID)
		apisFromDB, apiFetchRequests, err := s.apiProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.APIResources, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing apis for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing events for %s with id: %q", resource.Type, resource.ID)
		eventsFromDB, eventFetchRequests, err := s.eventProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, bundlesFromDB, packagesFromDB, doc.EventResources, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing events for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing entity types for %s with id: %q", resource.Type, resource.ID)
		entityTypesFromDB, err := s.entityTypeProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.EntityTypes, packagesFromDB, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing entity types for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing capabilities for %s with id: %q", resource.Type, resource.ID)
		capabilitiesFromDB, capabilitiesFetchRequests, err := s.capabilityProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, packagesFromDB, doc.Capabilities, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing capabilities for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing integration dependencies for %s with id: %q", resource.Type, resource.ID)
		integrationDependenciesFromDB, err := s.integrationDependencyProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, packagesFromDB, doc.IntegrationDependencies, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing integration dependencies for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing data products for %s with id: %q", resource.Type, resource.ID)
		dataProductsFromDB, err := s.dataProductProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, packagesFromDB, doc.DataProducts, resourceHashes)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing data products for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing tombstones for %s with id: %q", resource.Type, resource.ID)
		tombstonesFromDB, err := s.tombstoneProcessor.Process(ctx, resourceToAggregate.Type, resourceToAggregate.ID, doc.Tombstones)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing tombstones for %s with id: %q", resource.Type, resource.ID)

		fetchRequests := appendFetchRequests(apiFetchRequests, eventFetchRequests, capabilitiesFetchRequests)
		log.C(ctx).Infof("Starting deleting tombstoned resources for %s with id: %q", resource.Type, resource.ID)

		fetchRequests, err = s.tombstonedResourcesDeleter.Delete(ctx, resourceToAggregate.Type, vendorsFromDB, productsFromDB, packagesFromDB, bundlesFromDB, apisFromDB, eventsFromDB, entityTypesFromDB, capabilitiesFromDB, integrationDependenciesFromDB, dataProductsFromDB, tombstonesFromDB, fetchRequests)
		if err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished deleting tombstoned resources for %s with id: %q", resource.Type, resource.ID)

		log.C(ctx).Infof("Starting processing specs for %s with id: %q", resource.Type, resource.ID)
		if err := s.processSpecs(ctx, resourceToAggregate.Type, fetchRequests, ordRequestObject); err != nil {
			return validationErrors, err
		}
		log.C(ctx).Infof("Finished processing specs for %s with id: %q", resource.Type, resource.ID)
	}

	return validationErrors, nil
}

func (s *Service) processSpecs(ctx context.Context, resourceType directorresource.Type, ordFetchRequests []*processor.OrdFetchRequest, ordRequestObject requestobject.OpenResourceDiscoveryWebhookRequestObject) error {
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

func (s *Service) resyncBundle(ctx context.Context, resourceType directorresource.Type, resourceID string, bundlesFromDB []*model.Bundle, bndl model.BundleCreateInput, bndlHash uint64) error {
	ctx = addFieldToLogger(ctx, "bundle_ord_id", *bndl.OrdID)
	if i, found := searchInSlice(len(bundlesFromDB), func(i int) bool {
		return equalStrings(bundlesFromDB[i].OrdID, bndl.OrdID)
	}); found {
		log.C(ctx).Infof("Calculate the newest lastUpdate time for Consumption Bundle")
		newestLastUpdateTime, err := processor.NewestLastUpdateTimestamp(bndl.LastUpdate, bundlesFromDB[i].LastUpdate, bundlesFromDB[i].ResourceHash, bndlHash)
		if err != nil {
			return errors.Wrap(err, "error while calculating the newest lastUpdate time for Consumption Bundle")
		}

		bndl.LastUpdate = newestLastUpdateTime

		return s.bundleSvc.UpdateBundle(ctx, resourceType, bundlesFromDB[i].ID, bundleUpdateInputFromCreateInput(bndl), bndlHash)
	}

	currentTime := time.Now().Format(time.RFC3339)
	bndl.LastUpdate = &currentTime

	_, err := s.bundleSvc.CreateBundle(ctx, resourceType, resourceID, bndl, bndlHash)
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

func (s *Service) processWebhookAndDocuments(ctx context.Context, webhook *model.Webhook, resource Resource, globalResourcesOrdIDs map[string]bool, ordWebhookMapping application.ORDWebhookMapping) error {
	var (
		documents      Documents
		docsString     []string
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
		documents, docsString, webhookBaseURL, err = s.ordClient.FetchOpenResourceDiscoveryDocuments(ctx, resource, webhook, ordWebhookMapping, ordRequestObject)
		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			return errors.Wrapf(err, "error fetching ORD document for webhook with id %q: %v", webhook.ID, err)
		}
	}

	if len(documents) > 0 {
		log.C(ctx).Infof("Processing ORD documents for resource %s with ID %s", resource.Type, resource.ID)

		validationErrors, err := s.processDocuments(ctx, resource, webhookBaseURL, ordWebhookMapping.ProxyURL, ordRequestObject, documents, globalResourcesOrdIDs, docsString)
		if validationErrors != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)

			for _, e := range validationErrors {
				metricsPusher.ReportAggregationFailureORD(ctx, fmt.Sprintf("%s|%s|%s", e.Severity, e.OrdID, e.Description))
			}
			//
			//var errorNotification string
			//
			//for _, e := range validationErrors {
			//	errorNotification += fmt.Sprintf("%s|%s|%s", e.OrdID, e.Description, e.Type)
			//}
			//metricsPusher.ReportAggregationFailureORD(ctx, errorNotification)

			// TODO revisit
			log.C(ctx).WithError(errors.New("Some error")).WithField("validation_errors", validationErrors).Error(ValidationErrorMsg)
		}

		if err != nil {
			metricsPusher := metrics.NewAggregationFailurePusher(metricsCfg)
			metricsPusher.ReportAggregationFailureORD(ctx, err.Error())

			log.C(ctx).WithError(err).Errorf("%s: %v", ProcessingErrorMsg, err)
		}

		if err != nil {
			return &ProcessingError{
				ValidationErrors: nil,
				RuntimeError:     &RuntimeError{Message: err.Error()},
			}
		}

		if len(validationErrors) == 0 {
			return nil
		}

		return &ProcessingError{
			ValidationErrors: validationErrors,
			RuntimeError:     nil,
		}
	}

	log.C(ctx).Info("Successfully processed ORD documents")

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

	return s.processWebhookAndDocuments(ctx, webhook, resource, globalResourcesOrdIDs, ordWebhookMapping)
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

	return s.processWebhookAndDocuments(ctx, webhook, resource, globalResourcesOrdIDs, ordWebhookMapping)
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

		for _, dataProductInput := range doc.DataProducts {
			normalizedDataProducts, err := normalizeDataProduct(dataProductInput)
			if err != nil {
				return nil, err
			}

			hash, err := HashObject(normalizedDataProducts)
			if err != nil {
				return nil, errors.Wrapf(err, "while hashing data product with ORD ID: %s", str.PtrStrToStr(normalizedDataProducts.OrdID))
			}

			resourceHashes[str.PtrStrToStr(dataProductInput.OrdID)] = hash
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

func appendFetchRequests(fetchRequestsSlices ...[]*processor.OrdFetchRequest) []*processor.OrdFetchRequest {
	result := make([]*processor.OrdFetchRequest, 0)
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

func normalizeAPIDefinition(api *model.APIDefinitionInput) (model.APIDefinitionInput, error) {
	bytes, err := json.Marshal(api)
	if err != nil {
		return model.APIDefinitionInput{}, errors.Wrapf(err, "error while marshalling api definition with ID %s", str.PtrStrToStr(api.OrdID))
	}

	var normalizedAPIDefinition model.APIDefinitionInput
	if err := json.Unmarshal(bytes, &normalizedAPIDefinition); err != nil {
		return model.APIDefinitionInput{}, errors.Wrapf(err, "error while unmarshalling api definition with ID %s", str.PtrStrToStr(api.OrdID))
	}

	return normalizedAPIDefinition, nil
}

func normalizeEventDefinition(event *model.EventDefinitionInput) (model.EventDefinitionInput, error) {
	bytes, err := json.Marshal(event)
	if err != nil {
		return model.EventDefinitionInput{}, errors.Wrapf(err, "error while marshalling event definition with ID %s", str.PtrStrToStr(event.OrdID))
	}

	var normalizedEventDefinition model.EventDefinitionInput
	if err := json.Unmarshal(bytes, &normalizedEventDefinition); err != nil {
		return model.EventDefinitionInput{}, errors.Wrapf(err, "error while unmarshalling event definition with ID %s", str.PtrStrToStr(event.OrdID))
	}

	return normalizedEventDefinition, nil
}

func normalizeEntityType(entityType *model.EntityTypeInput) (model.EntityTypeInput, error) {
	bytes, err := json.Marshal(entityType)
	if err != nil {
		return model.EntityTypeInput{}, errors.Wrapf(err, "error while marshalling entity type with ID %s", entityType.OrdID)
	}

	var normalizedEntityType model.EntityTypeInput
	if err := json.Unmarshal(bytes, &normalizedEntityType); err != nil {
		return model.EntityTypeInput{}, errors.Wrapf(err, "error while unmarshalling entity type with ID %s", entityType.OrdID)
	}

	return normalizedEntityType, nil
}

func normalizeCapability(capability *model.CapabilityInput) (model.CapabilityInput, error) {
	bytes, err := json.Marshal(capability)
	if err != nil {
		return model.CapabilityInput{}, errors.Wrapf(err, "error while marshalling capability with ID %s", str.PtrStrToStr(capability.OrdID))
	}

	var normalizedCapability model.CapabilityInput
	if err := json.Unmarshal(bytes, &normalizedCapability); err != nil {
		return model.CapabilityInput{}, errors.Wrapf(err, "error while unmarshalling capability with ID %s", str.PtrStrToStr(capability.OrdID))
	}

	return normalizedCapability, nil
}

func normalizeIntegrationDependency(integrationDependency *model.IntegrationDependencyInput) (model.IntegrationDependencyInput, error) {
	bytes, err := json.Marshal(integrationDependency)
	if err != nil {
		return model.IntegrationDependencyInput{}, errors.Wrapf(err, "error while marshalling integration dependency with ID %s", str.PtrStrToStr(integrationDependency.OrdID))
	}

	var normalizedIntegrationDependency model.IntegrationDependencyInput
	if err := json.Unmarshal(bytes, &normalizedIntegrationDependency); err != nil {
		return model.IntegrationDependencyInput{}, errors.Wrapf(err, "error while unmarshalling integration dependency with ID %s", str.PtrStrToStr(integrationDependency.OrdID))
	}

	return normalizedIntegrationDependency, nil
}

func normalizeDataProduct(dataProduct *model.DataProductInput) (model.DataProductInput, error) {
	bytes, err := json.Marshal(dataProduct)
	if err != nil {
		return model.DataProductInput{}, errors.Wrapf(err, "error while marshalling data product with ID %s", str.PtrStrToStr(dataProduct.OrdID))
	}

	var normalizedDataProduct model.DataProductInput
	if err := json.Unmarshal(bytes, &normalizedDataProduct); err != nil {
		return model.DataProductInput{}, errors.Wrapf(err, "error while unmarshalling data product with ID %s", str.PtrStrToStr(dataProduct.OrdID))
	}

	return normalizedDataProduct, nil
}

func normalizePackage(pkg *model.PackageInput) (model.PackageInput, error) {
	bytes, err := json.Marshal(pkg)
	if err != nil {
		return model.PackageInput{}, errors.Wrapf(err, "error while marshalling package definition with ID %s", pkg.OrdID)
	}

	var normalizedPkgDefinition model.PackageInput
	if err := json.Unmarshal(bytes, &normalizedPkgDefinition); err != nil {
		return model.PackageInput{}, errors.Wrapf(err, "error while unmarshalling package definition with ID %s", pkg.OrdID)
	}

	return normalizedPkgDefinition, nil
}

func normalizeBundle(bndl *model.BundleCreateInput) (model.BundleCreateInput, error) {
	bytes, err := json.Marshal(bndl)
	if err != nil {
		return model.BundleCreateInput{}, errors.Wrapf(err, "error while marshalling bundle definition with ID %v", bndl.OrdID)
	}

	var normalizedBndlDefinition model.BundleCreateInput
	if err := json.Unmarshal(bytes, &normalizedBndlDefinition); err != nil {
		return model.BundleCreateInput{}, errors.Wrapf(err, "error while unmarshalling bundle definition with ID %v", bndl.OrdID)
	}

	return normalizedBndlDefinition, nil
}

// HashObject hashes the given object
func HashObject(obj interface{}) (uint64, error) {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, &hashstructure.HashOptions{SlicesAsSets: true})
	if err != nil {
		return 0, errors.New("failed to hash the given object")
	}

	return hash, nil
}
