package processor

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// APIService is responsible for the service-layer API operations.
//
//go:generate mockery --name=APIService --output=automock --outpkg=automock --case=underscore --disable-version-string
type APIService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput, targetURLsPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, packageID *string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundle map[string]string, defaultTargetURLPerBundleToBeCreated map[string]string, bundleIDsToBeDeleted []string, apiHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.APIDefinition, error)
}

// BundleReferenceService is responsible for the service-layer BundleReference operations.
//
//go:generate mockery --name=BundleReferenceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type BundleReferenceService interface {
	GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error)
}

// EntityTypeMappingService is responsible for processing of entity type entities.
//
//go:generate mockery --name=EntityTypeMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeMappingService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in *model.EntityTypeMappingInput) (string, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByOwnerResourceID(ctx context.Context, resourceID string, resourceType resource.Type) ([]*model.EntityTypeMapping, error)
}

// SpecService is responsible for the service-layer Specification operations.
//
//go:generate mockery --name=SpecService --output=automock --outpkg=automock --case=underscore --disable-version-string
type SpecService interface {
	CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, error)
	CreateByReferenceObjectIDWithDelayedFetchRequest(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, *model.FetchRequest, error)
	DeleteByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) error
	GetByID(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error)
	ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
	ListFetchRequestsByReferenceObjectIDsGlobal(ctx context.Context, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error)
	UpdateSpecOnly(ctx context.Context, spec model.Spec) error
	UpdateSpecOnlyGlobal(ctx context.Context, spec model.Spec) error
	ListIDByReferenceObjectID(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]string, error)
	GetByIDGlobal(ctx context.Context, id string) (*model.Spec, error)
}

// OrdFetchRequest defines fetch request
type OrdFetchRequest struct {
	*model.FetchRequest
	RefObjectOrdID string
}

// APIProcessor defines API processor
type APIProcessor struct {
	transact             persistence.Transactioner
	apiSvc               APIService
	entityTypeSvc        EntityTypeService
	entityTypeMappingSvc EntityTypeMappingService
	bundleReferenceSvc   BundleReferenceService
	specSvc              SpecService
}

// NewAPIProcessor creates new instance of APIProcessor
func NewAPIProcessor(transact persistence.Transactioner, apiSvc APIService, entityTypeSvc EntityTypeService, entityTypeMappingSvc EntityTypeMappingService, bundleReferenceSvc BundleReferenceService, specSvc SpecService) *APIProcessor {
	return &APIProcessor{
		transact:             transact,
		apiSvc:               apiSvc,
		entityTypeSvc:        entityTypeSvc,
		entityTypeMappingSvc: entityTypeMappingSvc,
		bundleReferenceSvc:   bundleReferenceSvc,
		specSvc:              specSvc,
	}
}

// Process re-syncs the apis passed as an argument.
func (ap *APIProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, apis []*model.APIDefinitionInput, resourceHashes map[string]uint64) ([]*model.APIDefinition, []*OrdFetchRequest, error) {
	apisFromDB, err := ap.listAPIsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*OrdFetchRequest, 0)
	for _, api := range apis {
		apiHash := resourceHashes[str.PtrStrToStr(api.OrdID)]
		apiFetchRequests, err := ap.resyncAPIInTx(ctx, resourceType, resourceID, apisFromDB, bundlesFromDB, packagesFromDB, api, apiHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range apiFetchRequests {
			fetchRequests = append(fetchRequests, &OrdFetchRequest{
				FetchRequest:   apiFetchRequests[i],
				RefObjectOrdID: *api.OrdID,
			})
		}
	}

	apisFromDB, err = ap.listAPIsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}
	return apisFromDB, fetchRequests, nil
}

func (ap *APIProcessor) listAPIsInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.APIDefinition, error) {
	tx, err := ap.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ap.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var apisFromDB []*model.APIDefinition

	switch resourceType {
	case resource.Application:
		apisFromDB, err = ap.apiSvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		apisFromDB, err = ap.apiSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing apis for %s with id %q", resourceType, resourceID)
	}

	return apisFromDB, tx.Commit()
}

func (ap *APIProcessor) resyncAPIInTx(ctx context.Context, resourceType resource.Type, resourceID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api *model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
	tx, err := ap.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ap.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := ap.resyncAPI(ctx, resourceType, resourceID, apisFromDB, bundlesFromDB, packagesFromDB, *api, apiHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing api with ORD ID %q", *api.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (ap *APIProcessor) resyncAPI(ctx context.Context, resourceType resource.Type, resourceID string, apisFromDB []*model.APIDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, api model.APIDefinitionInput, apiHash uint64) ([]*model.FetchRequest, error) {
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
		currentTime := time.Now().Format(time.RFC3339)
		api.LastUpdate = &currentTime

		apiID, err := ap.apiSvc.Create(ctx, resourceType, resourceID, nil, packageID, api, nil, defaultTargetURLPerBundle, apiHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}

		err = ap.resyncEntityTypeMappings(ctx, resource.API, apiID, api.EntityTypeMappings)
		if err != nil {
			return nil, err
		}

		fr, err := ap.createSpecs(ctx, model.APISpecReference, apiID, specs, resourceType)
		if err != nil {
			return nil, err
		}

		return fr, nil
	}

	log.C(ctx).Infof("Calculate the newest lastUpdate time for API")
	newestLastUpdateTime, err := NewestLastUpdateTimestamp(api.LastUpdate, apisFromDB[i].LastUpdate, apisFromDB[i].ResourceHash, apiHash)
	if err != nil {
		return nil, errors.Wrap(err, "error while calculating the newest lastUpdate time for API")
	}

	api.LastUpdate = newestLastUpdateTime

	err = ap.resyncEntityTypeMappings(ctx, resource.API, apisFromDB[i].ID, api.EntityTypeMappings)
	if err != nil {
		return nil, err
	}

	allBundleIDsForAPI, err := ap.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleAPIReference, &apisFromDB[i].ID)
	if err != nil {
		return nil, err
	}

	// in case of API update, we need to filter which ConsumptionBundleReferences should be deleted - those that are stored in db but not present in the input anymore
	bundleIDsForDeletion := extractBundleReferencesForDeletion(allBundleIDsForAPI, defaultTargetURLPerBundle)

	// in case of API update, we need to filter which ConsumptionBundleReferences should be created - those that are not present in db but are present in the input
	defaultTargetURLPerBundleForCreation := extractAllBundleReferencesForCreation(defaultTargetURLPerBundle, allBundleIDsForAPI)

	if err = ap.apiSvc.UpdateInManyBundles(ctx, resourceType, apisFromDB[i].ID, packageID, api, nil, defaultTargetURLPerBundle, defaultTargetURLPerBundleForCreation, bundleIDsForDeletion, apiHash, defaultConsumptionBundleID); err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest

	shouldFetchSpecs, err := checkIfShouldFetchSpecs(api.LastUpdate, apisFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = ap.resyncSpecs(ctx, model.APISpecReference, apisFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = ap.refetchFailedSpecs(ctx, resourceType, model.APISpecReference, apisFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return fetchRequests, nil
}

func (ap *APIProcessor) resyncEntityTypeMappings(ctx context.Context, resourceType resource.Type, resourceID string, entityTypeMappings []*model.EntityTypeMappingInput) error {
	entityTypeMappingsFromDB, err := ap.entityTypeMappingSvc.ListByOwnerResourceID(ctx, resourceID, resourceType)
	if err != nil {
		return errors.Wrapf(err, "error while listing entity type mappings for %s with id %q", resourceType, resourceID)
	}

	for _, entityTypeMappingFromDB := range entityTypeMappingsFromDB {
		err := ap.entityTypeMappingSvc.Delete(ctx, resourceType, entityTypeMappingFromDB.ID)
		if err != nil {
			return err
		}
	}
	for _, entityTypeMapping := range entityTypeMappings {
		_, err := ap.entityTypeMappingSvc.Create(ctx, resourceType, resourceID, entityTypeMapping)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ap *APIProcessor) createSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	fetchRequests := make([]*model.FetchRequest, 0)
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		_, fr, err := ap.specSvc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, *spec, resourceType, objectType, objectID)
		if err != nil {
			return nil, err
		}
		fetchRequests = append(fetchRequests, fr)
	}
	return fetchRequests, nil
}

func (ap *APIProcessor) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	if err := ap.specSvc.DeleteByReferenceObjectID(ctx, resourceType, objectType, objectID); err != nil {
		return nil, err
	}
	return ap.createSpecs(ctx, objectType, objectID, specs, resourceType)
}

func (ap *APIProcessor) refetchFailedSpecs(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]*model.FetchRequest, error) {
	specIDsFromDB, err := ap.specSvc.ListIDByReferenceObjectID(ctx, resourceType, objectType, objectID)
	if err != nil {
		return nil, err
	}

	var (
		fetchRequestsFromDB []*model.FetchRequest
		tnt                 string
	)
	if resourceType.IsTenantIgnorable() {
		fetchRequestsFromDB, err = ap.specSvc.ListFetchRequestsByReferenceObjectIDsGlobal(ctx, specIDsFromDB, objectType)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}

		fetchRequestsFromDB, err = ap.specSvc.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDsFromDB, objectType)
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
