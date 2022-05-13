package integrationsystem

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.integration_systems`

var tableColumns = []string{"id", "name", "description"}

// Converter missing godoc
//go:generate mockery --name=Converter --output=automock --outpkg=automock --case=underscore --disable-version-string
type Converter interface {
	ToEntity(in *model.IntegrationSystem) *Entity
	FromEntity(in *Entity) *model.IntegrationSystem
}

type pgRepository struct {
	creator               repo.CreatorGlobal
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal

	conv Converter
}

// NewRepository missing godoc
func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		creator:               repo.NewCreatorGlobal(resource.IntegrationSystem, tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(resource.IntegrationSystem, tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(resource.IntegrationSystem, tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(resource.IntegrationSystem, tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(resource.IntegrationSystem, tableName, []string{"name", "description"}, []string{"id"}),
		deleterGlobal:         repo.NewDeleterGlobal(resource.IntegrationSystem, tableName),
		conv:                  conv,
	}
}

// Create missing godoc
func (r *pgRepository) Create(ctx context.Context, item model.IntegrationSystem) error {
	return r.creator.Create(ctx, r.conv.ToEntity(&item))
}

// Get missing godoc
func (r *pgRepository) Get(ctx context.Context, id string) (*model.IntegrationSystem, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

// Exists missing godoc
func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

// List missing godoc
func (r *pgRepository) List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error) {
	var entityCollection Collection
	page, totalCount, err := r.pageableQuerierGlobal.ListGlobal(ctx, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return model.IntegrationSystemPage{}, err
	}

	items := make([]*model.IntegrationSystem, 0, len(entityCollection))

	for _, entity := range entityCollection {
		isModel := r.conv.FromEntity(&entity)
		items = append(items, isModel)
	}
	return model.IntegrationSystemPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

// Update missing godoc
func (r *pgRepository) Update(ctx context.Context, model model.IntegrationSystem) error {
	return r.updaterGlobal.UpdateSingleGlobal(ctx, r.conv.ToEntity(&model))
}

// Delete missing godoc
func (r *pgRepository) Delete(ctx context.Context, id string) error {
	return r.deleterGlobal.DeleteOneGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}
