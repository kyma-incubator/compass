package repo

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type ConditionOp string

const (
	EqualOp     ConditionOp = "="
	IsNotNullOp ConditionOp = "IS NOT NULL"
	InOp        ConditionOp = "IN"
	NotEqualOP  ConditionOp = "!="
)

type Conditions []Condition
type Condition struct {
	Field string
	Op    ConditionOp
	Val   interface{}
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
// For conditions like = there is a placeholder which value will be returned calling this func.
func (c *Condition) GetQueryArg() (interface{}, bool) {
	switch c.Op {
	case IsNotNullOp, InOp:
		return "", false
	default:
		return c.Val, true
	}
}

func NewEqualCondition(field string, val interface{}) Condition {
	return Condition{
		Field: field,
		Val:   val,
		Op:    EqualOp,
	}
}

func NewNotEqualCondition(field string, val interface{}) Condition {
	return Condition{
		Field: field,
		Val:   val,
		Op:    NotEqualOP,
	}
}

func NewNotNullCondition(field string) Condition {
	return Condition{
		Field: field,
		Op:    IsNotNullOp,
	}
}

func NewInConditionForSubQuery(field, subQuery string) Condition {
	return Condition{
		Field: field,
		Val:   subQuery,
		Op:    InOp,
	}
}

func NewInConditionForStringValues(field string, values []string) Condition {
	var escapedValues []string
	for _, value := range values {
		escapedValues = append(escapedValues, pq.QuoteLiteral(value))
	}

	return Condition{
		Field: field,
		Val:   strings.Join(escapedValues, ", "),
		Op:    InOp,
	}
}
