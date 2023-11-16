package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// EventService is responsible for the service-layer Event operations.
//
//go:generate mockery --name=EventService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EventService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID, packageID *string, in model.EventDefinitionInput, specs []*model.SpecInput, bundleIDs []string, eventHash uint64, defaultBundleID string) (string, error)
	UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, in model.EventDefinitionInput, specIn *model.SpecInput, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion []string, eventHash uint64, defaultBundleID string) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EventDefinition, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.EventDefinition, error)
}

// EventProcessor defines event processor
type EventProcessor struct {
	transact             persistence.Transactioner
	eventSvc             EventService
	entityTypeSvc        EntityTypeService
	entityTypeMappingSvc EntityTypeMappingService
	bundleReferenceSvc   BundleReferenceService
	specSvc              SpecService
}

// NewEventProcessor creates new instance of EventProcessor
func NewEventProcessor(transact persistence.Transactioner, eventSvc EventService, entityTypeSvc EntityTypeService, entityTypeMappingSvc EntityTypeMappingService, bundleReferenceSvc BundleReferenceService, specSvc SpecService) *EventProcessor {
	return &EventProcessor{
		transact:             transact,
		eventSvc:             eventSvc,
		entityTypeSvc:        entityTypeSvc,
		entityTypeMappingSvc: entityTypeMappingSvc,
		bundleReferenceSvc:   bundleReferenceSvc,
		specSvc:              specSvc,
	}
}

// Process re-syncs the events passed as an argument.
func (ep *EventProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, events []*model.EventDefinitionInput, resourceHashes map[string]uint64) ([]*model.EventDefinition, []*OrdFetchRequest, error) {
	eventsFromDB, err := ep.listEventsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*OrdFetchRequest, 0)
	for _, event := range events {
		eventHash := resourceHashes[str.PtrStrToStr(event.OrdID)]
		eventFetchRequests, err := ep.resyncEventInTx(ctx, resourceType, resourceID, eventsFromDB, bundlesFromDB, packagesFromDB, event, eventHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range eventFetchRequests {
			fetchRequests = append(fetchRequests, &OrdFetchRequest{
				FetchRequest:   eventFetchRequests[i],
				RefObjectOrdID: *event.OrdID,
			})
		}
	}

	eventsFromDB, err = ep.listEventsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}
	return eventsFromDB, fetchRequests, nil
}

func (ep *EventProcessor) listEventsInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.EventDefinition, error) {
	tx, err := ep.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ep.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var eventsFromDB []*model.EventDefinition
	switch resourceType {
	case resource.Application:
		eventsFromDB, err = ep.eventSvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		eventsFromDB, err = ep.eventSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing events for %s with id %q", resourceType, resourceID)
	}

	return eventsFromDB, tx.Commit()
}

func (ep *EventProcessor) resyncEventInTx(ctx context.Context, resourceType resource.Type, resourceID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event *model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
	tx, err := ep.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ep.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := ep.resyncEvent(ctx, resourceType, resourceID, eventsFromDB, bundlesFromDB, packagesFromDB, *event, eventHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing event with ORD ID %q", *event.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (ep *EventProcessor) resyncEvent(ctx context.Context, resourceType resource.Type, resourceID string, eventsFromDB []*model.EventDefinition, bundlesFromDB []*model.Bundle, packagesFromDB []*model.Package, event model.EventDefinitionInput, eventHash uint64) ([]*model.FetchRequest, error) {
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
		eventID, err := ep.eventSvc.Create(ctx, resourceType, resourceID, nil, packageID, event, nil, bundleIDsFromBundleReference, eventHash, defaultConsumptionBundleID)
		if err != nil {
			return nil, err
		}
		err = ep.resyncEntityTypeMappings(ctx, resource.EventDefinition, eventID, event.EntityTypeMappings)
		if err != nil {
			return nil, err
		}

		return ep.createSpecs(ctx, model.EventSpecReference, eventID, specs, resourceType)
	}
	err := ep.resyncEntityTypeMappings(ctx, resource.EventDefinition, eventsFromDB[i].ID, event.EntityTypeMappings)
	if err != nil {
		return nil, err
	}

	allBundleIDsForEvent, err := ep.bundleReferenceSvc.GetBundleIDsForObject(ctx, model.BundleEventReference, &eventsFromDB[i].ID)
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

	if err = ep.eventSvc.UpdateInManyBundles(ctx, resourceType, eventsFromDB[i].ID, event, nil, bundleIDsFromBundleReference, bundleIDsForCreation, bundleIDsForDeletion, eventHash, defaultConsumptionBundleID); err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	shouldFetchSpecs, err := checkIfShouldFetchSpecs(event.LastUpdate, eventsFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = ep.resyncSpecs(ctx, model.EventSpecReference, eventsFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = ep.refetchFailedSpecs(ctx, resourceType, model.EventSpecReference, eventsFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}

	return fetchRequests, nil
}

func (ep *EventProcessor) createSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	fetchRequests := make([]*model.FetchRequest, 0)
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		_, fr, err := ep.specSvc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, *spec, resourceType, objectType, objectID)
		if err != nil {
			return nil, err
		}
		fetchRequests = append(fetchRequests, fr)
	}
	return fetchRequests, nil
}

func (ep *EventProcessor) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	if err := ep.specSvc.DeleteByReferenceObjectID(ctx, resourceType, objectType, objectID); err != nil {
		return nil, err
	}
	return ep.createSpecs(ctx, objectType, objectID, specs, resourceType)
}

func (ep *EventProcessor) refetchFailedSpecs(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]*model.FetchRequest, error) {
	specIDsFromDB, err := ep.specSvc.ListIDByReferenceObjectID(ctx, resourceType, objectType, objectID)
	if err != nil {
		return nil, err
	}

	var (
		fetchRequestsFromDB []*model.FetchRequest
		tnt                 string
	)
	if resourceType.IsTenantIgnorable() {
		fetchRequestsFromDB, err = ep.specSvc.ListFetchRequestsByReferenceObjectIDsGlobal(ctx, specIDsFromDB, objectType)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}

		fetchRequestsFromDB, err = ep.specSvc.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDsFromDB, objectType)
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

func (ep *EventProcessor) resyncEntityTypeMappings(ctx context.Context, resourceType resource.Type, resourceID string, entityTypeMappings []*model.EntityTypeMappingInput) error {
	entityTypeMappingsFromDB, err := ep.entityTypeMappingSvc.ListByOwnerResourceID(ctx, resourceID, resourceType)
	if err != nil {
		return errors.Wrapf(err, "error while listing entity type mappings for %s with id %q", resourceType, resourceID)
	}

	for _, entityTypeMappingFromDB := range entityTypeMappingsFromDB {
		err := ep.entityTypeMappingSvc.Delete(ctx, resourceType, entityTypeMappingFromDB.ID)
		if err != nil {
			return err
		}
	}
	for _, entityTypeMapping := range entityTypeMappings {
		_, err := ep.entityTypeMappingSvc.Create(ctx, resourceType, resourceID, entityTypeMapping)
		if err != nil {
			return err
		}
	}

	return nil
}
