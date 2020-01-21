package repo

import (
	"errors"
	"fmt"
	"strings"
)

// OrderByDir is a type encapsulating the ORDER BY direction
type OrderByDir string

const (
	// AscOrderBy defines ascending order
	AscOrderBy OrderByDir = "ASC"
	// DescOrderBy defines descending order
	DescOrderBy OrderByDir = "DESC"
)

// OrderBy type that wraps the information about the ordering column and direction
type OrderBy struct {
	Field string
	Dir   OrderByDir
}

// NewAscOrderBy returns wrapping type for ascending order for a given column (field)
func NewAscOrderBy(field string) OrderBy {
	return OrderBy{
		Field: field,
		Dir:   AscOrderBy,
	}
}

// NewDescOrderBy returns wrapping type for descending orderd for a given column (field)
func NewDescOrderBy(field string) OrderBy {
	return OrderBy{
		Field: field,
		Dir:   DescOrderBy,
	}
}

// OrderByParams is a wrapping type for slice of OrderBy types
type OrderByParams []OrderBy

// NoOrderBy represents default ordering (no order specified)
var NoOrderBy OrderByParams = OrderByParams{}

type ConditionOp string

const (
	EqualOp     ConditionOp = "="
	IsNotNullOp ConditionOp = "IS NOT NULL"
	InOp        ConditionOp = "IN"
)

type Conditions []Condition
type Condition struct {
	Field string
	Op    ConditionOp
	Val   string
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *Condition) GetQueryPart(idx int) string {
	switch c.Op {
	case IsNotNullOp:
		return fmt.Sprintf("%s %s", c.Field, c.Op)
	case InOp:
		return fmt.Sprintf("%s %s (%s)", c.Field, c.Op, c.Val)
	default:
		return fmt.Sprintf("%s %s $%d", c.Field, c.Op, idx)
	}
}

// GetQueryArg returns a boolean flag if the condition contains argument and the argument value
//
// For conditions like IN and IS NOT NULL there are no arguments to be included in the query.
// For conditions like = there is a placeholder whcih value will be returned calling this func.
func (c *Condition) GetQueryArg() (string, bool) {
	switch c.Op {
	case IsNotNullOp, InOp:
		return "", false
	default:
		return c.Val, true
	}
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

func NewInCondition(field, val string) Condition {
	return Condition{
		Field: field,
		Val:   val,
		Op:    InOp,
	}
}

func getAllArgs(conditions Conditions) []interface{} {
	allArgs := []interface{}{}

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
