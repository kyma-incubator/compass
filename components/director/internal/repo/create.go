package repo

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/operation"
	"github.com/pkg/errors"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
)

// Creator missing godoc
type Creator interface {
	Create(ctx context.Context, tenant string, dbEntity interface{}) error
}

// CreatorGlobal missing godoc
type CreatorGlobal interface {
	Create(ctx context.Context, dbEntity interface{}) error
}

type universalCreator struct {
	tableName          string
	resourceType       resource.Type
	columns            []string
	matcherColumns     []string
	ownerCheckRequired bool
}

// NewCreator is a simplified constructor for child entities
func NewCreator(resourceType resource.Type, tableName string, columns []string) Creator {
	return &universalCreator{
		resourceType:       resourceType,
		tableName:          tableName,
		columns:            columns,
		ownerCheckRequired: true,
	}
}

// NewCreatorGlobal missing godoc
func NewCreatorGlobal(resourceType resource.Type, tableName string, columns []string) CreatorGlobal {
	return &globalCreator{
		resourceType: resourceType,
		tableName:    tableName,
		columns:      columns,
	}
}

// Create missing godoc
func (c *universalCreator) Create(ctx context.Context, tenant string, dbEntity interface{}) error {
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

	resourceType := c.resourceType
	if multiRefEntity, ok := dbEntity.(MultiRefEntity); ok {
		resourceType = multiRefEntity.GetRefSpecificResourceType()
	}

	if resourceType.IsTopLevel() {
		return c.unsafeCreateTopLevelEntity(ctx, id, tenant, dbEntity, resourceType)
	}

	return c.unsafeCreateChildEntity(ctx, id, tenant, dbEntity, resourceType)
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
		log.C(ctx).Infof("%s top level entity already exists based on matcher columns [%s]. Will proceed with only creating access record for tenant %s...", resourceType, strings.Join(c.matcherColumns, ", "), tenant)
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

func (c *universalCreator) unsafeCreateChildEntity(ctx context.Context, id string, tenant string, dbEntity interface{}, resourceType resource.Type) error {
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
		return errors.Errorf("Unkonw parentID for entity type %s and parentType %s", resourceType, parentResourceType)
	}

	conditions := Conditions{NewEqualCondition(M2MResourceIDColumn, parentID)}
	if c.ownerCheckRequired {
		conditions = append(conditions, NewEqualCondition(M2MOwnerColumn, true))
	}

	exister := NewExistQuerierWithEmbeddedTenant(parentResourceType, parentAccessTable, M2MTenantIDColumn)
	exists, err := exister.Exists(ctx, tenant, conditions)
	if err != nil {
		return errors.Wrap(err, "while checking for tenant access")
	}

	if !exists {
		return errors.Errorf("Tenant %s does not have access to the parent resource %s with ID %s.", tenant, parentResourceType, parentID)
	}

	return nil
}

type globalCreator struct {
	tableName    string
	resourceType resource.Type
	columns      []string
}

func (c *globalCreator) Create(ctx context.Context, dbEntity interface{}) error {
	if dbEntity == nil {
		return apperrors.NewInternalError("item cannot be nil")
	}

	persist, err := persistence.FromCtx(ctx)
	if err != nil {
		return err
	}

	var values []string
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
