package entitytypemapping

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	entityTypeMappingTable  = `public.entity_type_mappings`
	apiDefinitionIDColumn   = "api_definition_id"
	eventDefinitionIDColumn = "event_definition_id"
	idColumn                = "id"
)

var (
	entityTypeMappingColumns = []string{idColumn, "ready", "created_at", "updated_at", "deleted_at", "error", "api_definition_id", "event_definition_id",
		"api_model_selectors", "entity_type_targets"}
)

// EntityTypeMappingConverter missing godoc
//
//go:generate mockery --name=EntityTypeMappingConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeMappingConverter interface {
	ToEntity(in *model.EntityTypeMapping) *Entity
	FromEntity(entity *Entity) *model.EntityTypeMapping
}

type pgRepository struct {
	conv          EntityTypeMappingConverter
	lister        repo.Lister
	singleGetter  repo.SingleGetter
	deleter       repo.Deleter
	deleterGlobal repo.DeleterGlobal
	creator       repo.Creator
	creatorGlobal repo.CreatorGlobal
}

// NewRepository returns a repository instance
func NewRepository(conv EntityTypeMappingConverter) *pgRepository {
	return &pgRepository{
		conv:          conv,
		lister:        repo.NewLister(entityTypeMappingTable, entityTypeMappingColumns),
		singleGetter:  repo.NewSingleGetter(entityTypeMappingTable, entityTypeMappingColumns),
		deleter:       repo.NewDeleter(entityTypeMappingTable),
		deleterGlobal: repo.NewDeleterGlobal(resource.EntityTypeMapping, entityTypeMappingTable),
		creator:       repo.NewCreator(entityTypeMappingTable, entityTypeMappingColumns),
		creatorGlobal: repo.NewCreatorGlobal(resource.EntityTypeMapping, entityTypeMappingTable, entityTypeMappingColumns),
	}
}

// EntityTypeMappingCollection is an array of Entities
type EntityTypeMappingCollection []Entity

// Len returns the length of the collection
func (r EntityTypeMappingCollection) Len() int {
	return len(r)
}

// Create creates an Entity Type Mapping for a given resource.Type
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.EntityTypeMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting EntityTypeMapping entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.EntityTypeMapping, tenant, r.conv.ToEntity(model))
}

// CreateGlobal creates an entity type mapping globally without tenant isolation
func (r *pgRepository) CreateGlobal(ctx context.Context, model *model.EntityTypeMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting EntityTypeMapping entity with id %q", model.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(model))
}

// Delete deletes an Entity Type Mapping by ID
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting EntityTypeMapping entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.EntityTypeMapping, tenant, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// DeleteGlobal deletes an Entity Type Mapping without tenant isolation
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting EntityTypeMapping entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// GetByID returns an Entity Type Mapping by ID
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.EntityTypeMapping, error) {
	log.C(ctx).Debugf("Getting EntityTypeMapping entity with id %q", id)
	var entityTypeEnt Entity
	if err := r.singleGetter.Get(ctx, resource.EntityTypeMapping, tenant, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &entityTypeEnt); err != nil {
		return nil, err
	}

	entityTypeMappingModel := r.conv.FromEntity(&entityTypeEnt)

	return entityTypeMappingModel, nil
}

// ListByResourceID lists EntityTypeMappings by a given resource type and resource ID
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.EntityTypeMapping, error) {
	entityTypeMappingCollection := EntityTypeMappingCollection{}

	var condition repo.Condition
	switch resourceType {
	case resource.API:
		condition = repo.NewEqualCondition(apiDefinitionIDColumn, resourceID)
	case resource.EventDefinition:
		condition = repo.NewEqualCondition(eventDefinitionIDColumn, resourceID)
	default:
		return nil, errors.Errorf("unsupported resource type: %s", resourceType)
	}
	err := r.lister.ListWithSelectForUpdate(ctx, resource.EntityTypeMapping, tenantID, &entityTypeMappingCollection, condition)
	if err != nil {
		return nil, err
	}

	entityTypeMappings := make([]*model.EntityTypeMapping, 0, entityTypeMappingCollection.Len())
	for _, entityTypeMappingEnt := range entityTypeMappingCollection {
		entityTypeMappingModel := r.conv.FromEntity(&entityTypeMappingEnt)
		entityTypeMappings = append(entityTypeMappings, entityTypeMappingModel)
	}
	return entityTypeMappings, nil
}
