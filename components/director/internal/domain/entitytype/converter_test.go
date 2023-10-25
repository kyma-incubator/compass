package entitytype_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/entitytype"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entityTypeModel := fixEntityTypeModel(entityTypeID)
		require.NotNil(t, entityTypeModel)
		conv := entitytype.NewConverter(version.NewConverter())

		entity := conv.ToEntity(entityTypeModel)

		assert.Equal(t, fixEntityTypeEntity(entityTypeID), entity)
	})

	t.Run("Returns nil if entity type model is nil", func(t *testing.T) {
		conv := entitytype.NewConverter(version.NewConverter())

		entityTypeEntity := conv.ToEntity(nil)

		require.Nil(t, entityTypeEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityTypeEntity(entityTypeID)
		conv := entitytype.NewConverter(version.NewConverter())

		entityTypeModel := conv.FromEntity(entity)

		assert.Equal(t, fixEntityTypeModel(entityTypeID), entityTypeModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := entitytype.NewConverter(version.NewConverter())

		entityTypeModel := conv.FromEntity(nil)

		require.Nil(t, entityTypeModel)
	})
}
