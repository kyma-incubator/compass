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

// GlobalCreator missing godoc
type GlobalCreator interface {
	Create(ctx context.Context, dbEntity interface{}) error
}

type globalCreator struct {
	tableName    string
	resourceType resource.Type
	columns      []string
}

type universalCreator struct {
	tableName          string
	resourceType       resource.Type
	columns            []string
	matcherColumns     []string
	m2mTable           string
	m2mColumns         []string
	isTopLevelEntity   bool
	ownerCheckRequired bool
}

type creatorBuilder struct {
	tableName      string
	resourceType   resource.Type
	columns        []string
	matcherColumns []string
	m2mTable       string
	ownerCheck     bool
}

func NewCreatorBuilderFor(resType resource.Type) *creatorBuilder {
	return &creatorBuilder{
		resourceType: resType,
		ownerCheck:   true,
	}
}

func (cb *creatorBuilder) WithTable(tableName string) *creatorBuilder {
	cb.tableName = tableName
	return cb
}

func (cb *creatorBuilder) WithColumns(columns ...string) *creatorBuilder {
	cb.columns = columns
	return cb
}

func (cb *creatorBuilder) WithMatcherColumns(columns ...string) *creatorBuilder {
	cb.matcherColumns = columns
	return cb
}

func (cb *creatorBuilder) WithoutOwnerCheck() *creatorBuilder {
	cb.ownerCheck = false
	return cb
}

func (cb *creatorBuilder) Build() Creator {
	m2mTable, _ := cb.resourceType.TenantAccessTable()
	_, hasParent := cb.resourceType.Parent()
	return &universalCreator{
		tableName:          cb.tableName,
		resourceType:       cb.resourceType,
		columns:            cb.columns,
		matcherColumns:     cb.matcherColumns,
		m2mTable:           m2mTable,
		m2mColumns:         m2mColumns,
		isTopLevelEntity:   !hasParent,
		ownerCheckRequired: cb.ownerCheck,
	}
}

// NewCreator is a simplified constructor for child entities
func NewCreator(resourceType resource.Type, tableName string, columns []string) Creator {
	m2mTable, _ := resourceType.TenantAccessTable()
	_, hasParent := resourceType.Parent()
	return &universalCreator{
		resourceType:       resourceType,
		tableName:          tableName,
		columns:            columns,
		ownerCheckRequired: true,
		m2mTable:           m2mTable,
		m2mColumns:         m2mColumns,
		isTopLevelEntity:   !hasParent,
	}
}

// NewGlobalCreator missing godoc
func NewGlobalCreator(resourceType resource.Type, tableName string, columns []string) GlobalCreator {
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

	if c.isTopLevelEntity {
		return c.unsafeCreateTopLevelEntity(ctx, id, tenant, dbEntity)
	}

	return c.unsafeCreateChildEntity(ctx, id, tenant, dbEntity)
}

func (c *universalCreator) unsafeCreateTopLevelEntity(ctx context.Context, id string, tenant string, dbEntity interface{}) error {
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
		return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "while checking affected rows")
	}

	if affected == 0 {
		log.C(ctx).Infof("%s top level entity already exists based on matcher columns [%s]. Will proceed with only creating access record for tenant %s...", c.resourceType, strings.Join(c.matcherColumns, ", "), tenant)
	}

	vals := make([]string, 0, len(c.m2mColumns))
	for _, c := range c.m2mColumns {
		vals = append(vals, fmt.Sprintf(":%s", c))
	}
	insertTenantAccessStmt := fmt.Sprintf("INSERT INTO %s ( %s ) VALUES ( %s )", c.m2mTable, strings.Join(c.m2mColumns, ", "), strings.Join(vals, ", "))

	log.C(ctx).Debugf("Executing DB query: %s", insertTenantAccessStmt)
	_, err = persist.NamedExecContext(ctx, insertTenantAccessStmt, &TenantAccess{
		TenantID:   tenant,
		ResourceID: id,
		Owner:      true,
	})

	return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting tenant access record to '%s' table", c.m2mTable)
}

func (c *universalCreator) unsafeCreateChildEntity(ctx context.Context, id string, tenant string, dbEntity interface{}) error {
	if err := c.checkParentAccess(ctx, tenant, dbEntity); err != nil {
		return apperrors.NewUnauthorizedError(fmt.Sprintf("Tenant %s does not have access to the parent of the currently created %s: %v", tenant, c.resourceType, err))
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

	return persistence.MapSQLError(ctx, err, c.resourceType, resource.Create, "while inserting row to '%s' table", c.tableName)
}

func (c *universalCreator) checkParentAccess(ctx context.Context, tenant string, dbEntity interface{}) error {
	resourceType := c.resourceType
	if multiRefEntity, ok := dbEntity.(MultiRefEntity); ok {
		resourceType = multiRefEntity.GetRefSpecificResourceType()
	}

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

	exister := NewExistQuerier(parentResourceType, parentAccessTable, M2MTenantIDColumn)
	exists, err := exister.Exists(ctx, tenant, conditions)
	if err != nil {
		return errors.Wrap(err, "while checking for tenant access")
	}

	if !exists {
		return errors.Errorf("Tenant %s does not have access to the parent resource %s with ID %s.", tenant, parentResourceType, parentID)
	}

	return nil
}

func (c *globalCreator) Create(ctx context.Context, dbEntity interface{}) error { // TODO: <storage-redesign> remove duplication
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
