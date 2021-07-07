package repo

import (
	"fmt"
	"strings"
)

type Condition interface {
	// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
	GetQueryPart() string

	// GetQueryArg returns a boolean flag if the condition contains argument and the argument value
	//
	// For conditions like IN and IS NOT NULL there are no arguments to be included in the query.
	// For conditions like = there is a placeholder which value will be returned calling this func.
	GetQueryArgs() ([]interface{}, bool)
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

func (c *equalCondition) GetQueryPart() string {
	return fmt.Sprintf("%s = ?", c.field)
}

func (c *equalCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
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

func (c *notEqualCondition) GetQueryPart() string {
	return fmt.Sprintf("%s != ?", c.field)
}

func (c *notEqualCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

func NewNotNullCondition(field string) Condition {
	return &notNullCondition{
		field: field,
	}
}

type notNullCondition struct {
	field string
}

func (c *notNullCondition) GetQueryPart() string {
	return fmt.Sprintf("%s IS NOT NULL", c.field)
}

func (c *notNullCondition) GetQueryArgs() ([]interface{}, bool) {
	return nil, false
}

func NewInConditionForSubQuery(field, subQuery string, args []interface{}) Condition {
	return &inCondition{
		field:       field,
		parenthesis: subQuery,
		args:        args,
	}
}

type inCondition struct {
	field       string
	parenthesis string
	args        []interface{}
}

func (c *inCondition) GetQueryPart() string {
	return fmt.Sprintf("%s IN (%s)", c.field, c.parenthesis)
}

func (c *inCondition) GetQueryArgs() ([]interface{}, bool) {
	return c.args, true
}

func NewInConditionForStringValues(field string, values []string) Condition {
	var parenthesisParams []string
	var args []interface{}
	for _, value := range values {
		parenthesisParams = append(parenthesisParams, "?")
		args = append(args, value)
	}

	return &inCondition{
		field:       field,
		args:        args,
		parenthesis: strings.Join(parenthesisParams, ", "),
	}
}

type notRegexCondition struct {
	field string
	value string
}

func (c *notRegexCondition) GetQueryPart() string {
	return fmt.Sprintf("NOT %s ~ ?", c.field)
}

func (c *notRegexCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.value}, true
}

func NewNotRegexConditionString(field string, value string) Condition {
	return &notRegexCondition{
		field: field,
		value: value,
	}
}

func NewJSONCondition(field string, val interface{}) Condition {
	return &jsonCondition{
		field: field,
		val:   val,
	}
}

type jsonCondition struct {
	field string
	val   interface{}
}

func (c *jsonCondition) GetQueryPart() string {
	return fmt.Sprintf("%s @> ?", c.field)
}

func (c *jsonCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

type jsonArrAnyMatchCondition struct {
	field string
	val   []interface{}
}

func NewJSONArrAnyMatchCondition(field string, val []interface{}) Condition {
	return &jsonArrAnyMatchCondition{
		field: field,
		val:   val,
	}
}

func NewJSONArrMatchAnyStringCondition(field string, values ...string) Condition {
	valuesInterfaceSlice := make([]interface{}, 0, len(values))
	for _, v := range values {
		valuesInterfaceSlice = append(valuesInterfaceSlice, v)
	}

	return NewJSONArrAnyMatchCondition(field, valuesInterfaceSlice)
}

func (c *jsonArrAnyMatchCondition) GetQueryPart() string {
	valHolders := make([]string, 0, len(c.val))
	for range c.val {
		valHolders = append(valHolders, "?")
	}

	return fmt.Sprintf("%s ?| array[%s]", c.field, strings.Join(valHolders, ","))
}

func (c *jsonArrAnyMatchCondition) GetQueryArgs() ([]interface{}, bool) {
	return c.val, true
}

// NewTenantIsolationCondition returns a repo condition filtering all resources based on the tenant provided (recursively iterating over all the child tenants)
func NewTenantIsolationCondition(field string, val interface{}) Condition {
	return NewTenantIsolationConditionWithPlaceholder(field,"?", []interface{}{val})
}

// NewTenantIsolationConditionWithPlaceholder return tenant isolation repo condition with different tenant_id input placeholder.
// This is needed for update sql queries where the tenant_id to search for is not provided as an argument, but is used from the entity being updated.
func NewTenantIsolationConditionWithPlaceholder(field string, placeholder string, args []interface{}) Condition {
	const query = `
with recursive children AS
(SELECT t1.id, t1.parent
    FROM business_tenant_mappings t1
    WHERE id = %s
    UNION ALL
    SELECT t2.id, t2.parent
    FROM business_tenant_mappings t2
    INNER JOIN children t on t.id = t2.parent)
SELECT id from children
`
	return NewInConditionForSubQuery(field, fmt.Sprintf(query, placeholder), args)
}
