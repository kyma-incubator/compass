package processor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// EntityTypeService is responsible for the service-layer Entity Type operations.
//
//go:generate mockery --name=EntityTypeService --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeService interface {
	Create(ctx context.Context, resourceType resource.Type, resourceID string, in model.EntityTypeInput, entityTypeHash uint64) (string, error)
	Update(ctx context.Context, resourceType resource.Type, id string, in model.EntityTypeInput, entityTypeHash uint64) error
	Delete(ctx context.Context, resourceType resource.Type, id string) error
	ListByApplicationID(ctx context.Context, appID string) ([]*model.EntityType, error)
	ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.EntityType, error)
}

// EntityTypeProcessor defines entity type processor
type EntityTypeProcessor struct {
	transact      persistence.Transactioner
	entityTypeSvc EntityTypeService
}

// NewEntityTypeProcessor creates new instance of EntityTypeProcessor
func NewEntityTypeProcessor(transact persistence.Transactioner, entityTypeSvc EntityTypeService) *EntityTypeProcessor {
	return &EntityTypeProcessor{
		transact:      transact,
		entityTypeSvc: entityTypeSvc,
	}
}

func (ep *EntityTypeProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, packagesFromDB []*model.Package, entityTypes []*model.EntityTypeInput, resourceHashes map[string]uint64) ([]*model.EntityType, error) {
	entityTypesFromDB, err := ep.listEntityTypesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}

	for _, entityType := range entityTypes {
		entityTypeHash := resourceHashes[entityType.OrdID]
		err := ep.resyncEntityTypeInTx(ctx, resourceType, resourceID, entityTypesFromDB, packagesFromDB, entityType, entityTypeHash)
		if err != nil {
			return nil, err
		}
	}

	entityTypesFromDB, err = ep.listEntityTypesInTx(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	return entityTypesFromDB, nil
}

func (ep *EntityTypeProcessor) listEntityTypesInTx(ctx context.Context, resourceType resource.Type, resourceID string) ([]*model.EntityType, error) {
	tx, err := ep.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer ep.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	var entityTypesFromDB []*model.EntityType
	switch resourceType {
	case resource.Application:
		entityTypesFromDB, err = ep.entityTypeSvc.ListByApplicationID(ctx, resourceID)
	case resource.ApplicationTemplateVersion:
		entityTypesFromDB, err = ep.entityTypeSvc.ListByApplicationTemplateVersionID(ctx, resourceID)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "error while listing entity types for %s with id %q", resourceType, resourceID)
	}

	return entityTypesFromDB, tx.Commit()
}

func (ep *EntityTypeProcessor) resyncEntityTypeInTx(ctx context.Context, resourceType resource.Type, resourceID string, entityTypesFromDB []*model.EntityType, packagesFromDB []*model.Package, entityType *model.EntityTypeInput, entityTypeHash uint64) error {
	tx, err := ep.transact.Begin()
	if err != nil {
		return err
	}
	defer ep.transact.RollbackUnlessCommitted(ctx, tx)
	ctx = persistence.SaveToContext(ctx, tx)

	err = ep.resyncEntityType(ctx, resourceType, resourceID, entityTypesFromDB, packagesFromDB, *entityType, entityTypeHash)
	if err != nil {
		return errors.Wrapf(err, "error while resyncing entity type with ORD ID %q", entityType.OrdID)
	}
	return tx.Commit()
}

func (ep *EntityTypeProcessor) resyncEntityType(ctx context.Context, resourceType resource.Type, resourceID string, entityTypesFromDB []*model.EntityType, packagesFromDB []*model.Package, entityType model.EntityTypeInput, entityTypeHash uint64) error {
	ctx = addFieldToLogger(ctx, "entity_type_ord_id", entityType.OrdID)
	_, isEntityTypeFound := searchInSlice(len(entityTypesFromDB), func(i int) bool {
		return equalStrings(&entityTypesFromDB[i].OrdID, &entityType.OrdID)
	})

	if !isEntityTypeFound {
		_, err := ep.entityTypeSvc.Create(ctx, resourceType, resourceID, entityType, entityTypeHash)
		if err != nil {
			return err
		}
	} else {
		err := ep.entityTypeSvc.Update(ctx, resourceType, resourceID, entityType, entityTypeHash)
		if err != nil {
			return err
		}
	}
	return nil
}
