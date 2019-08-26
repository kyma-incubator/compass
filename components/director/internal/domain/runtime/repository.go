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

const runtimeTable string = `"public"."runtimes"`

var runtimeColumns = []string{"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"}

type pgRepository struct {
	*repo.ExistQuerier
	*repo.SingleGetter
	*repo.Deleter
	*repo.PageableQuerier
	*repo.Creator
	*repo.Updater
}

func NewRepository() *pgRepository {
	return &pgRepository{
		ExistQuerier:    repo.NewExistQuerier(runtimeTable, "tenant_id"),
		SingleGetter:    repo.NewSingleGetter(runtimeTable, "tenant_id", runtimeColumns),
		Deleter:         repo.NewDeleter(runtimeTable, "tenant_id"),
		PageableQuerier: repo.NewPageableQuerier(runtimeTable, "tenant_id", runtimeColumns),
		Creator:         repo.NewCreator(runtimeTable, runtimeColumns),
		Updater:         repo.NewUpdater(runtimeTable, []string{"name", "description", "status_condition", "status_timestamp"}, "tenant_id", []string{"id"}),
	}
}

func (r *pgRepository) Exists(ctx context.Context, tenant, id string) (bool, error) {
	return r.ExistQuerier.Exists(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *pgRepository) Delete(ctx context.Context, tenant string, id string) error {
	return r.Deleter.DeleteOne(ctx, tenant, repo.Conditions{{Field: "id", Val: id}})
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error) {
	var runtimeEnt Runtime
	if err := r.SingleGetter.Get(ctx, tenant, repo.Conditions{{Field: "id", Val: id}}, &runtimeEnt); err != nil {
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
	var additionalConditions string
	if filterSubquery != "" {
		additionalConditions = fmt.Sprintf(`"id" IN (%s)`, filterSubquery)
	}

	page, totalCount, err := r.PageableQuerier.List(ctx, tenant, pageSize, cursor, "id", &runtimesCollection, additionalConditions)

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

	return r.Creator.Create(ctx, runtimeEnt)
}

func (r *pgRepository) Update(ctx context.Context, item *model.Runtime) error {
	runtimeEnt, err := EntityFromRuntimeModel(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}
	return r.Updater.UpdateSingle(ctx, runtimeEnt)
}
