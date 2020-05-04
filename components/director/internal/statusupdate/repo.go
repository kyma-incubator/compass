package statusupdate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/timestamp"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

const (
	updateQuery = "UPDATE public.%s SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2"
	existsQuery = "SELECT 1 FROM public.%s WHERE id = $1 AND status_condition = 'CONNECTED'"
)

type repository struct {
	timestampGen timestamp.Generator
}

func NewRepository() *repository {
	return &repository{timestampGen: timestamp.DefaultGenerator()}
}

func (r *repository) UpdateStatus(ctx context.Context, id string, object WithStatusObject) error {

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return errors.Wrap(err, "while loading persistence from context")
	}

	stmt := fmt.Sprintf(updateQuery, object)

	_, err = persist.Exec(stmt, r.timestampGen(), id)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while updating %s status", object))
	}

	return nil
}

func (r *repository) IsConnected(ctx context.Context, id string, object WithStatusObject) (bool, error) {

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, errors.Wrap(err, "while loading persistence from context")
	}

	stmt := fmt.Sprintf(existsQuery, object)

	var count int
	err = persist.Get(&count, stmt, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}

	return true, nil

}
