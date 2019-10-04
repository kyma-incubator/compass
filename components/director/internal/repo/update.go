package repo

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/lib/pq"
	"github.com/pkg/errors"
)

type Updater interface {
	UpdateSingle(ctx context.Context, dbEntity interface{}) error
}

type UpdaterGlobal interface {
	UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error
}

type universalUpdater struct {
	tableName        string
	updatableColumns []string
	tenantColumn     *string
	idColumns        []string
}

func NewUpdater(tableName string, updatableColumns []string, tenantColumn string, idColumns []string) Updater {
	return &universalUpdater{
		tableName:        tableName,
		updatableColumns: updatableColumns,
		tenantColumn:     &tenantColumn,
		idColumns:        idColumns,
	}
}

func NewUpdaterGlobal(tableName string, updatableColumns []string, idColumns []string) UpdaterGlobal {
	return &universalUpdater{
		tableName:        tableName,
		updatableColumns: updatableColumns,
		idColumns:        idColumns,
	}
}

func (u *universalUpdater) UpdateSingle(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingle(ctx, dbEntity, false)
}

func (u *universalUpdater) UpdateSingleGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpdateSingle(ctx, dbEntity, true)
}

func (u *universalUpdater) unsafeUpdateSingle(ctx context.Context, dbEntity interface{}, isGlobal bool) error {
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

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("UPDATE %s SET %s", u.tableName, strings.Join(fieldsToSet, ", ")))
	if !isGlobal || len(u.idColumns) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}
	if !isGlobal {
		stmtBuilder.WriteString(fmt.Sprintf(" %s = :%s", *u.tenantColumn, *u.tenantColumn))
		if len(u.idColumns) > 0 {
			stmtBuilder.WriteString(" AND")
		}
	}
	if len(u.idColumns) > 0 {
		var preparedIDColumns []string
		for _, idCol := range u.idColumns {
			preparedIDColumns = append(preparedIDColumns, fmt.Sprintf("%s = :%s", idCol, idCol))
		}
		stmtBuilder.WriteString(fmt.Sprintf(" %s", strings.Join(preparedIDColumns, " AND ")))
	}

	res, err := persist.NamedExec(stmtBuilder.String(), dbEntity)
	if pqerr, ok := err.(*pq.Error); ok {
		if pqerr.Code == persistence.UniqueViolation {
			return apperrors.NewNotUniqueError(pqerr.Error())
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
