package data_product

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"
)

const dataProductsTable string = `"public"."data_products"`

var (
	bundleColumn       = "bundle_id"
	idColumn           = "id"
	dataProductColumns = []string{"id", "app_id", "ord_id", "local_id", "title", "short_description", "description", "version", "release_status", "visibility", "package_id",
		"tags", "industry", "line_of_business", "product_type", "data_product_owner"}
	idColumns        = []string{"id"}
	updatableColumns = []string{"ord_id", "local_id", "title", "short_description", "description", "version", "release_status", "visibility", "package_id",
		"tags", "industry", "line_of_business", "product_type", "data_product_owner"}
)

// DataProductConverter converts DataProducts between the model.DataProduct service-layer representation and the repo-layer representation Entity.
//
//go:generate mockery --name=DataProductConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type DataProductConverter interface {
	FromEntity(entity *Entity) *model.DataProduct
	ToEntity(apiModel *model.DataProduct) *Entity
}

type pgRepository struct {
	creator      repo.Creator
	singleGetter repo.SingleGetter
	lister       repo.Lister
	updater      repo.Updater
	conv         DataProductConverter
}

// NewRepository returns a new entity responsible for repo-layer DataProduct operations.
func NewRepository(conv DataProductConverter) *pgRepository {
	return &pgRepository{
		singleGetter: repo.NewSingleGetter(dataProductsTable, dataProductColumns),
		lister:       repo.NewLister(dataProductsTable, dataProductColumns),
		creator:      repo.NewCreator(dataProductsTable, dataProductColumns),
		updater:      repo.NewUpdater(dataProductsTable, updatableColumns, idColumns),
		conv:         conv,
	}
}

// DataProductCollection is an array of Entities
type DataProductCollection []Entity

// Len returns the length of the collection
func (r DataProductCollection) Len() int {
	return len(r)
}

// ListByApplicationID lists all DataProducts for a given application ID.
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.DataProduct, error) {
	dataProductCollection := DataProductCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.DataProduct, tenantID, &dataProductCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	dps := make([]*model.DataProduct, 0, dataProductCollection.Len())
	for _, dp := range dataProductCollection {
		dpModel := r.conv.FromEntity(&dp)
		dps = append(dps, dpModel)
	}
	return dps, nil
}

// GetByID retrieves the DataProduct with matching ID from the Compass storage.
func (r *pgRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.DataProduct, error) {
	var dataProductEntity Entity
	err := r.singleGetter.Get(ctx, resource.DataProduct, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &dataProductEntity)
	if err != nil {
		return nil, errors.Wrap(err, "while getting DataProduct")
	}

	dataProductModel := r.conv.FromEntity(&dataProductEntity)

	return dataProductModel, nil
}

// Create creates an DataProduct.
func (r *pgRepository) Create(ctx context.Context, tenant string, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)
	log.C(ctx).Errorf("Entity id: %s", entity.ID)
	log.C(ctx).Errorf("Entity application Id: %s", entity.ApplicationID)
	log.C(ctx).Errorf("Entity visibility: %s", entity.Visibility)
	err := r.creator.Create(ctx, resource.DataProduct, tenant, entity)
	if err != nil {
		return errors.Wrap(err, "while saving entity to db")
	}

	return nil
}

// Update updates an DataProduct.
func (r *pgRepository) Update(ctx context.Context, tenant string, item *model.DataProduct) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity := r.conv.ToEntity(item)

	return r.updater.UpdateSingle(ctx, resource.DataProduct, tenant, entity)
}
