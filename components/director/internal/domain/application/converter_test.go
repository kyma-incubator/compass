package application_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Application
		Expected *graphql.Application
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedModelApplication(t, "foo", "Foo", "Lorem ipsum"),
			Expected: fixDetailedGQLApplication(t, "foo", "Foo", "Lorem ipsum"),
		},
		{
			Name:  "Empty",
			Input: &model.Application{},
			Expected: &graphql.Application{
				Status: &graphql.ApplicationStatus{
					Condition: graphql.ApplicationStatusConditionInitial,
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
			converter := application.NewConverter()
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Application{
		fixModelApplication("foo", "Foo", "Lorem ipsum"),
		fixModelApplication("bar", "Bar", "Dolor sit amet"),
		{},
		nil,
	}
	expected := []*graphql.Application{
		fixGQLApplication("foo", "Foo", "Lorem ipsum"),
		fixGQLApplication("bar", "Bar", "Dolor sit amet"),
		{
			Status: &graphql.ApplicationStatus{
				Condition: graphql.ApplicationStatusConditionInitial,
			},
		},
	}

	// when
	converter := application.NewConverter()
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    graphql.ApplicationInput
		Expected model.ApplicationInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLApplicationInput("foo", "Lorem ipsum"),
			Expected: fixModelApplicationInput("foo", "Lorem ipsum"),
		},
		{
			Name:     "Empty",
			Input:    graphql.ApplicationInput{},
			Expected: model.ApplicationInput{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			converter := application.NewConverter()
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
