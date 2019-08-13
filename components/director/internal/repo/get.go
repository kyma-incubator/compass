package repo

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type SingleGetter struct {
	tableName       string
	idColumn        string
	tenantColumn    string
	selectedColumns string
}

func NewSingleGetter(tableName, tenantColumn, idColumn, selectedColumns string) *SingleGetter {
	return &SingleGetter{
		tableName:       tableName,
		idColumn:        idColumn,
		tenantColumn:    tenantColumn,
		selectedColumns: selectedColumns,
	}
}

func (g *SingleGetter) Get(ctx context.Context, tenant, id string, dest interface{}) error {
	if dest == nil {
		return errors.New("missing destination")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1 AND  %s = $2", g.selectedColumns, g.tableName, g.tenantColumn, g.idColumn)
	err = persist.Get(dest, q, tenant, id)
	if err != nil {
		if err != sql.ErrNoRows {
			return errors.Wrap(err, "while getting object from DB")
		}
		return &notFoundError{}
	}
	return nil
}
