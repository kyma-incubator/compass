package eventapi_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	modelEventAPIDefinition := fixDetailedModelEventAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")
	gqlEventAPIDefinition := fixDetailedGQLEventAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")
	emptyModelEventAPIDefinition := &model.EventAPIDefinition{}
	emptyGraphQLEventAPIDefinition := &graphql.EventAPIDefinition{}

	testCases := []struct {
		Name                  string
		Input                 *model.EventAPIDefinition
		Expected              *graphql.EventAPIDefinition
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    modelEventAPIDefinition,
			Expected: gqlEventAPIDefinition,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", modelEventAPIDefinition.Spec.FetchRequest).Return(gqlEventAPIDefinition.Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyModelEventAPIDefinition,
			Expected: emptyGraphQLEventAPIDefinition,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := eventapi.NewConverter(frConverter)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.EventAPIDefinition{
		fixModelEventAPIDefinition("foo", "Foo", "Lorem ipsum", "desc"),
		fixModelEventAPIDefinition("bar", "Bar", "Dolor sit amet", "desc"),
		{},
		nil,
	}

	expected := []*graphql.EventAPIDefinition{
		fixGQLEventAPIDefinition("foo", "Foo", "Lorem ipsum", "desc"),
		fixGQLEventAPIDefinition("bar", "Bar", "Dolor sit amet", "desc"),
		{},
	}

	frConverter := &automock.FetchRequestConverter{}

	// when
	converter := eventapi.NewConverter(frConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlEventAPIDefinitionInput := fixGQLEventAPIDefinitionInput("foo", "Lorem ipsum", "group")
	modelEventAPIDefinitionInput := fixModelEventAPIDefinitionInput("foo", "Lorem ipsum", "group")
	emptyGQLEventAPIDefinition := &graphql.EventAPIDefinitionInput{}
	emptyModelEventAPIDefinition := &model.EventAPIDefinitionInput{}
	testCases := []struct {
		Name                  string
		Input                 *graphql.EventAPIDefinitionInput
		Expected              *model.EventAPIDefinitionInput
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlEventAPIDefinitionInput,
			Expected: modelEventAPIDefinitionInput,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", gqlEventAPIDefinitionInput.Spec.FetchRequest).Return(modelEventAPIDefinitionInput.Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyGQLEventAPIDefinition,
			Expected: emptyModelEventAPIDefinition,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := eventapi.NewConverter(frConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	gqlApi1 := fixGQLEventAPIDefinitionInput("foo", "lorem", "group")
	gqlApi2 := fixGQLEventAPIDefinitionInput("bar", "ipsum", "group2")

	modelApi1 := fixModelEventAPIDefinitionInput("foo", "lorem", "group")
	modelApi2 := fixModelEventAPIDefinitionInput("bar", "ipsum", "group2")

	gqlAPIDefinitionInputs := []*graphql.EventAPIDefinitionInput{gqlApi1, gqlApi2}
	modelAPIDefinitionInputs := []*model.EventAPIDefinitionInput{modelApi1, modelApi2}
	testCases := []struct {
		Name                  string
		Input                 []*graphql.EventAPIDefinitionInput
		Expected              []*model.EventAPIDefinitionInput
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlAPIDefinitionInputs,
			Expected: modelAPIDefinitionInputs,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[0].Spec.FetchRequest).Return(modelAPIDefinitionInputs[0].Spec.FetchRequest).Once()
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[1].Spec.FetchRequest).Return(modelAPIDefinitionInputs[1].Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    []*graphql.EventAPIDefinitionInput{},
			Expected: nil,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := eventapi.NewConverter(frConverter)
			res := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
		})
	}
}
