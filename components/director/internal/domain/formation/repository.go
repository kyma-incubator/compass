package formation

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const tableName string = `public.formations`

var (
	updatableTableColumns = []string{"name"}
	idTableColumns        = []string{"id"}
	tableColumns          = []string{"id", "tenant_id", "formation_template_id", "name"}
	tenantColumn          = "tenant_id"
)

// EntityConverter converts between the internal model and entity
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.Formation, id, tenantID, formationTemplateID string) *Entity
	FromEntity(entity *Entity) *model.Formation
}

type repository struct {
	creator      repo.CreatorGlobal
	getter       repo.SingleGetter
	updater      repo.UpdaterGlobal
	deleter      repo.Deleter
	existQuerier repo.ExistQuerier
	conv         EntityConverter
}

// NewRepository creates a new Formation repository
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:      repo.NewCreatorGlobal(resource.Formations, tableName, tableColumns),
		getter:       repo.NewSingleGetterWithEmbeddedTenant(tableName, tenantColumn, tableColumns),
		updater:      repo.NewUpdaterWithEmbeddedTenant(resource.Formations, tableName, updatableTableColumns, tenantColumn, idTableColumns),
		deleter:      repo.NewDeleterWithEmbeddedTenant(tableName, tenantColumn),
		existQuerier: repo.NewExistQuerierWithEmbeddedTenant(tableName, tenantColumn),
		conv:         conv,
	}
}

// Create creates a Formation with a given input
func (r *repository) Create(ctx context.Context, item *model.Formation, id, tenant, formationTemplateID string) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Formation with name: %q to entity", item.Name)
	entity := r.conv.ToEntity(item, id, tenant, formationTemplateID)

	log.C(ctx).Debugf("Persisting Formation entity with name: %q to the DB", item.Name)
	return r.creator.Create(ctx, entity)
}

// Get returns a Formations by a given id
func (r *repository) Get(ctx context.Context, id, tenantID string) (*model.Formation, error) {
	var entity Entity
	if err := r.getter.Get(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		log.C(ctx).Errorf("An error occurred while getting formation with id: %q", id)
		return nil, errors.Wrapf(err, "An error occurred while getting formation with id: %q", id)
	}

	return r.conv.FromEntity(&entity), nil
}

// Update updates a Formation with the given input
func (r *repository) Update(ctx context.Context, model *model.Formation, id, tenantID, formationTemplateID string) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	exists, err := r.existQuerier.Exists(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
	if err != nil {
		return errors.Wrapf(err, "while ensuring Formation with ID: %q exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.Formations, id)
	}

	return r.updater.UpdateSingleGlobal(ctx, r.conv.ToEntity(model, id, tenantID, formationTemplateID))
}

// DeleteByName deletes a Formation with given name
func (r *repository) DeleteByName(ctx context.Context, name, tenantID string) error {
	log.C(ctx).Debugf("Deleting formation with name: %q...", name)
	return r.deleter.DeleteOne(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("name", name)})
}

// Exists check if a Formation with given ID exists
func (r *repository) Exists(ctx context.Context, id, tenantID string) (bool, error) {
	log.C(ctx).Debugf("Cheking if formation with ID: %q exists...", id)
	return r.existQuerier.Exists(ctx, resource.Formations, tenantID, repo.Conditions{repo.NewEqualCondition("id", id)})
}
