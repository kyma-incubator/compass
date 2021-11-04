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

// Deleter missing godoc
type Deleter interface {
	DeleteOne(ctx context.Context, tenant string, conditions Conditions) error
	DeleteMany(ctx context.Context, tenant string, conditions Conditions) error
}

// DeleterGlobal missing godoc
type DeleterGlobal interface {
	DeleteOneGlobal(ctx context.Context, conditions Conditions) error
	DeleteManyGlobal(ctx context.Context, conditions Conditions) error
}

type universalDeleter struct {
	tableName    string
	resourceType resource.Type
	tenantColumn *string
	isGlobal     bool
}

// NewDeleter missing godoc
func NewDeleter(resourceType resource.Type, tableName string) Deleter {
	return &universalDeleter{resourceType: resourceType, tableName: tableName, isGlobal: false}
}

// NewDeleterWithEmbeddedTenant missing godoc
func NewDeleterWithEmbeddedTenant(resourceType resource.Type, tableName string, tenantColumn string) Deleter {
	return &universalDeleter{resourceType: resourceType, tableName: tableName, isGlobal: false, tenantColumn: &tenantColumn}
}

// NewDeleterGlobal missing godoc
func NewDeleterGlobal(resourceType resource.Type, tableName string) DeleterGlobal {
	return &universalDeleter{tableName: tableName, resourceType: resourceType, isGlobal: true}
}

// DeleteOne missing godoc
func (g *universalDeleter) DeleteOne(ctx context.Context, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.unsafeDelete(ctx, conditions, true)
	}

	if g.resourceType.IsTopLevel() {
		return g.unsafeDeleteTenantAccess(ctx, tenant, conditions, true)
	}

	return g.unsafeDeleteChildEntity(ctx, tenant, conditions, true)
}

// DeleteMany missing godoc
func (g *universalDeleter) DeleteMany(ctx context.Context, tenant string, conditions Conditions) error {
	if tenant == "" {
		return apperrors.NewTenantRequiredError()
	}

	if g.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*g.tenantColumn, tenant)}, conditions...)
		return g.unsafeDelete(ctx, conditions, false)
	}

	if g.resourceType.IsTopLevel() {
		return g.unsafeDeleteTenantAccess(ctx, tenant, conditions, false)
	}

	return g.unsafeDeleteChildEntity(ctx, tenant, conditions, false)
}

// DeleteOneGlobal missing godoc
func (g *universalDeleter) DeleteOneGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, conditions, true)
}

// DeleteManyGlobal missing godoc
func (g *universalDeleter) DeleteManyGlobal(ctx context.Context, conditions Conditions) error {
	return g.unsafeDelete(ctx, conditions, false)
}

func (g *universalDeleter) unsafeDelete(ctx context.Context, conditions Conditions, requireSingleRemoval bool) error {
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
	if err = persistence.MapSQLError(ctx, err, g.resourceType, resource.Delete, "while deleting object from '%s' table", g.tableName); err != nil {
		return err
	}

	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if affected != 1 {
			return apperrors.NewInternalError("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}

func (g *universalDeleter) unsafeDeleteTenantAccess(ctx context.Context, tenant string, conditions Conditions, requireSingleRemoval bool) error {
	m2mTable, ok := g.resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %s does not have access table", g.resourceType)
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT id FROM %s WHERE", g.tableName))

	tenantIsolationSubquery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? AND %s = true", M2MResourceIDColumn, m2mTable, M2MTenantIDColumn, M2MOwnerColumn)
	conditions = append(conditions, NewInConditionForSubQuery("id", tenantIsolationSubquery, []interface{}{tenant}))

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	query := getQueryFromBuilder(stmtBuilder)
	log.C(ctx).Debugf("Executing DB query: %s", query)

	var ids IDs
	if requireSingleRemoval {
		var id string
		err = persist.GetContext(ctx, &id, query, allArgs...)
		ids = append(ids, id)
	} else {
		err = persist.SelectContext(ctx, &ids, query, allArgs...)
	}
	if err = persistence.MapSQLError(ctx, err, g.resourceType, resource.Delete, "while selecting objects from '%s' table by conditions", g.tableName); err != nil {
		return err
	}

	stmtBuilder.Reset()

	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s WHERE", m2mTable))

	deleteConditions := Conditions{NewEqualCondition(M2MTenantIDColumn, tenant), NewInConditionForStringValues(M2MResourceIDColumn, ids), NewEqualCondition(M2MOwnerColumn, true)}
	err = writeEnumeratedConditions(&stmtBuilder, deleteConditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}

	allArgs = getAllArgs(deleteConditions)

	query = getQueryFromBuilder(stmtBuilder)
	log.C(ctx).Debugf("Executing DB query: %s", query)

	_, err = persist.ExecContext(ctx, query, allArgs...)
	return persistence.MapSQLError(ctx, err, g.resourceType, resource.Delete, "while deleting object from '%s' table", m2mTable)
}

func (g *universalDeleter) unsafeDeleteChildEntity(ctx context.Context, tenant string, conditions Conditions, requireSingleRemoval bool) error {
	m2mTable, ok := g.resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %s does not have access table", g.resourceType)
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(fmt.Sprintf("DELETE FROM %s WHERE", g.tableName))

	tenantIsolationSubquery := fmt.Sprintf("SELECT %s FROM %s WHERE %s = ? AND %s = true", M2MResourceIDColumn, m2mTable, M2MTenantIDColumn, M2MOwnerColumn)
	conditions = append(conditions, NewInConditionForSubQuery("id", tenantIsolationSubquery, []interface{}{tenant}))

	err = writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return errors.Wrap(err, "while writing enumerated conditions")
	}
	allArgs := getAllArgs(conditions)

	if g.resourceType == resource.BundleInstanceAuth {
		stmtBuilder.WriteString(" OR owner_id = ?")
		allArgs = append(allArgs, tenant)
	}

	query := getQueryFromBuilder(stmtBuilder)
	log.C(ctx).Debugf("Executing DB query: %s", query)

	res, err := persist.ExecContext(ctx, query, allArgs...)
	if err = persistence.MapSQLError(ctx, err, g.resourceType, resource.Delete, "while deleting object from '%s' table", g.tableName); err != nil {
		return err
	}

	if requireSingleRemoval {
		affected, err := res.RowsAffected()
		if err != nil {
			return errors.Wrap(err, "while checking affected rows")
		}
		if affected != 1 {
			return apperrors.NewInternalError("delete should remove single row, but removed %d rows", affected)
		}
	}

	return nil
}

// IDs keeps IDs retrieved from the Compass storage.
type IDs []string

// Len returns the length of the IDs
func (i IDs) Len() int {
	return len(i)
}
