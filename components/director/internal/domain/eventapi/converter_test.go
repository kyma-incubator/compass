package eventapi_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/pkg/strings"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventapi/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	modelEventAPIDefinition := fixFullModelEventAPIDefinition("foo", "placeholder")
	gqlEventAPIDefinition := fixDetailedGQLEventAPIDefinition("foo", "placeholder")
	emptyModelEventAPIDefinition := &model.EventAPIDefinition{}
	emptyGraphQLEventAPIDefinition := &graphql.EventAPIDefinition{}

	testCases := []struct {
		Name                  string
		Input                 *model.EventAPIDefinition
		Expected              *graphql.EventAPIDefinition
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
	}{
		{
			Name:     "All properties given",
			Input:    &modelEventAPIDefinition,
			Expected: gqlEventAPIDefinition,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				return conv
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", modelEventAPIDefinition.Version).Return(gqlEventAPIDefinition.Version).Once()
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
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", emptyModelEventAPIDefinition.Version).Return(nil).Once()
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
			converter := eventapi.NewConverter(frConverter, versionConverter)
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
	input := []*model.EventAPIDefinition{
		fixMinModelEventAPIDefinition("foo", "placeholder"),
		fixMinModelEventAPIDefinition("bar", "placeholder"),
		{},
		nil,
	}

	expected := []*graphql.EventAPIDefinition{
		fixGQLEventAPIDefinition("foo", "placeholder"),
		fixGQLEventAPIDefinition("bar", "placeholder"),
		{},
	}

	frConverter := &automock.FetchRequestConverter{}
	versionConverter := &automock.VersionConverter{}

	for i, eventAPI := range input {
		if eventAPI == nil {
			continue
		}
		versionConverter.On("ToGraphQL", eventAPI.Version).Return(expected[i].Version).Once()
	}

	// when
	converter := eventapi.NewConverter(frConverter, versionConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConverter.AssertExpectations(t)
	versionConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlEventAPIDefinitionInput := fixGQLEventAPIDefinitionInput()
	modelEventAPIDefinitionInput := fixModelEventAPIDefinitionInput()
	emptyGQLEventAPIDefinition := &graphql.EventAPIDefinitionInput{}
	emptyModelEventAPIDefinition := &model.EventAPIDefinitionInput{}
	testCases := []struct {
		Name                  string
		Input                 *graphql.EventAPIDefinitionInput
		Expected              *model.EventAPIDefinitionInput
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
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
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", gqlEventAPIDefinitionInput.Version).Return(modelEventAPIDefinitionInput.Version).Once()
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
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", emptyGQLEventAPIDefinition.Version).Return(nil).Once()
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
			converter := eventapi.NewConverter(frConverter, versionConverter)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
			versionConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	gqlApi1 := fixGQLEventAPIDefinitionInput()
	gqlApi2 := fixGQLEventAPIDefinitionInput()
	gqlApi2.Group = strings.Ptr("group2")

	modelApi1 := fixModelEventAPIDefinitionInput()
	modelApi2 := fixModelEventAPIDefinitionInput()
	modelApi2.Group = strings.Ptr("group2")

	gqlEventAPIDefinitionInputs := []*graphql.EventAPIDefinitionInput{gqlApi1, gqlApi2}
	modelEventAPIDefinitionInputs := []*model.EventAPIDefinitionInput{modelApi1, modelApi2}
	testCases := []struct {
		Name                  string
		Input                 []*graphql.EventAPIDefinitionInput
		Expected              []*model.EventAPIDefinitionInput
		FetchRequestConverter func() *automock.FetchRequestConverter
		VersionConverter      func() *automock.VersionConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlEventAPIDefinitionInputs,
			Expected: modelEventAPIDefinitionInputs,
			FetchRequestConverter: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				for i, eventAPI := range gqlEventAPIDefinitionInputs {
					conv.On("InputFromGraphQL", eventAPI.Spec.FetchRequest).Return(modelEventAPIDefinitionInputs[i].Spec.FetchRequest).Once()
				}

				return conv
			},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				for i, eventAPI := range gqlEventAPIDefinitionInputs {
					conv.On("InputFromGraphQL", eventAPI.Version).Return(modelEventAPIDefinitionInputs[i].Version).Once()
				}

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
			converter := eventapi.NewConverter(frConverter, versionConverter)
			res := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConverter.AssertExpectations(t)
			versionConverter.AssertExpectations(t)
		})
	}
}

func TestEventApiSpecDataConversionNilStaysNil(t *testing.T) {
	// GIVEN
	mockFrConv := &automock.FetchRequestConverter{}
	defer mockFrConv.AssertExpectations(t)
	mockFrConv.On("InputFromGraphQL", mock.Anything).Return(nil)

	mockVersionConv := &automock.VersionConverter{}
	defer mockVersionConv.AssertExpectations(t)
	mockVersionConv.On("InputFromGraphQL", mock.Anything).Return(nil)
	mockVersionConv.On("ToGraphQL", mock.Anything).Return(nil)

	converter := eventapi.NewConverter(mockFrConv, mockVersionConv)
	// WHEN & THEN
	convertedInputModel := converter.InputFromGraphQL(&graphql.EventAPIDefinitionInput{Spec: &graphql.EventAPISpecInput{}})
	require.NotNil(t, convertedInputModel)
	require.NotNil(t, convertedInputModel.Spec)
	require.Nil(t, convertedInputModel.Spec.Data)
	convertedEvAPIDef := convertedInputModel.ToEventAPIDefinition("id", tenantID, "app_id")
	require.NotNil(t, convertedEvAPIDef)
	convertedGraphqlEvAPIDef := converter.ToGraphQL(convertedEvAPIDef)
	require.NotNil(t, convertedGraphqlEvAPIDef)
	assert.Nil(t, convertedGraphqlEvAPIDef.Spec.Data)
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("success when all nullable properties filled and converted", func(t *testing.T) {
		//GIVEN
		id := "id"
		eventModel := fixFullModelEventAPIDefinition(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		versionConv.On("ToEntity", fixVersionModel()).Return(fixVersionEntity()).Once()
		conv := eventapi.NewConverter(nil, versionConv)
		//WHEN
		eventAPIEnt, err := conv.ToEntity(eventModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixFullEventAPIDef(id, "placeholder"), eventAPIEnt)
		versionConv.AssertExpectations(t)
	})

	t.Run("success when all nullable properties empty and converter", func(t *testing.T) {
		// GIVEN
		id := "id"
		eventModel := fixMinModelEventAPIDefinition(id, "placeholder")
		require.NotNil(t, eventModel)
		versionConv := &automock.VersionConverter{}
		conv := eventapi.NewConverter(nil, versionConv)
		//WHEN
		eventEntity, err := conv.ToEntity(*eventModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixMinEntityEventAPIDef(id, "placeholder"), eventEntity)
		versionConv.AssertExpectations(t)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	t.Run("success when all nullable properties filled and converted", func(t *testing.T) {
		//GIVEN
		id := "id"
		eventEntity := fixFullEventAPIDef(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		exptectedModel := fixVersionModel()
		versionConv.On("FromEntity", fixVersionEntity()).Return(&exptectedModel).Once()
		conv := eventapi.NewConverter(nil, versionConv)
		//WHEN
		eventModel, err := conv.FromEntity(eventEntity)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventModel, fixFullModelEventAPIDefinition(id, "placeholder"))
		versionConv.AssertExpectations(t)
	})

	t.Run("success when all nullable properties empty and converted", func(t *testing.T) {
		// GIVEN
		id := "id"
		eventEntity := fixMinEntityEventAPIDef(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		versionConv.On("FromEntity", version.Version{}).Return(nil).Once()
		conv := eventapi.NewConverter(nil, versionConv)
		//WHEN
		eventModel, err := conv.FromEntity(eventEntity)
		//THEN
		require.NoError(t, err)
		expectedModel := fixMinModelEventAPIDefinition(id, "placeholder")
		require.NotNil(t, expectedModel)
		assert.Equal(t, *expectedModel, eventModel)
		versionConv.AssertExpectations(t)
	})
}
