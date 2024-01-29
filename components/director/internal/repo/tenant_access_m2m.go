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
	// M2MSourceColumn is the column name of the source in each tenant access table / view.
	M2MSourceColumn = "source"

	// CreateSingleTenantAccessQuery is a SQL query that creates a tenant access record
	CreateSingleTenantAccessQuery = `INSERT INTO %s ( %s ) VALUES ( %s ) ON CONFLICT ON CONSTRAINT tenant_applications_pkey DO NOTHING`

	// RecursiveCreateTenantAccessCTEQuery is a recursive SQL query that creates a tenant access record for a tenant and all its parents.
	RecursiveCreateTenantAccessCTEQuery = `WITH RECURSIVE parents AS
                   (SELECT t1.id, t1.type, tp1.parent_id, 0 AS depth, CAST(:source AS uuid) AS child_id
                    FROM business_tenant_mappings t1 LEFT JOIN tenant_parents tp1 on t1.id = tp1.tenant_id
                    WHERE id=:tenant_id
                    UNION ALL
                    SELECT t2.id, t2.type, tp2.parent_id, p.depth+ 1, p.id AS child_id
                    FROM business_tenant_mappings t2 LEFT JOIN tenant_parents tp2 on t2.id = tp2.tenant_id
                                                     INNER JOIN parents p on p.parent_id = t2.id)
			INSERT INTO %s ( %s )  (SELECT parents.id AS tenant_id, :id as id, :owner AS owner, parents.child_id as source FROM parents WHERE type != 'cost-object'
                                                                                                                 OR (type = 'cost-object' AND depth = (SELECT MIN(depth) FROM parents WHERE type = 'cost-object'))
					)
			ON CONFLICT ( tenant_id, id, source ) DO NOTHING`

	// RecursiveDeleteTenantAccessCTEQuery is a recursive SQL query that deletes tenant accesses based on given conditions for a tenant and all its parents.
	RecursiveDeleteTenantAccessCTEQuery = `WITH RECURSIVE
    parents AS
        (SELECT t1.id,
                t1.type,
                tp1.parent_id,
                0                                                    AS depth,
                CAST( ? AS uuid) AS child_id
         FROM business_tenant_mappings t1
                  LEFT JOIN tenant_parents tp1 ON t1.id = tp1.tenant_id
         WHERE id = ?
         UNION ALL
         SELECT t2.id, t2.type, tp2.parent_id, p.depth + 1, p.id AS child_id
         FROM business_tenant_mappings t2
                  LEFT JOIN tenant_parents tp2 ON t2.id = tp2.tenant_id
                  INNER JOIN parents p ON p.parent_id = t2.id),
    parent_access_records_count AS (SELECT pp.id    AS tenant_id,
                                           act.id   AS obj_id,
                                           pp.depth,
                                           COUNT(1) AS access_records_count
                                    FROM %s act
                                             JOIN parents pp ON act.tenant_id = pp.id
                                    WHERE act.%s
                                    GROUP BY pp.id, pp.depth, act.id),
    anchor AS (SELECT par.*
               FROM parent_access_records_count par
                        LEFT JOIN
                    parent_access_records_count par2 ON par.obj_id = par2.obj_id
                        AND par.depth > par2.depth AND par2.access_records_count > 1
               WHERE par.access_records_count > 1
                 AND par2.tenant_id IS NULL

               UNION ALL

               SELECT par.*
               FROM parent_access_records_count par
                        LEFT JOIN
                    parent_access_records_count par2 ON par.obj_id = par2.obj_id
                        AND par.depth > par2.depth AND par2.access_records_count > 1
                        LEFT JOIN
                    parent_access_records_count par3 ON par.obj_id = par3.obj_id
                        AND par.depth < par3.depth
               WHERE par.access_records_count = 1
                 AND par2.tenant_id IS NULL
                 AND par3.tenant_id IS NULL)
DELETE
FROM %s act
WHERE act.%s
  AND EXISTS (SELECT id
              FROM parents
              WHERE tenant_id = parents.id
                AND source = parents.child_id
                AND parents.depth <= ALL (SELECT a.depth FROM anchor a WHERE a.obj_id = act.id));`

	// DeleteDirectiveAccess is a query that deletes all the tenant accesses that have come from the directive.
	DeleteDirectiveAccess = `DELETE
FROM %s a
WHERE %s
  AND %s
  AND NOT EXISTS (SELECT 1 FROM tenant_applications ta WHERE ta.tenant_id = a.source AND ta.id = a.id);`

	// DeleteTenantAccessGrantedByParentQuery is a delete SQL query that deletes tenant accesses based on given tenant id and source.
	DeleteTenantAccessGrantedByParentQuery = `DELETE FROM %s WHERE tenant_id = ? AND SOURCE = ?`
)

// M2MColumns are the column names of the tenant access tables / views.
var M2MColumns = []string{M2MTenantIDColumn, M2MResourceIDColumn, M2MOwnerColumn, M2MSourceColumn}

// TenantAccess represents the tenant access table/views that are used for tenant isolation queries.
type TenantAccess struct {
	TenantID   string `db:"tenant_id"`
	ResourceID string `db:"id"`
	Owner      bool   `db:"owner"`
	Source     string `db:"source"`
}

// TenantAccessCollection is a wrapper type for slice of entities.
type TenantAccessCollection []TenantAccess

// Len returns the current number of entities in the collection.
func (tc TenantAccessCollection) Len() int {
	return len(tc)
}

// GetSingleTenantAccess gets a tenant access record for tenant with ID tenantID and resource with ID resourceID
func GetSingleTenantAccess(ctx context.Context, m2mTable string, tenantID, resourceID string) (*TenantAccess, error) {
	getter := NewSingleGetterGlobal(resource.TenantAccess, m2mTable, M2MColumns)

	tenantAccess := &TenantAccess{}
	err := getter.GetGlobal(ctx, Conditions{NewEqualCondition(M2MTenantIDColumn, tenantID), NewEqualCondition(M2MResourceIDColumn, resourceID)}, NoOrderBy, tenantAccess)
	if err != nil {
		log.C(ctx).Error(persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Get, "while fetching tenant access record from '%s' table", m2mTable))
		return nil, err
	}

	return tenantAccess, nil
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
// add it to all the parents of the given tenant. In case of conflict the entry is not updated
func CreateTenantAccessRecursively(ctx context.Context, m2mTable string, tenantAccess *TenantAccess) error {
	if tenantAccess.Source == "" {
		tenantAccess.Source = tenantAccess.TenantID
	}
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	insertTenantAccessStmt := fmt.Sprintf(RecursiveCreateTenantAccessCTEQuery, m2mTable, strings.Join(M2MColumns, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, tenantAccess)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Create, "while inserting tenant access record to '%s' table", m2mTable)
}

// DeleteTenantAccessRecursively deletes all the accesses to the provided resource IDs for the given tenant and all its parents.
func DeleteTenantAccessRecursively(ctx context.Context, m2mTable string, tenant string, resourceIDs []string, childTenant string) error {
	if len(resourceIDs) == 0 {
		return errors.New("resourceIDs cannot be empty")
	}

	if childTenant == "" {
		childTenant = tenant
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(resourceIDs)+1)
	args = append(args, childTenant)
	args = append(args, tenant)

	inCond := NewInConditionForStringValues(M2MResourceIDColumn, resourceIDs)
	if inArgs, ok := inCond.GetQueryArgs(); ok {
		args = append(args, inArgs...)
		args = append(args, inArgs...) //the same condition is used in the main delete query and in the generation of one of the CTEs
	}

	deleteTenantAccessStmt := fmt.Sprintf(RecursiveDeleteTenantAccessCTEQuery, m2mTable, inCond.GetQueryPart(), m2mTable, inCond.GetQueryPart())
	deleteTenantAccessStmt = sqlx.Rebind(sqlx.DOLLAR, deleteTenantAccessStmt)

	log.C(ctx).Debugf("Executing DB query: %s", deleteTenantAccessStmt)
	_, err = persist.ExecContext(ctx, deleteTenantAccessStmt, args...)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Delete, "while deleting tenant access record from '%s' table", m2mTable)
}

// DeleteTenantAccessFromDirective deletes all the accesses to the provided resource IDs created from the directive for which the root tenant no longer has access record
func DeleteTenantAccessFromDirective(ctx context.Context, m2mTable string, resourceIDs, rootTenantIDs []string) error {
	log.C(ctx).Infof("Deleting tenant access records for %s with source in %s where the source no longer has access to the object", resourceIDs, rootTenantIDs)
	if len(resourceIDs) == 0 {
		return errors.New("resourceIDs cannot be empty")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	args := make([]interface{}, 0, len(resourceIDs)+len(rootTenantIDs))

	inCondForResourceIDs := NewInConditionForStringValues(M2MResourceIDColumn, resourceIDs)
	if inArgs, ok := inCondForResourceIDs.GetQueryArgs(); ok {
		args = append(args, inArgs...)
	}

	inCondForRootTenantsIDs := NewInConditionForStringValues(M2MSourceColumn, rootTenantIDs)
	if inArgs, ok := inCondForRootTenantsIDs.GetQueryArgs(); ok {
		args = append(args, inArgs...)
	}

	deleteTenantAccessStmt := fmt.Sprintf(DeleteDirectiveAccess, m2mTable, inCondForResourceIDs.GetQueryPart(), inCondForRootTenantsIDs.GetQueryPart())
	deleteTenantAccessStmt = sqlx.Rebind(sqlx.DOLLAR, deleteTenantAccessStmt)

	log.C(ctx).Debugf("Executing DB query: %s", deleteTenantAccessStmt)
	_, err = persist.ExecContext(ctx, deleteTenantAccessStmt, args...)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Delete, "while deleting tenant access record from '%s' table", m2mTable)
}

// DeleteTenantAccessFromParent deletes all the accesses to resources that were granted to childTenantID from parentTenantID. Such tenant accesses are granted through the directive
func DeleteTenantAccessFromParent(ctx context.Context, m2mTable string, childTenantID, parentTenantID string) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	deleteTenantAccessStmt := fmt.Sprintf(DeleteTenantAccessGrantedByParentQuery, m2mTable)
	deleteTenantAccessStmt = sqlx.Rebind(sqlx.DOLLAR, deleteTenantAccessStmt)

	log.C(ctx).Debugf("Executing DB query: %s", deleteTenantAccessStmt)
	_, err = persist.ExecContext(ctx, deleteTenantAccessStmt, childTenantID, parentTenantID)

	return persistence.MapSQLError(ctx, err, resource.TenantAccess, resource.Delete, "while deleting tenant access records for tenant with ID %s granted by parent with ID %s from '%s' table", childTenantID, parentTenantID, m2mTable)
}
