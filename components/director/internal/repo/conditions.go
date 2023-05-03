package repo

import (
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// Condition represents an SQL condition
type Condition interface {
	// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
	GetQueryPart() string

	// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
	//
	// For conditions like IN and IS NOT NULL there are no arguments to be included in the query.
	// For conditions like = there is a placeholder which value will be returned calling this func.
	GetQueryArgs() ([]interface{}, bool)
}

// Conditions is a slice of conditions
type Conditions []Condition

// Operator represents an SQL operator
type Operator string

const (

	// OR represents an OR operator
	OR Operator = "OR"

	// AND represents an AND operator
	AND Operator = "AND"
)

// ConditionTree represents a tree of conditions. The tree is constructed bottom to top. Leafs are the actual conditions. Intermediate nodes are operators joining them.
type ConditionTree struct {
	// Operator represents an SQL operator used to join the children conditions. This is populated only for non-leaf nodes.
	Operator Operator
	// Operand is the actual condition. This is populated only for leaf nodes.
	Operand Condition
	// Children is a slice of conditions. This is populated only for non-leaf nodes.
	Children []*ConditionTree
}

// IsLeaf returns true if the node is a leaf node
func (t *ConditionTree) IsLeaf() bool {
	return len(t.Children) == 0
}

// BuildSubquery builds SQL subquery for a given condition tree
func (t *ConditionTree) BuildSubquery() (string, []interface{}) {
	if t.IsLeaf() {
		args, ok := t.Operand.GetQueryArgs()
		if !ok {
			args = []interface{}{}
		}
		return t.Operand.GetQueryPart(), args
	}

	args := make([]interface{}, 0)
	queryParts := make([]string, 0, len(t.Children))
	for _, child := range t.Children {
		queryPart, childArgs := child.BuildSubquery()
		queryParts = append(queryParts, queryPart)
		args = append(args, childArgs...)
	}

	sql := fmt.Sprintf("(%s)", strings.Join(queryParts, fmt.Sprintf(" %s ", t.Operator)))
	return sql, args
}

// ConditionTreesFromConditions builds a tree of conditions from a slice of conditions. The tree is constructed bottom to top.
func ConditionTreesFromConditions(conditions Conditions) []*ConditionTree {
	if len(conditions) == 0 {
		return nil
	}
	children := make([]*ConditionTree, 0, len(conditions))
	for i := range conditions {
		children = append(children, &ConditionTree{Operand: conditions[i]})
	}
	return children
}

// And joins given conditions with AND operator
func And(children ...*ConditionTree) *ConditionTree {
	if len(children) == 0 {
		return nil
	}
	if len(children) == 1 {
		return children[0]
	}
	return &ConditionTree{
		Operator: AND,
		Children: children,
	}
}

// Or joins given conditions with OR operator
func Or(children ...*ConditionTree) *ConditionTree {
	if len(children) == 0 {
		return nil
	}
	if len(children) == 1 {
		return children[0]
	}
	return &ConditionTree{
		Operator: OR,
		Children: children,
	}
}

// NewEqualCondition represents equal SQL condition (field = val)
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

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *equalCondition) GetQueryPart() string {
	return fmt.Sprintf("%s = ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *equalCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

// NewLessThanCondition represents less than SQL condition (field < val)
func NewLessThanCondition(field string, val interface{}) Condition {
	return &lessThanCondition{
		field: field,
		val:   val,
	}
}

type lessThanCondition struct {
	field string
	val   interface{}
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *lessThanCondition) GetQueryPart() string {
	return fmt.Sprintf("%s < ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *lessThanCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

// NewNotEqualCondition represents not equal SQL condition (field != val)
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

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *notEqualCondition) GetQueryPart() string {
	return fmt.Sprintf("%s != ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *notEqualCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

// NewNotNullCondition represents SQL not null condition (field IS NOT NULL)
func NewNotNullCondition(field string) Condition {
	return &notNullCondition{
		field: field,
	}
}

type notNullCondition struct {
	field string
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *notNullCondition) GetQueryPart() string {
	return fmt.Sprintf("%s IS NOT NULL", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *notNullCondition) GetQueryArgs() ([]interface{}, bool) {
	return nil, false
}

// NewNullCondition represents SQL null condition (field IS NULL)
func NewNullCondition(field string) Condition {
	return &nullCondition{
		field: field,
	}
}

type nullCondition struct {
	field string
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *nullCondition) GetQueryPart() string {
	return fmt.Sprintf("%s IS NULL", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *nullCondition) GetQueryArgs() ([]interface{}, bool) {
	return nil, false
}

// NewLikeCondition represents SQL like condition (field like val)
func NewLikeCondition(field string, val interface{}) Condition {
	return &likeCondition{
		field: field,
		val:   val,
	}
}

type likeCondition struct {
	field string
	val   interface{}
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *likeCondition) GetQueryPart() string {
	return fmt.Sprintf("%s ILIKE ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *likeCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

// NewInConditionForSubQuery represents SQL IN subquery (field IN (SELECT ...))
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

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *inCondition) GetQueryPart() string {
	return fmt.Sprintf("%s IN (%s)", c.field, c.parenthesis)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *inCondition) GetQueryArgs() ([]interface{}, bool) {
	return c.args, true
}

// NewInConditionForStringValues represents SQL IN condition (field IN (?, ?, ...))
func NewInConditionForStringValues(field string, values []string) Condition {
	parenthesisParams := make([]string, 0, len(values))
	args := make([]interface{}, 0, len(values))
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

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *notRegexCondition) GetQueryPart() string {
	return fmt.Sprintf("NOT %s ~ ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *notRegexCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.value}, true
}

// NewNotRegexConditionString represents SQL regex not match condition
func NewNotRegexConditionString(field string, value string) Condition {
	return &notRegexCondition{
		field: field,
		value: value,
	}
}

// NewJSONCondition represents PostgreSQL JSONB contains condition
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

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *jsonCondition) GetQueryPart() string {
	return fmt.Sprintf("%s @> ?", c.field)
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *jsonCondition) GetQueryArgs() ([]interface{}, bool) {
	return []interface{}{c.val}, true
}

type jsonArrAnyMatchCondition struct {
	field string
	val   []interface{}
}

// NewJSONArrAnyMatchCondition represents PostgreSQL JSONB array any element match condition
func NewJSONArrAnyMatchCondition(field string, val []interface{}) Condition {
	return &jsonArrAnyMatchCondition{
		field: field,
		val:   val,
	}
}

// NewJSONArrMatchAnyStringCondition represents PostgreSQL JSONB string array any element match condition
func NewJSONArrMatchAnyStringCondition(field string, values ...string) Condition {
	valuesInterfaceSlice := make([]interface{}, 0, len(values))
	for _, v := range values {
		valuesInterfaceSlice = append(valuesInterfaceSlice, v)
	}

	return NewJSONArrAnyMatchCondition(field, valuesInterfaceSlice)
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *jsonArrAnyMatchCondition) GetQueryPart() string {
	valHolders := make([]string, 0, len(c.val))
	for range c.val {
		valHolders = append(valHolders, "?")
	}

	return fmt.Sprintf("%s ?| array[%s]", c.field, strings.Join(valHolders, ","))
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *jsonArrAnyMatchCondition) GetQueryArgs() ([]interface{}, bool) {
	return c.val, true
}

type tenantIsolationCondition struct {
	sql  string
	args []interface{}
}

// GetQueryPart returns formatted string that will be included in the SQL query for a given condition
func (c *tenantIsolationCondition) GetQueryPart() string {
	return c.sql
}

// GetQueryArgs returns a boolean flag if the condition contain arguments and the actual arguments
func (c *tenantIsolationCondition) GetQueryArgs() ([]interface{}, bool) {
	return c.args, true
}

// NewTenantIsolationCondition is a tenant isolation SQL subquery for entities that have tenant accesses managed outside of
// the entity table (m2m table or view). Conditionally an owner check is added to the subquery.
// In case of resource.BundleInstanceAuth additional embedded owner check is added.
func NewTenantIsolationCondition(resourceType resource.Type, tenant string, ownerCheck bool) (Condition, error) {
	return newTenantIsolationConditionWithPlaceholder(resourceType, tenant, ownerCheck, true)
}

// NewTenantIsolationConditionForNamedArgs is the same as NewTenantIsolationCondition, but for update queries which use named args.
func NewTenantIsolationConditionForNamedArgs(resourceType resource.Type, tenant string, ownerCheck bool) (Condition, error) {
	return newTenantIsolationConditionWithPlaceholder(resourceType, tenant, ownerCheck, false)
}

func newTenantIsolationConditionWithPlaceholder(resourceType resource.Type, tenant string, ownerCheck bool, positionalArgs bool) (Condition, error) {
	m2mTable, ok := resourceType.TenantAccessTable()
	if !ok {
		return nil, errors.Errorf("entity %s does not have access table", resourceType)
	}

	var stmtBuilder strings.Builder
	var args []interface{}

	if positionalArgs {
		stmtBuilder.WriteString(fmt.Sprintf("(id IN (SELECT %s FROM %s WHERE %s = ?", M2MResourceIDColumn, m2mTable, M2MTenantIDColumn))
		args = append(args, tenant)
	} else {
		stmtBuilder.WriteString(fmt.Sprintf("(id IN (SELECT %s FROM %s WHERE %s = :tenant_id", M2MResourceIDColumn, m2mTable, M2MTenantIDColumn))
	}

	if ownerCheck {
		stmtBuilder.WriteString(fmt.Sprintf(" AND %s = true", M2MOwnerColumn))
	}
	stmtBuilder.WriteString(")")

	if resourceType == resource.BundleInstanceAuth {
		if positionalArgs {
			stmtBuilder.WriteString(" OR owner_id = ?")
			args = append(args, tenant)
		} else {
			stmtBuilder.WriteString(" OR owner_id = :owner_id")
		}
	}
	stmtBuilder.WriteString(")")

	return &tenantIsolationCondition{sql: stmtBuilder.String(), args: args}, nil
}
