package labelfilter_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestFromGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		query := "foo"
		in := &graphql.LabelFilter{
			Key:   "label",
			Query: &query,
		}

		expected := &labelfilter.LabelFilter{
			Key:   "label",
			Query: &query,
		}

		result := labelfilter.FromGraphQL(in)

		assert.Equal(t, expected, result)
	})

	t.Run("Empty query", func(t *testing.T) {
		in := &graphql.LabelFilter{
			Key:   "label",
			Query: nil,
		}

		expected := &labelfilter.LabelFilter{
			Key:   "label",
			Query: nil,
		}

		result := labelfilter.FromGraphQL(in)

		assert.Equal(t, expected, result)
	})
}

func TestMultipleFromGraphQL(t *testing.T) {
	queryFoo := "foo"
	queryBar := "bar"
	in := []*graphql.LabelFilter{
		{
			Key:   "label",
			Query: &queryFoo,
		},
		{
			Key:   "label2",
			Query: &queryBar,
		},
	}

	expected := []*labelfilter.LabelFilter{
		{
			Key:   "label",
			Query: &queryFoo,
		},
		{
			Key:   "label2",
			Query: &queryBar,
		},
	}

	result := labelfilter.MultipleFromGraphQL(in)

	assert.Equal(t, expected, result)
}
