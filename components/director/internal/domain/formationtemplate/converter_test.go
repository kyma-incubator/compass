package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_FromInputGraphQL(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *graphql.FormationTemplateInput
		Expected *model.FormationTemplateInput
	}{{
		Name:     "Success",
		Input:    &inputFormationTemplateGraphQLModel,
		Expected: &inputFormationTemplateModel,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter()
			// WHEN
			result := converter.FromInputGraphQL(testCase.Input)

			assert.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_FromModelInputToModel(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationTemplateInput
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &inputFormationTemplateModel,
		Expected: &formationTemplateModel,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter()
			// WHEN
			result := converter.FromModelInputToModel(testCase.Input, testID)

			assert.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_ToGraphQL(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *model.FormationTemplate
		Expected *graphql.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateModel,
		Expected: &formationTemplateGraphQLModel,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter()
			// WHEN
			result := converter.ToGraphQL(testCase.Input)

			assert.Equal(t, result, testCase.Expected)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    []*model.FormationTemplate
		Expected []*graphql.FormationTemplate
	}{{
		Name:     "Success",
		Input:    formationTemplateModelPage.Data,
		Expected: formationTemplateGraphQLModelPage.Data,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter()
			// WHEN
			result := converter.MultipleToGraphQL(testCase.Input)

			assert.ElementsMatch(t, result, testCase.Expected)
		})
	}
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.MultipleToGraphQL(formationTemplateModelPage.Data)

		// THEN
		assert.ElementsMatch(t, result, formationTemplateGraphQLModelPage.Data)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.MultipleToGraphQL(nil)

		assert.Nil(t, result)
	})
}

func TestConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result, err := converter.ToEntity(&formationTemplateModel)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, &formationTemplateEntity)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result, err := converter.ToEntity(nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestConverter_FromEntity(t *testing.T) {
	testCases := []struct {
		Name     string
		Input    *formationtemplate.Entity
		Expected *model.FormationTemplate
	}{{
		Name:     "Success",
		Input:    &formationTemplateEntity,
		Expected: &formationTemplateModel,
	}, {
		Name:     "Empty",
		Input:    nil,
		Expected: nil,
	},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			converter := formationtemplate.NewConverter()
			// WHEN
			result, err := converter.FromEntity(testCase.Input)

			assert.NoError(t, err)
			assert.Equal(t, result, testCase.Expected)
		})
	}
}
