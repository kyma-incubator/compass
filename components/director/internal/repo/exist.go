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
	idColumn     string
	tenantColumn string
}

func NewExistQuerier(tableName, idColumn, tenantColumn string) *ExistQuerier {
	return &ExistQuerier{tableName: tableName, idColumn: idColumn, tenantColumn: tenantColumn}
}

func (g *ExistQuerier) Exists(ctx context.Context, id, tenant string) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	q := fmt.Sprintf("select 1 from %s where %s=$1 and %s=$2", g.tableName, g.idColumn, g.tenantColumn)
	var count int
	err = persist.Get(&count, q, id, tenant)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}
	return true, nil
}
