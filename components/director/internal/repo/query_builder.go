package repo

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

type QueryBuilder interface {
	BuildQuery(tenantID string, isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error)
}

type universalQueryBuilder struct {
	tableName       string
	selectedColumns string
	tenantColumn    *string
	resourceType    resource.Type
}

func NewQueryBuilder(resourceType resource.Type, tableName string, tenantColumn string, selectedColumns []string) QueryBuilder {
	return &universalQueryBuilder{
		tableName:       tableName,
		selectedColumns: strings.Join(selectedColumns, ", "),
		tenantColumn:    &tenantColumn,
		resourceType:    resourceType,
	}
}

func (b *universalQueryBuilder) BuildQuery(tenantID string, isRebindingNeeded bool, conditions ...Condition) (string, []interface{}, error) {
	if tenantID == "" {
		return "", nil, apperrors.NewTenantRequiredError()
	}
	conditions = append(Conditions{NewTenantIsolationCondition(*b.tenantColumn, tenantID)}, conditions...)

	return buildSelectQuery(b.tableName, b.selectedColumns, conditions, OrderByParams{}, isRebindingNeeded)
}

// TODO: Refactor builder
func buildSelectQuery(tableName string, selectedColumns string, conditions Conditions, orderByParams OrderByParams, isRebindingNeeded bool) (string, []interface{}, error) {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectedColumns, tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err := writeEnumeratedConditions(&stmtBuilder, conditions)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing enumerated conditions.")
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

	var conditionsToJoin []string
	for _, cond := range conditions {
		conditionsToJoin = append(conditionsToJoin, cond.GetQueryPart())
	}

	builder.WriteString(fmt.Sprintf(" %s", strings.Join(conditionsToJoin, " AND ")))

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

	if orderByParams == nil || len(orderByParams) == 0 {
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
