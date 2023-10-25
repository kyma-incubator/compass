package entitytypemapping_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytypemapping"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entityTypeMappingModel := fixEntityTypeMappingModel(entityTypeMappingID)
		require.NotNil(t, entityTypeMappingModel)
		conv := entitytypemapping.NewConverter()

		entity := conv.ToEntity(entityTypeMappingModel)

		assert.Equal(t, fixEntityTypeMappingEntity(entityTypeMappingID), entity)
	})

	t.Run("Returns nil if entity type mapping model is nil", func(t *testing.T) {
		conv := entitytypemapping.NewConverter()

		entityTypeEntity := conv.ToEntity(nil)

		require.Nil(t, entityTypeEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityTypeMappingEntity(entityTypeMappingID)
		conv := entitytypemapping.NewConverter()

		entityTypeMappingModel := conv.FromEntity(entity)

		assert.Equal(t, fixEntityTypeMappingModel(entityTypeMappingID), entityTypeMappingModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := entitytypemapping.NewConverter()

		entityTypeModel := conv.FromEntity(nil)

		require.Nil(t, entityTypeModel)
	})
}
