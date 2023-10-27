package entitytype

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
	entityTypeTable            = `public.entity_types`
	appTemplateVersionIDColumn = "app_template_version_id"
	appIDColumn                = "app_id"
	idColumn                   = "id"
)

var (
	entityTypeColumns = []string{"id", "ready", "created_at", "updated_at", "deleted_at", "error", "app_id", "app_template_version_id", "ord_id", "local_tenant_id",
		"correlation_ids", "level", "title", "short_description", "description", "system_instance_aware", "changelog_entries", "package_id", "visibility",
		"links", "part_of_products", "last_update", "policy_level", "custom_policy_level", "release_status", "sunset_date", "successors", "extensible", "tags", "labels",
		"documentation_labels", "resource_hash", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
	updatableColumns = []string{"ready", "created_at", "updated_at", "deleted_at", "error", "ord_id", "local_tenant_id",
		"correlation_ids", "level", "title", "short_description", "description", "system_instance_aware", "changelog_entries", "package_id", "visibility",
		"links", "part_of_products", "last_update", "policy_level", "custom_policy_level", "release_status", "sunset_date", "successors", "extensible", "tags", "labels",
		"documentation_labels", "resource_hash", "version_value", "version_deprecated", "version_deprecated_since", "version_for_removal"}
)

// EntityTypeConverter missing godoc
//
//go:generate mockery --name=EntityTypeConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityTypeConverter interface {
	ToEntity(in *model.EntityType) *Entity
	FromEntity(entity *Entity) *model.EntityType
}

type pgRepository struct {
	conv               EntityTypeConverter
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
func NewRepository(conv EntityTypeConverter) *pgRepository {
	return &pgRepository{
		conv:               conv,
		existQuerier:       repo.NewExistQuerier(entityTypeTable),
		pageableQuerier:    repo.NewPageableQuerier(entityTypeTable, entityTypeColumns),
		lister:             repo.NewLister(entityTypeTable, entityTypeColumns),
		listerGlobal:       repo.NewListerGlobal(resource.EntityType, entityTypeTable, entityTypeColumns),
		singleGetter:       repo.NewSingleGetter(entityTypeTable, entityTypeColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.EntityType, entityTypeTable, entityTypeColumns),
		deleter:            repo.NewDeleter(entityTypeTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.EntityType, entityTypeTable),
		creator:            repo.NewCreator(entityTypeTable, entityTypeColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.EntityType, entityTypeTable, entityTypeColumns),
		updater:            repo.NewUpdater(entityTypeTable, updatableColumns, []string{idColumn}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.EntityType, entityTypeTable, updatableColumns, []string{idColumn}),
	}
}

// EntityTypeCollection is an array of Entities
type EntityTypeCollection []Entity

// Len returns the length of the collection
func (r EntityTypeCollection) Len() int {
	return len(r)
}

// Create creates an Entity Type for a given resource.Type
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.EntityType) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting EntityType entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.EntityType, tenant, r.conv.ToEntity(model))
}

// CreateGlobal creates an entity type globally without tenant isolation
func (r *pgRepository) CreateGlobal(ctx context.Context, model *model.EntityType) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting EntityType entity with id %q", model.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(model))
}

// Update updates an Entity Type by ID for a given resource.Type
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.EntityType) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating EntityType entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.EntityType, tenant, r.conv.ToEntity(model))
}

// UpdateGlobal updates entity type globally without tenant isolation
func (r *pgRepository) UpdateGlobal(ctx context.Context, model *model.EntityType) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating EntityType entity with id %q", model.ID)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete deletes an Entity Type by ID
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting EntityType entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.EntityType, tenant, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// DeleteGlobal deletes an Entity Type without tenant isolation
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting EntityType entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// Exists checks if an Entity Type with ID exists
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.EntityType, tenant, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// GetByID returns an Entity Type by ID
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.EntityType, error) {
	log.C(ctx).Debugf("Getting EntityType entity with id %q", id)
	var entityTypeEnt Entity
	if err := r.singleGetter.Get(ctx, resource.EntityType, tenant, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &entityTypeEnt); err != nil {
		return nil, err
	}

	entityTypeModel := r.conv.FromEntity(&entityTypeEnt)

	return entityTypeModel, nil
}

// GetByIDGlobal gets an entity type by ID without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.EntityType, error) {
	log.C(ctx).Debugf("Getting EntityType entity with id %q", id)
	var entityTypeEnt Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &entityTypeEnt); err != nil {
		return nil, err
	}

	entityTypeModel := r.conv.FromEntity(&entityTypeEnt)

	return entityTypeModel, nil
}

// GetByApplicationID retrieves the EntityType with matching ID and Application ID from the Compass storage.
func (r *pgRepository) GetByApplicationID(ctx context.Context, tenantID string, id, appID string) (*model.EntityType, error) {
	var entityTypeEntity Entity
	err := r.singleGetter.Get(ctx, resource.EntityType, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id), repo.NewEqualCondition(appIDColumn, appID)}, repo.NoOrderBy, &entityTypeEntity)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting EntityType for Application ID %s", appID)
	}

	entityTypeModel := r.conv.FromEntity(&entityTypeEntity)

	return entityTypeModel, nil
}

// ListByApplicationIDPage lists all EntityTypes for a given application ID with paging.
func (r *pgRepository) ListByApplicationIDPage(ctx context.Context, tenantID string, appID string, pageSize int, cursor string) (*model.EntityTypePage, error) {
	var entityTypeCollection EntityTypeCollection
	page, totalCount, err := r.pageableQuerier.List(ctx, resource.EntityType, tenantID, pageSize, cursor, idColumn, &entityTypeCollection, repo.NewEqualCondition(appIDColumn, appID))

	if err != nil {
		return nil, errors.Wrap(err, "while decoding page cursor")
	}

	items := make([]*model.EntityType, 0, len(entityTypeCollection))
	for _, entityType := range entityTypeCollection {
		m := r.conv.FromEntity(&entityType)
		items = append(items, m)
	}

	return &model.EntityTypePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// ListByResourceID lists EntityTypes by a given resource type and resource ID
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID, resourceID string, resourceType resource.Type) ([]*model.EntityType, error) {
	entityTypeCollection := EntityTypeCollection{}

	var condition repo.Condition
	var err error
	if resourceType == resource.Application {
		condition = repo.NewEqualCondition(appIDColumn, resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.EntityType, tenantID, &entityTypeCollection, condition)
	} else {
		condition = repo.NewEqualCondition(appTemplateVersionIDColumn, resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &entityTypeCollection, condition)
	}
	if err != nil {
		return nil, err
	}

	entityTypes := make([]*model.EntityType, 0, entityTypeCollection.Len())
	for _, entityTypeEnt := range entityTypeCollection {
		entityTypeModel := r.conv.FromEntity(&entityTypeEnt)
		entityTypes = append(entityTypes, entityTypeModel)
	}
	return entityTypes, nil
}
