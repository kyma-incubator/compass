package aspect_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/aspect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		aspectModel := fixAspectModel(aspectID)
		require.NotNil(t, aspectModel)
		conv := aspect.NewConverter()

		entity := conv.ToEntity(aspectModel)

		assert.Equal(t, fixEntityAspect(aspectID), entity)
	})

	t.Run("Returns nil if aspect model is nil", func(t *testing.T) {
		conv := aspect.NewConverter()

		aspectEntity := conv.ToEntity(nil)

		require.Nil(t, aspectEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityAspect(aspectID)
		conv := aspect.NewConverter()

		aspectModel := conv.FromEntity(entity)

		assert.Equal(t, fixAspectModel(aspectID), aspectModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := aspect.NewConverter()

		aspectModel := conv.FromEntity(nil)

		require.Nil(t, aspectModel)
	})
}
