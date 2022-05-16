package bundleinstanceauth

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const tableName string = `public.bundle_instance_auths`

var (
	idColumns        = []string{"id"}
	updatableColumns = []string{"auth_value", "status_condition", "status_timestamp", "status_message", "status_reason"}
	tableColumns     = []string{"id", "owner_id", "bundle_id", "context", "input_params", "auth_value", "status_condition", "status_timestamp", "status_message", "status_reason", "runtime_id", "runtime_context_id"}
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.BundleInstanceAuth) (*Entity, error)
	FromEntity(entity *Entity) (*model.BundleInstanceAuth, error)
}

type repository struct {
	creator      repo.CreatorGlobal
	singleGetter repo.SingleGetter
	lister       repo.Lister
	updater      repo.Updater
	deleter      repo.Deleter
	conv         EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		// TODO: We will use the default creator which does not ensures parent access. This is because in order to ensure that the caller
		//  tenant has access to the parent for BIAs we need to check for either owner or non-owner access. The straight-forward
		//  scenario will be to have a non-owner access to the bundle once a formation is created. Then we check for whatever access
		//  the caller has to the parent and allow it.
		//  However, this cannot be done before formations redesign and due to this the formation check will still take place
		//  in the pkg/scenario/directive.go. Once formation redesign in in place we can remove this directive and here we can use non-global creator.
		creator:      repo.NewCreatorGlobal(resource.BundleInstanceAuth, tableName, tableColumns),
		singleGetter: repo.NewSingleGetter(tableName, tableColumns),
		lister:       repo.NewLister(tableName, tableColumns),
		deleter:      repo.NewDeleter(tableName),
		updater:      repo.NewUpdater(tableName, updatableColumns, idColumns),
		conv:         conv,
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, item *model.BundleInstanceAuth) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting BundleInstanceAuth model to entity")
	}

	log.C(ctx).Debugf("Persisting BundleInstanceAuth entity with id %s to db", item.ID)
	err = r.creator.Create(ctx, entity)
	if err != nil {
		return errors.Wrapf(err, "while saving entity with id %s to db", item.ID)
	}

	return nil
}

// GetByID missing godoc
func (r *repository) GetByID(ctx context.Context, tenantID string, id string) (*model.BundleInstanceAuth, error) {
	var entity Entity
	if err := r.singleGetter.Get(ctx, resource.BundleInstanceAuth, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	itemModel, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrap(err, "while converting BundleInstanceAuth entity to model")
	}

	return itemModel, nil
}

// GetForBundle missing godoc
func (r *repository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.BundleInstanceAuth, error) {
	var ent Entity

	conditions := repo.Conditions{
		repo.NewEqualCondition("id", id),
		repo.NewEqualCondition("bundle_id", bundleID),
	}
	if err := r.singleGetter.Get(ctx, resource.BundleInstanceAuth, tenant, conditions, repo.NoOrderBy, &ent); err != nil {
		return nil, err
	}

	bndlModel, err := r.conv.FromEntity(&ent)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Bundle model from entity")
	}

	return bndlModel, nil
}

// ListByBundleID missing godoc
func (r *repository) ListByBundleID(ctx context.Context, tenantID string, bundleID string) ([]*model.BundleInstanceAuth, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("bundle_id", bundleID),
	}

	err := r.lister.List(ctx, resource.BundleInstanceAuth, tenantID, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// ListByRuntimeID missing godoc
func (r *repository) ListByRuntimeID(ctx context.Context, tenantID string, runtimeID string) ([]*model.BundleInstanceAuth, error) {
	var entities Collection

	conditions := repo.Conditions{
		repo.NewEqualCondition("runtime_id", runtimeID),
	}

	err := r.lister.List(ctx, resource.BundleInstanceAuth, tenantID, &entities, conditions...)

	if err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, tenant string, item *model.BundleInstanceAuth) error {
	if item == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrap(err, "while converting model to entity")
	}

	log.C(ctx).Debugf("Updating BundleInstanceAuth entity with id %s in db", item.ID)
	return r.updater.UpdateSingle(ctx, resource.BundleInstanceAuth, tenant, entity)
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, tenantID string, id string) error {
	return r.deleter.DeleteOne(ctx, resource.BundleInstanceAuth, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) multipleFromEntities(entities Collection) ([]*model.BundleInstanceAuth, error) {
	items := make([]*model.BundleInstanceAuth, 0, len(entities))
	for _, ent := range entities {
		m, err := r.conv.FromEntity(&ent)
		if err != nil {
			return nil, errors.Wrap(err, "while creating BundleInstanceAuth model from entity")
		}
		items = append(items, m)
	}
	return items, nil
}
