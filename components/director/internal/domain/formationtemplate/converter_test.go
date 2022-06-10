package formationtemplate_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConverter_FromInputGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.FromInputGraphQL(&inputFormationTemplateGraphQLModel)

		// THEN
		assert.Equal(t, *result, inputFormationTemplateModel)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.FromInputGraphQL(nil)

		assert.Nil(t, result)
	})
}

func TestConverter_FromModelInputToModel(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.FromModelInputToModel(&inputFormationTemplateModel, testID)

		// THEN
		assert.Equal(t, *result, formationTemplateModel)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.FromModelInputToModel(nil, "")

		assert.Nil(t, result)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.ToGraphQL(&formationTemplateModel)

		// THEN
		assert.Equal(t, *result, formationTemplateGraphQLModel)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result := converter.ToGraphQL(nil)

		assert.Nil(t, result)
	})
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
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
	t.Run("Success", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result, err := converter.FromEntity(&formationTemplateEntity)

		// THEN
		require.NoError(t, err)
		assert.Equal(t, result, &formationTemplateModel)
	})
	t.Run("Returns nil when given empty model", func(t *testing.T) {
		// GIVEN
		converter := formationtemplate.NewConverter()
		// WHEN
		result, err := converter.FromEntity(nil)

		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}
