package runtime

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const runtimeTable string = `public.runtimes`

var (
	runtimeColumns = []string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "creation_timestamp"}
	tenantColumn   = "tenant_id"
)

//go:generate mockery --name=EntityConverter --output=automock --outpkg=automock --case=underscore
type EntityConverter interface {
	MultipleFromEntities(entities RuntimeCollection) []*model.Runtime
}

type pgRepository struct {
	existQuerier       repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	pageableQuerier    repo.PageableQuerier
	lister             repo.Lister
	creator            repo.Creator
	updater            repo.Updater
	conv               EntityConverter
}

func NewRepository(conv EntityConverter) *pgRepository {
	return &pgRepository{
		existQuerier:       repo.NewExistQuerier(resource.Runtime, runtimeTable, tenantColumn),
		singleGetter:       repo.NewSingleGetter(resource.Runtime, runtimeTable, tenantColumn, runtimeColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.Runtime, runtimeTable, runtimeColumns),
		deleter:            repo.NewDeleter(resource.Runtime, runtimeTable, tenantColumn),
		pageableQuerier:    repo.NewPageableQuerier(resource.Runtime, runtimeTable, tenantColumn, runtimeColumns),
		lister:             repo.NewLister(resource.Runtime, runtimeTable, tenantColumn, runtimeColumns),
		creator:            repo.NewCreator(resource.Runtime, runtimeTable, runtimeColumns),
		updater:            repo.NewUpdater(resource.Runtime, runtimeTable, []string{"name", "description", "status_condition", "status_timestamp"}, tenantColumn, []string{"id"}),
		conv:               conv,
	}
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenant string, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error) {
	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := runtimeEnt.ToModel()

	return runtimeModel, nil
}

func (r *pgRepository) GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, tenant, additionalConditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := runtimeEnt.ToModel()

	return runtimeModel, nil
}

func (r *pgRepository) GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.RuntimeLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeEnt Runtime
	if err := r.singleGetterGlobal.GetGlobal(ctx, additionalConditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := runtimeEnt.ToModel()

	return runtimeModel, nil
}

func (r *pgRepository) ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	var entities RuntimeCollection

	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	err = r.lister.List(ctx, tenant, &entities, conditions...)
	if err != nil {
		return nil, err
	}

	return r.conv.MultipleFromEntities(entities), nil
}

type RuntimeCollection []Runtime

func (r RuntimeCollection) Len() int {
	return len(r)
}

func (r *pgRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error) {
	var runtimesCollection RuntimeCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var conditions repo.Conditions
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "name", &runtimesCollection, conditions...)

	if err != nil {
		return nil, err
	}

	items := r.conv.MultipleFromEntities(runtimesCollection)

	return &model.RuntimePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.Runtime) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}

	runtimeEnt, err := EntityFromRuntimeModel(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}

	return r.creator.Create(ctx, runtimeEnt)
}

func (r *pgRepository) Update(ctx context.Context, item *model.Runtime) error {
	runtimeEnt, err := EntityFromRuntimeModel(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}
	return r.updater.UpdateSingle(ctx, runtimeEnt)
}

func (r *pgRepository) UpdateTenantID(ctx context.Context, runtimeID, newTenantID string) error {
	updaterGlobal := repo.NewUpdaterGlobal(resource.Runtime, runtimeTable, []string{tenantColumn}, []string{"id"})

	runtimeEnt := &Runtime{
		ID:       runtimeID,
		TenantID: newTenantID,
	}
	return updaterGlobal.UpdateSingleGlobal(ctx, runtimeEnt)
}

func (r *pgRepository) GetOldestForFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	var additionalConditions repo.Conditions
	filterSubquery, args, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	orderByParams := repo.OrderByParams{repo.NewAscOrderBy("creation_timestamp")}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, tenant, additionalConditions, orderByParams, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel := runtimeEnt.ToModel()

	return runtimeModel, nil
}
