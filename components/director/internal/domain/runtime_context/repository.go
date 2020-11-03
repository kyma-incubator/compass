package runtime_context

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

const runtimeContextsTable string = `public.runtime_contexts`

var (
	runtimeContextColumns = []string{"id", "runtime_id", "tenant_id", "key", "value"}
	tenantColumn          = "tenant_id"
)

type pgRepository struct {
	existQuerier       repo.ExistQuerier
	singleGetter       repo.SingleGetter
	singleGetterGlobal repo.SingleGetterGlobal
	deleter            repo.Deleter
	pageableQuerier    repo.PageableQuerier
	creator            repo.Creator
	updater            repo.Updater
}

func NewRepository() *pgRepository {
	return &pgRepository{
		existQuerier:       repo.NewExistQuerier(resource.RuntimeContext, runtimeContextsTable, tenantColumn),
		singleGetter:       repo.NewSingleGetter(resource.RuntimeContext, runtimeContextsTable, tenantColumn, runtimeContextColumns),
		singleGetterGlobal: repo.NewSingleGetterGlobal(resource.RuntimeContext, runtimeContextsTable, runtimeContextColumns),
		deleter:            repo.NewDeleter(resource.RuntimeContext, runtimeContextsTable, tenantColumn),
		pageableQuerier:    repo.NewPageableQuerier(resource.RuntimeContext, runtimeContextsTable, tenantColumn, runtimeContextColumns),
		creator:            repo.NewCreator(resource.RuntimeContext, runtimeContextsTable, runtimeContextColumns),
		updater:            repo.NewUpdater(resource.RuntimeContext, runtimeContextsTable, []string{"key", "value"}, tenantColumn, []string{"id"}),
	}
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.existQuerier.Exists(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) Delete(ctx context.Context, tenant string, id string) error {
	return r.deleter.DeleteOne(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error) {
	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetter.Get(ctx, tenant, repo.Conditions{repo.NewEqualCondition("id", id)}, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return runtimeCtxEnt.ToModel(), nil
}

func (r *pgRepository) GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, args, err := label.FilterQuery(model.RuntimeContextLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetter.Get(ctx, tenant, additionalConditions, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return runtimeCtxEnt.ToModel(), nil
}

func (r *pgRepository) GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error) {
	filterSubquery, args, err := label.FilterQueryGlobal(model.RuntimeContextLabelableObject, label.IntersectSet, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var additionalConditions repo.Conditions
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	var runtimeCtxEnt RuntimeContext
	if err := r.singleGetterGlobal.GetGlobal(ctx, additionalConditions, repo.NoOrderBy, &runtimeCtxEnt); err != nil {
		return nil, err
	}

	return runtimeCtxEnt.ToModel(), nil
}

type RuntimeContextCollection []RuntimeContext

func (r RuntimeContextCollection) Len() int {
	return len(r)
}

func (r *pgRepository) List(ctx context.Context, runtimeID string, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error) {
	var runtimeCtxsCollection RuntimeContextCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}
	filterSubquery, args, err := label.FilterQuery(model.RuntimeContextLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	conditions := repo.Conditions{
		repo.NewEqualCondition("runtime_id", runtimeID),
	}
	if filterSubquery != "" {
		conditions = append(conditions, repo.NewInConditionForSubQuery("id", filterSubquery, args))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &runtimeCtxsCollection, conditions...)

	if err != nil {
		return nil, err
	}

	var items []*model.RuntimeContext

	for _, runtimeCtxEnt := range runtimeCtxsCollection {
		items = append(items, runtimeCtxEnt.ToModel())
	}
	return &model.RuntimeContextPage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page,
	}, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.RuntimeContext) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	return r.creator.Create(ctx, EntityFromRuntimeContextModel(item))
}

func (r *pgRepository) Update(ctx context.Context, item *model.RuntimeContext) error {
	if item == nil {
		return apperrors.NewInternalError("item can not be empty")
	}
	return r.updater.UpdateSingle(ctx, EntityFromRuntimeContextModel(item))
}
