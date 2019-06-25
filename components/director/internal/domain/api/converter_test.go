package api_test

import (
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.APIDefinition
		Expected *graphql.APIDefinition
	}{
		{
			Name:     "All properties given",
			Input:    fixDetailedModelApiDefinition(t, "foo", "Foo", "Lorem ipsum","group"),
			Expected: fixDetailedGQLApiDefinition(t, "foo", "Foo", "Lorem ipsum", "group"),
		},
		{
			Name:  "Empty",
			Input: &model.APIDefinition{},
			Expected: &graphql.APIDefinition{
				Spec: &graphql.APISpec{
					FetchRequest: &graphql.FetchRequest{},
				},
				Version: &graphql.Version{},
				Auth:&graphql.RuntimeAuth{},
				Auths: []*graphql.RuntimeAuth{},
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
			authConverter := &automock.AuthConverter{}
			frConverter := &automock.FetchRequestConverter{}

			if testCase.Input != nil {
				if testCase.Input.Spec != nil{
					frConverter.On("ToGraphQL",testCase.Input.Spec.FetchRequest).Return(testCase.Expected.Spec.FetchRequest).Once()
					authConverter.On("ToGraphQL",testCase.Input.Spec.FetchRequest.Auth).Return(testCase.Expected.Spec.FetchRequest.Auth).Once()

				}
				if len(testCase.Input.Auths) > 0 {
					authConverter.On("ToGraphQL", testCase.Input.Auths[0].Auth).Return(testCase.Expected.Auths[0].Auth).Once()
					authConverter.On("ToGraphQL", testCase.Input.Auths[1].Auth).Return(testCase.Expected.Auths[1].Auth).Once()
				}
				authConverter.On("ToGraphQL",testCase.Input.DefaultAuth).Return(testCase.Expected.DefaultAuth).Once()
				if testCase.Input.Auth != nil{
					authConverter.On("ToGraphQL",testCase.Input.Auth.Auth).Return(testCase.Expected.Auth.Auth).Once()
				}
				//authConverter.On("ToGraphQL",testCase.Input.Auths).Return(testCase.Expected.Auths).Once()
			}
			converter := api.NewConverter(authConverter,frConverter)

			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.APIDefinition{
		fixModelApiDefinition("foo", "Foo", "Lorem ipsum"),
		fixModelApiDefinition("bar", "Bar", "Dolor sit amet"),
		{},
		nil,
	}

	fetchRequest := graphql.FetchRequest{}
	spec := graphql.APISpec{
		FetchRequest:&fetchRequest,
	}
	version := graphql.Version{}

	expected := []*graphql.APIDefinition{
		fixGQLApiDefinition("foo", "Foo", "Lorem ipsum"),
		fixGQLApiDefinition("bar", "Bar", "Dolor sit amet"),
		{
			Spec: &spec,
			Version: &version,
			DefaultAuth: &graphql.Auth{
					Credential:            nil,
					AdditionalHeaders:     nil,
					AdditionalQueryParams: nil,
					RequestAuth:           nil,
				},
				Auth:&graphql.RuntimeAuth{},
				Auths: []*graphql.RuntimeAuth{},
		},
	}

	// when
	authConverter := &automock.AuthConverter{}
	frConverter := &automock.FetchRequestConverter{}

	auth :=graphql.Auth{}

	frConverter.On("ToGraphQL",input[0].Spec.FetchRequest).Return(expected[0].Spec.FetchRequest).Once()
	frConverter.On("ToGraphQL",input[1].Spec.FetchRequest).Return(expected[1].Spec.FetchRequest).Once()

	authConverter.On("ToGraphQL",input[0].DefaultAuth).Return(expected[0].DefaultAuth).Once()
	authConverter.On("ToGraphQL",input[1].DefaultAuth).Return(expected[1].DefaultAuth).Once()
	authConverter.On("ToGraphQL",input[2].DefaultAuth).Return(&auth).Once()

	converter := api.NewConverter(authConverter,frConverter)
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
			Input:    fixGQLApiDefinitionInput("foo", "Lorem ipsum","group"),
			Expected: fixModelApiDefinitionInput("foo", "Lorem ipsum","group"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.APIDefinitionInput{},
			Expected: &model.APIDefinitionInput{
				Spec: &model.APISpecInput{
					FetchRequest: &model.FetchRequestInput{},
				},
				Version: &model.VersionInput{},
			},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when

			authConverter := &automock.AuthConverter{}
			frConverter := &automock.FetchRequestConverter{}

			if testCase.Input != nil {
				if testCase.Input.Spec != nil {
					frConverter.On("InputFromGraphQL",testCase.Input.Spec.FetchRequest).Return(testCase.Expected.Spec.FetchRequest).Once()
				}
				authConverter.On("InputFromGraphQL",testCase.Input.DefaultAuth).Return(testCase.Expected.DefaultAuth).Once()
			}

			converter := api.NewConverter(authConverter,frConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given

	gqlApi1 := fixGQLApiDefinitionInput("foo", "lorem","group")
	gqlApi2 := fixGQLApiDefinitionInput("bar", "ipsum","group2")

	modelApi1 := fixModelApiDefinitionInput("foo", "lorem","group")
	modelApi2 := fixModelApiDefinitionInput("bar", "ipsum","group2")

	testCases := []struct {
		Name     string
		Input    []*graphql.APIDefinitionInput
		Expected []*model.APIDefinitionInput
	}{
		{
			Name:     "All properties given",
			Input:    []*graphql.APIDefinitionInput{gqlApi1,gqlApi2},
			Expected: []*model.APIDefinitionInput{modelApi1,modelApi2},
		},
		{
			Name:     "Empty",
			Input:    []*graphql.APIDefinitionInput{},
			Expected: []*model.APIDefinitionInput{},
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when

			authConverter := &automock.AuthConverter{}
			frConverter := &automock.FetchRequestConverter{}

			if len(testCase.Input) > 0 {
				frConverter.On("InputFromGraphQL",testCase.Input[0].Spec.FetchRequest).Return(testCase.Expected[0].Spec.FetchRequest).Once()
				frConverter.On("InputFromGraphQL",testCase.Input[1].Spec.FetchRequest).Return(testCase.Expected[1].Spec.FetchRequest).Once()
				authConverter.On("InputFromGraphQL", testCase.Input[0].DefaultAuth).Return(testCase.Expected[0].DefaultAuth).Once()
				authConverter.On("InputFromGraphQL", testCase.Input[1].DefaultAuth).Return(testCase.Expected[1].DefaultAuth).Once()
			}
			converter := api.NewConverter(authConverter,frConverter)
			res := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}