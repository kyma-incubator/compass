package apptemplateversion

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const tableName string = `public.app_template_versions`

var (
	idTableColumns        = []string{"id"}
	updatableTableColumns = []string{"title", "correlation_ids"}
	tableColumns          = append(idTableColumns, []string{"version", "title", "correlation_ids", "release_date", "created_at", "app_template_id"}...)
)

// EntityConverter missing godoc
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.ApplicationTemplateVersion) (*Entity, error)
	FromEntity(entity *Entity) (*model.ApplicationTemplateVersion, error)
}

type repository struct {
	creator               repo.CreatorGlobal
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal
	listerGlobal          repo.ListerGlobal
	conv                  EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:               repo.NewCreatorGlobal(resource.Application, tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(resource.ApplicationTemplateVersion, tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.ApplicationTemplateVersion, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(resource.ApplicationTemplateVersion, tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.ApplicationTemplateVersion, tableName, updatableTableColumns, idTableColumns),
		deleterGlobal:         repo.NewDeleterGlobal(resource.ApplicationTemplateVersion, tableName),
		listerGlobal:          repo.NewListerGlobal(resource.ApplicationTemplateVersion, tableName, tableColumns),
		conv:                  conv,
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, item model.ApplicationTemplateVersion) error {
	log.C(ctx).Debugf("Converting Application Template with id %s to entity", item.ID)
	entity, err := r.conv.ToEntity(&item)
	if err != nil {
		return errors.Wrapf(err, "while converting Application Template with ID %s", item.ID)
	}

	log.C(ctx).Debugf("Persisting Application Template entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

// Get missing godoc
func (r *repository) Get(ctx context.Context, id string) (*model.ApplicationTemplateVersion, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	result, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Application Template with ID %s", id)
	}

	return result, nil
}

func (r *repository) ListByIDs(ctx context.Context, ids []string) ([]*model.ApplicationTemplateVersion, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var entities EntityCollection
	if err := r.listerGlobal.ListGlobal(ctx, &entities, repo.NewInConditionForStringValues("id", ids)); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, model model.ApplicationTemplateVersion) error {
	entity, err := r.conv.ToEntity(&model)
	if err != nil {
		return errors.Wrapf(err, "while converting Application Template with ID %s", model.ID)
	}

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

// Delete missing godoc
func (r *repository) Delete(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) multipleFromEntities(entities EntityCollection) ([]*model.ApplicationTemplateVersion, error) {
	items := make([]*model.ApplicationTemplateVersion, 0, len(entities))
	for _, ent := range entities {
		m, err := r.conv.FromEntity(&ent)
		if err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}
