package product_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/product"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		productModel := fixProductModelForApp()
		require.NotNil(t, productModel)
		conv := product.NewConverter()

		entity := conv.ToEntity(productModel)

		assert.Equal(t, fixEntityProductForApp(), entity)
	})

	t.Run("Returns nil if package model is nil", func(t *testing.T) {
		conv := product.NewConverter()

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityProductForApp()
		conv := product.NewConverter()

		productModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixProductModelForApp(), productModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := product.NewConverter()

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
