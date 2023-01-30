package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

const (
	// M2MTenantIDColumn is the column name of the tenant_id in each tenant access table / view.
	M2MTenantIDColumn = "tenant_id"
	// M2MResourceIDColumn is the column name of the resource id in each tenant access table / view.
	M2MResourceIDColumn = "id"
	// M2MOwnerColumn is the column name of the owner in each tenant access table / view.
	M2MOwnerColumn = "owner"

	// CreateSingleTenantAccessQuery sda
	CreateSingleTenantAccessQuery = `INSERT INTO %s ( %s ) VALUES ( %s ) ON CONFLICT ON CONSTRAINT tenant_applications_pkey DO NOTHING`

	// RecursiveCreateTenantAccessCTEQuery is a recursive SQL query that creates a tenant access record for a tenant and all its parents.
	RecursiveCreateTenantAccessCTEQuery = `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.parent
                    FROM business_tenant_mappings t1
                    WHERE id = :tenant_id
                    UNION ALL
                    SELECT t2.id, t2.parent
                    FROM business_tenant_mappings t2
                             INNER JOIN parents t on t2.id = t.parent)
			INSERT INTO %s ( %s ) (SELECT parents.id AS tenant_id, :id as id, :owner AS owner FROM parents)`

	// RecursiveUpsertTenantAccessCTEQuery is a recursive SQL query that creates a tenant access record for a tenant and all its parents.
	RecursiveUpsertTenantAccessCTEQuery = RecursiveCreateTenantAccessCTEQuery + " ON CONFLICT ( tenant_id, id ) DO NOTHING"

	// RecursiveDeleteTenantAccessCTEQuery is a recursive SQL query that deletes tenant accesses based on given conditions for a tenant and all its parents.
	RecursiveDeleteTenantAccessCTEQuery = `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.parent
                    FROM business_tenant_mappings t1
                    WHERE id = ?
                    UNION ALL
                    SELECT t2.id, t2.parent
                    FROM business_tenant_mappings t2
                             INNER JOIN parents t on t2.id = t.parent)
			DELETE FROM %s WHERE %s AND owner = true AND tenant_id IN (SELECT id FROM parents)`
)

// M2MColumns are the column names of the tenant access tables / views.
var M2MColumns = []string{M2MTenantIDColumn, M2MResourceIDColumn, M2MOwnerColumn}

// TenantAccess represents the tenant access table/views that are used for tenant isolation queries.
type TenantAccess struct {
	TenantID   string `db:"tenant_id"`
	ResourceID string `db:"id"`
	Owner      bool   `db:"owner"`
}

// TenantAccessCollection is a wrapper type for slice of entities.
type TenantAccessCollection []TenantAccess

// Len returns the current number of entities in the collection.
func (tc TenantAccessCollection) Len() int {
	return len(tc)
}

// CreateSingleTenantAccess create a tenant access for a single entity
func CreateSingleTenantAccess(ctx context.Context, m2mTable string, tenantAccess *TenantAccess) error {
	values := make([]string, 0, len(M2MColumns))
	for _, c := range M2MColumns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertTenantAccessStmt := fmt.Sprintf(CreateSingleTenantAccessQuery, m2mTable, strings.Join(M2MColumns, ", "), strings.Join(values, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, *tenantAccess)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Create, "while inserting tenant access record to '%s' table", m2mTable)
}

// CreateTenantAccessRecursively creates the given tenantAccess in the provided m2mTable while making sure to recursively
// add it to all the parents of the given tenant.
func CreateTenantAccessRecursively(ctx context.Context, m2mTable string, tenantAccess *TenantAccess) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertTenantAccessStmt := fmt.Sprintf(RecursiveCreateTenantAccessCTEQuery, m2mTable, strings.Join(M2MColumns, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, tenantAccess)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Create, "while inserting tenant access record to '%s' table", m2mTable)
}

// UpsertTenantAccessRecursively upserts the given tenantAccess in the provided m2mTable while making sure to recursively
// add it to all the parents of the given tenant.
func UpsertTenantAccessRecursively(ctx context.Context, m2mTable string, tenantAccess *TenantAccess) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertTenantAccessStmt := fmt.Sprintf(RecursiveUpsertTenantAccessCTEQuery, m2mTable, strings.Join(M2MColumns, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, tenantAccess)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Create, "while upserting tenant access record to '%s' table", m2mTable)
}

// DeleteTenantAccessRecursively deletes all the accesses to the provided resource IDs for the given tenant and all its parents.
func DeleteTenantAccessRecursively(ctx context.Context, m2mTable string, tenant string, resourceIDs []string) error {
	if len(resourceIDs) == 0 {
		return errors.New("resourceIDs cannot be empty")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(resourceIDs)+1)
	args = append(args, tenant)

	inCond := NewInConditionForStringValues(M2MResourceIDColumn, resourceIDs)
	if inArgs, ok := inCond.GetQueryArgs(); ok {
		args = append(args, inArgs...)
	}

	deleteTenantAccessStmt := fmt.Sprintf(RecursiveDeleteTenantAccessCTEQuery, m2mTable, inCond.GetQueryPart())
	deleteTenantAccessStmt = sqlx.Rebind(sqlx.DOLLAR, deleteTenantAccessStmt)

	log.C(ctx).Debugf("Executing DB query: %s", deleteTenantAccessStmt)
	_, err = persist.ExecContext(ctx, deleteTenantAccessStmt, args...)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Delete, "while deleting tenant access record from '%s' table", m2mTable)
}
