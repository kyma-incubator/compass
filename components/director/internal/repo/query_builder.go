package repo

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

func getAllArgs(conditions Conditions) []interface{} {
	var allArgs []interface{}

	for _, cond := range conditions {
		if argVal, ok := cond.GetQueryArg(); ok {
			allArgs = append(allArgs, argVal)
		}
	}
	return allArgs
}

func writeEnumeratedConditions(builder *strings.Builder, startIdx int, conditions Conditions) error {
	if builder == nil {
		return errors.New("builder cannot be nil")
	}

	var conditionsToJoin []string
	for idx, cond := range conditions {
		conditionsToJoin = append(conditionsToJoin, cond.GetQueryPart(idx+startIdx))
	}

	builder.WriteString(fmt.Sprintf(" %s", strings.Join(conditionsToJoin, " AND ")))

	return nil
}

// TODO: Refactor builder
func buildSelectQuery(tableName string, selectedColumns string, conditions Conditions, orderByParams OrderByParams) (string, []interface{}, error) {
	var stmtBuilder strings.Builder
	startIdx := 1

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectedColumns, tableName))
	if len(conditions) > 0 {
		stmtBuilder.WriteString(" WHERE")
	}

	err := writeEnumeratedConditions(&stmtBuilder, startIdx, conditions)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing enumerated conditions")
	}

	err = writeOrderByPart(&stmtBuilder, orderByParams)
	if err != nil {
		return "", nil, errors.Wrap(err, "while writing order by part")
	}

	allArgs := getAllArgs(conditions)

	return stmtBuilder.String(), allArgs, nil
}

func writeOrderByPart(builder *strings.Builder, orderByParams OrderByParams) error {
	if builder == nil {
		return errors.New("builder cannot be nil")
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
