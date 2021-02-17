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
	tenantColumn     = "tenant_id"
	productColumns   = []string{"ord_id", tenantColumn, "app_id", "title", "short_description", "vendor", "parent", "sap_ppms_object_id", "labels"}
	updatableColumns = []string{"title", "short_description", "vendor", "parent", "sap_ppms_object_id", "labels"}
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Product) *Entity
	FromEntity(entity *Entity) (*model.Product, error)
}

type pgRepository struct {
	conv         EntityConverter
	existQuerier repo.ExistQuerier
	singleGetter repo.SingleGetter
	deleter      repo.Deleter
	creator      repo.Creator
	updater      repo.Updater
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:         conv,
		existQuerier: repo.NewExistQuerier(resource.Product, productTable, tenantColumn),
		singleGetter: repo.NewSingleGetter(resource.Product, productTable, tenantColumn, productColumns),
		deleter:      repo.NewDeleter(resource.Product, productTable, tenantColumn),
		creator:      repo.NewCreator(resource.Product, productTable, productColumns),
		updater:      repo.NewUpdater(resource.Product, productTable, updatableColumns, tenantColumn, []string{"ord_id"}),
	}
}

func (r *pgRepository) Create(ctx context.Context, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Product entity with id %q", model.OrdID)
	return r.creator.Create(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Update(ctx context.Context, model *model.Product) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Product entity with id %q", model.OrdID)
	return r.updater.UpdateSingle(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Product entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Product, error) {
	log.C(ctx).Debugf("Getting Product entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Product from Entity")
	}

	return productModel, nil
}
