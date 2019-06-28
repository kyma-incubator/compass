package api_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	modelAPIDefinition := fixDetailedModelAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")
	gqlAPIDefinition := fixDetailedGQLAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")
	emptyModelAPIDefinition := &model.APIDefinition{}
	emptyGraphQLAPIDefinition := &graphql.APIDefinition{}

	testCases := []struct {
		Name                  string
		Input                 *model.APIDefinition
		Expected              *graphql.APIDefinition
		AuthConverterFn       func() *automock.AuthConverter
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    modelAPIDefinition,
			Expected: gqlAPIDefinition,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelAPIDefinition.DefaultAuth).Return(gqlAPIDefinition.DefaultAuth).Once()
				conv.On("ToGraphQL", modelAPIDefinition.Auths[0].Auth).Return(gqlAPIDefinition.Auths[0].Auth).Once()
				conv.On("ToGraphQL", modelAPIDefinition.Auths[1].Auth).Return(gqlAPIDefinition.Auths[1].Auth).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", modelAPIDefinition.Spec.FetchRequest).Return(gqlAPIDefinition.Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyModelAPIDefinition,
			Expected: emptyGraphQLAPIDefinition,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", emptyModelAPIDefinition.DefaultAuth).Return(nil).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := api.NewConverter(authConverter, frConverter)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
			authConverter.AssertExpectations(t)
			frConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.APIDefinition{
		fixModelAPIDefinition("foo", "1", "Foo", "Lorem ipsum"),
		fixModelAPIDefinition("bar", "1", "Bar", "Dolor sit amet"),
		{},
		nil,
	}

	expected := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", "1", "Foo", "Lorem ipsum"),
		fixGQLAPIDefinition("bar", "1", "Bar", "Dolor sit amet"),
		{},
	}

	authConverter := &automock.AuthConverter{}
	frConverter := &automock.FetchRequestConverter{}

	authConverter.On("ToGraphQL", input[0].DefaultAuth).Return(expected[0].DefaultAuth).Once()
	authConverter.On("ToGraphQL", input[1].DefaultAuth).Return(expected[1].DefaultAuth).Once()
	authConverter.On("ToGraphQL", input[2].DefaultAuth).Return(nil).Once()

	// when
	converter := api.NewConverter(authConverter, frConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	authConverter.AssertExpectations(t)
	frConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput("foo", "Lorem ipsum", "group")
	modelAPIDefinitionInput := fixModelAPIDefinitionInput("foo", "Lorem ipsum", "group")
	emptyGQLAPIDefinition := &graphql.APIDefinitionInput{}
	emptyModelAPIDefinition := &model.APIDefinitionInput{}
	testCases := []struct {
		Name                  string
		Input                 *graphql.APIDefinitionInput
		Expected              *model.APIDefinitionInput
		AuthConverterFn       func() *automock.AuthConverter
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlAPIDefinitionInput,
			Expected: modelAPIDefinitionInput,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.DefaultAuth).Return(modelAPIDefinitionInput.DefaultAuth).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.Spec.FetchRequest).Return(modelAPIDefinitionInput.Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    &graphql.APIDefinitionInput{},
			Expected: &model.APIDefinitionInput{},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", emptyGQLAPIDefinition.DefaultAuth).Return(emptyModelAPIDefinition.DefaultAuth).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := api.NewConverter(authConverter, frConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConverter.AssertExpectations(t)
			frConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	gqlAPI1 := fixGQLAPIDefinitionInput("foo", "lorem", "group")
	gqlAPI2 := fixGQLAPIDefinitionInput("bar", "ipsum", "group2")

	modelAPI1 := fixModelAPIDefinitionInput("foo", "lorem", "group")
	modelAPI2 := fixModelAPIDefinitionInput("bar", "ipsum", "group2")

	gqlAPIDefinitionInputs := []*graphql.APIDefinitionInput{gqlAPI1, gqlAPI2}
	modelAPIDefinitionInputs := []*model.APIDefinitionInput{modelAPI1, modelAPI2}
	testCases := []struct {
		Name                  string
		Input                 []*graphql.APIDefinitionInput
		Expected              []*model.APIDefinitionInput
		AuthConverterFn       func() *automock.AuthConverter
		FetchRequestConverter func() *automock.FetchRequestConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlAPIDefinitionInputs,
			Expected: modelAPIDefinitionInputs,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[0].DefaultAuth).Return(modelAPIDefinitionInputs[0].DefaultAuth).Once()
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[1].DefaultAuth).Return(modelAPIDefinitionInputs[1].DefaultAuth).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[0].Spec.FetchRequest).Return(modelAPIDefinitionInputs[0].Spec.FetchRequest).Once()
				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[1].Spec.FetchRequest).Return(modelAPIDefinitionInputs[1].Spec.FetchRequest).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    []*graphql.APIDefinitionInput{},
			Expected: nil,
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			authConverter := testCase.AuthConverterFn()
			frConverter := testCase.FetchRequestConverter()

			// when
			converter := api.NewConverter(authConverter, frConverter)
			res := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			authConverter.AssertExpectations(t)
			frConverter.AssertExpectations(t)
		})
	}
}
