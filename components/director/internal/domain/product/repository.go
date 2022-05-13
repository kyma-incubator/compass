package product

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const productTable string = `public.products`

var (
	productColumns   = []string{"ord_id", "app_id", "title", "short_description", "vendor", "parent", "labels", "correlation_ids", "id", "documentation_labels"}
	updatableColumns = []string{"title", "short_description", "vendor", "parent", "labels", "correlation_ids", "documentation_labels"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Product) *Entity
	FromEntity(entity *Entity) (*model.Product, error)
}

type pgRepository struct {
	conv               EntityConverter
	existQuerier       repo.ExistQuerier
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

// NewRepository creates a new instance of product repository
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:               conv,
		existQuerier:       repo.NewExistQuerier(productTable),
		singleGetter:       repo.NewSingleGetter(productTable, productColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Product, productTable, productColumns),
		lister:             repo.NewLister(productTable, productColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Product, productTable, productColumns),
		deleter:            repo.NewDeleter(productTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Product, productTable),
		creator:            repo.NewCreator(productTable, productColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.Product, productTable, productColumns),
		updater:            repo.NewUpdater(productTable, updatableColumns, []string{"id"}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Product, productTable, updatableColumns, []string{"id"}),
	}
}

// Create creates a new product
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Product entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.Product, tenant, r.conv.ToEntity(model))
}

// CreateGlobal creates a new product without tenant isolation
func (r *pgRepository) CreateGlobal(ctx context.Context, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Product entity with id %q", model.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(model))
}

// Update updates an existing product
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Product entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.Product, tenant, r.conv.ToEntity(model))
}

// UpdateGlobal updates an existing product without tenant isolation
func (r *pgRepository) UpdateGlobal(ctx context.Context, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Product entity with id %q", model.ID)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete deletes an existing product
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Product entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.Product, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes an existing product without tenant isolation
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting Product entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists checks if a product exists
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Product, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID gets a product by its id
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Product, error) {
	log.C(ctx).Debugf("Getting Product entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Product, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Product from Entity")
	}

	return productModel, nil
}

// GetByIDGlobal gets a product by its id without tenant isolation
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Product, error) {
	log.C(ctx).Debugf("Getting Product entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Product from Entity")
	}

	return productModel, nil
}

// ListByApplicationID gets all products for a given application id
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Product, error) {
	productCollection := productCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.Product, tenantID, &productCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	products := make([]*model.Product, 0, productCollection.Len())
	for _, product := range productCollection {
		productModel, err := r.conv.FromEntity(&product)
		if err != nil {
			return nil, err
		}
		products = append(products, productModel)
	}
	return products, nil
}

// ListGlobal gets all global products (with NULL app_id) without tenant isolation
func (r *pgRepository) ListGlobal(ctx context.Context) ([]*model.Product, error) {
	productCollection := productCollection{}
	if err := r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &productCollection, repo.NewNullCondition("app_id")); err != nil {
		return nil, err
	}
	products := make([]*model.Product, 0, productCollection.Len())
	for _, product := range productCollection {
		productModel, err := r.conv.FromEntity(&product)
		if err != nil {
			return nil, err
		}
		products = append(products, productModel)
	}
	return products, nil
}

type productCollection []Entity

// Len missing godoc
func (pc productCollection) Len() int {
	return len(pc)
}
