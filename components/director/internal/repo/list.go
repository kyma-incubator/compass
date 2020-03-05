package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Lister interface {
	List(ctx context.Context, tenant string, dest Collection, additionalConditions ...string) error
}

type ListerGlobal interface {
	ListGlobal(ctx context.Context, dest Collection, additionalConditions ...string) error
}

type universalLister struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
}

func NewLister(tableName string, tenantColumn string, selectedColumns []string) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
	}
}

func NewListerGlobal(tableName string, selectedColumns []string) ListerGlobal {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

func (l *universalLister) List(ctx context.Context, tenant string, dest Collection, additionalConditions ...string) error {
	return l.unsafeList(ctx, str.Ptr(tenant), dest, additionalConditions...)
}

func (l *universalLister) ListGlobal(ctx context.Context, dest Collection, additionalConditions ...string) error {
	return l.unsafeList(ctx, nil, dest, additionalConditions...)
}

func (l *universalLister) unsafeList(ctx context.Context, tenant *string, dest Collection, additionalConditions ...string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	stmt := buildSelectStatement(l.selectedColumns, l.tableName, l.tenantColumn, additionalConditions)

	var args []interface{}
	if tenant != nil {
		args = append(args, *tenant)
	}
	err = persist.Select(dest, stmt, args...)
	if err != nil {
		return errors.Wrap(err, "while fetching list of objects from DB")
	}

	return nil
}
