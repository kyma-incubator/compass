package repo

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Lister is an interface for listing tenant scoped entities with either externally managed tenant accesses (m2m table or view) or embedded tenant in them.
type Lister interface {
	List(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error
	ListWithSelectForUpdate(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error
	SetSelectedColumns(selectedColumns []string)
	Clone() *universalLister
}

// ListerGlobal is an interface for listing global entities.
type ListerGlobal interface {
	ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error
	ListGlobalWithSelectForUpdate(ctx context.Context, dest Collection, additionalConditions ...Condition) error
	SetSelectedColumns(selectedColumns []string)
	Clone() *universalLister
}

type universalLister struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
	ownerCheck      bool

	orderByParams OrderByParams
}

// NewListerWithEmbeddedTenant is a constructor for Lister about entities with tenant embedded in them.
func NewListerWithEmbeddedTenant(tableName string, tenantColumn string, selectedColumns []string) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		orderByParams:   NoOrderBy,
	}
}

// NewLister is a constructor for Lister about entities with externally managed tenant accesses (m2m table or view)
func NewLister(tableName string, selectedColumns []string) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   NoOrderBy,
	}
}

// NewOwnerLister is a constructor for Lister about entities with externally managed tenant accesses (m2m table or view) with owner access check
func NewOwnerLister(tableName string, selectedColumns []string, ownerCheck bool) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		ownerCheck:      ownerCheck,
		orderByParams:   NoOrderBy,
	}
}

// NewListerWithOrderBy is a constructor for Lister about entities with externally managed tenant accesses (m2m table or view) with additional order by clause.
func NewListerWithOrderBy(tableName string, selectedColumns []string, orderByParams OrderByParams) Lister {
	return &universalLister{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   orderByParams,
	}
}

// NewListerGlobalWithOrderBy is a constructor for ListerGlobal about global entities with additional order by clause.
func NewListerGlobalWithOrderBy(resourceType resource.Type, tableName string, selectedColumns []string, orderByParams OrderByParams) ListerGlobal {
	return &universalLister{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   orderByParams,
	}
}

// NewListerGlobal is a constructor for ListerGlobal about global entities.
func NewListerGlobal(resourceType resource.Type, tableName string, selectedColumns []string) ListerGlobal {
	return &universalLister{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		orderByParams:   NoOrderBy,
	}
}

// List lists tenant scoped entities with tenant isolation subquery.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
// If the tenantColumn is not configured an entity with externally managed tenant accesses in m2m table / view is assumed.
func (l *universalLister) List(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error {
	return l.listWithTenantScope(ctx, resourceType, tenant, dest, NoLock, additionalConditions)
}

// ListWithSelectForUpdate lists tenant scoped entities with tenant isolation subquery and
func (l *universalLister) ListWithSelectForUpdate(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, additionalConditions ...Condition) error {
	return l.listWithTenantScope(ctx, resourceType, tenant, dest, ForUpdateLock, additionalConditions)
}

func (l *universalLister) listWithTenantScope(ctx context.Context, resourceType resource.Type, tenant string, dest Collection, lockClause string, additionalConditions []Condition) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if l.tenantColumn != nil {
		additionalConditions = append(Conditions{NewEqualCondition(*l.tenantColumn, tenant)}, additionalConditions...)
		return l.list(ctx, resourceType, dest, lockClause, additionalConditions...)
	}

	tenantIsolation, err :=
		newTenantIsolationConditionWithPlaceholder(resourceType, tenant, l.ownerCheck, true)
	if err != nil {
		return err
	}

	additionalConditions = append(additionalConditions, tenantIsolation)

	return l.list(ctx, resourceType, dest, lockClause, additionalConditions...)
}

// SetSelectedColumns sets the selected columns for the query.
func (l *universalLister) SetSelectedColumns(selectedColumns []string) {
	l.selectedColumns = strings.Join(selectedColumns, ", ")
}

// Clone creates a new instance of the lister with the same configuration.
func (l *universalLister) Clone() *universalLister {
	var clonedLister universalLister

	clonedLister.resourceType = l.resourceType
	clonedLister.tableName = l.tableName
	clonedLister.selectedColumns = l.selectedColumns
	clonedLister.tenantColumn = l.tenantColumn
	clonedLister.orderByParams = append(clonedLister.orderByParams, l.orderByParams...)

	return &clonedLister
}

// ListGlobal lists global entities without tenant isolation.
func (l *universalLister) ListGlobal(ctx context.Context, dest Collection, additionalConditions ...Condition) error {
	return l.list(ctx, l.resourceType, dest, NoLock, additionalConditions...)
}

// ListGlobalWithSelectForUpdate lists global entities without tenant isolation.
func (l *universalLister) ListGlobalWithSelectForUpdate(ctx context.Context, dest Collection, additionalConditions ...Condition) error {
	return l.list(ctx, l.resourceType, dest, ForUpdateLock, additionalConditions...)
}

func (l *universalLister) list(ctx context.Context, resourceType resource.Type, dest Collection, lockClause string, conditions ...Condition) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args, err := buildSelectQuery(l.tableName, l.selectedColumns, conditions, l.orderByParams, lockClause, true)
	if err != nil {
		return errors.Wrap(err, "while building list query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	err = persist.SelectContext(ctx, dest, query, args...)

	return persistence.MapSQLError(ctx, err, resourceType, resource.List, "while fetching list of objects from '%s' table", l.tableName)
}
