package bundlereferences_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success for API BundleReference", func(t *testing.T) {
		// GIVEN
		apiBundleReferenceModel := fixAPIBundleReferenceModel()
		conv := bundlereferences.NewConverter()

		// WHEN
		entity := conv.ToEntity(apiBundleReferenceModel)

		// THEN
		assert.Equal(t, fixAPIBundleReferenceEntity(), entity)
	})
	t.Run("success for Event BundleReference", func(t *testing.T) {
		// GIVEN
		eventBundleReferenceModel := fixEventBundleReferenceModel()
		conv := bundlereferences.NewConverter()

		// WHEN
		entity := conv.ToEntity(eventBundleReferenceModel)

		// THEN
		assert.Equal(t, fixEventBundleReferenceEntity(), entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success for API BundleReference", func(t *testing.T) {
		// GIVEN
		apiBundleReferenceEntity := fixAPIBundleReferenceEntity()
		conv := bundlereferences.NewConverter()

		// WHEN
		model, err := conv.FromEntity(apiBundleReferenceEntity)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, fixAPIBundleReferenceModel(), model)
	})
	t.Run("error for API BundleReference", func(t *testing.T) {
		// GIVEN
		invalidAPIBundleReferenceEntity := fixInvalidAPIBundleReferenceEntity()
		conv := bundlereferences.NewConverter()

		// WHEN
		_, err := conv.FromEntity(invalidAPIBundleReferenceEntity)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while determining object reference")
	})
	t.Run("success for Event BundleReference", func(t *testing.T) {
		// GIVEN
		eventBundleReferenceEntity := fixEventBundleReferenceEntity()
		conv := bundlereferences.NewConverter()

		// WHEN
		model, err := conv.FromEntity(eventBundleReferenceEntity)

		// THEN
		assert.NoError(t, err)
		assert.Equal(t, fixEventBundleReferenceModel(), model)
	})
	t.Run("error for Event BundleReference", func(t *testing.T) {
		// GIVEN
		invalidEventBundleReferenceEntity := fixInvalidEventBundleReferenceEntity()
		conv := bundlereferences.NewConverter()

		// WHEN
		_, err := conv.FromEntity(invalidEventBundleReferenceEntity)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while determining object reference")
	})
}
