package tenant_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter(t *testing.T) {
	c := tenant.NewConverter()
	t.Run("all fields", func(t *testing.T) {
		input := model.BusinessTenantMapping{
			ID:             uuid.New().String(),
			Name:           "tenant",
			ExternalTenant: "external-tenant",
			Provider:       "SAP",
			Status:         model.Active,
		}
		//when
		entity := c.ToEntity(&input)
		outputModel := c.FromEntity(entity)

		//then
		assert.Equal(t, &input, outputModel)
	})

	t.Run("nil model", func(t *testing.T) {
		//When
		entity := c.ToEntity(nil)

		//Then
		require.Nil(t, entity)
	})

	t.Run("nil entity", func(t *testing.T) {
		//When
		tenantModel := c.FromEntity(nil)

		//Then
		require.Nil(t, tenantModel)
	})
}
