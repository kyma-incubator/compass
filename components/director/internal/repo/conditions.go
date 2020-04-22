package repo

import (
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Condition interface {
	// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
	GetQueryPart(idx int) string

	// GetQueryArg returns a boolean flag if the condition contains argument and the argument value
	//
	// For conditions like IN and IS NOT NULL there are no arguments to be included in the query.
	// For conditions like = there is a placeholder which value will be returned calling this func.
	GetQueryArg() (interface{}, bool)
}

type Conditions []Condition

func NewEqualCondition(field string, val interface{}) Condition {
	return &equalCondition{
		field: field,
		val:   val,
	}
}

type equalCondition struct {
	field string
	val   interface{}
}

func (c *equalCondition) GetQueryPart(idx int) string {
	return fmt.Sprintf("%s = $%d", c.field, idx)
}

func (c *equalCondition) GetQueryArg() (interface{}, bool) {
	return c.val, true
}

func NewNotEqualCondition(field string, val interface{}) Condition {
	return &notEqualCondition{
		field: field,
		val:   val,
	}
}

type notEqualCondition struct {
	field string
	val   interface{}
}

func (c *notEqualCondition) GetQueryPart(idx int) string {
	return fmt.Sprintf("%s != $%d", c.field, idx)
}

func (c *notEqualCondition) GetQueryArg() (interface{}, bool) {
	return c.val, true
}

func NewNotNullCondition(field string) Condition {
	return &notNullCondition{
		field: field,
	}
}

type notNullCondition struct {
	field string
}

func (c *notNullCondition) GetQueryPart(idx int) string {
	return fmt.Sprintf("%s IS NOT NULL", c.field)
}

func (c *notNullCondition) GetQueryArg() (interface{}, bool) {
	return nil, false
}

func NewInConditionForSubQuery(field, subQuery string) Condition {
	return &inCondition{
		field:       field,
		parenthesis: subQuery,
	}
}

func NewInConditionForStringValues(field string, values []string) Condition {
	var escapedValues []string
	for _, value := range values {
		escapedValues = append(escapedValues, pq.QuoteLiteral(value))
	}

	return &inCondition{
		field:       field,
		parenthesis: strings.Join(escapedValues, ", "),
	}
}

type inCondition struct {
	field       string
	parenthesis string
}

func (c *inCondition) GetQueryPart(idx int) string {
	return fmt.Sprintf("%s IN (%s)", c.field, c.parenthesis)
}

func (c *inCondition) GetQueryArg() (interface{}, bool) {
	return nil, false
}
