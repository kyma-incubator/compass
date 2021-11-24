package repo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// Creator is an interface for creating entities with externally managed tenant accesses (m2m table or view)
type Creator interface {
	Create(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error
}

// CreatorGlobal is an interface for creating global entities without tenant or entities with tenant embedded in them.
type CreatorGlobal interface {
	Create(ctx context.Context, dbEntity interface{}) error
}

type universalCreator struct {
	tableName          string
	columns            []string
	matcherColumns     []string
	ownerCheckRequired bool
}

// NewCreator is a constructor for Creator about entities with externally managed tenant accesses (m2m table or view)
func NewCreator(tableName string, columns []string) Creator {
	return &universalCreator{
		tableName:          tableName,
		columns:            columns,
		ownerCheckRequired: true,
	}
}

// NewCreatorWithMatchingColumns is a constructor for Creator about entities with externally managed tenant accesses (m2m table or view).
// In addition, matcherColumns can be added in order to identify already existing top-level entities and prevent their duplicate creation.
func NewCreatorWithMatchingColumns(tableName string, columns []string, matcherColumns []string) Creator {
	return &universalCreator{
		tableName:          tableName,
		columns:            columns,
		matcherColumns:     matcherColumns,
		ownerCheckRequired: true,
	}
}

// NewCreatorGlobal is a constructor for GlobalCreator about entities without tenant or entities with tenant embedded in them.
func NewCreatorGlobal(resourceType resource.Type, tableName string, columns []string) CreatorGlobal {
	return &globalCreator{
		resourceType: resourceType,
		tableName:    tableName,
		columns:      columns,
	}
}

// Create is a method for creating entities with externally managed tenant accesses (m2m table or view)
// In case of top level entity it creates tenant access record in the m2m table as well.
// In case of child entity first it checks if the calling tenant has access to the parent entity and then creates the child entity.
func (c *universalCreator) Create(ctx context.Context, resourceType resource.Type, tenant string, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	var id string
	if identifiable, ok := dbEntity.(Identifiable); ok {
		id = identifiable.GetID()
	}

	if len(id) == 0 {
		return apperrors.NewInternalError("id cannot be empty, check if the entity implements Identifiable")
	}

	entity, ok := dbEntity.(Entity)
	if ok && entity.GetCreatedAt().IsZero() { // This zero check is needed to mock the Create tests
		now := time.Now()
		entity.SetCreatedAt(now)
		entity.SetReady(true)
		entity.SetError(NewValidNullableString(""))

		if operation.ModeFromCtx(ctx) == graphql.OperationModeAsync {
			entity.SetReady(false)
		}

		dbEntity = entity
	}

	if resourceType.IsTopLevel() {
		return c.unsafeCreateTopLevelEntity(ctx, id, tenant, dbEntity, resourceType)
	}

	return c.unsafeCreateChildEntity(ctx, tenant, dbEntity, resourceType)
}

func (c *universalCreator) unsafeCreateTopLevelEntity(ctx context.Context, id string, tenant string, dbEntity interface{}, resourceType resource.Type) error {
	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	values := make([]string, 0, len(c.columns))
	for _, c := range c.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))
	if len(c.matcherColumns) > 0 {
		stmt = fmt.Sprintf("%s ON CONFLICT ( %s ) DO NOTHING", stmt, strings.Join(c.matcherColumns, ", "))
	}

	log.C(ctx).Debugf("Executing DB query: %s", stmt)
	res, err := persist.NamedExecContext(ctx, stmt, dbEntity)
	if err != nil {
		return persistence.MapSQLError(ctx, err, resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}

	if affected == 0 {
		log.C(ctx).Warnf("%s top level entity already exists based on matcher columns [%s]. Returning not-unique error for calling tenant %s...", resourceType, strings.Join(c.matcherColumns, ", "), tenant)
		return apperrors.NewNotUniqueError(resourceType)
	}

	vals := make([]string, 0, len(m2mColumns))
	for _, c := range m2mColumns {
		vals = append(vals, fmt.Sprintf(":%s", c))
	}

	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("entity %s does not have access table", resourceType)
	}

	insertTenantAccessStmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", m2mTable, strings.Join(m2mColumns, ", "), strings.Join(vals, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, &TenantAccess{
		TenantID:   tenant,
		ResourceID: id,
		Owner:      true,
	})

	return persistence.MapSQLError(ctx, err, resourceType, resource.Create, "while inserting tenant access record to '%s' table", m2mTable)
}

func (c *universalCreator) unsafeCreateChildEntity(ctx context.Context, tenant string, dbEntity interface{}, resourceType resource.Type) error {
	if err := c.checkParentAccess(ctx, tenant, dbEntity, resourceType); err != nil {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Tenant %s does not have access to the parent of the currently created %s: %v", tenant, resourceType, err))
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	values := make([]string, 0, len(c.columns))
	for _, c := range c.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	insertStmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertStmt)
	_, err = persist.NamedExecContext(ctx, insertStmt, dbEntity)

	return persistence.MapSQLError(ctx, err, resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
}

func (c *universalCreator) checkParentAccess(ctx context.Context, tenant string, dbEntity interface{}, resourceType resource.Type) error {
	parentResourceType, ok := resourceType.Parent()
	if !ok {
		return errors.Errorf("entity %s does not have parent", resourceType)
	}

	parentAccessTable, ok := parentResourceType.TenantAccessTable()
	if !ok {
		return errors.Errorf("parent entity %s does not have access table", parentResourceType)
	}

	var parentID string
	if childEntity, ok := dbEntity.(ChildEntity); ok {
		parentID = childEntity.GetParentID()
	}

	if len(parentID) == 0 {
		return errors.Errorf("unknown parentID for entity type %s and parentType %s", resourceType, parentResourceType)
	}

	conditions := Conditions{NewEqualCondition(M2MResourceIDColumn, parentID)}
	if c.ownerCheckRequired {
		conditions = append(conditions, NewEqualCondition(M2MOwnerColumn, true))
	}

	exister := NewExistQuerierWithEmbeddedTenant(parentAccessTable, M2MTenantIDColumn)
	exists, err := exister.Exists(ctx, resource.TenantAccess, tenant, conditions)
	if err != nil {
		return errors.Wrap(err, "while checking for tenant access")
	}

	if !exists {
		return errors.Errorf("tenant %s does not have access to the parent resource %s with ID %s", tenant, parentResourceType, parentID)
	}

	return nil
}

type globalCreator struct {
	tableName    string
	resourceType resource.Type
	columns      []string
}

// Create creates a new global entity or entity with embedded tenant in it.
func (c *globalCreator) Create(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	values := make([]string, 0, len(c.columns))
	for _, c := range c.columns {
		values = append(values, fmt.Sprintf(":%s", c))
	}

	entity, ok := dbEntity.(Entity)
	if ok && entity.GetCreatedAt().IsZero() { // This zero check is needed to mock the Create tests
		now := time.Now()
		entity.SetCreatedAt(now)
		entity.SetReady(true)
		entity.SetError(NewValidNullableString(""))

		if operation.ModeFromCtx(ctx) == graphql.OperationModeAsync {
			entity.SetReady(false)
		}

		dbEntity = entity
	}

	stmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.tableName, strings.Join(c.columns, ", "), strings.Join(values, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", stmt)
	_, err = persist.NamedExecContext(ctx, stmt, dbEntity)

	return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
}
