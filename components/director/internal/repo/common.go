package repo

import (
	"fmt"
	"strings"
)

type Conditions []Condition
type Condition struct {
	Field string
	Val   string
}

func getAllArgs(tenant string, conditions Conditions) []interface{} {
	allArgs := []interface{}{tenant}
	for _, idAndVal := range conditions {
		allArgs = append(allArgs, idAndVal.Val)
	}
	return allArgs
}

func appendEnumeratedConditions(query string, startIdx int, conditions Conditions) string {
	out := query
	for idx, idAndVal := range conditions {
		out = fmt.Sprintf("%s AND %s = $%d", out, idAndVal.Field, idx+startIdx)
	}
	return out
}

func buildSelectStatement(selectedColumns, tableName, tenantColumn string, additionalConditions []string) string {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s WHERE %s=$1", selectedColumns, tableName, tenantColumn))

	for _, cond := range additionalConditions {
		if strings.TrimSpace(cond) != "" {
			stmtBuilder.WriteString(fmt.Sprintf(` AND %s`, cond))
		}
	}

	return stmtBuilder.String()
}
