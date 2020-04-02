package statusupdate

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

const query = "UPDATE %s SET status_condition = 'CONNECTED', status_timestamp = %s WHERE id = '%s'"

func (u *update) UpdateStatus(ctx context.Context, id string) error {
	tx, err := u.transact.Begin()
	if err != nil {
		return err
	}
	defer u.transact.RollbackUnlessCommited(tx)

	stmt := fmt.Sprintf(query, u.table, u.timestampGen(), id)

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	_, err := persist.NamedExec()
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
