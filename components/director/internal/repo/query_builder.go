package repo

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

const (
	// ForUpdateLock Represents FOR UPDATE lock clause in SELECT queries.
	ForUpdateLock string = "FOR UPDATE"
	// NoLock Represents missing lock clause in SELECT queries.
	NoLock string = ""
)

// QueryBuilder is an interface for building queries about tenant scoped entities with either externally managed tenant accesses (m2m table or view) or embedded tenant in them.
type QueryBuilder interface {
	BuildQuery(resourceType resource.Type, tenantID string, isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error)
}

// QueryBuilderGlobal is an interface for building queries about global entities.
type QueryBuilderGlobal interface {
	BuildQueryGlobal(isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error)
}

type universalQueryBuilder struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
}

// NewQueryBuilderWithEmbeddedTenant is a constructor for QueryBuilder about entities with tenant embedded in them.
func NewQueryBuilderWithEmbeddedTenant(tableName string, tenantColumn string, selectedColumns []string) QueryBuilder {
	return &universalQueryBuilder{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
	}
}

// NewQueryBuilder is a constructor for QueryBuilder about entities with externally managed tenant accesses (m2m table or view)
func NewQueryBuilder(tableName string, selectedColumns []string) QueryBuilder {
	return &universalQueryBuilder{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

// NewQueryBuilderGlobal is a constructor for QueryBuilderGlobal about global entities.
func NewQueryBuilderGlobal(resourceType resource.Type, tableName string, selectedColumns []string) QueryBuilderGlobal {
	return &universalQueryBuilder{
		resourceType:    resourceType,
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
	}
}

// BuildQueryGlobal builds a SQL query for global entities without tenant isolation.
func (b *universalQueryBuilder) BuildQueryGlobal(isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error) {
	return buildSelectQuery(b.tableName, b.selectedColumns, conditions, OrderByParams{}, NoLock, isRebindingNeeded)
}

// BuildQuery builds a SQL query for tenant scoped entities with tenant isolation subquery.
// If the tenantColumn is configured the isolation is based on equal condition on tenantColumn.
// If the tenantColumn is not configured an entity with externally managed tenant accesses in m2m table / view is assumed.
func (b *universalQueryBuilder) BuildQuery(resourceType resource.Type, tenantID string, isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error) {
	if tenantID == "" {
		return "", nil, apperrors.NewTenantRequiredError()
	}

	if b.tenantColumn != nil {
		conditions = append(Conditions{NewEqualCondition(*b.tenantColumn, tenantID)}, conditions...)
		return buildSelectQuery(b.tableName, b.selectedColumns, conditions, OrderByParams{}, NoLock, isRebindingNeeded)
	}

	tenantIsolation, err := NewTenantIsolationCondition(resourceType, tenantID, false)
	if err != nil {
		return "", nil, err
	}

	conditions = append(conditions, tenantIsolation)

	return buildSelectQuery(b.tableName, b.selectedColumns, conditions, OrderByParams{}, NoLock, isRebindingNeeded)
}

func buildSelectQueryFromTree(tableName string, selectedColumns string, conditions *ConditionTree, orderByParams OrderByParams, lockClause string, isRebindingNeeded bool) (string, []interface{}, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectedColumns, tableName))
	var allArgs []interface{}
	if conditions != nil {
		stmtBuilder.WriteString(" WHERE ")
		var subquery string
		subquery, allArgs = conditions.BuildSubquery()
		stmtBuilder.WriteString(subquery)
	}

	err := writeOrderByPart(&stmtBuilder, orderByParams)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing order by part")
	}

	writeLockClause(&stmtBuilder, lockClause)

	if isRebindingNeeded {
		return getQueryFromBuilder(stmtBuilder), allArgs, nil
	}
	return stmtBuilder.String(), allArgs, nil
}

// TODO: Refactor builder
func buildSelectQuery(tableName string, selectedColumns string, conditions Conditions, orderByParams OrderByParams, lockClause string, isRebindingNeeded bool) (string, []interface{}, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectedColumns, tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err := writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing enumerated conditions.")
	}

	allArgs := getAllArgs(conditions)

	err = writeOrderByPart(&stmtBuilder, orderByParams)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing order by part")
	}

	writeLockClause(&stmtBuilder, lockClause)

	if isRebindingNeeded {
		return getQueryFromBuilder(stmtBuilder), allArgs, nil
	}
	return stmtBuilder.String(), allArgs, nil
}

func buildUnionQuery(queries []string) string {
	if len(queries) == 0 {
		return ""
	}

	for i := range queries {
		queries[i] = "(" + queries[i] + ")"
	}

	unionQuery := strings.Join(queries, " UNION ")
	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(unionQuery)

	return getQueryFromBuilder(stmtBuilder)
}

func buildCountQuery(tableName string, idColumn string, conditions Conditions, groupByParams GroupByParams, orderByParams OrderByParams, isRebindingNeeded bool) (string, []interface{}, error) {
	isGroupByParam := false
	for _, s := range groupByParams {
		if idColumn == s {
			isGroupByParam = true
		}
	}
	if !isGroupByParam {
		return "", nil, errors.New("id column is not in group by params")
	}

	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s AS id, COUNT(*) AS total_count FROM %s", idColumn, tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err := writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing enumerated conditions.")
	}

	err = writeGroupByPart(&stmtBuilder, groupByParams)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing order by part")
	}

	err = writeOrderByPart(&stmtBuilder, orderByParams)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing order by part")
	}

	allArgs := getAllArgs(conditions)

	if isRebindingNeeded {
		return getQueryFromBuilder(stmtBuilder), allArgs, nil
	}
	return stmtBuilder.String(), allArgs, nil
}

func buildSelectQueryWithLimitAndOffset(tableName string, selectedColumns string, conditions Conditions, orderByParams OrderByParams, limit, offset int, isRebindingNeeded bool) (string, []interface{}, error) {
	query, args, err := buildSelectQuery(tableName, selectedColumns, conditions, orderByParams, NoLock, isRebindingNeeded)
	if err != nil {
		return "", nil, err
	}

	var stmtBuilder strings.Builder
	stmtBuilder.WriteString(query)

	err = writeLimitPart(&stmtBuilder)
	if err != nil {
		return "", nil, err
	}
	args = append(args, limit)

	err = writeOffsetPart(&stmtBuilder)
	if err != nil {
		return "", nil, err
	}
	args = append(args, offset)

	if isRebindingNeeded {
		return getQueryFromBuilder(stmtBuilder), args, nil
	}
	return stmtBuilder.String(), args, nil
}

func getAllArgs(conditions Conditions) []interface{} {
	var allArgs []interface{}

	for _, cond := range conditions {
		if argVal, ok := cond.GetQueryArgs(); ok {
			allArgs = append(allArgs, argVal...)
		}
	}
	return allArgs
}

func writeEnumeratedConditions(builder *strings.Builder, conditions Conditions) error {
	if builder == nil {
		return apperrors.NewInternalError("builder cannot be nil")
	}

	conditionsToJoin := make([]string, 0, len(conditions))
	for _, cond := range conditions {
		conditionsToJoin = append(conditionsToJoin, cond.GetQueryPart())
	}

	builder.WriteString(" ")
	builder.WriteString(strings.Join(conditionsToJoin, " AND "))

	return nil
}

const anyKeyExistsOp = "?|"
const anyKeyExistsOpPlaceholder = "{{anyKeyExistsOp}}"

const allKeysExistOp = "?&"
const allKeysExistOpPlaceholder = "{{allKeysExistOp}}"

const singleKeyExistsOp = "] ? "
const singleKeyExistsOpPlaceholder = "{{singleKeyExistsOp}}"

var tempReplace = []string{
	anyKeyExistsOp, anyKeyExistsOpPlaceholder,
	allKeysExistOp, allKeysExistOpPlaceholder,
	singleKeyExistsOp, singleKeyExistsOpPlaceholder,
}

var reverseTempReplace = []string{
	anyKeyExistsOpPlaceholder, anyKeyExistsOp,
	allKeysExistOpPlaceholder, allKeysExistOp,
	singleKeyExistsOpPlaceholder, singleKeyExistsOp,
}

// sqlx doesn't detect ?| and ?& operators properly
func getQueryFromBuilder(builder strings.Builder) string {
	strToRebind := strings.NewReplacer(tempReplace...).Replace(builder.String())
	strAfterRebind := sqlx.Rebind(sqlx.DOLLAR, strToRebind)

	query := strings.NewReplacer(reverseTempReplace...).Replace(strAfterRebind)

	return query
}

func writeOrderByPart(builder *strings.Builder, orderByParams OrderByParams) error {
	if builder == nil {
		return apperrors.NewInternalError("builder cannot be nil")
	}

	if len(orderByParams) == 0 {
		return nil
	}

	builder.WriteString(" ORDER BY")
	for idx, orderBy := range orderByParams {
		if idx > 0 {
			builder.WriteString(",")
		}
		builder.WriteString(fmt.Sprintf(" %s %s", orderBy.Field, orderBy.Dir))
	}

	return nil
}

// GroupByParams missing godoc
type GroupByParams []string

func writeGroupByPart(builder *strings.Builder, groupByParams GroupByParams) error {
	if builder == nil {
		return apperrors.NewInternalError("builder cannot be nil")
	}

	if len(groupByParams) == 0 {
		return nil
	}

	builder.WriteString(" GROUP BY ")
	builder.WriteString(strings.Join(groupByParams, " ,"))

	return nil
}

func writeLimitPart(builder *strings.Builder) error {
	if builder == nil {
		return apperrors.NewInternalError("builder cannot be nil")
	}

	builder.WriteString(" LIMIT ?")
	return nil
}

func writeOffsetPart(builder *strings.Builder) error {
	if builder == nil {
		return apperrors.NewInternalError("builder cannot be nil")
	}

	builder.WriteString(" OFFSET ?")
	return nil
}

func writeLockClause(builder *strings.Builder, lockClause string) {
	lock := strings.TrimSpace(lockClause)
	if lock != NoLock {
		builder.WriteString(" " + lock)
	}
}
