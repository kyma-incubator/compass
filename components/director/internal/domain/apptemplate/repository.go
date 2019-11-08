package apptemplate

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"
)

const tableName string = `public.application_templates`

var (
	updatableTableColumns = []string{"name", "description", "application_input", "placeholders", "access_level"}
	idTableColumns        = []string{"id"}
	tableColumns          = append(idTableColumns, updatableTableColumns...)
)

//go:generate mockery -name=EntityConverter -output=automock -outpkg=automock -case=underscore
type EntityConverter interface {
	ToEntity(in *model.ApplicationTemplate) (*Entity, error)
	FromEntity(entity *Entity) *model.ApplicationTemplate
}

type repository struct {
	creator               repo.Creator
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal
	conv                  EntityConverter
}

func NewRepository(conv EntityConverter) *repository {
	return &repository{
		creator:               repo.NewCreator(tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(tableName, updatableTableColumns, idTableColumns),
		deleterGlobal:         repo.NewDeleterGlobal(tableName),
		conv:                  conv,
	}
}

func (r *repository) Create(ctx context.Context, item model.ApplicationTemplate) error {
	entity, err := r.conv.ToEntity(&item)
	if err != nil {
		return errors.Wrapf(err, "while converting Application Template with ID %s", item.ID)
	}

	return r.creator.Create(ctx, entity)
}

func (r *repository) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

func (r *repository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *repository) List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error) {
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.ListGlobal(ctx, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return model.ApplicationTemplatePage{}, err
	}

	var items []*model.ApplicationTemplate

	for _, entity := range entityCollection {
		isModel := r.conv.FromEntity(&entity)
		items = append(items, isModel)
	}
	return model.ApplicationTemplatePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *repository) Update(ctx context.Context, model model.ApplicationTemplate) error {
	entity, err := r.conv.ToEntity(&model)
	if err != nil {
		return errors.Wrapf(err, "while converting Application Template with ID %s", model.ID)
	}

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

func (r *repository) Delete(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

