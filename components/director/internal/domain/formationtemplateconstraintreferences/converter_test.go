package formationtemplateconstraintreferences_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationtemplateconstraintreferences"
	"github.com/stretchr/testify/require"
	"testing"
)

var converter = formationtemplateconstraintreferences.NewConverter()

func TestToEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.ToEntity(constraintReference)

		// THEN
		require.Equal(t, entity, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.ToEntity(nil)

		// THEN
		require.Nil(t, actual)
	})
}

func TestFromEntity(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// WHEN
		actual := converter.FromEntity(entity)

		// THEN
		require.Equal(t, constraintReference, actual)
	})
	t.Run("Nil input", func(t *testing.T) {
		// WHEN
		actual := converter.FromEntity(nil)

		// THEN
		require.Nil(t, actual)
	})
}
