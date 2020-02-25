package eventdef_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	modelEventAPIDefinition := fixFullModelEventDefinition("foo", "placeholder")
	gqlEventAPIDefinition := fixDetailedGQLEventDefinition("foo", "placeholder")
	emptyModelEventAPIDefinition := &model.EventDefinition{}
	emptyGraphQLEventDefinition := &graphql.EventDefinition{}

	testCases := []struct {
		Name                  string
		Input                 *model.EventDefinition
		Expected              *graphql.EventDefinition
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
			Expected: emptyGraphQLEventDefinition,
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
			converter := eventdef.NewConverter(frConverter, versionConverter)
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
	input := []*model.EventDefinition{
		fixMinModelEventAPIDefinition("foo", "placeholder"),
		fixMinModelEventAPIDefinition("bar", "placeholder"),
		{},
		nil,
	}

	expected := []*graphql.EventDefinition{
		fixGQLEventDefinition("foo", "placeholder"),
		fixGQLEventDefinition("bar", "placeholder"),
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
	converter := eventdef.NewConverter(frConverter, versionConverter)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConverter.AssertExpectations(t)
	versionConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlEventAPIDefinitionInput := fixGQLEventDefinitionInput()
	modelEventAPIDefinitionInput := fixModelEventDefinitionInput()
	emptyGQLEventAPIDefinition := &graphql.EventDefinitionInput{}
	emptyModelEventAPIDefinition := &model.EventDefinitionInput{}
	testCases := []struct {
		Name                  string
		Input                 *graphql.EventDefinitionInput
		Expected              *model.EventDefinitionInput
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
			converter := eventdef.NewConverter(frConverter, versionConverter)
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
	gqlApi1 := fixGQLEventDefinitionInput()
	gqlApi2 := fixGQLEventDefinitionInput()
	gqlApi2.Group = str.Ptr("group2")

	modelApi1 := fixModelEventDefinitionInput()
	modelApi2 := fixModelEventDefinitionInput()
	modelApi2.Group = str.Ptr("group2")

	gqlEventAPIDefinitionInputs := []*graphql.EventDefinitionInput{gqlApi1, gqlApi2}
	modelEventAPIDefinitionInputs := []*model.EventDefinitionInput{modelApi1, modelApi2}
	testCases := []struct {
		Name                  string
		Input                 []*graphql.EventDefinitionInput
		Expected              []*model.EventDefinitionInput
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
			Input:    []*graphql.EventDefinitionInput{},
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
			converter := eventdef.NewConverter(frConverter, versionConverter)
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

	converter := eventdef.NewConverter(mockFrConv, mockVersionConv)
	// WHEN & THEN
	convertedInputModel := converter.InputFromGraphQL(&graphql.EventDefinitionInput{Spec: &graphql.EventSpecInput{}})
	require.NotNil(t, convertedInputModel)
	require.NotNil(t, convertedInputModel.Spec)
	require.Nil(t, convertedInputModel.Spec.Data)
	convertedEvAPIDef := convertedInputModel.ToEventDefinition("id", str.Ptr("app_id"), tenantID)
	require.NotNil(t, convertedEvAPIDef)
	convertedGraphqlEvAPIDef := converter.ToGraphQL(convertedEvAPIDef)
	require.NotNil(t, convertedGraphqlEvAPIDef)
	assert.Nil(t, convertedGraphqlEvAPIDef.Spec.Data)
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("success when all nullable properties filled and converted", func(t *testing.T) {
		//GIVEN
		id := "id"
		eventModel := fixFullModelEventDefinition(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		versionConv.On("ToEntity", fixVersionModel()).Return(fixVersionEntity()).Once()
		conv := eventdef.NewConverter(nil, versionConv)
		//WHEN
		eventAPIEnt, err := conv.ToEntity(eventModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixFullEventDef(id, "placeholder"), eventAPIEnt)
		versionConv.AssertExpectations(t)
	})

	t.Run("success when all nullable properties empty and converter", func(t *testing.T) {
		// GIVEN
		id := "id"
		eventModel := fixMinModelEventAPIDefinition(id, "placeholder")
		require.NotNil(t, eventModel)
		versionConv := &automock.VersionConverter{}
		conv := eventdef.NewConverter(nil, versionConv)
		//WHEN
		eventEntity, err := conv.ToEntity(*eventModel)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, fixMinEntityEventDef(id, "placeholder"), eventEntity)
		versionConv.AssertExpectations(t)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	t.Run("success when all nullable properties filled and converted", func(t *testing.T) {
		//GIVEN
		id := "id"
		eventEntity := fixFullEventDef(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		exptectedModel := fixVersionModel()
		versionConv.On("FromEntity", fixVersionEntity()).Return(&exptectedModel).Once()
		conv := eventdef.NewConverter(nil, versionConv)
		//WHEN
		eventModel, err := conv.FromEntity(eventEntity)
		//THEN
		require.NoError(t, err)
		assert.Equal(t, eventModel, fixFullModelEventDefinition(id, "placeholder"))
		versionConv.AssertExpectations(t)
	})

	t.Run("success when all nullable properties empty and converted", func(t *testing.T) {
		// GIVEN
		id := "id"
		eventEntity := fixMinEntityEventDef(id, "placeholder")
		versionConv := &automock.VersionConverter{}
		versionConv.On("FromEntity", version.Version{}).Return(nil).Once()
		conv := eventdef.NewConverter(nil, versionConv)
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
