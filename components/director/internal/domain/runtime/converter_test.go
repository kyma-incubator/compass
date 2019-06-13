package runtime_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Runtime
		Expected *graphql.Runtime
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum"),
			Expected: fixDetailedGQLRuntime(t, "foo", "Foo", "Lorem ipsum"),
		},
		{
			Name:  "Empty",
			Input: &model.Runtime{},
			Expected: &graphql.Runtime{
				Status: &graphql.RuntimeStatus{
					Condition: graphql.RuntimeStatusConditionInitial,
				},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			converter := runtime.Converter{}
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Runtime{
		fixModelRuntime("foo", "Foo", "Lorem ipsum"),
		{},
		nil,
	}
	expected := []*graphql.Runtime{
		fixGQLRuntime("foo", "Foo", "Lorem ipsum"),
		{
			Status: &graphql.RuntimeStatus{
				Condition: graphql.RuntimeStatusConditionInitial,
			},
		},
	}

	// when
	converter := runtime.Converter{}
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    graphql.RuntimeInput
		Expected model.RuntimeInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLRuntimeInput("foo", "Lorem ipsum"),
			Expected: fixModelRuntimeInput("foo", "Lorem ipsum"),
		},
		{
			Name:     "Empty",
			Input:    graphql.RuntimeInput{},
			Expected: model.RuntimeInput{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			converter := runtime.Converter{}
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
