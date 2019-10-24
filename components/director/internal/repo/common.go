package repo

import (
	"errors"
	"fmt"
	"strings"
)

type ConditionOp string

const (
	EqualOp     ConditionOp = "="
	IsNotNullOp ConditionOp = "IS NOT NULL"
)

type Conditions []Condition
type Condition struct {
	Field string
	Op    ConditionOp
	Val   string
}

func NewEqualCondition(field, val string) Condition {
	return Condition{
		Field: field,
		Val:   val,
		Op:    EqualOp,
	}
}

func NewNotNullCondition(field string) Condition {
	return Condition{
		Field: field,
		Op:    IsNotNullOp,
	}
}

func isNoArgCondition(cond Condition) bool {
	return cond.Op == IsNotNullOp
}

func getAllArgs(tenant *string, conditions Conditions) []interface{} {
	allArgs := []interface{}{}
	if tenant != nil {
		allArgs = append(allArgs, tenant)
	}
	for _, idAndVal := range conditions {
		if isNoArgCondition(idAndVal) {
			continue
		}
		allArgs = append(allArgs, idAndVal.Val)
	}
	return allArgs
}

func writeEnumeratedConditions(builder *strings.Builder, startIdx int, conditions Conditions) error {
	if builder == nil {
		return errors.New("builder cannot be nil")
	}

	var conditionsToJoin []string
	for idx, idAndVal := range conditions {
		if isNoArgCondition(idAndVal) {
			conditionsToJoin = append(conditionsToJoin, fmt.Sprintf("%s %s", idAndVal.Field, idAndVal.Op))
		} else {
			conditionsToJoin = append(conditionsToJoin, fmt.Sprintf("%s %s $%d", idAndVal.Field, idAndVal.Op, idx+startIdx))
		}
	}

	builder.WriteString(fmt.Sprintf(" %s", strings.Join(conditionsToJoin, " AND ")))

	return nil
}

func buildSelectStatement(selectedColumns, tableName string, tenantColumn *string, additionalConditions []string) string {
	var stmtBuilder strings.Builder

	stmtBuilder.WriteString(fmt.Sprintf("SELECT %s FROM %s", selectedColumns, tableName))

	if tenantColumn != nil || len(additionalConditions) > 0 {
		stmtBuilder.WriteString(" WHERE ")
	}

	if tenantColumn != nil {
		stmtBuilder.WriteString(fmt.Sprintf("%s=$1", *tenantColumn))
	}

	if len(additionalConditions) > 0 && tenantColumn != nil {
		stmtBuilder.WriteString(" AND ")
	}

	stmtBuilder.WriteString(strings.Join(additionalConditions, " AND "))

	return stmtBuilder.String()
}
