package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// UpserterGlobal is an interface for upserting global entities without tenant or entities with tenant embedded in them.
type UpserterGlobal interface {
	UpsertGlobal(ctx context.Context, dbEntity interface{}) error
}

// Upserter is an interface for upderting entities with externally managed tenant accesses (m2m table or view)
type Upserter interface {
	Upsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
}

type universalUpserter struct {
	tableName          string
	resourceType       resource.Type
	insertColumns      []string
	conflictingColumns []string
	updateColumns      []string
	tenantColumn       *string
}

// NewUpserter is a constructor for Upserter about entities with externally managed tenant accesses (m2m table or view)
func NewUpserter(tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) Upserter {
	return &universalUpserter{
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// NewUpserterGlobal is a constructor for UpdaterGlobal about global entities without tenant.
func NewUpserterGlobal(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) UpserterGlobal {
	return &universalUpserter{
		resourceType:       resourceType,
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// NewUpserterWithEmbeddedTenant is a constructor for Upserter about entities with tenant embedded in them.
func NewUpserterWithEmbeddedTenant(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string, tenantColumn string) Upserter {
	return &universalUpserter{
		resourceType:       resourceType,
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
		tenantColumn:       &tenantColumn,
	}
}

// Upsert adds a new entity in the Compass DB in case it does not exist. If it already exists, updates it.
// This upserter is not suitable for resources that have m2m tenant relation as it does not maintain tenant accesses.
// Use it for global scoped resources or resources with embedded tenant_id only.
func (u *universalUpserter) Upsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	return u.unsafeUpsert(ctx, resourceType, tenant, dbEntity)
}

func (u *universalUpserter) UpsertGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpsert(ctx, u.resourceType, "", dbEntity)
}

func (u *universalUpserter) buildQuery() string {
	var stmtBuilder strings.Builder

	values := make([]string, 0, len(u.insertColumns))
	for _, c := range u.insertColumns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	update := make([]string, 0, len(u.updateColumns))
	for _, c := range u.updateColumns {
		update = append(update, fmt.Sprintf("%[1]s=EXCLUDED.%[1]s", c))
	}
	stmtWithoutUpsert := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", u.tableName, strings.Join(u.insertColumns, ", "), strings.Join(values, ", "))
	stmtWithUpsert := fmt.Sprintf("%s ON CONFLICT ( %s ) DO UPDATE SET %s", stmtWithoutUpsert, strings.Join(u.conflictingColumns, ", "), strings.Join(update, ", "))

	stmtBuilder.WriteString(stmtWithUpsert)
	return stmtBuilder.String()
}

func (u *universalUpserter) addTenantIsolation(query string, resourceType resource.Type, tenant string) (string, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(query)
	if u.tenantColumn != nil { // if embedded tenant
		stmtBuilder.WriteString(" WHERE ")
		stmtBuilder.WriteString(fmt.Sprintf(" %s.%s = :%s", u.tableName, *u.tenantColumn, *u.tenantColumn))
	} else if len(tenant) > 0 { // if not global
		tenantIsolationCondition, err := NewTenantIsolationConditionForNamedArgs(resourceType, tenant, true)
		if err != nil {
			return "", err
		}

		tenantIsolationStatement := strings.Replace(tenantIsolationCondition.GetQueryPart(), "(", fmt.Sprintf("(%s.", u.tableName), 1)
		stmtBuilder.WriteString(" WHERE ")
		stmtBuilder.WriteString(tenantIsolationStatement)
	}

	return stmtBuilder.String(), nil
}

func (u *universalUpserter) unsafeUpsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query := u.buildQuery()
	queryWithTenantIsolation, err := u.addTenantIsolation(query, resourceType, tenant)
	if err != nil {
		return err
	}

	if entityWithExternalTenant, ok := dbEntity.(EntityWithExternalTenant); ok && (u.tenantColumn == nil && len(tenant) > 0) {
		dbEntity = entityWithExternalTenant.DecorateWithTenantID(tenant)
	}

	log.C(ctx).Warnf("Executing DB query: %s", queryWithTenantIsolation)
	_, err = persist.NamedExecContext(ctx, queryWithTenantIsolation, dbEntity)
	return persistence.MapSQLError(ctx, err, u.resourceType, resource.Upsert, "while upserting row to '%s' table", u.tableName)
}
