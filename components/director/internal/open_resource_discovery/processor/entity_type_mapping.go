package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// EntityTypeMappingService is responsible for the service-layer Entity Type Mapping operations.
//
//go:generate mockery --name=EntityTypeMappingService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeMappingService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.EntityTypeMappingInput) (string, error)
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByAPIDefinitionID(ctx context.Context, apiDefinitionID string) ([]*model.EntityTypeMapping, error)
	ListByEventDefinitionID(ctx context.Context, eventDefinitionID string) ([]*model.EntityTypeMapping, error)
}

// EntityTypeMappingProcessor defines entity type mapping processor
type EntityTypeMappingProcessor struct {
	transact             persistence.Transactioner
	entityTypeMappingSvc EntityTypeMappingService
}

// NewEntityTypeMappingProcessor creates new instance of EntityTypeMappingProcessor
func NewEntityTypeMappingProcessor(transact persistence.Transactioner, entityTypeMappingSvc EntityTypeMappingService) *EntityTypeMappingProcessor {
	return &EntityTypeMappingProcessor{
		transact:             transact,
		entityTypeMappingSvc: entityTypeMappingSvc,
	}
}

// Process re-syncs the entity types mappings passed as an argument.
func (etmp *EntityTypeMappingProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, entityTypeMappings []*model.EntityTypeMappingInput) ([]*model.EntityTypeMapping, error) {
	entityTypeMappingsFromDB, err := etmp.listEntityTypeMappingsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	err = etmp.resyncEntityTypeMappingsInTx(ctx, resourceType, resourceID, entityTypeMappingsFromDB, entityTypeMappings)
	if err != nil {
		return nil, err
	}

	entityTypeMappingsFromDB, err = etmp.listEntityTypeMappingsInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return entityTypeMappingsFromDB, nil
}

func (etmp *EntityTypeMappingProcessor) listEntityTypeMappingsInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.EntityTypeMapping, error) {
	tx, err := etmp.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer etmp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var entityTypeMappingsFromDB []*model.EntityTypeMapping
	switch resourceType {
	case resource.API:
		entityTypeMappingsFromDB, err = etmp.entityTypeMappingSvc.ListByAPIDefinitionID(ctx, resourceID)
	case resource.EventDefinition:
		entityTypeMappingsFromDB, err = etmp.entityTypeMappingSvc.ListByEventDefinitionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing entity type mappings for %s with id %q", resourceType, resourceID)
	}

	return entityTypeMappingsFromDB, tx.Commit()
}

func (etmp *EntityTypeMappingProcessor) resyncEntityTypeMappingsInTx(ctx context.Context, resourceType resource.Type, resourceID string, entityTypeMappingsFromDB []*model.EntityTypeMapping, entityTypeMappings []*model.EntityTypeMappingInput) error {
	tx, err := etmp.transact.Begin()
	if err != nil {
		return err
	}
	defer etmp.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	err = etmp.resyncEntityTypeMapping(ctx, resourceType, resourceID, entityTypeMappingsFromDB, entityTypeMappings)
	if err != nil {
		return errors.Wrapf(err, "error while resyncing entity type mapping")
	}
	return tx.Commit()
}

func (etmp *EntityTypeMappingProcessor) resyncEntityTypeMapping(ctx context.Context, resourceType resource.Type, resourceID string, entityTypeMappingsFromDB []*model.EntityTypeMapping, entityTypeMappings []*model.EntityTypeMappingInput) error {
	for _, entityTypeMappingFromDB := range entityTypeMappingsFromDB {
		err := etmp.entityTypeMappingSvc.Delete(ctx, resourceType, entityTypeMappingFromDB.ID)
		if err != nil {
			return err
		}
	}
	for _, entityTypeMapping := range entityTypeMappings {
		_, err := etmp.entityTypeMappingSvc.Create(ctx, resourceType, resourceID, *entityTypeMapping)
		if err != nil {
			return err
		}
	}

	return nil
}
