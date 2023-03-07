package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/pkg/errors"
)

// Deleter is an interface for deleting tenant scoped entities with either externally managed tenant accesses (m2m table or view) or embedded tenant in them.
type Deleter interface {
	DeleteOne(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) error
	DeleteMany(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) error
}

// DeleterGlobal is an interface for deleting global entities.
type DeleterGlobal interface {
	DeleteOneGlobal(ctx context.Context, conditions Conditions) error
	DeleteManyGlobal(ctx context.Context, conditions Conditions) error
}

// DeleteConditionTree deletes tenant scoped entities matching the provided condition tree with tenant isolation.
type DeleteConditionTree interface {
	DeleteConditionTree(ctx context.Context, resourceType resource.Type, tenant string, conditionTree *ConditionTree) error
}

type universalDeleter struct {
	tableName    string
	resourceType resource.Type
	tenantColumn *string
}

// NewDeleter is a constructor for Deleter about entities with externally managed tenant accesses (m2m table or view)
func NewDeleter(tableName string) Deleter {
	return &universalDeleter{tableName: tableName}
}

// NewDeleteConditionTreeWithEmbeddedTenant is a constructor for Deleter about entities with externally managed tenant accesses (m2m table or view)
func NewDeleteConditionTreeWithEmbeddedTenant(tableName string, tenantColumn string) DeleteConditionTree {
	return &universalDeleter{tableName: tableName, tenantColumn: &tenantColumn}
}

// NewDeleterWithEmbeddedTenant is a constructor for Deleter about entities with tenant embedded in them.
func NewDeleterWithEmbeddedTenant(tableName string, tenantColumn string) Deleter {
	return &universalDeleter{tableName: tableName, tenantColumn: &tenantColumn}
}

// NewDeleterGlobal is a constructor for DeleterGlobal about global entities.
func NewDeleterGlobal(resourceType resource.Type, tableName string) DeleterGlobal {
	return &universalDeleter{tableName: tableName, resourceType: resourceType}
}

// DeleteOne deletes exactly one entity from the database if the calling tenant has owner access to it. It returns an error if more than one entity matches the provided conditions.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
func (g *universalDeleter) DeleteOne(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.delete(ctx, resourceType, conditions, true)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, true)
	if err != nil {
		return err
	}

	conditions = append(conditions, tenantIsolation)

	return g.delete(ctx, resourceType, conditions, true)
}

// DeleteMany deletes all the entities that match the provided conditions from the database if the calling tenant has owner access to them.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
func (g *universalDeleter) DeleteMany(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.delete(ctx, resourceType, conditions, false)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, true)
	if err != nil {
		return err
	}

	conditions = append(conditions, tenantIsolation)

	return g.delete(ctx, resourceType, conditions, false)
}

// DeleteOneGlobal deletes exactly one entity from the database. It returns an error if more than one entity matches the provided conditions.
func (g *universalDeleter) DeleteOneGlobal(ctx context.Context, conditions Conditions) error {
	return g.delete(ctx, g.resourceType, conditions, true)
}

// DeleteManyGlobal deletes all the entities that match the provided conditions from the database.
func (g *universalDeleter) DeleteManyGlobal(ctx context.Context, conditions Conditions) error {
	return g.delete(ctx, g.resourceType, conditions, false)
}

// DeleteConditionTree lists tenant scoped entities matching the provided condition tree with tenant isolation.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
// If the tenantColumn is not configured an entity with externally managed tenant accesses in m2m table / view is assumed.
func (g *universalDeleter) DeleteConditionTree(ctx context.Context, resourceType resource.Type, tenant string, conditionTree *ConditionTree) error {
	return g.treeConditionDeleterWithTenantScope(ctx, resourceType, tenant, NoLock, conditionTree)
}

func (g *universalDeleter) treeConditionDeleterWithTenantScope(ctx context.Context, resourceType resource.Type, tenant string, lockClause string, conditionTree *ConditionTree) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions := And(&ConditionTree{Operand: NewEqualCondition(*g.tenantColumn, tenant)}, conditionTree)
		return g.deleteWithConditionTree(ctx, resourceType, lockClause, conditions)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, true)
	if err != nil {
		return err
	}

	conditions := And(&ConditionTree{Operand: tenantIsolation}, conditionTree)

	return g.deleteWithConditionTree(ctx, resourceType, lockClause, conditions)
}

func (g *universalDeleter) deleteWithConditionTree(ctx context.Context, resourceType resource.Type, lockClause string, conditionTree *ConditionTree) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query, args := buildDeleteQueryFromTree(g.tableName, conditionTree, lockClause, true)

	log.C(ctx).Debugf("Executing DB query: %s", query)
	_, err = persist.ExecContext(ctx, query, args...)

	return persistence.MapSQLError(ctx, err, resourceType, resource.List, "while fetching list of objects from '%s' table", g.tableName)
}

func buildDeleteQueryFromTree(tableName string, conditions *ConditionTree, lockClause string, isRebindingNeeded bool) (string, []interface{}) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s", tableName))
	var allArgs []interface{}
	if conditions != nil {
		stmtBuilder.WriteString(" WHERE ")
		var subquery string
		subquery, allArgs = conditions.BuildSubquery()
		stmtBuilder.WriteString(subquery)
	}

	writeLockClause(&stmtBuilder, lockClause)

	if isRebindingNeeded {
		return getQueryFromBuilder(stmtBuilder), allArgs
	}
	return stmtBuilder.String(), allArgs
}

func (g *universalDeleter) delete(ctx context.Context, resourceType resource.Type, conditions Conditions, requireSingleRemoval bool) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s", g.tableName))

	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	query := getQueryFromBuilder(stmtBuilder)
	log.C(ctx).Debugf("Executing DB query: %s", query)
	res, err := persist.ExecContext(ctx, query, allArgs...)
	if err = persistence.MapSQLError(ctx, err, resourceType, resource.Delete, "while deleting object from '%s' table", g.tableName); err != nil {
		return err
	}

	isTenantScopedDelete := g.tenantColumn != nil

	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if (affected == 0 && hasTenantIsolationCondition(conditions)) || (affected == 0 && isTenantScopedDelete) {
			return apperrors.NewUnauthorizedError(apperrors.ShouldBeOwnerMsg)
		}
		if affected != 1 {
			return apperrors.NewInternalError("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}

func hasTenantIsolationCondition(conditions Conditions) bool {
	for _, cond := range conditions {
		if _, ok := cond.(*tenantIsolationCondition); ok {
			return true
		}
	}
	return false
}

// IDs keeps IDs retrieved from the Compass storage.
type IDs []string

// Len returns the length of the IDs
func (i IDs) Len() int {
	return len(i)
}
