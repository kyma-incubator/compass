package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// IntegrationDependencyService is responsible for the service-layer Integration Dependency operations.
//
//go:generate mockery --name=IntegrationDependencyService --output=automock --outpkg=automock --case=underscore --disable-version-string
type IntegrationDependencyService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.IntegrationDependency, error)
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
}

// AspectService is responsible for the service-layer Aspect operations.
//
//go:generate mockery --name=AspectService --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, in model.AspectInput) (string, error)
	DeleteByIntegrationDependencyID(ctx context.Context, integrationDependencyID string) error
}

// AspectEventResourceService is responsible for the service-layer Aspect Event Resource operations.
//
//go:generate mockery --name=AspectEventResourceService --output=automock --outpkg=automock --case=underscore --disable-version-string
type AspectEventResourceService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, aspectID string, in model.AspectEventResourceInput) (string, error)
}

// IntegrationDependencyProcessor defines Integration Dependency processor
type IntegrationDependencyProcessor struct {
	transact                 persistence.Transactioner
	integrationDependencySvc IntegrationDependencyService
	aspectSvc                AspectService
	aspectEventResourceSvc   AspectEventResourceService
}

// NewIntegrationDependencyProcessor creates new instance of IntegrationDependencyProcessor
func NewIntegrationDependencyProcessor(transact persistence.Transactioner, integrationDependencySvc IntegrationDependencyService, aspectSvc AspectService, aspectEventResourceSvc AspectEventResourceService) *IntegrationDependencyProcessor {
	return &IntegrationDependencyProcessor{
		transact:                 transact,
		integrationDependencySvc: integrationDependencySvc,
		aspectSvc:                aspectSvc,
		aspectEventResourceSvc:   aspectEventResourceSvc,
	}
}

// Process re-syncs the integration dependencies passed as an argument.
func (id *IntegrationDependencyProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, packagesFromDB []*model.Package, integrationDependencies []*model.IntegrationDependencyInput, resourceHashes map[string]uint64) ([]*model.IntegrationDependency, error) {
	integrationDependenciesFromDB, err := id.listIntegrationDependenciesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, integrationDependency := range integrationDependencies {
		integrationDependencyHash := resourceHashes[str.PtrStrToStr(integrationDependency.OrdID)]
		if err := id.resyncIntegrationDependencyInTx(ctx, resourceType, resourceID, integrationDependenciesFromDB, packagesFromDB, integrationDependency, integrationDependencyHash); err != nil {
			return nil, err
		}
	}

	integrationDependenciesFromDB, err = id.listIntegrationDependenciesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return integrationDependenciesFromDB, nil
}

func (id *IntegrationDependencyProcessor) listIntegrationDependenciesInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.IntegrationDependency, error) {
	tx, err := id.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer id.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var integrationDependenciesFromDB []*model.IntegrationDependency
	switch resourceType {
	case resource.Application:
		integrationDependenciesFromDB, err = id.integrationDependencySvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		integrationDependenciesFromDB, err = id.integrationDependencySvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing integration dependencies for %s with id %q", resourceType, resourceID)
	}

	return integrationDependenciesFromDB, tx.Commit()
}

func (id *IntegrationDependencyProcessor) resyncIntegrationDependencyInTx(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependenciesFromDB []*model.IntegrationDependency, packagesFromDB []*model.Package, integrationDependency *model.IntegrationDependencyInput, integrationDependencyHash uint64) error {
	tx, err := id.transact.Begin()
	if err != nil {
		return err
	}
	defer id.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	if err := id.resyncIntegrationDependency(ctx, resourceType, resourceID, integrationDependenciesFromDB, packagesFromDB, *integrationDependency, integrationDependencyHash); err != nil {
		return errors.Wrapf(err, "error while resyncing integration dependency for resource with ORD ID %q", *integrationDependency.OrdID)
	}
	return tx.Commit()
}

func (id *IntegrationDependencyProcessor) resyncIntegrationDependency(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependenciesFromDB []*model.IntegrationDependency, packagesFromDB []*model.Package, integrationDependency model.IntegrationDependencyInput, integrationDependencyHash uint64) error {
	ctx = addFieldToLogger(ctx, "integration_dependency_ord_id", *integrationDependency.OrdID)
	i, isIntegrationDependencyFound := searchInSlice(len(integrationDependenciesFromDB), func(i int) bool {
		return equalStrings(integrationDependenciesFromDB[i].OrdID, integrationDependency.OrdID)
	})

	var packageID *string
	if i, found := searchInSlice(len(packagesFromDB), func(i int) bool {
		return equalStrings(&packagesFromDB[i].OrdID, integrationDependency.OrdPackageID)
	}); found {
		packageID = &packagesFromDB[i].ID
	}

	if !isIntegrationDependencyFound {
		integrationDependencyID, err := id.integrationDependencySvc.Create(ctx, resourceType, resourceID, packageID, integrationDependency, integrationDependencyHash)
		if err != nil {
			return err
		}

		aspectEventResourcesByAspectIDInput, err := id.createAspects(ctx, resourceType, resourceID, integrationDependencyID, integrationDependency.Aspects)
		if err != nil {
			return err
		}

		for aspectID, aspectEventResourcesInput := range aspectEventResourcesByAspectIDInput {
			err = id.createAspectEventResources(ctx, resourceType, resourceID, aspectID, aspectEventResourcesInput)
			if err != nil {
				return err
			}
		}

		return nil
	}

	log.C(ctx).Infof("Calculate the newest lastUpdate time for Integration Dependency")
	newestLastUpdateTime, err := NewestLastUpdateTimestamp(integrationDependency.LastUpdate, integrationDependenciesFromDB[i].LastUpdate, integrationDependenciesFromDB[i].ResourceHash, integrationDependencyHash)
	if err != nil {
		return errors.Wrap(err, "error while calculating the newest lastUpdate time for Integration Dependency")
	}

	integrationDependency.LastUpdate = newestLastUpdateTime

	err = id.integrationDependencySvc.Update(ctx, resourceType, resourceID, integrationDependenciesFromDB[i].ID, packageID, integrationDependency, integrationDependencyHash)
	if err != nil {
		return err
	}

	return id.resyncAspects(ctx, resourceType, resourceID, integrationDependenciesFromDB[i].ID, integrationDependency.Aspects)
}

func (id *IntegrationDependencyProcessor) createAspects(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, aspects []*model.AspectInput) (map[string][]*model.AspectEventResourceInput, error) {
	aspectEventResourcesByAspectIDInput := make(map[string][]*model.AspectEventResourceInput, 0)
	for _, aspect := range aspects {
		id, err := id.aspectSvc.Create(ctx, resourceType, resourceID, integrationDependencyID, *aspect)
		if err != nil {
			return nil, err
		}
		aspectEventResourcesByAspectIDInput[id] = aspect.EventResources
	}

	return aspectEventResourcesByAspectIDInput, nil
}

func (id *IntegrationDependencyProcessor) resyncAspects(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, aspects []*model.AspectInput) error {
	// this has to delete aspects and its event resources as well, otherwise use deleteByAspectID method
	if err := id.aspectSvc.DeleteByIntegrationDependencyID(ctx, integrationDependencyID); err != nil {
		return err
	}

	aspectEventResourcesByAspectIDInput, err := id.createAspects(ctx, resourceType, resourceID, integrationDependencyID, aspects)
	if err != nil {
		return err
	}

	for aspectID, aspectEventResourcesInput := range aspectEventResourcesByAspectIDInput {
		err := id.createAspectEventResources(ctx, resourceType, resourceID, aspectID, aspectEventResourcesInput)
		if err != nil {
			return err
		}
	}

	return nil
}

func (id *IntegrationDependencyProcessor) createAspectEventResources(ctx context.Context, resourceType resource.Type, resourceID string, aspectID string, aspectEventResourcesInput []*model.AspectEventResourceInput) error {
	for _, aspectEventResourceInput := range aspectEventResourcesInput {
		_, err := id.aspectEventResourceSvc.Create(ctx, resourceType, resourceID, aspectID, *aspectEventResourceInput)
		if err != nil {
			return err
		}
	}

	return nil
}
