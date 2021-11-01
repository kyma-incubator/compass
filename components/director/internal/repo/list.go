package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Lister missing godoc
type Lister interface {
	List(ctx context.Context, tenant string, dest Collection, additionalConditions ...Condition) error
	SetSelectedColumns(selectedColumns []string)
	Clone() Lister
}

// ListerGlobal missing godoc
type ListerGlobal interface {
	ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error
}

type universalLister struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type

	orderByParams OrderByParams
}

// NewLister missing godoc
func NewLister(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string) Lister {
	return &universalLister{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		orderByParams:   NoOrderBy,
	}
}

// NewListerWithOrderBy missing godoc
func NewListerWithOrderBy(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string, orderByParams OrderByParams) Lister {
	return &universalLister{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		orderByParams:   orderByParams,
	}
}

// NewListerGlobal missing godoc
func NewListerGlobal(resourceType resource.Type, tableName string, selectedColumns []string) ListerGlobal {
	return &universalLister{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   NoOrderBy,
	}
}

// List missing godoc
func (l *universalLister) List(ctx context.Context, tenant string, dest Collection, additionalConditions ...Condition) error {
	/*if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}*/
	//additionalConditions = append(Conditions{NewTenantIsolationCondition(*l.tenantColumn, tenant)}, additionalConditions...)
	return l.unsafeList(ctx, dest, additionalConditions...)
}

// SetSelectedColumns missing godoc
func (l *universalLister) SetSelectedColumns(selectedColumns []string) {
	l.selectedColumns = strings.Join(selectedColumns, ", ")
}

// Clone missing godoc
func (l *universalLister) Clone() Lister {
	var clonedLister universalLister

	clonedLister.resourceType = l.resourceType
	clonedLister.tableName = l.tableName
	clonedLister.selectedColumns = l.selectedColumns
	clonedLister.tenantColumn = l.tenantColumn
	clonedLister.orderByParams = append(clonedLister.orderByParams, l.orderByParams...)

	return &clonedLister
}

// ListGlobal missing godoc
func (l *universalLister) ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error {
	return l.unsafeList(ctx, dest, additionalConditions...)
}

func (l *universalLister) unsafeList(ctx context.Context, dest Collection, conditions ...Condition) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(l.tableName, l.selectedColumns, conditions, l.orderByParams, true)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	err = persist.SelectContext(ctx, dest, query, args...)

	return persistence.MapSQLError(ctx, err, l.resourceType, resource.List, "while fetching list of objects from '%s' table", l.tableName)
}
