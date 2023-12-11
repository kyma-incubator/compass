package aspecteventresource_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aspectEventResourceModel := fixAspectEventResourceModel(aspectEventResourceID)
		require.NotNil(t, aspectEventResourceModel)
		conv := aspecteventresource.NewConverter()

		entity := conv.ToEntity(aspectEventResourceModel)

		assert.Equal(t, fixEntityAspectEventResource(aspectEventResourceID, appID, aspectID), entity)
	})

	t.Run("Returns nil if aspect event resource model is nil", func(t *testing.T) {
		conv := aspecteventresource.NewConverter()

		aspectEventResourceEntity := conv.ToEntity(nil)

		require.Nil(t, aspectEventResourceEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityAspectEventResource(aspectEventResourceID, appID, aspectID)
		conv := aspecteventresource.NewConverter()

		aspectEventResourceModel := conv.FromEntity(entity)

		assert.Equal(t, fixAspectEventResourceModel(aspectEventResourceID), aspectEventResourceModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := aspecteventresource.NewConverter()

		aspectEventResourceModel := conv.FromEntity(nil)

		require.Nil(t, aspectEventResourceModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	modelAspectEventResource := fixAspectEventResourceModel(aspectEventResourceID)
	gqlAspectEventResource := fixGQLAspectEventDefinition(aspectEventResourceID)

	testCases := []struct {
		Name     string
		Input    *model.AspectEventResource
		Expected *graphql.AspectEventDefinition
	}{
		{
			Name:     "All properties given",
			Input:    modelAspectEventResource,
			Expected: gqlAspectEventResource,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVE
			converter := aspecteventresource.NewConverter()

			// WHEN
			res, err := converter.ToGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	aspectEventResource1 := fixAspectEventResourceModel("id1")
	aspectEventResource2 := fixAspectEventResourceModel("id2")

	inputAspectEventResources := []*model.AspectEventResource{aspectEventResource1, aspectEventResource2, nil}

	gqlAspectEventDef1 := fixGQLAspectEventDefinition("id1")
	gqlAspectEventDef2 := fixGQLAspectEventDefinition("id2")

	expected := []*graphql.AspectEventDefinition{gqlAspectEventDef1, gqlAspectEventDef2}

	// WHEN
	converter := aspecteventresource.NewConverter()
	res, err := converter.MultipleToGraphQL(inputAspectEventResources)
	assert.NoError(t, err)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlAspectEventDefInput := fixGQLAspectEventDefinitionInput()
	modelAspectEventResourceInput := fixAspectEventResourceInputModel()
	modelAspectEventResourceInput.MinVersion = nil

	testCases := []struct {
		Name     string
		Input    *graphql.AspectEventDefinitionInput
		Expected *model.AspectEventResourceInput
	}{
		{
			Name:     "All properties given",
			Input:    gqlAspectEventDefInput,
			Expected: &modelAspectEventResourceInput,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVE
			converter := aspecteventresource.NewConverter()

			// WHEN
			res, err := converter.InputFromGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlAspectEventDef1 := fixGQLAspectEventDefinitionInput()
	gqlAspectEventDef2 := fixGQLAspectEventDefinitionInput()

	modelAspectEventResource1 := fixAspectEventResourceInputModel()
	modelAspectEventResource1.MinVersion = nil
	modelAspectEventResource2 := fixAspectEventResourceInputModel()
	modelAspectEventResource2.MinVersion = nil

	gqlAspectEventDefInputs := []*graphql.AspectEventDefinitionInput{gqlAspectEventDef1, gqlAspectEventDef2, nil}
	modelAspectEventResourceInputs := []*model.AspectEventResourceInput{&modelAspectEventResource1, &modelAspectEventResource2}
	testCases := []struct {
		Name     string
		Input    []*graphql.AspectEventDefinitionInput
		Expected []*model.AspectEventResourceInput
	}{
		{
			Name:     "All properties given",
			Input:    gqlAspectEventDefInputs,
			Expected: modelAspectEventResourceInputs,
		},
		{
			Name:     "Nil",
			Expected: []*model.AspectEventResourceInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := aspecteventresource.NewConverter()

			// WHEN
			res, err := converter.MultipleInputFromGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
