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
	tenantColumn     = "tenant_id" // TODO: <storage-redesign> delete this column
	vendorColumns    = []string{"ord_id", "app_id", "title", "labels", "partners", "id"}
	updatableColumns = []string{"title", "labels", "partners"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in *model.Vendor) *Entity
	FromEntity(entity *Entity) (*model.Vendor, error)
}

type pgRepository struct {
	conv         EntityConverter
	existQuerier repo.ExistQuerier
	singleGetter repo.SingleGetter
	lister       repo.Lister
	deleter      repo.Deleter
	creator      repo.Creator
	updater      repo.Updater
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:         conv,
		existQuerier: repo.NewExistQuerier(resource.Vendor, vendorTable, tenantColumn),
		singleGetter: repo.NewSingleGetter(resource.Vendor, vendorTable, tenantColumn, vendorColumns),
		lister:       repo.NewLister(resource.Vendor, vendorTable, tenantColumn, vendorColumns),
		deleter:      repo.NewDeleter(resource.Vendor, vendorTable, tenantColumn),
		creator:      repo.NewCreator(resource.Vendor, vendorTable, vendorColumns),
		updater:      repo.NewUpdater(resource.Vendor, vendorTable, updatableColumns, []string{"id"}),
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Vendor entity with id %q", model.ID)
	return r.creator.Create(ctx, tenant, r.conv.ToEntity(model))
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Vendor entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, tenant, r.conv.ToEntity(model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Vendor entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Vendor, error) {
	log.C(ctx).Debugf("Getting Vendor entity with id %q", id)
	var productEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &productEnt); err != nil {
		return nil, err
	}

	productModel, err := r.conv.FromEntity(&productEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Vendor from Entity")
	}

	return productModel, nil
}

// ListByApplicationID missing godoc
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Vendor, error) {
	vendorCollection := vendorCollection{}
	if err := r.lister.List(ctx, tenantID, &vendorCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	vendors := make([]*model.Vendor, 0, vendorCollection.Len())
	for _, vendor := range vendorCollection {
		vendorModel, err := r.conv.FromEntity(&vendor)
		if err != nil {
			return nil, err
		}
		vendors = append(vendors, vendorModel)
	}
	return vendors, nil
}

type vendorCollection []Entity

// Len missing godoc
func (pc vendorCollection) Len() int {
	return len(pc)
}
