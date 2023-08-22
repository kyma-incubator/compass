package api_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/api"
	"github.com/kyma-incubator/compass/components/director/internal/domain/api/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	placeholder := "test"
	modelAPIDefinition, modelSpec, modelBundleReference := fixFullAPIDefinitionModelWithAppID(placeholder)
	gqlAPIDefinition := fixFullGQLAPIDefinition(placeholder)
	emptyModelAPIDefinition := &model.APIDefinition{BaseEntity: &model.BaseEntity{}}
	emptyGraphQLAPIDefinition := &graphql.APIDefinition{BaseEntity: &graphql.BaseEntity{}}

	expectedErr := errors.New("error")

	testCases := []struct {
		Name                 string
		Input                *model.APIDefinition
		SpecInput            *model.Spec
		BundleReferenceInput *model.BundleReference
		Expected             *graphql.APIDefinition
		VersionConverter     func() *automock.VersionConverter
		SpecConverter        func() *automock.SpecConverter
		ExpectedErr          error
	}{
		{
			Name:                 "All properties given",
			Input:                &modelAPIDefinition,
			SpecInput:            &modelSpec,
			BundleReferenceInput: &modelBundleReference,
			Expected:             gqlAPIDefinition,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", modelAPIDefinition.Version).Return(gqlAPIDefinition.Version).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", &modelSpec).Return(gqlAPIDefinition.Spec, nil).Once()
				return conv
			},
		},
		{
			Name:                 "Error while converting spec",
			Input:                &modelAPIDefinition,
			SpecInput:            &modelSpec,
			BundleReferenceInput: &modelBundleReference,
			Expected:             nil,
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", &modelSpec).Return(nil, expectedErr).Once()
				return conv
			},
			ExpectedErr: expectedErr,
		},
		{
			Name:                 "Empty",
			Input:                emptyModelAPIDefinition,
			SpecInput:            &model.Spec{},
			BundleReferenceInput: &model.BundleReference{},
			Expected:             emptyGraphQLAPIDefinition,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("ToGraphQL", emptyModelAPIDefinition.Version).Return(nil).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("ToGraphQLAPISpec", &model.Spec{}).Return(nil, nil).Once()
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
		t.Run(testCase.Name, func(t *testing.T) {
			//give
			versionConverter := testCase.VersionConverter()
			specConverter := testCase.SpecConverter()

			// WHEN
			converter := api.NewConverter(versionConverter, specConverter)
			res, err := converter.ToGraphQL(testCase.Input, testCase.SpecInput, testCase.BundleReferenceInput)
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
	// GIVEN
	api1, spec1, bundleRef1 := fixFullAPIDefinitionModelWithAppID("test1")
	api2, spec2, bundleRef2 := fixFullAPIDefinitionModelWithAppID("test2")

	inputAPIs := []*model.APIDefinition{
		&api1, &api2, {BaseEntity: &model.BaseEntity{}}, nil,
	}

	inputSpecs := []*model.Spec{
		&spec1, &spec2, {}, nil,
	}

	inputBundleRefs := []*model.BundleReference{
		&bundleRef1, &bundleRef2, {}, nil,
	}

	expected := []*graphql.APIDefinition{
		fixFullGQLAPIDefinition("test1"),
		fixFullGQLAPIDefinition("test2"),
		{BaseEntity: &graphql.BaseEntity{}},
	}

	versionConverter := &automock.VersionConverter{}
	specConverter := &automock.SpecConverter{}

	for i, api := range inputAPIs {
		if api == nil {
			continue
		}
		versionConverter.On("ToGraphQL", api.Version).Return(expected[i].Version).Once()
		specConverter.On("ToGraphQLAPISpec", inputSpecs[i]).Return(expected[i].Spec, nil).Once()
	}

	// WHEN
	converter := api.NewConverter(versionConverter, specConverter)
	res, err := converter.MultipleToGraphQL(inputAPIs, inputSpecs, inputBundleRefs)
	assert.NoError(t, err)

	// then
	assert.Equal(t, expected, res)
	specConverter.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlAPIDefinitionInput := fixGQLAPIDefinitionInput("foo", "Lorem ipsum", "group")
	modelAPIDefinitionInput, modelSpec := fixModelAPIDefinitionInput("foo", "Lorem ipsum", "group")
	emptyGQLAPIDefinition := &graphql.APIDefinitionInput{}

	expectedErr := errors.New("error")

	testCases := []struct {
		Name             string
		Input            *graphql.APIDefinitionInput
		Expected         *model.APIDefinitionInput
		ExpectedSpec     *model.SpecInput
		VersionConverter func() *automock.VersionConverter
		SpecConverter    func() *automock.SpecConverter
		ExpectedErr      error
	}{
		{
			Name:         "All properties given",
			Input:        gqlAPIDefinitionInput,
			Expected:     modelAPIDefinitionInput,
			ExpectedSpec: modelSpec,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", gqlAPIDefinitionInput.Version).Return(modelAPIDefinitionInput.VersionInput).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLAPISpec", gqlAPIDefinitionInput.Spec).Return(modelSpec, nil).Once()
				return conv
			},
		},
		{
			Name:  "Error while converting spec",
			Input: gqlAPIDefinitionInput,
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLAPISpec", gqlAPIDefinitionInput.Spec).Return(nil, expectedErr).Once()
				return conv
			},
			ExpectedErr: expectedErr,
		},
		{
			Name:         "Empty",
			Input:        &graphql.APIDefinitionInput{},
			Expected:     &model.APIDefinitionInput{},
			ExpectedSpec: &model.SpecInput{},
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				conv.On("InputFromGraphQL", emptyGQLAPIDefinition.Version).Return(nil).Once()
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				conv.On("InputFromGraphQLAPISpec", emptyGQLAPIDefinition.Spec).Return(&model.SpecInput{}, nil).Once()
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
		t.Run(testCase.Name, func(t *testing.T) {
			//give
			versionConverter := testCase.VersionConverter()
			specConverter := testCase.SpecConverter()

			// WHEN
			converter := api.NewConverter(versionConverter, specConverter)
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
	// GIVEN
	gqlAPI1 := fixGQLAPIDefinitionInput("foo", "lorem", "group")
	gqlAPI2 := fixGQLAPIDefinitionInput("bar", "ipsum", "group2")

	modelAPI1, modelSpec1 := fixModelAPIDefinitionInput("foo", "lorem", "group")
	modelAPI2, modelSpec2 := fixModelAPIDefinitionInput("bar", "ipsum", "group2")

	gqlAPIDefinitionInputs := []*graphql.APIDefinitionInput{gqlAPI1, gqlAPI2}
	modelAPIDefinitionInputs := []*model.APIDefinitionInput{modelAPI1, modelAPI2}
	modelSpecInputs := []*model.SpecInput{modelSpec1, modelSpec2}
	testCases := []struct {
		Name             string
		Input            []*graphql.APIDefinitionInput
		Expected         []*model.APIDefinitionInput
		ExpectedSpecs    []*model.SpecInput
		VersionConverter func() *automock.VersionConverter
		SpecConverter    func() *automock.SpecConverter
	}{
		{
			Name:          "All properties given",
			Input:         gqlAPIDefinitionInputs,
			Expected:      modelAPIDefinitionInputs,
			ExpectedSpecs: modelSpecInputs,
			VersionConverter: func() *automock.VersionConverter {
				conv := &automock.VersionConverter{}
				for i, apiDef := range gqlAPIDefinitionInputs {
					conv.On("InputFromGraphQL", apiDef.Version).Return(modelAPIDefinitionInputs[i].VersionInput).Once()
				}
				return conv
			},
			SpecConverter: func() *automock.SpecConverter {
				conv := &automock.SpecConverter{}
				for i, apiDef := range gqlAPIDefinitionInputs {
					conv.On("InputFromGraphQLAPISpec", apiDef.Spec).Return(modelSpecInputs[i], nil).Once()
				}
				return conv
			},
		},
		{
			Name:          "Empty",
			Input:         []*graphql.APIDefinitionInput{},
			Expected:      []*model.APIDefinitionInput{},
			ExpectedSpecs: []*model.SpecInput{},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
		{
			Name:          "Nil",
			Expected:      []*model.APIDefinitionInput{},
			ExpectedSpecs: []*model.SpecInput{},
			VersionConverter: func() *automock.VersionConverter {
				return &automock.VersionConverter{}
			},
			SpecConverter: func() *automock.SpecConverter {
				return &automock.SpecConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			versionConverter := testCase.VersionConverter()
			specCovnerter := testCase.SpecConverter()

			// WHEN
			converter := api.NewConverter(versionConverter, specCovnerter)
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
	t.Run("success all nullable properties filled with Application ID", func(t *testing.T) {
		// GIVEN
		apiModel, _, _ := fixFullAPIDefinitionModelWithAppID("foo")

		versionConv := version.NewConverter()
		conv := api.NewConverter(versionConv, nil)
		// WHEN
		entity := conv.ToEntity(&apiModel)
		// THEN
		assert.Equal(t, fixFullEntityAPIDefinitionWithAppID(apiDefID, "foo"), *entity)
	})

	t.Run("success all nullable properties filled with Application Template Version ID", func(t *testing.T) {
		// GIVEN
		apiModel, _, _ := fixFullAPIDefinitionModelWithAppTemplateVersionID("foo")

		versionConv := version.NewConverter()
		conv := api.NewConverter(versionConv, nil)
		// WHEN
		entity := conv.ToEntity(&apiModel)
		// THEN
		assert.Equal(t, fixFullEntityAPIDefinitionWithAppTemplateVersionID(apiDefID, "foo"), *entity)
	})

	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		apiModel := fixAPIDefinitionModel("id", "name", "target_url")
		require.NotNil(t, apiModel)
		versionConv := version.NewConverter()
		conv := api.NewConverter(versionConv, nil)
		// WHEN
		entity := conv.ToEntity(apiModel)
		// THEN
		assert.Equal(t, fixEntityAPIDefinition("id", "name", "target_url"), entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		entity := fixFullEntityAPIDefinitionWithAppID(apiDefID, "placeholder")
		versionConv := version.NewConverter()
		conv := api.NewConverter(versionConv, nil)
		// WHEN
		apiModel := conv.FromEntity(&entity)
		// THEN
		expectedModel, _, _ := fixFullAPIDefinitionModelWithAppID("placeholder")
		assert.Equal(t, &expectedModel, apiModel)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		entity := fixEntityAPIDefinition("id", "name", "target_url")
		versionConv := version.NewConverter()
		conv := api.NewConverter(versionConv, nil)
		// WHEN
		apiModel := conv.FromEntity(entity)
		// THEN
		expectedModel := fixAPIDefinitionModel("id", "name", "target_url")
		require.NotNil(t, expectedModel)
		assert.Equal(t, expectedModel, apiModel)
	})
}
