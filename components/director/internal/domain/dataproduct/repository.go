package dataproduct

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const (
	dataProductTable           string = `public.data_products`
	idColumn                   string = "id"
	appIDColumn                string = "app_id"
	appTemplateVersionIDColumn string = "app_template_version_id"
)

var (
	idColumns          = []string{"id"}
	dataProductColumns = []string{"id", "app_id", "app_template_version_id", "ord_id", "local_tenant_id", "correlation_ids", "title", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "disabled", "deprecation_date", "sunset_date", "successors", "changelog_entries", "type", "category", "entity_types", "input_ports", "output_ports", "responsible", "data_product_links",
		"links", "industry", "line_of_business", "tags", "labels", "documentation_labels", "policy_level", "custom_policy_level", "system_instance_aware",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
	updatableColumns = []string{"ord_id", "local_tenant_id", "correlation_ids", "title", "short_description", "description", "package_id", "last_update", "visibility",
		"release_status", "disabled", "deprecation_date", "sunset_date", "successors", "changelog_entries", "type", "category", "entity_types", "input_ports", "output_ports", "responsible", "data_product_links",
		"links", "industry", "line_of_business", "tags", "labels", "documentation_labels", "policy_level", "custom_policy_level", "system_instance_aware",
		"version_value", "version_deprecated", "version_deprecated_since", "version_for_removal", "ready", "created_at", "updated_at", "deleted_at", "error", "resource_hash"}
)

// DataProductConverter converts Data Products between the model.DataProduct service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=DataProductConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataProductConverter interface {
	FromEntity(entity *Entity) *model.DataProduct
	ToEntity(dataProductModel *model.DataProduct) *Entity
}

type pgRepository struct {
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal

	conv DataProductConverter
}

// NewRepository returns a new entity responsible for repo-layer Data Products operations.
func NewRepository(conv DataProductConverter) *pgRepository {
	return &pgRepository{
		lister:             repo.NewLister(dataProductTable, dataProductColumns),
		listerGlobal:       repo.NewListerGlobal(resource.DataProduct, dataProductTable, dataProductColumns),
		creator:            repo.NewCreator(dataProductTable, dataProductColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.DataProduct, dataProductTable, dataProductColumns),
		singleGetter:       repo.NewSingleGetter(dataProductTable, dataProductColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.DataProduct, dataProductTable, dataProductColumns),
		updater:            repo.NewUpdater(dataProductTable, updatableColumns, idColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.DataProduct, dataProductTable, updatableColumns, idColumns),
		deleter:            repo.NewDeleter(dataProductTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.DataProduct, dataProductTable),

		conv: conv,
	}
}

// DataProductCollection is an array of Entities
type DataProductCollection []Entity

// Len returns the length of the collection
func (r DataProductCollection) Len() int {
	return len(r)
}

// ListByResourceID lists all Data Products for a given resource ID and resource type.
func (r *pgRepository) ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.DataProduct, error) {
	dataProductCollection := DataProductCollection{}

	var condition repo.Condition
	var err error
	switch resourceType {
	case resource.Application:
		condition = repo.NewEqualCondition(appIDColumn, resourceID)
		err = r.lister.ListWithSelectForUpdate(ctx, resource.DataProduct, tenantID, &dataProductCollection, condition)
	case resource.ApplicationTemplateVersion:
		condition = repo.NewEqualCondition(appTemplateVersionIDColumn, resourceID)
		err = r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &dataProductCollection, condition)
	}
	if err != nil {
		return nil, err
	}

	dataProducts := make([]*model.DataProduct, 0, dataProductCollection.Len())
	for _, dataProduct := range dataProductCollection {
		dataProductModel := r.conv.FromEntity(&dataProduct)
		dataProducts = append(dataProducts, dataProductModel)
	}

	return dataProducts, nil
}

// Create creates a Data Product.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creator.Create(ctx, resource.DataProduct, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// CreateGlobal creates a Data Product without tenant isolation.
func (r *pgRepository) CreateGlobal(ctx context.Context, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	err := r.creatorGlobal.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// GetByID gets a Data Product by ID from the Compass storage.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.DataProduct, error) {
	var dataProductEntity Entity
	err := r.singleGetter.Get(ctx, resource.DataProduct, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &dataProductEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Data Product")
	}

	dataProductModel := r.conv.FromEntity(&dataProductEntity)

	return dataProductModel, nil
}

// GetByIDGlobal gets a Data Product by ID without tenant isolation.
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.DataProduct, error) {
	var dataProductEntity Entity
	err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)}, repo.NoOrderBy, &dataProductEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Data Product")
	}

	dataProductModel := r.conv.FromEntity(&dataProductEntity)

	return dataProductModel, nil
}

// Update updates a Data Product.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.DataProduct, tenant, entity)
}

// UpdateGlobal updates an existing Data Product without tenant isolation.
func (r *pgRepository) UpdateGlobal(ctx context.Context, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete deletes a Data Product by its ID.
func (r *pgRepository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.DataProduct, tenantID, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

// DeleteGlobal deletes a Data Product by its ID without tenant isolation.
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}
