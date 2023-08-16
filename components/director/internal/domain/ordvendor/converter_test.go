package ordvendor_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/ordvendor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		vendorModel := fixVendorModelForApp()
		require.NotNil(t, vendorModel)
		conv := ordvendor.NewConverter()

		entity := conv.ToEntity(vendorModel)

		assert.Equal(t, fixEntityVendor(), entity)
	})

	t.Run("Returns nil if vendor model is nil", func(t *testing.T) {
		conv := ordvendor.NewConverter()

		ent := conv.ToEntity(nil)

		require.Nil(t, ent)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixEntityVendor()
		conv := ordvendor.NewConverter()

		vendorModel, err := conv.FromEntity(entity)

		require.NoError(t, err)
		assert.Equal(t, fixVendorModelForApp(), vendorModel)
	})

	t.Run("Returns error if Entity is nil", func(t *testing.T) {
		conv := ordvendor.NewConverter()

		_, err := conv.FromEntity(nil)

		require.Error(t, err)
	})
}
