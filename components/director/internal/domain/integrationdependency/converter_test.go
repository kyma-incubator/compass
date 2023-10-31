package integrationdependency_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/integrationdependency"
	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		integrationDependencyModel := fixIntegrationDependencyModel(integrationDependencyID)
		require.NotNil(t, integrationDependencyModel)
		conv := integrationdependency.NewConverter(version.NewConverter())

		entity := conv.ToEntity(integrationDependencyModel)

		assert.Equal(t, fixIntegrationDependencyEntity(integrationDependencyID), entity)
	})

	t.Run("Returns nil if integration dependency model is nil", func(t *testing.T) {
		conv := integrationdependency.NewConverter(version.NewConverter())

		integrationDependencyEntity := conv.ToEntity(nil)

		require.Nil(t, integrationDependencyEntity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		entity := fixIntegrationDependencyEntity(integrationDependencyID)
		conv := integrationdependency.NewConverter(version.NewConverter())

		integrationDependencyModel := conv.FromEntity(entity)

		assert.Equal(t, fixIntegrationDependencyModel(integrationDependencyID), integrationDependencyModel)
	})

	t.Run("Returns nil if Entity is nil", func(t *testing.T) {
		conv := integrationdependency.NewConverter(version.NewConverter())

		integrationDependencyModel := conv.FromEntity(nil)

		require.Nil(t, integrationDependencyModel)
	})
}
