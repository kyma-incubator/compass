package repo

import (
	"context"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type Deleter struct {
	tableName    string
	tenantColumn string
}

func NewDeleter(tableName, tenantColumn string) *Deleter {
	return &Deleter{tableName: tableName, tenantColumn: tenantColumn}
}

func (g *Deleter) DeleteOne(ctx context.Context, tenant string, conditions Conditions) error {
	return g.delete(ctx, tenant, conditions, true)
}

func (g *Deleter) DeleteMany(ctx context.Context, tenant string, conditions Conditions) error {
	return g.delete(ctx, tenant, conditions, false)
}

func (g *Deleter) delete(ctx context.Context, tenant string, conditions Conditions, requireSingleRemoval bool) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("DELETE FROM %s WHERE %s = $1", g.tableName, g.tenantColumn)
	q = appendEnumeratedConditions(q, 2, conditions)
	allArgs := getAllArgs(tenant, conditions)
	res, err := persist.Exec(q, allArgs...)
	if err != nil {
		return errors.Wrap(err, "while deleting from database")
	}
	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if affected != 1 {
			return fmt.Errorf("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}
