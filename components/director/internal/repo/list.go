package repo

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
)

type Lister struct {
	tableName       string
	selectedColumns string
	tenantColumn    string
}

func NewLister(tableName, tenantColumn string, selectedColumns []string) *Lister {
	return &Lister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    tenantColumn,
	}
}

func (l *Lister) List(ctx context.Context, tenant string, dest Collection, additionalConditions ...string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	stmt := fixSelectStatement(l.selectedColumns, l.tableName, l.tenantColumn, additionalConditions)

	err = persist.Select(dest, stmt, tenant)
	if err != nil {
		return errors.Wrap(err, "while fetching list of objects from DB")
	}

	return nil
}
