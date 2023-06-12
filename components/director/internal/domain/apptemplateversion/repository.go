package apptemplateversion

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.app_template_versions`

var (
	idTableColumns        = []string{"id"}
	updatableTableColumns = []string{"title", "correlation_ids"}
	tableColumns          = append(idTableColumns, []string{"version", "title", "correlation_ids", "release_date", "created_at", "app_template_id"}...)
)

// EntityConverter converts between internal model and entity
//
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.ApplicationTemplateVersion) *Entity
	FromEntity(entity *Entity) *model.ApplicationTemplateVersion
}

type repository struct {
	creator            repo.CreatorGlobal
	existQuerierGlobal repo.ExistQuerierGlobal
	singleGetterGlobal repo.SingleGetterGlobal
	updaterGlobal      repo.UpdaterGlobal
	listerGlobal       repo.ListerGlobal
	conv               EntityConverter
}

// NewRepository returns a new entity responsible for repo-layer ApplicationTemplateVersion operations.
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:            repo.NewCreatorGlobal(resource.Application, tableName, tableColumns),
		existQuerierGlobal: repo.NewExistQuerierGlobal(resource.ApplicationTemplateVersion, tableName),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.ApplicationTemplateVersion, tableName, tableColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(resource.ApplicationTemplateVersion, tableName, updatableTableColumns, idTableColumns),
		listerGlobal:       repo.NewListerGlobal(resource.ApplicationTemplateVersion, tableName, tableColumns),
		conv:               conv,
	}
}

// Create persists a model.ApplicationTemplateVersion
func (r *repository) Create(ctx context.Context, item model.ApplicationTemplateVersion) error {
	log.C(ctx).Debugf("Converting Application Template Version with id %s to entity", item.ID)
	entity := r.conv.ToEntity(&item)

	log.C(ctx).Debugf("Persisting Application Template Version entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

// GetByAppTemplateIDAndVersion gets globally a model.ApplicationTemplateVersion based on Application Template ID and version
func (r *repository) GetByAppTemplateIDAndVersion(ctx context.Context, appTemplateID, version string) (*model.ApplicationTemplateVersion, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{
		repo.NewEqualCondition("app_template_id", appTemplateID),
		repo.NewEqualCondition("version", version),
	}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

// ListByAppTemplateID lists multiple model.ApplicationTemplateVersion based on Application Template ID
func (r *repository) ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error) {
	var entities EntityCollection
	if err := r.listerGlobal.ListGlobal(ctx, &entities, repo.NewEqualCondition("app_template_id", appTemplateID)); err != nil {
		return nil, err
	}

	return r.multipleFromEntities(entities)
}

// Exists checks if a ApplicationTemplateVersion with a given ID is in the database
func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Update updates a model.ApplicationTemplateVersion
func (r *repository) Update(ctx context.Context, model model.ApplicationTemplateVersion) error {
	entity := r.conv.ToEntity(&model)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

func (r *repository) multipleFromEntities(entities EntityCollection) ([]*model.ApplicationTemplateVersion, error) {
	items := make([]*model.ApplicationTemplateVersion, 0, len(entities))
	for _, ent := range entities {
		m := r.conv.FromEntity(&ent)
		items = append(items, m)
	}
	return items, nil
}
