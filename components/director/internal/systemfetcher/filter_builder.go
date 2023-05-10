package systemfetcher

import (
	"strings"
)

// Expression represents the single check expression
type Expression struct {
	Key       string
	Operation string
	Value     string
}

func (e *Expression) getValue() string {
	return "'" + e.Value + "'"
}

func (e *Expression) buildExpression() string {
	return strings.Join([]string{e.Key, e.Operation, e.getValue()}, " ")
}

// Filter is the set of single expressions which will be concatenated when building the query
type Filter struct {
	Expressions []Expression
}

func (f *Filter) addExpression(e Expression) {
	f.Expressions = append(f.Expressions, e)
}

func (f *Filter) buildFilter() string {
	var filter strings.Builder

	for i, expr := range f.Expressions {
		filter.WriteString(expr.buildExpression())

		if i < len(f.Expressions)-1 {
			filter.WriteString(" and ")
		}
	}

	return filter.String()
}

// FilterBuilder builds the filter query from given expressions
type FilterBuilder struct {
	Filters []Filter
}

// NewExpression returns new filter expression based on the given arguments
func (b *FilterBuilder) NewExpression(key, operation, value string) Expression {
	return Expression{
		Key:       key,
		Operation: operation,
		Value:     value,
	}
}

func (b *FilterBuilder) addFilter(expr ...Expression) {
	var newFilter Filter

	for _, e := range expr {
		newFilter.addExpression(e)
	}

	b.Filters = append(b.Filters, newFilter)
}

func (b *FilterBuilder) buildFilterQuery() string {
	var filterQuery strings.Builder

	for i, f := range b.Filters {
		filterQuery.WriteString("(" + f.buildFilter() + ")")

		if i < len(b.Filters)-1 {
			filterQuery.WriteString(" or ")
		}
	}

	return filterQuery.String()
}
