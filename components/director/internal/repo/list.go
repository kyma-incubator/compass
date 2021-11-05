package repo

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Lister missing godoc
type Lister interface {
	List(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error
	SetSelectedColumns(selectedColumns []string)
	Clone() *universalLister
}

// ListerGlobal missing godoc
type ListerGlobal interface {
	ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error
	SetSelectedColumns(selectedColumns []string)
	Clone() *universalLister
}

type universalLister struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type

	orderByParams OrderByParams
}

// NewListerWithEmbeddedTenant missing godoc
func NewListerWithEmbeddedTenant(tableName string, tenantColumn string, selectedColumns []string) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		orderByParams:   NoOrderBy,
	}
}

// NewLister missing godoc
func NewLister(tableName string, selectedColumns []string) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   NoOrderBy,
	}
}

// NewListerWithOrderBy missing godoc
func NewListerWithOrderBy(tableName string, selectedColumns []string, orderByParams OrderByParams) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
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
func (l *universalLister) List(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if l.tenantColumn != nil {
		additionalConditions = append(Conditions{NewEqualCondition(*l.tenantColumn, tenant)}, additionalConditions...)
		return l.unsafeList(ctx, tenant, resourceType, dest, additionalConditions...)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, false)
	if err != nil {
		return err
	}

	additionalConditions = append(additionalConditions, tenantIsolation)

	return l.unsafeList(ctx, tenant, resourceType, dest, additionalConditions...)
}

// SetSelectedColumns missing godoc
func (l *universalLister) SetSelectedColumns(selectedColumns []string) {
	l.selectedColumns = strings.Join(selectedColumns, ", ")
}

// Clone missing godoc
func (l *universalLister) Clone() *universalLister {
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
	return l.unsafeList(ctx, "", l.resourceType, dest, additionalConditions...)
}

func (l *universalLister) unsafeList(ctx context.Context, tenant string, resourceType resource.Type, dest Collection, conditions ...Condition) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(resourceType, l.tableName, l.selectedColumns, tenant, conditions, l.orderByParams, true)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	err = persist.SelectContext(ctx, dest, query, args...)

	return persistence.MapSQLError(ctx, err, resourceType, resource.List, "while fetching list of objects from '%s' table", l.tableName)
}
