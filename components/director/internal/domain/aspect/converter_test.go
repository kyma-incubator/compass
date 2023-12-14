package aspect_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspecteventresource"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aspectModel := fixAspectModel(aspectID)
		require.NotNil(t, aspectModel)
		conv := aspect.NewConverter(aspecteventresource.NewConverter())

		entity := conv.ToEntity(aspectModel)

		assert.Equal(t, fixEntityAspect(aspectID, appID, integrationDependencyID), entity)
	})

	t.Run("Returns nil if aspect model is nil", func(t *testing.T) {
		conv := aspect.NewConverter(nil)

		aspectEntity := conv.ToEntity(nil)

		require.Nil(t, aspectEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityAspect(aspectID, appID, integrationDependencyID)
		conv := aspect.NewConverter(aspecteventresource.NewConverter())

		aspectModel := conv.FromEntity(entity)

		assert.Equal(t, fixAspectModel(aspectID), aspectModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := aspect.NewConverter(nil)

		aspectModel := conv.FromEntity(nil)

		require.Nil(t, aspectModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	modelAspect := fixAspectModel(aspectID)
	gqlAspect := fixGQLAspect(aspectID)

	testCases := []struct {
		Name                      string
		InputAspect               *model.Aspect
		InputAspectEventResources []*model.AspectEventResource
		Expected                  *graphql.Aspect
	}{
		{
			Name:        "All properties given",
			InputAspect: modelAspect,
			Expected:    gqlAspect,
		},
		{
			Name:        "Nil",
			InputAspect: nil,
			Expected:    nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVE
			converter := aspect.NewConverter(aspecteventresource.NewConverter())

			// WHEN
			res, err := converter.ToGraphQL(testCase.InputAspect, testCase.InputAspectEventResources)

			// THEN
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	aspect1 := fixAspectModel("id1")
	aspect2 := fixAspectModel("id2")

	inputAspects := []*model.Aspect{aspect1, aspect2, nil}

	gqlAspect1 := fixGQLAspect("id1")
	gqlAspect2 := fixGQLAspect("id2")

	expected := []*graphql.Aspect{gqlAspect1, gqlAspect2}

	// WHEN
	converter := aspect.NewConverter(aspecteventresource.NewConverter())
	res, err := converter.MultipleToGraphQL(inputAspects, map[string][]*model.AspectEventResource{})
	assert.NoError(t, err)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlAspectInput := fixGQLAspectInput()
	gqlAspectInput.Mandatory = nil
	modelAspectInput := fixAspectInputModel()
	modelAspectInput.SupportMultipleProviders = nil

	testCases := []struct {
		Name     string
		Input    *graphql.AspectInput
		Expected *model.AspectInput
	}{
		{
			Name:     "All properties given",
			Input:    gqlAspectInput,
			Expected: &modelAspectInput,
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
			converter := aspect.NewConverter(aspecteventresource.NewConverter())

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
	gqlAspect1 := fixGQLAspectInput()
	gqlAspect2 := fixGQLAspectInput()

	modelAspect1 := fixAspectInputModel()
	modelAspect1.SupportMultipleProviders = nil
	modelAspect2 := fixAspectInputModel()
	modelAspect2.SupportMultipleProviders = nil

	gqlAspectInputs := []*graphql.AspectInput{gqlAspect1, gqlAspect2, nil}
	modelAspectInputs := []*model.AspectInput{&modelAspect1, &modelAspect2}
	testCases := []struct {
		Name     string
		Input    []*graphql.AspectInput
		Expected []*model.AspectInput
	}{
		{
			Name:     "All properties given",
			Input:    gqlAspectInputs,
			Expected: modelAspectInputs,
		},
		{
			Name:     "Nil",
			Expected: []*model.AspectInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := aspect.NewConverter(aspecteventresource.NewConverter())

			// WHEN
			res, err := converter.MultipleInputFromGraphQL(testCase.Input)

			// THEN
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
