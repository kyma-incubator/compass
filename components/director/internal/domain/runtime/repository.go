package runtime

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/lib/pq"
)

const runtimeTable string = `"public"."runtimes"`

type pgRepository struct{}

func NewPostgresRepository() *pgRepository {
	return &pgRepository{}
}

func (r *pgRepository) GetByID(ctx context.Context, tenant, id string) (*model.Runtime, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT "id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth" FROM %s WHERE "id" = $1 AND "tenant_id" = $2`,
		runtimeTable)

	var runtimeEnt Runtime
	err = persist.Get(&runtimeEnt, stmt, id, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtime from DB")
	}

	runtimeModel, err := runtimeEnt.ToModel()

	return runtimeModel, errors.Wrap(err, "while creating runtime model from entity")
}

// TODO: Make filtering and paging
func (r *pgRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching DB from context")
	}

	stmt := fmt.Sprintf(`SELECT "id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth" FROM %s WHERE "tenant_id" = $1`,
		runtimeTable)

	var runtimesEnt []Runtime
	err = persist.Select(&runtimesEnt, stmt, tenant)
	if err != nil {
		return nil, errors.Wrap(err, "while fetching runtimes from DB")
	}

	var items []*model.Runtime

	for _, runtimeEnt := range runtimesEnt {
		model, err := runtimeEnt.ToModel()
		if err != nil {
			return nil, errors.Wrap(err, "while creating runtime model from entity")
		}

		items = append(items, model)
	}

	return &model.RuntimePage{
		Data:       items,
		TotalCount: len(items),
		PageInfo: &pagination.Page{
			StartCursor: "",
			EndCursor:   "",
			HasNextPage: false,
		},
	}, nil
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

	stmt := fmt.Sprintf(`INSERT INTO %s ("id", "tenant_id", "name", "description", "status_condition", "status_timestamp", "auth") VALUES (:id, :tenant_id, :name, :description, :status_condition, :status_timestamp, :auth)`,
		runtimeTable)

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

func (r *pgRepository) Delete(ctx context.Context, id string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while fetching persistence from context")
	}

	stmt := fmt.Sprintf(`DELETE FROM %s WHERE "id" = $1`, runtimeTable)
	_, err = persist.Exec(stmt, id)

	return errors.Wrap(err, "while deleting the runtime entity from database")
}
