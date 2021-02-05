package mp_package_test

import (
	"testing"

	mp_package "github.com/kyma-incubator/compass/components/director/internal/domain/package"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		pkgModel := fixPackageModel()
		require.NotNil(t, pkgModel)
		conv := mp_package.NewConverter()

		entity := conv.ToEntity(pkgModel)

		assert.Equal(t, fixEntityPackage(), entity)
	})

	t.Run("Returns nil if package model is nil", func(t *testing.T) {
		conv := mp_package.NewConverter()

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		entity := fixEntityPackage()
		conv := mp_package.NewConverter()

		pkgModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixPackageModel(), pkgModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := mp_package.NewConverter()

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
