package capability_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/capability"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled with Application ID", func(t *testing.T) {
		// GIVEN
		capabilityModel, _ := fixFullCapabilityModelWithAppID("test-capability")
		require.NotNil(t, capabilityModel)

		versionConv := version.NewConverter()
		conv := capability.NewConverter(versionConv)
		// WHEN
		entity := conv.ToEntity(&capabilityModel)
		// THEN
		assert.Equal(t, fixFullEntityCapabilityWithAppID(capabilityID, "test-capability"), *entity)
	})

	t.Run("success all nullable properties filled with Application Template Version ID", func(t *testing.T) {
		// GIVEN
		capabilityModel, _ := fixFullCapabilityModelWithAppTemplateVersionID("test-capability")
		require.NotNil(t, capabilityModel)

		versionConv := version.NewConverter()
		conv := capability.NewConverter(versionConv)
		// WHEN
		entity := conv.ToEntity(&capabilityModel)
		// THEN
		assert.Equal(t, fixFullEntityCapabilityWithAppTemplateVersionID(capabilityID, "test-capability"), *entity)
	})

	t.Run("success all nullable properties are empty", func(t *testing.T) {
		// GIVEN
		capabilityModel := fixCapabilityModel("id", "test-capability")
		require.NotNil(t, capabilityModel)

		versionConv := version.NewConverter()
		conv := capability.NewConverter(versionConv)
		// WHEN
		entity := conv.ToEntity(capabilityModel)
		// THEN
		assert.Equal(t, fixEntityCapability("id", "test-capability"), entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		entity := fixFullEntityCapabilityWithAppID(capabilityID, "test-capability")

		versionConv := version.NewConverter()
		conv := capability.NewConverter(versionConv)
		// WHEN
		capabilityModel := conv.FromEntity(&entity)
		// THEN
		expectedModel, _ := fixFullCapabilityModelWithAppID("test-capability")
		assert.Equal(t, &expectedModel, capabilityModel)
	})

	t.Run("success all nullable properties are empty", func(t *testing.T) {
		// GIVEN
		entity := fixEntityCapability("id", "test-capability")

		versionConv := version.NewConverter()
		conv := capability.NewConverter(versionConv)
		// WHEN
		capabilityModel := conv.FromEntity(entity)
		// THEN
		expectedModel := fixCapabilityModel("id", "test-capability")
		require.NotNil(t, expectedModel)
		assert.Equal(t, expectedModel, capabilityModel)
	})
}
