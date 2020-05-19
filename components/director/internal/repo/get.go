package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/customerrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

type SingleGetter interface {
	Get(ctx context.Context, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
}

type SingleGetterGlobal interface {
	GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error
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

func (g *universalSingleGetter) Get(ctx context.Context, tenant string, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if tenant == "" {
		return errors.New("tenant cannot be empty")
	}
	conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
	return g.unsafeGet(ctx, conditions, orderByParams, dest)
}

func (g *universalSingleGetter) GetGlobal(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	return g.unsafeGet(ctx, conditions, orderByParams, dest)
}

func (g *universalSingleGetter) unsafeGet(ctx context.Context, conditions Conditions, orderByParams OrderByParams, dest interface{}) error {
	if dest == nil {
		return errors.New("item cannot be nil")
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(g.tableName, g.selectedColumns, conditions, orderByParams)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	err = persist.Get(dest, query, args...)

	if err != nil {
		return customerrors.MapSQLError(err)
	}
	return nil
}
