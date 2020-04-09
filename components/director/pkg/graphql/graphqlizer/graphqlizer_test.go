package graphqlizer_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraphqlizer_LabelsToGQL(t *testing.T) {
	// GIVEN
	g := graphqlizer.Graphqlizer{}

	testCases := []struct {
		Name          string
		Input         graphql.Labels
		Expected      string
		ExpectedError error
	}{
		{
			Name: "Success when slice of strings",
			Input: graphql.Labels{
				"foo": []string{"test", "best", "asdf"},
			},
			Expected:      "{foo:[\"test\",\"best\",\"asdf\"],}",
			ExpectedError: nil,
		},
		{
			Name: "Success when string",
			Input: graphql.Labels{
				"bar": "test",
			},
			Expected:      "{bar:\"test\",}",
			ExpectedError: nil,
		},
		{
			Name: "Success when bool",
			Input: graphql.Labels{
				"baz": true,
			},
			Expected:      "{baz:true,}",
			ExpectedError: nil,
		},
		{
			Name: "Success when map of strings",
			Input: graphql.Labels{
				"biz": map[string]string{"test": "a", "best": "b", "asdf": "c"},
			},
			Expected:      "{biz:{asdf:\"c\",best:\"b\",test:\"a\",},}",
			ExpectedError: nil,
		},
		{
			Name: "Success when number",
			Input: graphql.Labels{
				"buz": 10,
			},
			Expected:      "{buz:10,}",
			ExpectedError: nil,
		},
		{
			Name: "Success when mixed",
			Input: graphql.Labels{
				"foo": []string{"test", "best", "asdf"},
				"bar": "test",
				"baz": true,
				"biz": map[string]string{"test": "a", "best": "b", "asdf": "c"},
				"buz": 10,
			},
			Expected:      "{bar:\"test\",baz:true,biz:{asdf:\"c\",best:\"b\",test:\"a\",},buz:10,foo:[\"test\",\"best\",\"asdf\"],}",
			ExpectedError: nil,
		},
		{
			Name: "Success when nested iterables",
			Input: graphql.Labels{
				"foo": []interface{}{"test", map[string]string{"asdf": "fdsa", "fdsa": "asdf"}, []string{"aaaa", "bbbb"}},
			},
			Expected:      "{foo:[\"test\",{asdf:\"fdsa\",fdsa:\"asdf\",},[\"aaaa\",\"bbbb\"]],}",
			ExpectedError: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			result, err := g.LabelsToGQL(testCase.Input)
			// THEN
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, result)
		})
	}
}
