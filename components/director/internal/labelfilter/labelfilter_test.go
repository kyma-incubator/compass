package labelfilter_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestFromGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		inOp := graphql.FilterOperatorAny
		in := &graphql.LabelFilter{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: &inOp,
		}

		expected := &labelfilter.LabelFilter{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: labelfilter.FilterOperatorAny,
		}

		result := labelfilter.FromGraphQL(in)

		assert.Equal(t, expected, result)
	})

	t.Run("Default value", func(t *testing.T) {
		in := &graphql.LabelFilter{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: nil,
		}

		expected := &labelfilter.LabelFilter{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: labelfilter.FilterOperatorAll,
		}

		result := labelfilter.FromGraphQL(in)

		assert.Equal(t, expected, result)
	})
}

func TestMultipleFromGraphQL(t *testing.T) {
	inOp := graphql.FilterOperatorAll
	in := []*graphql.LabelFilter{
		{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: &inOp,
		},
		{
			Label: "label2",
			Values: []string{
				"test",
			},
			Operator: &inOp,
		},
	}

	expected := []*labelfilter.LabelFilter{
		{
			Label: "label",
			Values: []string{
				"foo",
				"bar",
			},
			Operator: labelfilter.FilterOperatorAll,
		},
		{
			Label: "label2",
			Values: []string{
				"test",
			},
			Operator: labelfilter.FilterOperatorAll,
		},
	}

	result := labelfilter.MultipleFromGraphQL(in)

	assert.Equal(t, expected, result)
}
