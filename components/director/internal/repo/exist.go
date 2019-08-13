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

func NewExistQuerier(tableName, tenantColumn, idColumn string) *ExistQuerier {
	return &ExistQuerier{tableName: tableName, idColumn: idColumn, tenantColumn: tenantColumn}
}

func (g *ExistQuerier) Exists(ctx context.Context, tenant, id string) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	q := fmt.Sprintf("SELECT 1 FROM %s WHERE %s=$1 AND %s=$2", g.tableName, g.tenantColumn, g.idColumn)
	var count int
	err = persist.Get(&count, q, tenant, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, errors.Wrap(err, "while getting object from DB")
	}
	return true, nil
}
