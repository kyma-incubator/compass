package repo

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/jmoiron/sqlx"

	"github.com/pkg/errors"
)

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

// TODO: Refactor builder
func buildSelectQuery(tableName string, selectedColumns string, conditions Conditions, orderByParams OrderByParams) (string, []interface{}, error) {
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

	return getQueryFromBuilder(stmtBuilder), allArgs, nil
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
