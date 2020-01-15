package tenant

import (
	"context"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.tenant_mapping`

var tableColumns = []string{"id", "name", "external_tenant", "internal_tenant", "provider_name", "status"}

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in *model.TenantMapping) *Entity
	FromEntity(in *Entity) *model.TenantMapping
}

type pgRepository struct {
	creator               repo.Creator
	existQuerierGlobal    repo.ExistQuerierGlobal
	singleGetterGlobal    repo.SingleGetterGlobal
	pageableQuerierGlobal repo.PageableQuerierGlobal
	updaterGlobal         repo.UpdaterGlobal
	deleterGlobal         repo.DeleterGlobal

	conv Converter
}

func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		creator:               repo.NewCreator(tableName, tableColumns),
		existQuerierGlobal:    repo.NewExistQuerierGlobal(tableName),
		singleGetterGlobal:    repo.NewSingleGetterGlobal(tableName, tableColumns),
		pageableQuerierGlobal: repo.NewPageableQuerierGlobal(tableName, tableColumns),
		updaterGlobal:         repo.NewUpdaterGlobal(tableName, []string{"name", "external_tenant", "internal_tenant", "provider_name", "status"}, []string{"id"}),
		deleterGlobal:         repo.NewDeleterGlobal(tableName),
		conv:                  conv,
	}
}

func (r *pgRepository) Create(ctx context.Context, item model.TenantMapping) error {
	return r.creator.Create(ctx, r.conv.ToEntity(&item))
}

func (r *pgRepository) CreateMany(ctx context.Context, items []model.TenantMapping) error {
	for index, item := range items {
		err := r.creator.Create(ctx, r.conv.ToEntity(&item))
		if err != nil {
			return errors.Wrapf(err, "while persisting %d item", index)
		}
	}
	return nil
}

func (r *pgRepository) Get(ctx context.Context, id string) (*model.TenantMapping, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id),
		repo.NewNotEqualCondition("status", string(Inactive))}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

func (r *pgRepository) GetByExternalTenant(ctx context.Context, externalTenant, provider string) (*model.TenantMapping, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("external_tenant", externalTenant),
		repo.NewEqualCondition("provider_name", provider), repo.NewNotEqualCondition("status", string(Inactive))}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

func (r *pgRepository) GetByInternalTenant(ctx context.Context, internalTenant string) (*model.TenantMapping, error) {
	var entity Entity
	if err := r.singleGetterGlobal.GetGlobal(ctx, repo.Conditions{repo.NewEqualCondition("internal_tenant", internalTenant),
		repo.NewNotEqualCondition("status", string(Inactive))}, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) ExistsByExternalTenant(ctx context.Context, externalTenant, provider string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("external_tenant", externalTenant),
		repo.NewEqualCondition("provider_name", provider)})
}

func (r *pgRepository) ExistsByInternalTenant(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition("internal_tenant", id)})
}

func (r *pgRepository) List(ctx context.Context, pageSize int, cursor string) (*model.TenantMappingPage, error) {
	var entityCollection EntityCollection
	page, totalCount, err := r.pageableQuerierGlobal.ListGlobal(ctx, pageSize, cursor, "id", &entityCollection)
	if err != nil {
		return &model.TenantMappingPage{}, err
	}

	var items []*model.TenantMapping

	for _, entity := range entityCollection {
		tmModel := r.conv.FromEntity(&entity)
		items = append(items, tmModel)
	}
	return &model.TenantMappingPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) Update(ctx context.Context, model *model.TenantMapping) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	entity := r.conv.ToEntity(model)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}
