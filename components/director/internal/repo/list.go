package repo

import (
	"context"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

type Lister interface {
	List(ctx context.Context, tenant string, dest Collection, additionalConditions ...Condition) error
}

type ListerGlobal interface {
	ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error
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

func (l *universalLister) List(ctx context.Context, tenant string, dest Collection, additionalConditions ...Condition) error {
	if tenant == "" {
		return errors.New("tenant cannot be empty")
	}
	additionalConditions = append(Conditions{NewEqualCondition(*l.tenantColumn, tenant)}, additionalConditions...)
	return l.unsafeList(ctx, dest, additionalConditions...)
}

func (l *universalLister) ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error {
	return l.unsafeList(ctx, dest, additionalConditions...)
}

func (l *universalLister) unsafeList(ctx context.Context, dest Collection, conditions ...Condition) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(l.tableName, l.selectedColumns, conditions, OrderByParams{})
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	err = persist.Select(dest, query, args...)
	if err != nil {
		return errors.Wrap(err, "while fetching list of objects from DB")
	}

	return nil
}
