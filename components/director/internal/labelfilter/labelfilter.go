package labelfilter

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

type LabelFilter struct {
	Label string
	Query *string
}

func FromGraphQL(in *graphql.LabelFilter) *LabelFilter {
	return &LabelFilter{
		Label: in.Label,
		Query: in.Query,
	}
}

func MultipleFromGraphQL(in []*graphql.LabelFilter) []*LabelFilter {
	var filters []*LabelFilter

	for _, f := range in {
		filters = append(filters, FromGraphQL(f))
	}

	return filters
}
