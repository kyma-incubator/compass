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
	tenantColumn     = "tenant_id"
	tombstoneColumns = []string{"ord_id", tenantColumn, "app_id", "removal_date"}
	updatableColumns = []string{"removal_date"}
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.Tombstone) *Entity
	FromEntity(entity *Entity) (*model.Tombstone, error)
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
		existQuerier: repo.NewExistQuerier(resource.Tombstone, tombstoneTable, tenantColumn),
		singleGetter: repo.NewSingleGetter(resource.Tombstone, tombstoneTable, tenantColumn, tombstoneColumns),
		deleter:      repo.NewDeleter(resource.Tombstone, tombstoneTable, tenantColumn),
		creator:      repo.NewCreator(resource.Tombstone, tombstoneTable, tombstoneColumns),
		updater:      repo.NewUpdater(resource.Tombstone, tombstoneTable, updatableColumns, tenantColumn, []string{"ord_id"}),
	}
}

func (r *pgRepository) Create(ctx context.Context, model *model.Tombstone) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}

	log.C(ctx).Debugf("Persisting Tombstone entity with id %q", model.OrdID)
	return r.creator.Create(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Update(ctx context.Context, model *model.Tombstone) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be nil")
	}
	log.C(ctx).Debugf("Updating Tombstone entity with id %q", model.OrdID)
	return r.updater.UpdateSingle(ctx, r.conv.ToEntity(model))
}

func (r *pgRepository) Delete(ctx context.Context, tenant, id string) error {
	log.C(ctx).Debugf("Deleting Tombstone entity with id %q", id)
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Tombstone, error) {
	log.C(ctx).Debugf("Getting Tombstone entity with id %q", id)
	var tombstoneEnt Entity
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("ord_id", id)}, repo.NoOrderBy, &tombstoneEnt); err != nil {
		return nil, err
	}

	tombstoneModel, err := r.conv.FromEntity(&tombstoneEnt)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Tombstone from Entity")
	}

	return tombstoneModel, nil
}
