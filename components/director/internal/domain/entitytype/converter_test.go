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
		etModel := fixEntityTypeModel(ID)
		require.NotNil(t, etModel)
		conv := entitytype.NewConverter(version.NewConverter())

		entity := conv.ToEntity(etModel)

		assert.Equal(t, fixEntityTypeEntity(ID), entity)
	})

	t.Run("Returns nil if package model is nil", func(t *testing.T) {
		conv := entitytype.NewConverter(version.NewConverter())

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityTypeEntity(ID)
		conv := entitytype.NewConverter(version.NewConverter())

		etModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixEntityTypeModel(ID), etModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := entitytype.NewConverter(version.NewConverter())

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
