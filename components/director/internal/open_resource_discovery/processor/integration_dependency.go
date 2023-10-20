package processor

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

// IntegrationDependencyService is responsible for the service-layer Integration Dependency operations.
//
//go:generate mockery --name=IntegrationDependencyService --ouidut=automock --ouidkg=automock --case=underscore --disable-version-string
type IntegrationDependencyService interface {
	ListByApplicationID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error)
	Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) error
	Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) error
}

// IntegrationDependencyProcessor defines Integration Dependency processor
type IntegrationDependencyProcessor struct {
	transact                 persistence.Transactioner
	integrationDependencySvc IntegrationDependencyService
}

// NewIntegrationDependencyProcessor creates new instance of IntegrationDependencyProcessor
func NewIntegrationDependencyProcessor(transact persistence.Transactioner, integrationDependencySvc IntegrationDependencyService) *IntegrationDependencyProcessor {
	return &IntegrationDependencyProcessor{
		transact:                 transact,
		integrationDependencySvc: integrationDependencySvc,
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
		return errors.Wrapf(err, "error while resyncing integration dependency for resource with ORD ID %q", integrationDependency.OrdID)
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
		err := id.integrationDependencySvc.Create(ctx, resourceType, resourceID, packageID, integrationDependency, integrationDependencyHash)
		if err != nil {
			return err
		}

		return nil
	}

	return id.integrationDependencySvc.Update(ctx, resourceType, resourceID, integrationDependenciesFromDB[i].ID, integrationDependency, integrationDependencyHash)
}
