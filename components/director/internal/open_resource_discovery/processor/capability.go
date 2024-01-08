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

// CapabilityService is responsible for the service-layer Capability operations.
//
//go:generate mockery --name=CapabilityService --output=automock --outpkg=automock --case=underscore --disable-version-string
type CapabilityService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.Capability, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.Capability, error)
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.CapabilityInput, spec []*model.SpecInput, capabilityHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, packageID *string, in model.CapabilityInput, capabilityHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// CapabilityProcessor defines capability processor
type CapabilityProcessor struct {
	transact      persistence.Transactioner
	capabilitySvc CapabilityService
	specSvc       SpecService
}

// NewCapabilityProcessor creates new instance of CapabilityProcessor
func NewCapabilityProcessor(transact persistence.Transactioner, capabilitySvc CapabilityService, specSvc SpecService) *CapabilityProcessor {
	return &CapabilityProcessor{
		transact:      transact,
		capabilitySvc: capabilitySvc,
		specSvc:       specSvc,
	}
}

// Process re-syncs the capabilities passed as an argument.
func (cp *CapabilityProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, packagesFromDB []*model.Package, capabilities []*model.CapabilityInput, resourceHashes map[string]uint64) ([]*model.Capability, []*OrdFetchRequest, error) {
	capabilitiesFromDB, err := cp.listCapabilitiesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	fetchRequests := make([]*OrdFetchRequest, 0)
	for _, capability := range capabilities {
		capabilityHash := resourceHashes[str.PtrStrToStr(capability.OrdID)]
		capabilityFetchRequests, err := cp.resyncCapabilitiesInTx(ctx, resourceType, resourceID, capabilitiesFromDB, packagesFromDB, capability, capabilityHash)
		if err != nil {
			return nil, nil, err
		}

		for i := range capabilityFetchRequests {
			fetchRequests = append(fetchRequests, &OrdFetchRequest{
				FetchRequest:   capabilityFetchRequests[i],
				RefObjectOrdID: *capability.OrdID,
			})
		}
	}

	capabilitiesFromDB, err = cp.listCapabilitiesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, nil, err
	}

	return capabilitiesFromDB, fetchRequests, nil
}

func (cp *CapabilityProcessor) listCapabilitiesInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.Capability, error) {
	tx, err := cp.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer cp.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	var capabilitiesFromDB []*model.Capability

	switch resourceType {
	case resource.Application:
		capabilitiesFromDB, err = cp.capabilitySvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		capabilitiesFromDB, err = cp.capabilitySvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing capabilities for %s with id %q", resourceType, resourceID)
	}

	return capabilitiesFromDB, nil
}

func (cp *CapabilityProcessor) resyncCapabilitiesInTx(ctx context.Context, resourceType resource.Type, resourceID string, capabilitiesFromDB []*model.Capability, packagesFromDB []*model.Package, capability *model.CapabilityInput, capabilityHash uint64) ([]*model.FetchRequest, error) {
	tx, err := cp.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer cp.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	fetchRequests, err := cp.resyncCapability(ctx, resourceType, resourceID, capabilitiesFromDB, packagesFromDB, *capability, capabilityHash)
	if err != nil {
		return nil, errors.Wrapf(err, "error while resyncing capability with ORD ID %q", *capability.OrdID)
	}
	return fetchRequests, tx.Commit()
}

func (cp *CapabilityProcessor) resyncCapability(ctx context.Context, resourceType resource.Type, resourceID string, capabilitiesFromDB []*model.Capability, packagesFromDB []*model.Package, capability model.CapabilityInput, capabilityHash uint64) ([]*model.FetchRequest, error) {
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
		capabilityID, err := cp.capabilitySvc.Create(ctx, resourceType, resourceID, packageID, capability, nil, capabilityHash)
		if err != nil {
			return nil, err
		}

		fetchRequests, err := cp.createSpecs(ctx, model.CapabilitySpecReference, capabilityID, specs, resourceType)
		if err != nil {
			return nil, err
		}

		return fetchRequests, nil
	}

	err := cp.capabilitySvc.Update(ctx, resourceType, capabilitiesFromDB[i].ID, packageID, capability, capabilityHash)
	if err != nil {
		return nil, err
	}

	var fetchRequests []*model.FetchRequest
	shouldFetchSpecs, err := checkIfShouldFetchSpecs(capability.LastUpdate, capabilitiesFromDB[i].LastUpdate)
	if err != nil {
		return nil, err
	}

	if shouldFetchSpecs {
		fetchRequests, err = cp.resyncSpecs(ctx, model.CapabilitySpecReference, capabilitiesFromDB[i].ID, specs, resourceType)
		if err != nil {
			return nil, err
		}
	} else {
		fetchRequests, err = cp.refetchFailedSpecs(ctx, resourceType, model.CapabilitySpecReference, capabilitiesFromDB[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return fetchRequests, nil
}

func (cp *CapabilityProcessor) createSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	fetchRequests := make([]*model.FetchRequest, 0)
	for _, spec := range specs {
		if spec == nil {
			continue
		}

		_, fr, err := cp.specSvc.CreateByReferenceObjectIDWithDelayedFetchRequest(ctx, *spec, resourceType, objectType, objectID)
		if err != nil {
			return nil, err
		}
		fetchRequests = append(fetchRequests, fr)
	}
	return fetchRequests, nil
}

func (cp *CapabilityProcessor) resyncSpecs(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string, specs []*model.SpecInput, resourceType resource.Type) ([]*model.FetchRequest, error) {
	if err := cp.specSvc.DeleteByReferenceObjectID(ctx, resourceType, objectType, objectID); err != nil {
		return nil, err
	}
	return cp.createSpecs(ctx, objectType, objectID, specs, resourceType)
}

func (cp *CapabilityProcessor) refetchFailedSpecs(ctx context.Context, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) ([]*model.FetchRequest, error) {
	specIDsFromDB, err := cp.specSvc.ListIDByReferenceObjectID(ctx, resourceType, objectType, objectID)
	if err != nil {
		return nil, err
	}

	var (
		fetchRequestsFromDB []*model.FetchRequest
		tnt                 string
	)
	if resourceType.IsTenantIgnorable() {
		fetchRequestsFromDB, err = cp.specSvc.ListFetchRequestsByReferenceObjectIDsGlobal(ctx, specIDsFromDB, objectType)
	} else {
		tnt, err = tenant.LoadFromContext(ctx)
		if err != nil {
			return nil, err
		}

		fetchRequestsFromDB, err = cp.specSvc.ListFetchRequestsByReferenceObjectIDs(ctx, tnt, specIDsFromDB, objectType)
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
