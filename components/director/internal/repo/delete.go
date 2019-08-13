package repo

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type Deleter struct {
	tableName    string
	idColumn     string
	tenantColumn string
}

func NewDeleter(tableName, tenantColumn, idColumn string) *Deleter {
	return &Deleter{tableName: tableName, idColumn: idColumn, tenantColumn: tenantColumn}
}

func (g *Deleter) Delete(ctx context.Context, tenant string, id string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("DELETE FROM %s WHERE %s=$1 AND %s=$2", g.tableName, g.tenantColumn, g.idColumn)
	res, err := persist.Exec(q, tenant, id)
	if err != nil {
		return errors.Wrap(err, "while deleting from database")
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}
	if affected != 1 {
		return fmt.Errorf("delete should remove single row, but removed %d rows", affected)
	}

	return errors.Wrap(err, "while deleting the runtime entity from database")
}
