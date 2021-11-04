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
	productColumns   = []string{"ord_id", "app_id", "title", "short_description", "vendor", "parent", "labels", "correlation_ids", "id"}
	updatableColumns = []string{"title", "short_description", "vendor", "parent", "labels", "correlation_ids"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in *model.Product) *Entity
	FromEntity(entity *Entity) (*model.Product, error)
}

type pgRepository struct {
	conv         EntityConverter
	existQuerier repo.ExistQuerier
	lister       repo.Lister
	singleGetter repo.SingleGetter
	deleter      repo.Deleter
	creator      repo.Creator
	updater      repo.Updater
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:         conv,
		existQuerier: repo.NewExistQuerier(resource.Product, productTable),
		singleGetter: repo.NewSingleGetter(resource.Product, productTable, productColumns),
		lister:       repo.NewLister(resource.Product, productTable, productColumns),
		deleter:      repo.NewDeleter(resource.Product, productTable),
		creator:      repo.NewCreator(resource.Product, productTable, productColumns),
		updater:      repo.NewUpdater(resource.Product, productTable, updatableColumns, []string{"id"}),
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Product entity with id %q", model.ID)
	return r.creator.Create(ctx, tenant, r.conv.ToEntity(model))
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Product entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, tenant, r.conv.ToEntity(model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Product entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Product, error) {
	log.C(ctx).Debugf("Getting Product entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Product from Entity")
	}

	return productModel, nil
}

// ListByApplicationID missing godoc
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Product, error) {
	productCollection := productCollection{}
	if err := r.lister.List(ctx, tenantID, &productCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
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
