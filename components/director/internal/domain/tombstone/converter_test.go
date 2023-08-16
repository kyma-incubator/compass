package tombstone_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		tombstoneModel := fixTombstoneModelForApp()
		require.NotNil(t, tombstoneModel)
		conv := tombstone.NewConverter()

		entity := conv.ToEntity(tombstoneModel)

		assert.Equal(t, fixEntityTombstoneForApp(), entity)
	})

	t.Run("Returns nil if tombstone model is nil", func(t *testing.T) {
		conv := tombstone.NewConverter()

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityTombstoneForApp()
		conv := tombstone.NewConverter()

		tombstoneModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixTombstoneModelForApp(), tombstoneModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := tombstone.NewConverter()

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
