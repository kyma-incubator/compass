package api_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	placeholder := "test"
	modelAPIDefinition := fixFullAPIDefinitionModel(placeholder)
	gqlAPIDefinition := fixFullGQLAPIDefinition(placeholder)
	emptyModelAPIDefinition := &model.APIDefinition{}
	emptyGraphQLAPIDefinition := &graphql.APIDefinition{}

	testCases := []struct {
		Name                  string
		Input                 *model.APIDefinition
		Expected              *graphql.APIDefinition
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
	}{
		{
			Name:     "All properties given",
			Input:    &modelAPIDefinition,
			Expected: gqlAPIDefinition,

			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", modelAPIDefinition.Version).Return(gqlAPIDefinition.Version).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyModelAPIDefinition,
			Expected: emptyGraphQLAPIDefinition,

			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", emptyModelAPIDefinition.Version).Return(nil).Once()
				return conv
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,

			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()
			versionConverter := testCase.VersionConverter()

			// when
			converter := api.NewConverter(frConverter, versionConverter)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
			versionConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.APIDefinition{
		fixAPIDefinitionModel("foo", "1", "Foo", "Lorem ipsum"),
		fixAPIDefinitionModel("bar", "1", "Bar", "Dolor sit amet"),
		{},
		nil,
	}

	expected := []*graphql.APIDefinition{
		fixGQLAPIDefinition("foo", "1", "Foo", "Lorem ipsum"),
		fixGQLAPIDefinition("bar", "1", "Bar", "Dolor sit amet"),
		{},
	}

	frConverter := &automock.FetchRequestConverter{}
	versionConverter := &automock.VersionConverter{}

	for i, api := range input {
		if api == nil {
			continue
		}
		versionConverter.On("ToGraphQL", api.Version).Return(expected[i].Version).Once()
	}

	// when
	converter := api.NewConverter(frConverter, versionConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput("foo", "Lorem ipsum", "group")
	modelAPIDefinitionInput := fixModelAPIDefinitionInput("foo", "Lorem ipsum", "group")
	emptyGQLAPIDefinition := &graphql.APIDefinitionInput{}
	testCases := []struct {
		Name                  string
		Input                 *graphql.APIDefinitionInput
		Expected              *model.APIDefinitionInput
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlAPIDefinitionInput,
			Expected: modelAPIDefinitionInput,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.Spec.FetchRequest).Return(modelAPIDefinitionInput.Spec.FetchRequest, nil).Once()
				return conv
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.Version).Return(modelAPIDefinitionInput.Version).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    &graphql.APIDefinitionInput{},
			Expected: &model.APIDefinitionInput{},
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", emptyGQLAPIDefinition.Version).Return(nil).Once()
				return conv
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()
			versionConverter := testCase.VersionConverter()

			// when
			converter := api.NewConverter(frConverter, versionConverter)
			res, err := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
			versionConverter.AssertExpectations(t)
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
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlAPIDefinitionInputs,
			Expected: modelAPIDefinitionInputs,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				for i, apiDef := range gqlAPIDefinitionInputs {
					conv.On("InputFromGraphQL", apiDef.Spec.FetchRequest).Return(modelAPIDefinitionInputs[i].Spec.FetchRequest, nil).Once()
				}

				return conv
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				for i, apiDef := range gqlAPIDefinitionInputs {
					conv.On("InputFromGraphQL", apiDef.Version).Return(modelAPIDefinitionInputs[i].Version).Once()
				}
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    []*graphql.APIDefinitionInput{},
			Expected: nil,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			frConverter := testCase.FetchRequestConverter()
			versionConverter := testCase.VersionConverter()

			// when
			converter := api.NewConverter(frConverter, versionConverter)
			res, err := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
			versionConverter.AssertExpectations(t)
		})
	}
}

func TestApiSpecDataConversionNilStaysNil(t *testing.T) {
	// GIVEN

	mockFrConv := &automock.FetchRequestConverter{}
	defer mockFrConv.AssertExpectations(t)
	mockFrConv.On("InputFromGraphQL", mock.Anything).Return(nil, nil)

	mockVersionConv := &automock.VersionConverter{}
	defer mockVersionConv.AssertExpectations(t)
	mockVersionConv.On("InputFromGraphQL", mock.Anything).Return(nil)
	mockVersionConv.On("ToGraphQL", mock.Anything).Return(nil)

	converter := api.NewConverter(mockFrConv, mockVersionConv)
	// WHEN & THEN
	convertedInputModel, err := converter.InputFromGraphQL(&graphql.APIDefinitionInput{Spec: &graphql.APISpecInput{}})
	require.NoError(t, err)
	require.NotNil(t, convertedInputModel)
	require.NotNil(t, convertedInputModel.Spec)
	require.Nil(t, convertedInputModel.Spec.Data)
	convertedAPIDef := convertedInputModel.ToAPIDefinitionWithinBundle("id", "app_id", tenantID)
	require.NotNil(t, convertedAPIDef)
	convertedGraphqlAPIDef := converter.ToGraphQL(convertedAPIDef)
	require.NotNil(t, convertedGraphqlAPIDef)
	assert.Nil(t, convertedGraphqlAPIDef.Spec.Data)
}

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		apiModel := fixFullAPIDefinitionModel("foo")
		require.NotNil(t, apiModel)
		versionConv := version.NewConverter()
		conv := api.NewConverter(nil, versionConv)
		//WHEN
		entity := conv.ToEntity(apiModel)
		//THEN
		assert.Equal(t, fixFullEntityAPIDefinition(apiDefID, "foo"), entity)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		apiModel := fixAPIDefinitionModel("id", "bundle_id", "name", "target_url")
		require.NotNil(t, apiModel)
		versionConv := version.NewConverter()
		conv := api.NewConverter(nil, versionConv)
		//WHEN
		entity := conv.ToEntity(*apiModel)
		//THEN
		assert.Equal(t, fixEntityAPIDefinition("id", "bundle_id", "name", "target_url"), entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		entity := fixFullEntityAPIDefinition(apiDefID, "placeholder")
		versionConv := version.NewConverter()
		conv := api.NewConverter(nil, versionConv)
		//WHEN
		apiModel := conv.FromEntity(entity)
		//THEN
		assert.Equal(t, fixFullAPIDefinitionModel("placeholder"), apiModel)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		entity := fixEntityAPIDefinition("id", "bundle_id", "name", "target_url")
		versionConv := version.NewConverter()
		conv := api.NewConverter(nil, versionConv)
		//WHEN
		apiModel := conv.FromEntity(entity)
		//THEN
		expectedModel := fixAPIDefinitionModel("id", "bundle_id", "name", "target_url")
		require.NotNil(t, expectedModel)
		assert.Equal(t, *expectedModel, apiModel)
	})
}
