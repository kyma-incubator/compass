package labeldef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const (
	tableName     = "public.label_definitions"
	tenantColumn  = "tenant_id"
	keyColumn     = "key"
	schemaColumn  = "schema"
	versionColumn = "version"
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	ToEntity(in model.LabelDefinition) (Entity, error)
	FromEntity(in Entity) (model.LabelDefinition, error)
}

var (
	idColumns          = []string{"id"}
	versionedIDColumns = append(idColumns, versionColumn)
	labeldefColumns    = []string{"id", tenantColumn, keyColumn, schemaColumn, versionColumn}
	updatableColumns   = []string{"schema"}
)

type repository struct {
	conv             EntityConverter
	creator          repo.Creator
	getter           repo.SingleGetter
	lister           repo.Lister
	existQuerier     repo.ExistQuerier
	deleter          repo.Deleter
	updater          repo.Updater
	versionedUpdater repo.Updater
	upserter         repo.Upserter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{conv: conv,
		creator:          repo.NewCreator(resource.LabelDefinition, tableName, labeldefColumns),
		getter:           repo.NewSingleGetter(resource.LabelDefinition, tableName, tenantColumn, labeldefColumns),
		existQuerier:     repo.NewExistQuerier(resource.LabelDefinition, tableName, tenantColumn),
		lister:           repo.NewLister(resource.LabelDefinition, tableName, tenantColumn, labeldefColumns),
		deleter:          repo.NewDeleter(resource.LabelDefinition, tableName, tenantColumn),
		updater:          repo.NewUpdater(resource.LabelDefinition, tableName, updatableColumns, tenantColumn, idColumns),
		versionedUpdater: repo.NewUpdater(resource.LabelDefinition, tableName, updatableColumns, tenantColumn, versionedIDColumns),
		upserter:         repo.NewUpserter(resource.LabelDefinition, tableName, labeldefColumns, []string{tenantColumn, keyColumn}, []string{schemaColumn}),
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, def model.LabelDefinition) error {
	entity, err := r.conv.ToEntity(def)
	if err != nil {
		return errors.Wrap(err, "while converting Label Definition to insert")
	}

	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrap(err, "while inserting Label Definition")
	}
	return nil
}

// Upsert missing godoc
func (r *repository) Upsert(ctx context.Context, label model.LabelDefinition) error {
	labelEntity, err := r.conv.ToEntity(label)
	if err != nil {
		return errors.Wrap(err, "while creating label definition entity from model")
	}

	return r.upserter.Upsert(ctx, labelEntity)
}

// GetByKey missing godoc
func (r *repository) GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	conds := repo.Conditions{repo.NewEqualCondition("key", key)}
	dest := Entity{}

	err := r.getter.Get(ctx, tenant, conds, repo.NoOrderBy, &dest)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Label Definition by key=%s", key)
	}

	out, err := r.conv.FromEntity(dest)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Label Definition")
	}
	return &out, nil
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, tenant string, key string) (bool, error) {
	conds := repo.Conditions{repo.NewEqualCondition("key", key)}
	return r.existQuerier.Exists(ctx, tenant, conds)
}

// List missing godoc
func (r *repository) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	var dest EntityCollection

	err := r.lister.List(ctx, tenant, &dest)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Label Definitions")
	}
	out := make([]model.LabelDefinition, 0, len(dest))
	for _, entity := range dest {
		ld, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Label Definition [key=%s]", entity.Key)
		}
		out = append(out, ld)
	}
	return out, nil
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, def model.LabelDefinition) error {
	entity, err := r.conv.ToEntity(def)
	if err != nil {
		return errors.Wrap(err, "while creating Label Definition entity from model")
	}
	return r.updater.UpdateSingleWithVersion(ctx, entity)
}

// UpdateWithVersion missing godoc
func (r *repository) UpdateWithVersion(ctx context.Context, def model.LabelDefinition) error {
	entity, err := r.conv.ToEntity(def)
	if err != nil {
		return errors.Wrap(err, "while creating Label Definition entity from model")
	}
	return r.versionedUpdater.UpdateSingleWithVersion(ctx, entity)
}

// DeleteByKey missing godoc
func (r *repository) DeleteByKey(ctx context.Context, tenant, key string) error {
	conds := repo.Conditions{repo.NewEqualCondition("key", key)}
	return r.deleter.DeleteOne(ctx, tenant, conds)
}
