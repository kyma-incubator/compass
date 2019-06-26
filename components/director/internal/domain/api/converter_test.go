package api_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given


	//data := graphql.CLOB("")
	modelAPIDefinition := fixDetailedModelAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")
	gqlAPIDefinition := fixDetailedGQLAPIDefinition(t, "foo", "Foo", "Lorem ipsum", "group")

	testCases := []struct {
		Name     string
		Input    *model.APIDefinition
		Expected *graphql.APIDefinition
		AuthConverterFn func() *automock.AuthConverter
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
				conv.On("ToGraphQL", modelAPIDefinition.Spec.FetchRequest.Auth).Return(gqlAPIDefinition.Spec.FetchRequest.Auth).Once()
				return conv
			},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("ToGraphQL", modelAPIDefinition.Spec.FetchRequest).Return(gqlAPIDefinition.Spec.FetchRequest).Once()
				return conv
			},
		},
		//{
		//	Name:  "Empty",
		//	Input: &model.APIDefinition{
		//		Spec: &model.APISpec{
		//			FetchRequest:&model.FetchRequest{},
		//		},
		//		Auths:[]*model.RuntimeAuth{},
		//	},
		//	Expected: &graphql.APIDefinition{
		//		Spec:&graphql.APISpec{
		//			Data: &data,
		//			FetchRequest:&graphql.FetchRequest{},
		//		},
		//	},
		//	AuthConverterFn: func() *automock.AuthConverter {
		//		conv := &automock.AuthConverter{}
		//		conv.On("ToGraphQL", modelAPIDefinition.Spec.FetchRequest.Auth).Return(nil).Once()
		//		conv.On("ToGraphQL", modelAPIDefinition.Auths[0].Auth).Return(gqlAPIDefinition.Auths[0].Auth).Once()
		//
		//		return conv
		//	},
		//	FetchRequestConverter: func() *automock.FetchRequestConverter {
		//		conv := &automock.FetchRequestConverter{}
		//		conv.On("ToGraphQL", modelAPIDefinition.Spec.FetchRequest).Return(nil).Once()
		//		return conv
		//	},
		//},
		//{
		//	Name:     "Nil",
		//	Input:    nil,
		//	Expected: nil,
		//	AuthConverterFn:nil,
		//	FetchRequestConverter:nil,
		//},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// when
			//authConverter := &automock.AuthConverter{}
			//frConverter := &automock.FetchRequestConverter{}

			//if testCase.Input != nil {
			//	frConverter.On("ToGraphQL", testCase.Input.Spec.FetchRequest).Return(testCase.Expected.Spec.FetchRequest).Once()
			//	authConverter.On("ToGraphQL", testCase.Input.Spec.FetchRequest.Auth).Return(testCase.Expected.Spec.FetchRequest.Auth).Once()
			//
			//		//for _, runtimeAuth := range testCase.Input.Auths {
			//		//	authConverter.On("ToGraphQL", runtimeAuth.Auth).Return(runtimeAuth.Auth).Once()
			//		//}
			//	if len(testCase.Input.Auths) > 0 {
			//		authConverter.On("ToGraphQL", testCase.Input.Auths[0].Auth).Return(testCase.Expected.Auths[0].Auth).Once()
			//		authConverter.On("ToGraphQL", testCase.Input.Auths[1].Auth).Return(testCase.Expected.Auths[1].Auth).Once()
			//	}
			//	authConverter.On("ToGraphQL", testCase.Input.DefaultAuth).Return(testCase.Expected.DefaultAuth).Once()
			//}
			converter := api.NewConverter(testCase.AuthConverterFn(), testCase.FetchRequestConverter())

			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.APIDefinition{
		fixModelAPIDefinition("foo", "Foo", "Lorem ipsum"),
		fixModelAPIDefinition("bar", "Bar", "Dolor sit amet"),
		{},
		nil,
	}

	expected := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", "Foo", "Lorem ipsum"),
		fixGQLAPIDefinition("bar", "Bar", "Dolor sit amet"),
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
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.APIDefinitionInput
		Expected *model.APIDefinitionInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLAPIDefinitionInput("foo", "Lorem ipsum", "group"),
			Expected: fixModelAPIDefinitionInput("foo", "Lorem ipsum", "group"),
		},
		{},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			authConverter := &automock.AuthConverter{}
			frConverter := &automock.FetchRequestConverter{}

			if testCase.Input != nil {
				if testCase.Input.Spec != nil {
					frConverter.On("InputFromGraphQL", testCase.Input.Spec.FetchRequest).Return(testCase.Expected.Spec.FetchRequest).Once()
				}
				authConverter.On("InputFromGraphQL", testCase.Input.DefaultAuth).Return(testCase.Expected.DefaultAuth).Once()
			}
			// when
			converter := api.NewConverter(authConverter, frConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	gqlApi1 := fixGQLAPIDefinitionInput("foo", "lorem", "group")
	gqlApi2 := fixGQLAPIDefinitionInput("bar", "ipsum", "group2")

	modelApi1 := fixModelAPIDefinitionInput("foo", "lorem", "group")
	modelApi2 := fixModelAPIDefinitionInput("bar", "ipsum", "group2")

	testCases := []struct {
		Name     string
		Input    []*graphql.APIDefinitionInput
		Expected []*model.APIDefinitionInput
	}{
		{
			Name:     "All properties given",
			Input:    []*graphql.APIDefinitionInput{gqlApi1, gqlApi2},
			Expected: []*model.APIDefinitionInput{modelApi1, modelApi2},
		},
		{
			Name:     "Empty",
			Input:    []*graphql.APIDefinitionInput{},
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			// given
			authConverter := &automock.AuthConverter{}
			frConverter := &automock.FetchRequestConverter{}

			if len(testCase.Input) > 0 {
				frConverter.On("InputFromGraphQL", testCase.Input[0].Spec.FetchRequest).Return(testCase.Expected[0].Spec.FetchRequest).Once()
				frConverter.On("InputFromGraphQL", testCase.Input[1].Spec.FetchRequest).Return(testCase.Expected[1].Spec.FetchRequest).Once()
				authConverter.On("InputFromGraphQL", testCase.Input[0].DefaultAuth).Return(testCase.Expected[0].DefaultAuth).Once()
				authConverter.On("InputFromGraphQL", testCase.Input[1].DefaultAuth).Return(testCase.Expected[1].DefaultAuth).Once()
			}

			// when
			converter := api.NewConverter(authConverter, frConverter)
			res := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
