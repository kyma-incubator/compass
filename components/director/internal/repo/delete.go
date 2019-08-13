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

func NewDeleter(tableName, idColumn, tenantColumn string) *Deleter {
	return &Deleter{tableName: tableName, idColumn: idColumn, tenantColumn: tenantColumn}
}

func (g *Deleter) Delete(ctx context.Context, id string, tenant string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("delete from %s where %s=$1 and %s=$2", g.tableName, g.idColumn, g.tenantColumn)
	res, err := persist.Exec(q, id, tenant)
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
