package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Updater struct {
	tableName        string
	updatableColumns []string
	tenantColumn     string
	idColumns        []string
}

func NewUpdater(tableName string, updatableColumns []string, tenantColumn string, idColumns []string) *Updater {
	return &Updater{
		tableName:        tableName,
		updatableColumns: updatableColumns,
		tenantColumn:     tenantColumn,
		idColumns:        idColumns,
	}
}

func (u *Updater) UpdateSingle(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return errors.New("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var fieldsToSet []string
	for _, c := range u.updatableColumns {
		fieldsToSet = append(fieldsToSet, fmt.Sprintf("%s = :%s", c, c))
	}

	stmt := fmt.Sprintf("UPDATE %s SET %s WHERE %s = :%s", u.tableName, strings.Join(fieldsToSet, ", "), u.tenantColumn, u.tenantColumn)
	for _, idCol := range u.idColumns {
		stmt += fmt.Sprintf(" AND %s = :%s", idCol, idCol)
	}

	res, err := persist.NamedExec(stmt, dbEntity)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return &notUniqueError{}
		}
	}
	if err != nil {
		return errors.Wrap(err, "while updating single entity")
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}
	if affected != 1 {
		return fmt.Errorf("should update single row, but updated %d rows", affected)
	}

	return nil
}
