package tenant_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter(t *testing.T) {
	// Given
	id := "foo"
	name := "bar"

	t.Run("all fields", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		input := newModelBusinessTenantMapping(id, name)
		entity := c.ToEntity(input)
		outputModel := c.FromEntity(entity)

		//then
		assert.Equal(t, input, outputModel)
	})

	t.Run("initialized from entity", func(t *testing.T) {
		c := tenant.NewConverter()
		initialized := true
		input := newEntityBusinessTenantMapping(id, name)
		input.Initialized = &initialized

		// When
		outputModel := c.FromEntity(input)

		// Then
		assert.Equal(t, input.Initialized, outputModel.Initialized)
	})

	t.Run("nil model", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		entity := c.ToEntity(nil)

		// Then
		require.Nil(t, entity)
	})

	t.Run("nil entity", func(t *testing.T) {
		c := tenant.NewConverter()

		// When
		tenantModel := c.FromEntity(nil)

		// Then
		require.Nil(t, tenantModel)
	})
}
