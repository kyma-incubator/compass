package labelfilter

import "github.com/kyma-incubator/compass/components/director/internal/graphql"

type LabelFilter struct {
	Label    string
	Values   []string
	Operator FilterOperator
}

type FilterOperator string

const (
	FilterOperatorAll FilterOperator = "ALL"
	FilterOperatorAny FilterOperator = "ANY"
)

func FromGraphQL(in *graphql.LabelFilter) *LabelFilter {
	return &LabelFilter{
		Values:   in.Values,
		Label:    in.Label,
		Operator: convertFilterOperatorFromGraphQL(in.Operator),
	}
}

func MultipleFromGraphQL(in []*graphql.LabelFilter) []*LabelFilter {
	var filters []*LabelFilter

	for _, f := range in {
		filters = append(filters, FromGraphQL(f))
	}

	return filters
}

func convertFilterOperatorFromGraphQL(in *graphql.FilterOperator) FilterOperator {
	if in == nil {
		return FilterOperatorAll
	}

	switch *in {
	case graphql.FilterOperatorAny:
		return FilterOperatorAny
	case graphql.FilterOperatorAll:
		return FilterOperatorAll
	}

	return FilterOperatorAll
}
