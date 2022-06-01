package apptemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const tableName string = `public.app_templates`

var (
	updatableTableColumns = []string{"name", "description", "application_input", "placeholders", "access_level"}
	idTableColumns        = []string{"id"}
	tableColumns          = append(idTableColumns, updatableTableColumns...)
)

// EntityConverter missing godoc
//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type EntityConverter interface {
	ToEntity(in *model.ApplicationTemplate) (*Entity, error)
	FromEntity(entity *Entity) (*model.ApplicationTemplate, error)
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
		creator:               repo.NewCreatorGlobal(resource.ApplicationTemplate, tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(resource.ApplicationTemplate, tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.ApplicationTemplate, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(resource.ApplicationTemplate, tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.ApplicationTemplate, tableName, updatableTableColumns, idTableColumns),
		deleterGlobal:         repo.NewDeleterGlobal(resource.ApplicationTemplate, tableName),
		conv:                  conv,
	}
}

// Create missing godoc
func (r *repository) Create(ctx context.Context, item model.ApplicationTemplate) error {
	log.C(ctx).Debugf("Converting Application Template with id %s to entity", item.ID)
	entity, err := r.conv.ToEntity(&item)
	if err != nil {
		return errors.Wrapf(err, "while converting Application Template with ID %s", item.ID)
	}

	log.C(ctx).Debugf("Persisting Application Template entity with id %s to db", item.ID)
	return r.creator.Create(ctx, entity)
}

// Get missing godoc
func (r *repository) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
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

// TODO missing godoc
func (r *repository) GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.AppTemplateLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, additionalConditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	model, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Application Template with ID %s", entity.ID)
	}

	return model, nil
}

// GetByName missing godoc
func (r *repository) GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("name", name)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	result, err := r.conv.FromEntity(&entity)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Application Template with [name=%s]", name)
	}

	return result, nil
}

// Exists missing godoc
func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// List missing godoc
func (r *repository) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (model.ApplicationTemplatePage, error) {
	var entityCollection EntityCollection

	filterSubquery, args, err := label.FilterQueryGlobal(model.AppTemplateLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return model.ApplicationTemplatePage{}, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	conditionsTree := repo.And(repo.ConditionTreesFromConditions(conditions)...)

	page, totalCount, err := r.pageableQuerierGlobal.ListGlobalWithAdditionalConditions(ctx, pageSize, cursor, "id", &entityCollection, conditionsTree)
	if err != nil {
		return model.ApplicationTemplatePage{}, err
	}

	items := make([]*model.ApplicationTemplate, 0, len(entityCollection))

	for _, entity := range entityCollection {
		isModel, err := r.conv.FromEntity(&entity)
		if err != nil {
			return model.ApplicationTemplatePage{}, errors.Wrapf(err, "while converting Application Template entity with ID %s", entity.ID)
		}

		items = append(items, isModel)
	}
	return model.ApplicationTemplatePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// Update missing godoc
func (r *repository) Update(ctx context.Context, model model.ApplicationTemplate) error {
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
