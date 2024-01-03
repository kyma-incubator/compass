package dataproduct_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/dataproduct"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		dataProductModel := fixDataProductModel(dataProductID)
		require.NotNil(t, dataProductModel)

		conv := dataproduct.NewConverter(version.NewConverter())

		entity := conv.ToEntity(dataProductModel)

		assert.Equal(t, fixDataProductEntity(dataProductID, appID), entity)
	})

	t.Run("Returns nil if data product model is nil", func(t *testing.T) {
		conv := dataproduct.NewConverter(version.NewConverter())

		dataProductEntity := conv.ToEntity(nil)

		require.Nil(t, dataProductEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixDataProductEntity(dataProductID, appID)
		conv := dataproduct.NewConverter(version.NewConverter())

		dataProductModel := conv.FromEntity(entity)

		assert.Equal(t, fixDataProductModel(dataProductID), dataProductModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := dataproduct.NewConverter(version.NewConverter())

		dataProductModel := conv.FromEntity(nil)

		require.Nil(t, dataProductModel)
	})
}
