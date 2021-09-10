package labelfilter

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"

// LabelFilter missing godoc
type LabelFilter struct {
	Key   string
	Query *string
}

// FromGraphQL missing godoc
func FromGraphQL(in *graphql.LabelFilter) *LabelFilter {
	return &LabelFilter{
		Key:   in.Key,
		Query: in.Query,
	}
}

// MultipleFromGraphQL missing godoc
func MultipleFromGraphQL(in []*graphql.LabelFilter) []*LabelFilter {
	filters := make([]*LabelFilter, 0, len(in))

	for _, f := range in {
		filters = append(filters, FromGraphQL(f))
	}

	return filters
}

// NewForKey missing godoc
func NewForKey(key string) *LabelFilter {
	return &LabelFilter{key, nil}
}

// NewForKeyWithQuery missing godoc
func NewForKeyWithQuery(key, query string) *LabelFilter {
	return &LabelFilter{key, &query}
}
