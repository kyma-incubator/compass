package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type ExistQuerier struct {
	tableName    string
	tenantColumn string
}

func NewExistQuerier(tableName, tenantColumn string) *ExistQuerier {
	return &ExistQuerier{tableName: tableName, tenantColumn: tenantColumn}
}

func (g *ExistQuerier) Exists(ctx context.Context, tenant string, conditions Conditions) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	q := fmt.Sprintf("SELECT 1 FROM %s WHERE %s = $1", g.tableName, g.tenantColumn)
	q = appendEnumeratedConditions(q, 2, conditions)

	allArgs := getAllArgs(tenant, conditions)
	var count int
	err = persist.Get(&count, q, allArgs...)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}
	return true, nil
}
