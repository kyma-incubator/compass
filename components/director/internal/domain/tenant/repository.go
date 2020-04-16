package tenant

import (
	"context"
	"fmt"

	"github.com/lib/pq"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

const tableName string = `public.business_tenant_mappings`

var tableColumns = []string{idColumn, externalNameColumn, externalTenantColumn, providerNameColumn, statusColumn}
var (
	idColumn             = "id"
	externalNameColumn   = "external_name"
	externalTenantColumn = "external_tenant"
	providerNameColumn   = "provider_name"
	statusColumn         = "status"
)

//go:generate mockery -name=Converter -output=automock -outpkg=automock -case=underscore
type Converter interface {
	ToEntity(in *model.BusinessTenantMapping) *Entity
	FromEntity(in *Entity) *model.BusinessTenantMapping
}

type pgRepository struct {
	creator            repo.Creator
	existQuerierGlobal repo.ExistQuerierGlobal
	singleGetterGlobal repo.SingleGetterGlobal
	listerGlobal       repo.ListerGlobal
	updaterGlobal      repo.UpdaterGlobal
	deleterGlobal      repo.DeleterGlobal

	conv Converter
}

func NewRepository(conv Converter) *pgRepository {
	return &pgRepository{
		creator:            repo.NewCreator(tableName, tableColumns),
		existQuerierGlobal: repo.NewExistQuerierGlobal(tableName),
		singleGetterGlobal: repo.NewSingleGetterGlobal(tableName, tableColumns),
		listerGlobal:       repo.NewListerGlobal(tableName, tableColumns),
		updaterGlobal:      repo.NewUpdaterGlobal(tableName, []string{externalNameColumn, externalTenantColumn, providerNameColumn, statusColumn}, []string{idColumn}),
		deleterGlobal:      repo.NewDeleterGlobal(tableName),
		conv:               conv,
	}
}

func (r *pgRepository) Create(ctx context.Context, item model.BusinessTenantMapping) error {
	return r.creator.Create(ctx, r.conv.ToEntity(&item))
}

func (r *pgRepository) Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	var entity Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(idColumn, id),
		repo.NewNotEqualCondition(statusColumn, string(Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}

	return r.conv.FromEntity(&entity), nil
}

func (r *pgRepository) GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error) {
	var entity Entity
	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
		repo.NewNotEqualCondition(statusColumn, string(Inactive))}
	if err := r.singleGetterGlobal.GetGlobal(ctx, conditions, repo.NoOrderBy, &entity); err != nil {
		return nil, err
	}
	return r.conv.FromEntity(&entity), nil
}

func (r *pgRepository) Exists(ctx context.Context, id string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(idColumn, id)})
}

func (r *pgRepository) ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error) {
	return r.existQuerierGlobal.ExistsGlobal(ctx, repo.Conditions{repo.NewEqualCondition(externalTenantColumn, externalTenant)})
}

func (r *pgRepository) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	var entityCollection EntityCollection

	condition := fmt.Sprintf("%s = %s", statusColumn, pq.QuoteLiteral(string(Active)))
	err := r.listerGlobal.ListGlobal(ctx, &entityCollection, condition)
	if err != nil {
		return nil, err
	}

	var items []*model.BusinessTenantMapping

	for _, entity := range entityCollection {
		tmModel := r.conv.FromEntity(&entity)
		items = append(items, tmModel)
	}
	return items, nil
}

func (r *pgRepository) Update(ctx context.Context, model *model.BusinessTenantMapping) error {
	if model == nil {
		return errors.New("model can not be empty")
	}

	entity := r.conv.ToEntity(model)

	return r.updaterGlobal.UpdateSingleGlobal(ctx, entity)
}

func (r *pgRepository) DeleteByExternalTenant(ctx context.Context, externalTenant string) error {
	conditions := repo.Conditions{
		repo.NewEqualCondition(externalTenantColumn, externalTenant),
	}

	return r.deleterGlobal.DeleteManyGlobal(ctx, conditions)
}
