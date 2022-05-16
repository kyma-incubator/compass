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
	vendorColumns    = []string{"ord_id", "app_id", "title", "labels", "partners", "id", "documentation_labels"}
	updatableColumns = []string{"title", "labels", "partners", "documentation_labels"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Vendor) *Entity
	FromEntity(entity *Entity) (*model.Vendor, error)
}

type pgRepository struct {
	conv               EntityConverter
	existQuerier       repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	lister             repo.Lister
	listerGlobal       repo.ListerGlobal
	deleter            repo.Deleter
	deleterGlobal      repo.DeleterGlobal
	creator            repo.Creator
	creatorGlobal      repo.CreatorGlobal
	updater            repo.Updater
	updaterGlobal      repo.UpdaterGlobal
}

// NewRepository creates a new instance of repository
func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		conv:               conv,
		existQuerier:       repo.NewExistQuerier(vendorTable),
		singleGetter:       repo.NewSingleGetter(vendorTable, vendorColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Vendor, vendorTable, vendorColumns),
		lister:             repo.NewLister(vendorTable, vendorColumns),
		listerGlobal:       repo.NewListerGlobal(resource.Vendor, vendorTable, vendorColumns),
		deleter:            repo.NewDeleter(vendorTable),
		deleterGlobal:      repo.NewDeleterGlobal(resource.Vendor, vendorTable),
		creator:            repo.NewCreator(vendorTable, vendorColumns),
		creatorGlobal:      repo.NewCreatorGlobal(resource.Vendor, vendorTable, vendorColumns),
		updater:            repo.NewUpdater(vendorTable, updatableColumns, []string{"id"}),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.Vendor, vendorTable, updatableColumns, []string{"id"}),
	}
}

// Create creates a new Vendor
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Vendor entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.Vendor, tenant, r.conv.ToEntity(model))
}

// CreateGlobal creates a new Vendor without tenant isolation.
func (r *pgRepository) CreateGlobal(ctx context.Context, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Vendor entity with id %q", model.ID)
	return r.creatorGlobal.Create(ctx, r.conv.ToEntity(model))
}

// Update updates an existing Vendor
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Vendor entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.Vendor, tenant, r.conv.ToEntity(model))
}

// UpdateGlobal updates a Vendor without tenant isolation.
func (r *pgRepository) UpdateGlobal(ctx context.Context, model *model.Vendor) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Vendor entity with id %q", model.ID)
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(model))
}

// Delete deletes an existing Vendor
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Vendor entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.Vendor, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// DeleteGlobal deletes a Vendor without tenant isolation.
func (r *pgRepository) DeleteGlobal(ctx context.Context, id string) error {
	log.C(ctx).Debugf("Deleting Vendor entity with id %q", id)
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists checks if a Vendor exists
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Vendor, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID gets a Vendor by its id
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Vendor, error) {
	log.C(ctx).Debugf("Getting Vendor entity with id %q", id)
	var vendorEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Vendor, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &vendorEnt); err != nil {
		return nil, err
	}

	vendorModel, err := r.conv.FromEntity(&vendorEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Vendor from Entity")
	}

	return vendorModel, nil
}

// GetByIDGlobal gets a Vendor without tenant isolation.
func (r *pgRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Vendor, error) {
	log.C(ctx).Debugf("Getting Vendor entity with id %q", id)
	var vendorEnt Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &vendorEnt); err != nil {
		return nil, err
	}

	vendorModel, err := r.conv.FromEntity(&vendorEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Vendor from Entity")
	}

	return vendorModel, nil
}

// ListByApplicationID gets a list of Vendors by given application id
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Vendor, error) {
	vendorCollection := vendorCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.Vendor, tenantID, &vendorCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
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

// ListGlobal lists all Global Vendors (with NULL app_id) without tenant isolation.
func (r *pgRepository) ListGlobal(ctx context.Context) ([]*model.Vendor, error) {
	vendorCollection := vendorCollection{}
	if err := r.listerGlobal.ListGlobalWithSelectForUpdate(ctx, &vendorCollection, repo.NewNullCondition("app_id")); err != nil {
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
