package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/pkg/errors"
)

type SingleGetter struct {
	tableName       string
	tenantColumn    string
	selectedColumns string
}

func NewSingleGetter(tableName, tenantColumn string, selectedColumns []string) *SingleGetter {
	return &SingleGetter{
		tableName:       tableName,
		tenantColumn:    tenantColumn,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func (g *SingleGetter) Get(ctx context.Context, tenant string, conditions Conditions, dest interface{}) error {
	if dest == nil {
		return errors.New("item cannot be nil")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s = $1", g.selectedColumns, g.tableName, g.tenantColumn)
	q = appendEnumeratedConditions(q, 2, conditions)
	allArgs := getAllArgs(tenant, conditions)
	err = persist.Get(dest, q, allArgs...)
	switch {
	case err == sql.ErrNoRows:
		return &notFoundError{}
	case err != nil:
		return errors.Wrap(err, "while getting object from DB")
	}
	return nil
}
