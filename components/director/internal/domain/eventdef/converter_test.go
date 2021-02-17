package eventdef_test

import (
	"fmt"
	"testing"

	event "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/eventdef/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	placeholder := "test"
	modelEventDefinition, modelSpec := fixFullEventDefinitionModel(placeholder)
	gqlEventDefinition := fixFullGQLEventDefinition(placeholder)
	emptyModelEventDefinition := &model.EventDefinition{BaseEntity: &model.BaseEntity{}}
	emptyGraphQLEventDefinition := &graphql.EventDefinition{BaseEntity: &graphql.BaseEntity{}}

	expectedErr := errors.New("error")

	testCases := []struct {
		Name             string
		Input            *model.EventDefinition
		SpecInput        *model.Spec
		Expected         *graphql.EventDefinition
		VersionConverter func() *automock.VersionConverter
		SpecConverter    func() *automock.SpecConverter
		ExpectedErr      error
	}{
		{
			Name:      "All properties given",
			Input:     &modelEventDefinition,
			SpecInput: &modelSpec,
			Expected:  gqlEventDefinition,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", modelEventDefinition.Version).Return(gqlEventDefinition.Version).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", &modelSpec).Return(gqlEventDefinition.Spec, nil).Once()
				return conv
			},
		},
		{
			Name:      "Error while converting spec",
			Input:     &modelEventDefinition,
			SpecInput: &modelSpec,
			Expected:  nil,
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", &modelSpec).Return(nil, expectedErr).Once()
				return conv
			},
			ExpectedErr: expectedErr,
		},
		{
			Name:      "Empty",
			Input:     emptyModelEventDefinition,
			SpecInput: &model.Spec{},
			Expected:  emptyGraphQLEventDefinition,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", emptyModelEventDefinition.Version).Return(nil).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLEventSpec", &model.Spec{}).Return(nil, nil).Once()
				return conv
			},
		},
		{
			Name: "Nil",
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//give
			versionConverter := testCase.VersionConverter()
			specConverter := testCase.SpecConverter()

			// when
			converter := event.NewConverter(versionConverter, specConverter)
			res, err := converter.ToGraphQL(testCase.Input, testCase.SpecInput)
			// then
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			// then
			assert.EqualValues(t, testCase.Expected, res)
			versionConverter.AssertExpectations(t)
			specConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	event1, spec1 := fixFullEventDefinitionModel("test1")
	event2, spec2 := fixFullEventDefinitionModel("test2")

	inputApis := []*model.EventDefinition{
		&event1, &event2, {BaseEntity: &model.BaseEntity{}}, nil,
	}

	inputSpecs := []*model.Spec{
		&spec1, &spec2, {}, nil,
	}

	expected := []*graphql.EventDefinition{
		fixFullGQLEventDefinition("test1"),
		fixFullGQLEventDefinition("test2"),
		{BaseEntity: &graphql.BaseEntity{}},
	}

	versionConverter := &automock.VersionConverter{}
	specConverter := &automock.SpecConverter{}

	for i, event := range inputApis {
		if event == nil {
			continue
		}
		versionConverter.On("ToGraphQL", event.Version).Return(expected[i].Version).Once()
		specConverter.On("ToGraphQLEventSpec", inputSpecs[i]).Return(expected[i].Spec, nil).Once()
	}

	// when
	converter := event.NewConverter(versionConverter, specConverter)
	res, err := converter.MultipleToGraphQL(inputApis, inputSpecs)
	assert.NoError(t, err)

	// then
	assert.Equal(t, expected, res)
	specConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	gqlEventDefinitionInput := fixGQLEventDefinitionInput("foo", "Lorem ipsum", "group")
	modelEventDefinitionInput, modelSpec := fixModelEventDefinitionInput("foo", "Lorem ipsum", "group")
	emptyGQLEventDefinition := &graphql.EventDefinitionInput{}

	expectedErr := errors.New("error")

	testCases := []struct {
		Name             string
		Input            *graphql.EventDefinitionInput
		Expected         *model.EventDefinitionInput
		ExpectedSpec     *model.SpecInput
		VersionConverter func() *automock.VersionConverter
		SpecConverter    func() *automock.SpecConverter
		ExpectedErr      error
	}{
		{
			Name:         "All properties given",
			Input:        gqlEventDefinitionInput,
			Expected:     modelEventDefinitionInput,
			ExpectedSpec: modelSpec,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", gqlEventDefinitionInput.Version).Return(modelEventDefinitionInput.Version).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLEventSpec", gqlEventDefinitionInput.Spec).Return(modelSpec, nil).Once()
				return conv
			},
		},
		{
			Name:  "Error while converting spec",
			Input: gqlEventDefinitionInput,
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLEventSpec", gqlEventDefinitionInput.Spec).Return(nil, expectedErr).Once()
				return conv
			},
			ExpectedErr: expectedErr,
		},
		{
			Name:         "Empty",
			Input:        &graphql.EventDefinitionInput{},
			Expected:     &model.EventDefinitionInput{},
			ExpectedSpec: &model.SpecInput{},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", emptyGQLEventDefinition.Version).Return(nil).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLEventSpec", emptyGQLEventDefinition.Spec).Return(&model.SpecInput{}, nil).Once()
				return conv
			},
		},
		{
			Name: "Nil",
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//give
			versionConverter := testCase.VersionConverter()
			specConverter := testCase.SpecConverter()

			// when
			converter := event.NewConverter(versionConverter, specConverter)
			res, spec, err := converter.InputFromGraphQL(testCase.Input)
			// then
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			// then
			assert.Equal(t, testCase.Expected, res)
			assert.Equal(t, testCase.ExpectedSpec, spec)
			versionConverter.AssertExpectations(t)
			specConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	gqlEvent1 := fixGQLEventDefinitionInput("foo", "lorem", "group")
	gqlEvent2 := fixGQLEventDefinitionInput("bar", "ipsum", "group2")

	modelEvent1, modelSpec1 := fixModelEventDefinitionInput("foo", "lorem", "group")
	modelEvent2, modelSpec2 := fixModelEventDefinitionInput("bar", "ipsum", "group2")

	gqlEventDefinitionInputs := []*graphql.EventDefinitionInput{gqlEvent1, gqlEvent2}
	modelEventDefinitionInputs := []*model.EventDefinitionInput{modelEvent1, modelEvent2}
	modelSpecInputs := []*model.SpecInput{modelSpec1, modelSpec2}
	testCases := []struct {
		Name             string
		Input            []*graphql.EventDefinitionInput
		Expected         []*model.EventDefinitionInput
		ExpectedSpecs    []*model.SpecInput
		VersionConverter func() *automock.VersionConverter
		SpecConverter    func() *automock.SpecConverter
	}{
		{
			Name:          "All properties given",
			Input:         gqlEventDefinitionInputs,
			Expected:      modelEventDefinitionInputs,
			ExpectedSpecs: modelSpecInputs,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				for i, eventDef := range gqlEventDefinitionInputs {
					conv.On("InputFromGraphQL", eventDef.Version).Return(modelEventDefinitionInputs[i].Version).Once()
				}
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				for i, eventDef := range gqlEventDefinitionInputs {
					conv.On("InputFromGraphQLEventSpec", eventDef.Spec).Return(modelSpecInputs[i], nil).Once()
				}
				return conv
			},
		},
		{
			Name:          "Empty",
			Input:         []*graphql.EventDefinitionInput{},
			Expected:      nil,
			ExpectedSpecs: nil,
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
		{
			Name: "Nil",
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%s", testCase.Name), func(t *testing.T) {
			//given
			versionConverter := testCase.VersionConverter()
			specCovnerter := testCase.SpecConverter()

			// when
			converter := event.NewConverter(versionConverter, specCovnerter)
			res, specs, err := converter.MultipleInputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			assert.Equal(t, testCase.ExpectedSpecs, specs)
			versionConverter.AssertExpectations(t)
		})
	}
}

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		eventModel, _ := fixFullEventDefinitionModel("foo")

		versionConv := version.NewConverter()
		conv := event.NewConverter(versionConv, nil)
		//WHEN
		entity := conv.ToEntity(eventModel)
		//THEN
		assert.Equal(t, fixFullEntityEventDefinition(eventID, "foo"), entity)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		eventModel := fixEventDefinitionModel("id", "bndl_id", "name")
		require.NotNil(t, eventModel)
		versionConv := version.NewConverter()
		conv := event.NewConverter(versionConv, nil)
		//WHEN
		entity := conv.ToEntity(*eventModel)
		//THEN
		assert.Equal(t, fixEntityEventDefinition("id", "bndl_id", "name"), entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		//GIVEN
		entity := fixFullEntityEventDefinition(eventID, "placeholder")
		versionConv := version.NewConverter()
		conv := event.NewConverter(versionConv, nil)
		//WHEN
		eventModel := conv.FromEntity(entity)
		//THEN
		expectedModel, _ := fixFullEventDefinitionModel("placeholder")
		assert.Equal(t, expectedModel, eventModel)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		//GIVEN
		entity := fixEntityEventDefinition("id", "bndl_id", "name")
		versionConv := version.NewConverter()
		conv := event.NewConverter(versionConv, nil)
		//WHEN
		eventModel := conv.FromEntity(entity)
		//THEN
		expectedModel := fixEventDefinitionModel("id", "bndl_id", "name")
		require.NotNil(t, expectedModel)
		assert.Equal(t, *expectedModel, eventModel)
	})
}
