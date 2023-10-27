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
	entityTypeMappingColumns = []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "api_definition_id", "event_definition_id",
		"api_model_selectors", "entity_type_targets"}
	updatableColumns = []string{"ready", "created_at", "updated_at", "deleted_at", "error", "api_definition_id", "event_definition_id",
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
	conv               EntityTypeMappingConverter
	existQuerier       repo.ExistQuerier
	pageableQuerier    repo.PageableQuerier
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
}

// NewRepository returns a repository instance
func NewRepository(conv EntityTypeMappingConverter) *pgRepository {
	return &pgRepository{
		conv:               conv,
		existQuerier:       repo.NewExistQuerier(entityTypeMappingTable),
		pageableQuerier:    repo.NewPageableQuerier(entityTypeMappingTable, entityTypeMappingColumns),
		lister:             repo.NewLister(entityTypeMappingTable, entityTypeMappingColumns),
		listerGlobal:       repo.NewListerGlobal(resource.EntityTypeMapping, entityTypeMappingTable, entityTypeMappingColumns),
		singleGetter:       repo.NewSingleGetter(entityTypeMappingTable, entityTypeMappingColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.EntityTypeMapping, entityTypeMappingTable, entityTypeMappingColumns),
		deleter:            repo.NewDeleter(entityTypeMappingTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.EntityTypeMapping, entityTypeMappingTable),
		creator:            repo.NewCreator(entityTypeMappingTable, entityTypeMappingColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.EntityTypeMapping, entityTypeMappingTable, entityTypeMappingColumns),
		updater:            repo.NewUpdater(entityTypeMappingTable, updatableColumns, []string{"id"}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.EntityTypeMapping, entityTypeMappingTable, updatableColumns, []string{"id"}),
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

// Update updates an Entity Type Mapping by ID for a given resource.Type
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.EntityTypeMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating EntityTypeMapping entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.EntityTypeMapping, tenant, r.conv.ToEntity(model))
}

// UpdateGlobal updates entity type mapping globally without tenant isolation
func (r *pgRepository) UpdateGlobal(ctx context.Context, model *model.EntityTypeMapping) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating EntityTypeMapping entity with id %q", model.ID)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete deletes an Entity Type Mapping by ID
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting EntityTypeMapping entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.EntityTypeMapping, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes an Entity Type Mapping without tenant isolation
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting EntityTypeMapping entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists checks if an Entity Type Mapping with ID exists
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.EntityTypeMapping, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID returns an Entity Type Mapping by ID
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.EntityTypeMapping, error) {
	log.C(ctx).Debugf("Getting EntityTypeMapping entity with id %q", id)
	var entityTypeEnt Entity
	if err := r.singleGetter.Get(ctx, resource.EntityTypeMapping, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entityTypeEnt); err != nil {
		return nil, err
	}

	entityTypeMappingModel := r.conv.FromEntity(&entityTypeEnt)

	return entityTypeMappingModel, nil
}

// GetByIDGlobal gets an entity type by ID without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.EntityTypeMapping, error) {
	log.C(ctx).Debugf("Getting EntityTypeMapping entity with id %q", id)
	var entityTypeMappingEnt Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entityTypeMappingEnt); err != nil {
		return nil, err
	}

	entityTypeMappingModel := r.conv.FromEntity(&entityTypeMappingEnt)

	return entityTypeMappingModel, nil
}

// GetByAPIDefinitionID retrieves the EntityTypeMapping with matching ID and API Definition ID from the Compass storage.
func (r *pgRepository) GetByAPIDefinitionID(ctx context.Context, tenantID string, id, apiDefinitionID string) (*model.EntityTypeMapping, error) {
	var entityTypeMappingEntity Entity
	err := r.singleGetter.Get(ctx, resource.EntityTypeMapping, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id), repo.NewEqualCondition(apiDefinitionIDColumn, apiDefinitionID)}, repo.NoOrderBy, &entityTypeMappingEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EntityTypeMapping for API Definition ID %s", apiDefinitionID)
	}

	entityTypeModel := r.conv.FromEntity(&entityTypeMappingEntity)

	return entityTypeModel, nil
}

// ListByAPIDefinitionIDPage lists all EntityTypeMapping for a given API Definition ID with paging.
func (r *pgRepository) ListByAPIDefinitionIDPage(ctx context.Context, tenantID string, apiDefinitionID string, pageSize int, cursor string) (*model.EntityTypeMappingPage, error) {
	var entityTypeMappingCollection EntityTypeMappingCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.EntityTypeMapping, tenantID, pageSize, cursor, idColumn, &entityTypeMappingCollection, repo.NewEqualCondition(apiDefinitionIDColumn, apiDefinitionID))

	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	items := make([]*model.EntityTypeMapping, 0, len(entityTypeMappingCollection))
	for _, entityTypeMapping := range entityTypeMappingCollection {
		m := r.conv.FromEntity(&entityTypeMapping)
		items = append(items, m)
	}

	return &model.EntityTypeMappingPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// GetByAPIDefinitionID retrieves the EntityTypeMapping with matching ID and Event Definition ID from the Compass storage.
func (r *pgRepository) GetByEventDefinitionID(ctx context.Context, tenantID string, id, eventDefinitionID string) (*model.EntityTypeMapping, error) {
	var entityTypeMappingEntity Entity
	err := r.singleGetter.Get(ctx, resource.EntityTypeMapping, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id), repo.NewEqualCondition(eventDefinitionIDColumn, eventDefinitionID)}, repo.NoOrderBy, &entityTypeMappingEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EntityTypeMapping for Event Definition ID %s", eventDefinitionID)
	}

	entityTypeModel := r.conv.FromEntity(&entityTypeMappingEntity)

	return entityTypeModel, nil
}

// ListByEventDefinitionIDPage lists all EntityTypeMapping for a given Event Definition ID with paging.
func (r *pgRepository) ListByEventDefinitionIDPage(ctx context.Context, tenantID string, eventDefinitionID string, pageSize int, cursor string) (*model.EntityTypeMappingPage, error) {
	var entityTypeMappingCollection EntityTypeMappingCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.EntityTypeMapping, tenantID, pageSize, cursor, idColumn, &entityTypeMappingCollection, repo.NewEqualCondition(eventDefinitionIDColumn, eventDefinitionID))

	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	items := make([]*model.EntityTypeMapping, 0, len(entityTypeMappingCollection))
	for _, entityTypeMapping := range entityTypeMappingCollection {
		m := r.conv.FromEntity(&entityTypeMapping)
		items = append(items, m)
	}

	return &model.EntityTypeMappingPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
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
