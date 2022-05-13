package tombstone

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tombstoneTable string = `public.tombstones`

var (
	tombstoneColumns = []string{"ord_id", "app_id", "removal_date", "id"}
	updatableColumns = []string{"removal_date"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Tombstone) *Entity
	FromEntity(entity *Entity) (*model.Tombstone, error)
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
		existQuerier: repo.NewExistQuerier(tombstoneTable),
		singleGetter: repo.NewSingleGetter(tombstoneTable, tombstoneColumns),
		lister:       repo.NewLister(tombstoneTable, tombstoneColumns),
		deleter:      repo.NewDeleter(tombstoneTable),
		creator:      repo.NewCreator(tombstoneTable, tombstoneColumns),
		updater:      repo.NewUpdater(tombstoneTable, updatableColumns, []string{"id"}),
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, tenant string, model *model.Tombstone) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Tombstone entity with id %q", model.ID)
	return r.creator.Create(ctx, resource.Tombstone, tenant, r.conv.ToEntity(model))
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, tenant string, model *model.Tombstone) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Tombstone entity with id %q", model.ID)
	return r.updater.UpdateSingle(ctx, resource.Tombstone, tenant, r.conv.ToEntity(model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Tombstone entity with id %q", id)
	return r.deleter.DeleteOne(ctx, resource.Tombstone, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, resource.Tombstone, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// GetByID missing godoc
func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Tombstone, error) {
	log.C(ctx).Debugf("Getting Tombstone entity with id %q", id)
	var tombstoneEnt Entity
	if err := r.singleGetter.Get(ctx, resource.Tombstone, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &tombstoneEnt); err != nil {
		return nil, err
	}

	tombstoneModel, err := r.conv.FromEntity(&tombstoneEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Tombstone from Entity")
	}

	return tombstoneModel, nil
}

// ListByApplicationID missing godoc
func (r *pgRepository) ListByApplicationID(ctx context.Context, tenantID, appID string) ([]*model.Tombstone, error) {
	tombstoneCollection := tombstoneCollection{}
	if err := r.lister.ListWithSelectForUpdate(ctx, resource.Tombstone, tenantID, &tombstoneCollection, repo.NewEqualCondition("app_id", appID)); err != nil {
		return nil, err
	}
	tombstones := make([]*model.Tombstone, 0, tombstoneCollection.Len())
	for _, tombstone := range tombstoneCollection {
		tombstoneModel, err := r.conv.FromEntity(&tombstone)
		if err != nil {
			return nil, err
		}
		tombstones = append(tombstones, tombstoneModel)
	}
	return tombstones, nil
}

type tombstoneCollection []Entity

// Len missing godoc
func (pc tombstoneCollection) Len() int {
	return len(pc)
}
