package runtime

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
)

const runtimeTable string = `"public"."runtimes"`

const runtimeFields string = `"id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth"`

type pgRepository struct {
	repo.ExistQuerier
	repo.SingleGetter
	repo.Deleter
	repo.PageableQuerier
}

func NewPostgresRepository() *pgRepository {
	return &pgRepository{
		ExistQuerier: repo.ExistQuerier{
			Query: fmt.Sprintf(`SELECT 1 FROM %s WHERE "id" = $1 AND "tenant_id" = $2`, runtimeTable),
		},
		SingleGetter: repo.SingleGetter{
			Query: fmt.Sprintf(`SELECT %s FROM %s WHERE "id" = $1 AND "tenant_id" = $2`, runtimeFields, runtimeTable),
		},
		Deleter: repo.Deleter{
			Query: fmt.Sprintf(`DELETE FROM %s WHERE "id" = $1`, runtimeTable),
		},
		PageableQuerier: repo.PageableQuerier{
			query:        fmt.Sprintf(`SELECT %s FROM %s WHERE "tenant_id"  = $1`, runtimeFields, runtimeTable),
			columns:      runtimeFields,
			RelationName: "runtimes",
		},
	}
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error) {
	var runtimeEnt Runtime
	if err := r.SingleGetter.Get(ctx, tenant, id, &runtimeEnt); err != nil {
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
	var runtimesEnt RuntimeCollection
	tenantID, err := uuid.Parse(tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing tenant as UUID") //TODO
	}
	filterSubquery, err := label.FilterQuery(model.RuntimeLabelableObject, label.IntersectSet, tenantID, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while building filter query")
	}

	var page *pagination.Page
	var totalCount int

	if filterSubquery != "" {
		page, totalCount, err = r.PageableQuerier.List(ctx, tenant, pageSize, cursor, &runtimesEnt, fmt.Sprintf(`"id" IN (%s)`, filterSubquery))
	} else {
		page, totalCount, err = r.PageableQuerier.List(ctx, tenant, pageSize, cursor, &runtimesEnt)
	}

	if err != nil {
		return nil, err
	}

	var items []*model.Runtime

	for _, runtimeEnt := range runtimesEnt {
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

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	runtimeEnt, err := EntityFromRuntimeModel(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}

	stmt := fmt.Sprintf(`INSERT INTO %s ( %s )
								VALUES (:id, :tenant_id, :name, :description, :status_condition, :status_timestamp, :auth)`,
		runtimeTable, runtimeFields)

	_, err = persist.NamedExec(stmt, runtimeEnt)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return errors.New("runtime name is not unique within tenant")
		}
	}

	return errors.Wrap(err, "while inserting the runtime entity to database")
}

func (r *pgRepository) Update(ctx context.Context, item *model.Runtime) error {
	if item == nil {
		return errors.New("item can not be empty")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	runtimeEnt, err := EntityFromRuntimeModel(item)
	if err != nil {
		return errors.Wrap(err, "while creating runtime entity from model")
	}

	stmt := fmt.Sprintf(`UPDATE %s SET "name" = :name, "description" = :description, "status_condition" = :status_condition, "status_timestamp" = :status_timestamp WHERE "id" = :id`,
		runtimeTable)
	_, err = persist.NamedExec(stmt, runtimeEnt)

	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return errors.New("runtime name is not unique within tenant")
		}
	}

	return errors.Wrap(err, "while updating the runtime entity in database")
}
