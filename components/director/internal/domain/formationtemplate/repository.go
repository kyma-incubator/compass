package formationtemplate

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

const tableName string = `public.formation_templates`

var (
	updatableTableColumns = []string{"name", "application_types", "runtime_types", "missing_artifact_info_message", "missing_artifact_warning_message"}
	idTableColumns        = []string{"id"}
	tableColumns          = append(idTableColumns, updatableTableColumns...)
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.FormationTemplate) (*Entity, error)
	FromEntity(entity *Entity) (*model.FormationTemplate, error)
}

type repository struct {
	creator               repo.CreatorGlobal
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal
	conv                  EntityConverter
}

// NewRepository missing godoc
func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:               repo.NewCreatorGlobal(resource.FormationTemplate, tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(resource.FormationTemplate, tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.FormationTemplate, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(resource.FormationTemplate, tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.FormationTemplate, tableName, updatableTableColumns, idTableColumns),
		deleterGlobal:         repo.NewDeleterGlobal(resource.FormationTemplate, tableName),
		conv:                  conv,
	}
}

func (r *repository) Create(ctx context.Context, item *model.FormationTemplate) error {
	if item == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	log.C(ctx).Debugf("Converting Formation Template with id %s to entity", item.ID)
	entity, err := r.conv.ToEntity(item)
	if err != nil {
		return errors.Wrapf(err, "while converting Template Template with ID %s", item.ID)
	}

	log.C(ctx).Debugf("Persisting Template Template entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

func (r *repository) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
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

func (r *repository) List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.ListGlobal(ctx, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return nil, err
	}

	items := make([]*model.FormationTemplate, 0, len(entityCollection))

	for _, entity := range entityCollection {
		isModel, err := r.conv.FromEntity(entity)
		if err != nil {
			return nil, errors.Wrapf(err, "while converting Formation Template entity with ID %s", entity.ID)
		}

		items = append(items, isModel)
	}
	return &model.FormationTemplatePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *repository) Update(ctx context.Context, model *model.FormationTemplate) error {
	if model == nil {
		return apperrors.NewInternalError("model can not be empty")
	}

	entity, err := r.conv.ToEntity(model)
	if err != nil {
		return errors.Wrapf(err, "while converting Formation Template with ID %s", model.ID)
	}

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}
