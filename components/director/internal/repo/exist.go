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

// ExistQuerier is an interface for checking existence of tenant scoped entities with either externally managed tenant accesses (m2m table or view) or embedded tenant in them.
type ExistQuerier interface {
	Exists(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) (bool, error)
}

// ExistQuerierGlobal is an interface for checking existence of global entities.
type ExistQuerierGlobal interface {
	ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error)
}

// ExistQuerierGlobalWithConditionTree is an interface for checking existence of global entities.
type ExistQuerierGlobalWithConditionTree interface {
	ExistsGlobalWithConditionTree(ctx context.Context, conditionTree *ConditionTree) (bool, error)
}

type universalExistQuerier struct {
	tableName    string
	tenantColumn *string
	resourceType resource.Type
	ownerCheck   bool
}

// NewExistQuerier is a constructor for ExistQuerier about entities with externally managed tenant accesses (m2m table or view)
func NewExistQuerier(tableName string) ExistQuerier {
	return &universalExistQuerier{tableName: tableName}
}

// NewExistQuerierWithOwnerCheck is a constructor for ExistQuerier about entities with externally managed tenant accesses (m2m table or view) with additional owner check.
func NewExistQuerierWithOwnerCheck(tableName string) ExistQuerier {
	return &universalExistQuerier{tableName: tableName, ownerCheck: true}
}

// NewExistQuerierWithEmbeddedTenant is a constructor for ExistQuerier about entities with tenant embedded in them.
func NewExistQuerierWithEmbeddedTenant(tableName string, tenantColumn string) ExistQuerier {
	return &universalExistQuerier{tableName: tableName, tenantColumn: &tenantColumn}
}

// NewExistQuerierGlobal is a constructor for ExistQuerierGlobal about global entities.
func NewExistQuerierGlobal(resourceType resource.Type, tableName string) ExistQuerierGlobal {
	return &universalExistQuerier{tableName: tableName, resourceType: resourceType}
}

// NewExistsQuerierGlobalWithConditionTree is a constructor for ExistQuerierGlobalWithConditionTree about global entities.
func NewExistsQuerierGlobalWithConditionTree(resourceType resource.Type, tableName string) ExistQuerierGlobalWithConditionTree {
	return &universalExistQuerier{tableName: tableName, resourceType: resourceType}
}

// Exists checks for existence of tenant scoped entities with tenant isolation subquery.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
// If the tenantColumn is not configured an entity with externally managed tenant accesses in m2m table / view is assumed.
func (g *universalExistQuerier) Exists(ctx context.Context, resourceType resource.Type, tenant string, conditions Conditions) (bool, error) {
	if tenant == "" {
		return false, apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.exists(ctx, resourceType, conditions)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenant, g.ownerCheck)
	if err != nil {
		return false, err
	}

	conditions = append(conditions, tenantIsolation)

	return g.exists(ctx, resourceType, conditions)
}

// ExistsGlobal checks for existence of global entities without tenant isolation.
func (g *universalExistQuerier) ExistsGlobal(ctx context.Context, conditions Conditions) (bool, error) {
	return g.exists(ctx, g.resourceType, conditions)
}

func (g *universalExistQuerier) exists(ctx context.Context, resourceType resource.Type, conditions Conditions) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT 1 FROM %s", g.tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return false, errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	query := getQueryFromBuilder(stmtBuilder)

	log.C(ctx).Debugf("Executing DB query: %s", query)
	var count int
	err = persist.GetContext(ctx, &count, query, allArgs...)
	err = persistence.MapSQLError(ctx, err, resourceType, resource.Exists, "while getting object from '%s' table", g.tableName)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ExistsGlobalWithConditionTree checks for existence of global entities without tenant isolation.
func (g *universalExistQuerier) ExistsGlobalWithConditionTree(ctx context.Context, conditionTree *ConditionTree) (bool, error) {
	return g.existsWithConditionTree(ctx, g.resourceType, conditionTree)
}

func (g *universalExistQuerier) existsWithConditionTree(ctx context.Context, resourceType resource.Type, conditionTree *ConditionTree) (bool, error) {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return false, err
	}

	query, args, err := buildSelectQueryFromTree(g.tableName, "1", conditionTree, NoOrderBy, NoLock, true)
	if err != nil {
		return false, errors.Wrap(err, "while building exist query")
	}

	log.C(ctx).Debugf("Executing DB query: %s", query)
	var count int
	err = persist.GetContext(ctx, &count, query, args...)
	err = persistence.MapSQLError(ctx, err, resourceType, resource.Exists, "while getting object from '%s' table", g.tableName)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
