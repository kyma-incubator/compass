package runtime_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	allDetailsInput := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	allDetailsExpected := fixDetailedGQLRuntime(t, "foo", "Foo", "Lorem ipsum")

	// given
	testCases := []struct {
		Name     string
		Input    *model.Runtime
		Expected *graphql.Runtime
	}{
		{
			Name:     "All properties given",
			Input:    allDetailsInput,
			Expected: allDetailsExpected,
		},
		{
			Name:  "Empty",
			Input: &model.Runtime{},
			Expected: &graphql.Runtime{
				Status: &graphql.RuntimeStatus{
					Condition: graphql.RuntimeStatusConditionInitial,
				},
				Metadata: &graphql.RuntimeMetadata{
					CreationTimestamp: graphql.Timestamp{},
				},
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := runtime.NewConverter()
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Runtime{
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem ipsum"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Dolor sit amet"),
		{},
		nil,
	}
	expected := []*graphql.Runtime{
		fixGQLRuntime(t, "foo", "Foo", "Lorem ipsum"),
		fixGQLRuntime(t, "bar", "Bar", "Dolor sit amet"),
		{
			Status: &graphql.RuntimeStatus{
				Condition: graphql.RuntimeStatusConditionInitial,
			},
			Metadata: &graphql.RuntimeMetadata{
				CreationTimestamp: graphql.Timestamp{},
			},
		},
	}

	// when
	converter := runtime.NewConverter()
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// when
			converter := runtime.NewConverter()
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
