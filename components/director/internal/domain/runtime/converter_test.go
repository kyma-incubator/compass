package runtime_test

import (
	"fmt"
	rtmautomock "github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	allDetailsInput := fixDetailedModelRuntime(t, "foo", "Foo", "Lorem ipsum")
	allDetailsExpected := fixDetailedGQLRuntime(t, "foo", "Foo", "Lorem ipsum")
	var modelAuth *model.Auth

	// given
	testCases := []struct {
		Name            string
		Input           *model.Runtime
		Expected        *graphql.Runtime
		AuthConverterFn func() *rtmautomock.AuthConverter
	}{
		{
			Name:     "All properties given",
			Input:    allDetailsInput,
			Expected: allDetailsExpected,
			AuthConverterFn: func() *rtmautomock.AuthConverter {
				conv := &rtmautomock.AuthConverter{}
				conv.On("ToGraphQL", allDetailsInput.AgentAuth).Return(allDetailsExpected.AgentAuth).Once()
				return conv
			},
		},
		{
			Name:  "Empty",
			Input: &model.Runtime{},
			Expected: &graphql.Runtime{
				Status: &graphql.RuntimeStatus{
					Condition: graphql.RuntimeStatusConditionInitial,
				},
			},
			AuthConverterFn: func() *rtmautomock.AuthConverter {
				conv := &rtmautomock.AuthConverter{}
				conv.On("ToGraphQL", modelAuth).Return(nil).Once()
				return conv
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			AuthConverterFn: func() *rtmautomock.AuthConverter {
				conv := &rtmautomock.AuthConverter{}
				return conv
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			authConverter := testCase.AuthConverterFn()

			// when
			converter := runtime.NewConverter(authConverter)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)

			authConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	var modelAuth *model.Auth
	authConverter := &rtmautomock.AuthConverter{}
	authConverter.On("ToGraphQL", modelAuth).Return(nil)

	input := []*model.Runtime{
		fixModelRuntime("foo", "Foo", "Lorem ipsum"),
		fixModelRuntime("bar", "Bar", "Dolor sit amet"),
		{},
		nil,
	}
	expected := []*graphql.Runtime{
		fixGQLRuntime("foo", "Foo", "Lorem ipsum"),
		fixGQLRuntime("bar", "Bar", "Dolor sit amet"),
		{
			Status: &graphql.RuntimeStatus{
				Condition: graphql.RuntimeStatusConditionInitial,
			},
		},
	}

	// when
	converter := runtime.NewConverter(authConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	authConverter := &rtmautomock.AuthConverter{}
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
			converter := runtime.NewConverter(authConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
