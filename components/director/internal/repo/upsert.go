package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// UpserterGlobal is an interface for upserting global entities without tenant or entities with tenant embedded in them.
type UpserterGlobal interface {
	UpsertGlobal(ctx context.Context, dbEntity interface{}) error
}

// Upserter is an interface for upserting entities with externally managed tenant accesses (m2m table or view)
type Upserter interface {
	Upsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
}

type upserter struct {
	tableName          string
	insertColumns      []string
	conflictingColumns []string
	updateColumns      []string
}

type upserterGlobal struct {
	tableName          string
	resourceType       resource.Type
	tenantColumn       *string
	insertColumns      []string
	conflictingColumns []string
	updateColumns      []string
}

// NewUpserter is a constructor for Upserter about entities with externally managed tenant accesses (m2m table or view)
func NewUpserter(tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) Upserter {
	return &upserter{
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// NewUpserterGlobal is a constructor for UpserterGlobal about global entities without tenant.
func NewUpserterGlobal(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) UpserterGlobal {
	return &upserterGlobal{
		resourceType:       resourceType,
		tableName:          tableName,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// NewUpserterWithEmbeddedTenant is a constructor for Upserter about entities with tenant embedded in them.
func NewUpserterWithEmbeddedTenant(resourceType resource.Type, tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string, tenantColumn string) UpserterGlobal {
	return &upserterGlobal{
		resourceType:       resourceType,
		tableName:          tableName,
		tenantColumn:       &tenantColumn,
		insertColumns:      insertColumns,
		conflictingColumns: conflictingColumns,
		updateColumns:      updateColumns,
	}
}

// Upsert adds a new entity in the Compass DB in case it does not exist. If it already exists, updates it.
// This upserter is suitable for resources that have m2m tenant relation as it does maintain tenant accesses.
func (u *upserter) Upsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	return u.unsafeUpsert(ctx, resourceType, tenant, dbEntity)
}

func (u *upserter) unsafeUpsert(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query := buildQuery(u.tableName, u.insertColumns, u.conflictingColumns, u.updateColumns)
	queryWithTenantIsolation, err := u.addTenantIsolation(query, resourceType, tenant)
	if err != nil {
		return err
	}

	if entityWithExternalTenant, ok := dbEntity.(EntityWithExternalTenant); ok {
		dbEntity = entityWithExternalTenant.DecorateWithTenantID(tenant)
	}

	log.C(ctx).Debugf("Executing DB query: %s", queryWithTenantIsolation)
	res, err := persist.NamedExecContext(ctx, queryWithTenantIsolation, dbEntity)
	if err = persistence.MapSQLError(ctx, err, resourceType, resource.Upsert, "while upserting row to '%s' table", u.tableName); err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}

	if err = assertSingleRowAffectedWhenUpserting(affected, true); err != nil {
		return err
	}

	if resourceType.IsTopLevel() {
		if err = u.upsertTenantAccess(ctx, resourceType, dbEntity, tenant); err != nil {
			return err
		}
	}

	return nil
}

func (u *upserter) upsertTenantAccess(ctx context.Context, resourceType resource.Type, dbEntity interface{}, tenant string) error {
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %s does not have access table", resourceType)
	}

	var id string
	if identifiable, ok := dbEntity.(Identifiable); ok {
		id = identifiable.GetID()
	}

	if len(id) == 0 {
		return apperrors.NewInternalError("id cannot be empty, check if the entity implements Identifiable")
	}

	return UpsertTenantAccessRecursively(ctx, m2mTable, &TenantAccess{
		TenantID:   tenant,
		ResourceID: id,
		Owner:      true,
	})
}

func (u *upserter) addTenantIsolation(query string, resourceType resource.Type, tenant string) (string, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(query)

	tenantIsolationCondition, err := NewTenantIsolationConditionForNamedArgs(resourceType, tenant, true)
	if err != nil {
		return "", err
	}

	tenantIsolationStatement := strings.Replace(tenantIsolationCondition.GetQueryPart(), "(", fmt.Sprintf("(%s.", u.tableName), 1)
	stmtBuilder.WriteString(" WHERE ")
	stmtBuilder.WriteString(tenantIsolationStatement)

	return stmtBuilder.String(), nil
}

// UpsertGlobal adds a new entity in the Compass DB in case it does not exist. If it already exists, updates it.
// This upserter is not suitable for resources that have m2m tenant relation as it does not maintain tenant accesses.
// Use it for global scoped resources or resources with embedded tenant_id only.
func (u *upserterGlobal) UpsertGlobal(ctx context.Context, dbEntity interface{}) error {
	return u.unsafeUpsert(ctx, u.resourceType, dbEntity)
}

func (u *upserterGlobal) unsafeUpsert(ctx context.Context, resourceType resource.Type, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	query := buildQuery(u.tableName, u.insertColumns, u.conflictingColumns, u.updateColumns)
	if u.tenantColumn != nil {
		query = u.addTenantIsolation(query)
	}

	log.C(ctx).Warnf("Executing DB query: %s", query)
	res, err := persist.NamedExecContext(ctx, query, dbEntity)
	if err = persistence.MapSQLError(ctx, err, resourceType, resource.Upsert, "while upserting row to '%s' table", u.tableName); err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}

	isTenantScopedUpsert := u.tenantColumn != nil
	return assertSingleRowAffectedWhenUpserting(affected, isTenantScopedUpsert)
}

func (u *upserterGlobal) addTenantIsolation(query string) string {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(query)

	stmtBuilder.WriteString(" WHERE ")
	stmtBuilder.WriteString(fmt.Sprintf(" %s.%s = :%s", u.tableName, *u.tenantColumn, *u.tenantColumn))

	return stmtBuilder.String()
}

func buildQuery(tableName string, insertColumns []string, conflictingColumns []string, updateColumns []string) string {
	var stmtBuilder strings.Builder

	values := make([]string, 0, len(insertColumns))
	for _, c := range insertColumns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	update := make([]string, 0, len(updateColumns))
	for _, c := range updateColumns {
		update = append(update, fmt.Sprintf("%[1]s=EXCLUDED.%[1]s", c))
	}
	stmtWithoutUpsert := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", tableName, strings.Join(insertColumns, ", "), strings.Join(values, ", "))
	stmtWithUpsert := fmt.Sprintf("%s ON CONFLICT ( %s ) DO UPDATE SET %s", stmtWithoutUpsert, strings.Join(conflictingColumns, ", "), strings.Join(update, ", "))

	stmtBuilder.WriteString(stmtWithUpsert)
	return stmtBuilder.String()
}

func assertSingleRowAffectedWhenUpserting(affected int64, isTenantScopedUpsert bool) error {
	if affected == 0 && isTenantScopedUpsert {
		return apperrors.NewUnauthorizedError(apperrors.ShouldBeOwnerMsg)
	}

	if affected != 1 {
		return apperrors.NewInternalError(apperrors.ShouldUpsertSingleRowButUpsertedMsgF, affected)
	}
	return nil
}
