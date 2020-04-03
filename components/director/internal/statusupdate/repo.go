package statusupdate

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
)

const (
	updateQuery = "UPDATE %s SET status_condition = 'CONNECTED', status_timestamp = $1 WHERE id = $2"
	existsQuery = "SELECT 1 FROM %s WHERE id = $1 AND status_condition = 'CONNECTED'"
)

func (u *update) UpdateStatus(id string) error {
	tx, err := u.transact.Begin()
	if err != nil {
		return errors.Wrap(err, "while opening transaction")
	}
	defer u.transact.RollbackUnlessCommited(tx)

	stmt := fmt.Sprintf(updateQuery, u.table)

	_, err = tx.Exec(stmt, u.timestampGen(), id)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("while updating %s status", u.table))
	}

	err = tx.Commit()
	if err != nil {
		return errors.Wrap(err, "while committing the automatic update")
	}
	return nil
}

func (u *update) IsConnected(id string) (bool, error) {
	tx, err := u.transact.Begin()
	if err != nil {
		return false, errors.Wrap(err, "while opening transaction")
	}
	defer u.transact.RollbackUnlessCommited(tx)

	stmt := fmt.Sprintf(existsQuery, u.table)

	var count int
	err = tx.Get(&count, stmt, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}

	err = tx.Commit()
	if err != nil {
		return false, errors.Wrap(err, "while committing transaction")
	}
	return true, nil

}
