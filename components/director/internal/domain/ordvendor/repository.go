package ordvendor

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const vendorTable string = `public.vendors`

var (
	tenantColumn     = "tenant_id"
	vendorColumns    = []string{"ord_id", tenantColumn, "app_id", "title", "type", "labels"}
	updatableColumns = []string{"title", "type", "labels"}
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Vendor) *Entity
	FromEntity(entity *Entity) (*model.Vendor, error)
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
		existQuerier: repo.NewExistQuerier(resource.Vendor, vendorTable, tenantColumn),
		singleGetter: repo.NewSingleGetter(resource.Vendor, vendorTable, tenantColumn, vendorColumns),
		deleter:      repo.NewDeleter(resource.Vendor, vendorTable, tenantColumn),
		creator:      repo.NewCreator(resource.Vendor, vendorTable, vendorColumns),
		updater:      repo.NewUpdater(resource.Vendor, vendorTable, updatableColumns, tenantColumn, []string{"ord_id"}),
	}
}

func (r *pgRepository) Create(ctx context.Context, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Vendor entity with id %q", model.OrdID)
	return r.creator.Create(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Update(ctx context.Context, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Vendor entity with id %q", model.OrdID)
	return r.updater.UpdateSingle(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Vendor entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Vendor, error) {
	log.C(ctx).Debugf("Getting Vendor entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Vendor from Entity")
	}

	return productModel, nil
}
