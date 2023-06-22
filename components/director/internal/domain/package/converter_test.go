package ordpackage_test

import (
	"testing"

	ordpackage "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pkgModel := fixPackageModelForApp()
		require.NotNil(t, pkgModel)
		conv := ordpackage.NewConverter()

		entity := conv.ToEntity(pkgModel)

		assert.Equal(t, fixEntityPackageForApp(), entity)
	})

	t.Run("Returns nil if package model is nil", func(t *testing.T) {
		conv := ordpackage.NewConverter()

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityPackageForApp()
		conv := ordpackage.NewConverter()

		pkgModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixPackageModelForApp(), pkgModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := ordpackage.NewConverter()

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
