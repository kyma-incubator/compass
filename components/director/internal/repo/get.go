package repo

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"
)

type SingleGetter interface {
	Get(ctx context.Context, tenant string, conditions Conditions, dest interface{}) error
}

type SingleGetterGlobal interface {
	GetGlobal(ctx context.Context, conditions Conditions, dest interface{}) error
}

type universalSingleGetter struct {
	tableName       string
	tenantColumn    *string
	selectedColumns string
}

func NewSingleGetter(tableName string, tenantColumn string, selectedColumns []string) SingleGetter {
	return &universalSingleGetter{
		tableName:       tableName,
		tenantColumn:    &tenantColumn,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func NewSingleGetterGlobal(tableName string, selectedColumns []string) SingleGetterGlobal {
	return &universalSingleGetter{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func (g *universalSingleGetter) Get(ctx context.Context, tenant string, conditions Conditions, dest interface{}) error {
	return g.unsafeGet(ctx, str.Ptr(tenant), conditions, dest)
}

func (g *universalSingleGetter) GetGlobal(ctx context.Context, conditions Conditions, dest interface{}) error {
	return g.unsafeGet(ctx, nil, conditions, dest)
}

func (g *universalSingleGetter) unsafeGet(ctx context.Context, tenant *string, conditions Conditions, dest interface{}) error {
	if dest == nil {
		return errors.New("item cannot be nil")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder
	startIdx := 1

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", g.selectedColumns, g.tableName))
	if tenant != nil || len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}
	if tenant != nil {
		stmtBuilder.WriteString(fmt.Sprintf(" %s = $1", *g.tenantColumn))
		if len(conditions) > 0 {
			stmtBuilder.WriteString(" AND")
		}
		startIdx = 2
	}
	err = writeEnumeratedConditions(&stmtBuilder, startIdx, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(tenant, conditions)
	err = persist.Get(dest, stmtBuilder.String(), allArgs...)
	switch {
	case err == sql.ErrNoRows:
		return apperrors.NewNotFoundError("")
	case err != nil:
		return errors.Wrap(err, "while getting object from DB")
	}
	return nil
}
