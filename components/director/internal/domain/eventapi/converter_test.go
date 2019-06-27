package eventapi_test


import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	"testing"

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
//
//func TestConverter_MultipleToGraphQL(t *testing.T) {
//	// given
//	input := []*model.APIDefinition{
//		fixModelAPIDefinition("foo", "Foo", "Lorem ipsum"),
//		fixModelAPIDefinition("bar", "Bar", "Dolor sit amet"),
//		{},
//		nil,
//	}
//
//	expected := []*graphql.APIDefinition{
//		fixGQLAPIDefinition("foo", "Foo", "Lorem ipsum"),
//		fixGQLAPIDefinition("bar", "Bar", "Dolor sit amet"),
//		{},
//	}
//
//	authConverter := &automock.AuthConverter{}
//	frConverter := &automock.FetchRequestConverter{}
//
//	authConverter.On("ToGraphQL", input[0].DefaultAuth).Return(expected[0].DefaultAuth).Once()
//	authConverter.On("ToGraphQL", input[1].DefaultAuth).Return(expected[1].DefaultAuth).Once()
//	authConverter.On("ToGraphQL", input[2].DefaultAuth).Return(nil).Once()
//
//	// when
//	converter := api.NewConverter(authConverter, frConverter)
//	res := converter.MultipleToGraphQL(input)
//
//	// then
//	assert.Equal(t, expected, res)
//	authConverter.AssertExpectations(t)
//	frConverter.AssertExpectations(t)
//}
//
//func TestConverter_InputFromGraphQL(t *testing.T) {
//	// given
//	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput("foo", "Lorem ipsum", "group")
//	modelAPIDefinitionInput := fixModelAPIDefinitionInput("foo", "Lorem ipsum", "group")
//	emptyGQLAPIDefinition := &graphql.APIDefinitionInput{}
//	emptyModelAPIDefinition := &model.APIDefinitionInput{}
//	testCases := []struct {
//		Name                  string
//		Input                 *graphql.APIDefinitionInput
//		Expected              *model.APIDefinitionInput
//		AuthConverterFn       func() *automock.AuthConverter
//		FetchRequestConverter func() *automock.FetchRequestConverter
//	}{
//		{
//			Name:     "All properties given",
//			Input:    gqlAPIDefinitionInput,
//			Expected: modelAPIDefinitionInput,
//			AuthConverterFn: func() *automock.AuthConverter {
//				conv := &automock.AuthConverter{}
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.DefaultAuth).Return(modelAPIDefinitionInput.DefaultAuth).Once()
//				return conv
//			},
//			FetchRequestConverter: func() *automock.FetchRequestConverter {
//				conv := &automock.FetchRequestConverter{}
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.Spec.FetchRequest).Return(modelAPIDefinitionInput.Spec.FetchRequest).Once()
//				return conv
//			},
//		},
//		{
//			Name:     "Empty",
//			Input:    &graphql.APIDefinitionInput{},
//			Expected: &model.APIDefinitionInput{},
//			AuthConverterFn: func() *automock.AuthConverter {
//				conv := &automock.AuthConverter{}
//				conv.On("InputFromGraphQL", emptyGQLAPIDefinition.DefaultAuth).Return(emptyModelAPIDefinition.DefaultAuth).Once()
//				return conv
//			},
//			FetchRequestConverter: func() *automock.FetchRequestConverter {
//				return &automock.FetchRequestConverter{}
//			},
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
//			//given
//			authConverter := testCase.AuthConverterFn()
//			frConverter := testCase.FetchRequestConverter()
//
//			// when
//			converter := api.NewConverter(authConverter, frConverter)
//			res := converter.InputFromGraphQL(testCase.Input)
//
//			// then
//			assert.Equal(t, testCase.Expected, res)
//			authConverter.AssertExpectations(t)
//			frConverter.AssertExpectations(t)
//		})
//	}
//}
//
//func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
//	// given
//	gqlApi1 := fixGQLAPIDefinitionInput("foo", "lorem", "group")
//	gqlApi2 := fixGQLAPIDefinitionInput("bar", "ipsum", "group2")
//
//	modelApi1 := fixModelAPIDefinitionInput("foo", "lorem", "group")
//	modelApi2 := fixModelAPIDefinitionInput("bar", "ipsum", "group2")
//
//	gqlAPIDefinitionInputs := []*graphql.APIDefinitionInput{gqlApi1, gqlApi2}
//	modelAPIDefinitionInputs := []*model.APIDefinitionInput{modelApi1, modelApi2}
//	testCases := []struct {
//		Name                  string
//		Input                 []*graphql.APIDefinitionInput
//		Expected              []*model.APIDefinitionInput
//		AuthConverterFn       func() *automock.AuthConverter
//		FetchRequestConverter func() *automock.FetchRequestConverter
//	}{
//		{
//			Name:     "All properties given",
//			Input:    gqlAPIDefinitionInputs,
//			Expected: modelAPIDefinitionInputs,
//			AuthConverterFn: func() *automock.AuthConverter {
//				conv := &automock.AuthConverter{}
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[0].DefaultAuth).Return(modelAPIDefinitionInputs[0].DefaultAuth).Once()
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[1].DefaultAuth).Return(modelAPIDefinitionInputs[1].DefaultAuth).Once()
//				return conv
//			},
//			FetchRequestConverter: func() *automock.FetchRequestConverter {
//				conv := &automock.FetchRequestConverter{}
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[0].Spec.FetchRequest).Return(modelAPIDefinitionInputs[0].Spec.FetchRequest).Once()
//				conv.On("InputFromGraphQL", gqlAPIDefinitionInputs[1].Spec.FetchRequest).Return(modelAPIDefinitionInputs[1].Spec.FetchRequest).Once()
//				return conv
//			},
//		},
//		{
//			Name:     "Empty",
//			Input:    []*graphql.APIDefinitionInput{},
//			Expected: nil,
//			AuthConverterFn: func() *automock.AuthConverter {
//				return &automock.AuthConverter{}
//			},
//			FetchRequestConverter: func() *automock.FetchRequestConverter {
//				return &automock.FetchRequestConverter{}
//			},
//		},
//	}
//
//	for _, testCase := range testCases {
//		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
//			//given
//			authConverter := testCase.AuthConverterFn()
//			frConverter := testCase.FetchRequestConverter()
//
//			// when
//			converter := api.NewConverter(authConverter, frConverter)
//			res := converter.MultipleInputFromGraphQL(testCase.Input)
//
//			// then
//			assert.Equal(t, testCase.Expected, res)
//			authConverter.AssertExpectations(t)
//			frConverter.AssertExpectations(t)
//		})
//	}
//}