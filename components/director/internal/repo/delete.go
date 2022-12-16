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

type universalDeleter struct {
	tableName    string
	resourceType resource.Type
	tenantColumn *string
}

// NewDeleter is a constructor for Deleter about entities with externally managed tenant accesses (m2m table or view)
func NewDeleter(tableName string) Deleter {
	return &universalDeleter{tableName: tableName}
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

	isTenantScopedUpdate := g.tenantColumn != nil

	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if (affected == 0 && hasTenantIsolationCondition(conditions)) || (affected == 0 && isTenantScopedUpdate) {
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
