package runtime

import (
	"context"
	"fmt"

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

type pgRepository struct {
	existQuerier    repo.ExistQuerier
	singleGetter    repo.SingleGetter
	deleter         repo.Deleter
	pageableQuerier repo.PageableQuerier
	creator         repo.Creator
	updater         repo.Updater
}

func NewRepository() *pgRepository {
	return &pgRepository{
		existQuerier:    repo.NewExistQuerier(runtimeTable, tenantColumn),
		singleGetter:    repo.NewSingleGetter(runtimeTable, tenantColumn, runtimeColumns),
		deleter:         repo.NewDeleter(runtimeTable, tenantColumn),
		pageableQuerier: repo.NewPageableQuerier(runtimeTable, tenantColumn, runtimeColumns),
		creator:         repo.NewCreator(runtimeTable, runtimeColumns),
		updater:         repo.NewUpdater(runtimeTable, []string{"name", "description", "status_condition", "status_timestamp"}, tenantColumn, []string{"id"}),
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

	runtimeModel, err := runtimeEnt.ToModel()
	if err != nil {
		return nil, errors.Wrap(err, "while creating runtime model from entity")
	}

	return runtimeModel, nil
}

func (r *pgRepository) GetByFiltersAndID(ctx context.Context, tenant, id string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	additionalConditions := repo.Conditions{repo.NewEqualCondition("id", id)}

	filterSubquery, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInCondition("id", filterSubquery))
	}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, tenant, additionalConditions, repo.NoOrderBy, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel, err := runtimeEnt.ToModel()
	if err != nil {
		return nil, errors.Wrap(err, "while creating runtime model from entity")
	}

	return runtimeModel, nil
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
	filterSubquery, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	var additionalConditions []string
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, fmt.Sprintf(`"id" IN (%s)`, filterSubquery))
	}

	page, totalCount, err := r.pageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &runtimesCollection, additionalConditions...)

	if err != nil {
		return nil, err
	}

	var items []*model.Runtime

	for _, runtimeEnt := range runtimesCollection {
		m, err := runtimeEnt.ToModel()
		if err != nil {
			return nil, errors.Wrap(err, "while creating runtime model from entity")
		}

		items = append(items, m)
	}
	return &model.RuntimePage{
		Data:       items,
		TotalCount: totalCount,
		PageInfo:   page}, nil
}

func (r *pgRepository) Create(ctx context.Context, item *model.Runtime) error {
	if item == nil {
		return errors.New("item can not be empty")
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

func (r *pgRepository) GetOldestForFilters(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Runtime, error) {
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID")
	}

	var additionalConditions repo.Conditions
	filterSubquery, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}
	if filterSubquery != "" {
		additionalConditions = append(additionalConditions, repo.NewInCondition("id", filterSubquery))
	}

	orderByParams := repo.OrderByParams{repo.NewAscOrderBy("creation_timestamp")}

	var runtimeEnt Runtime
	if err := r.singleGetter.Get(ctx, tenant, additionalConditions, orderByParams, &runtimeEnt); err != nil {
		return nil, err
	}

	runtimeModel, err := runtimeEnt.ToModel()
	if err != nil {
		return nil, errors.Wrap(err, "while creating runtime model from entity")
	}

	return runtimeModel, nil
}
