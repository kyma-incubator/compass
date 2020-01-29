package externaltenant_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTenants(t *testing.T) {
	//given
	validSrcPath := "./testdata/valid_tenants.json"
	pathToInvalidJSON := "./testdata/invalid.json"
	invalidPath := "foo"

	provider := "testProvider"
	expectedTenants := []model.BusinessTenantMappingInput{
		{
			Name:           "default",
			ExternalTenant: "id-default",
			Provider:       provider,
		},
		{
			Name:           "foo",
			ExternalTenant: "id-foo",
			Provider:       provider,
		},
		{
			Name:           "bar",
			ExternalTenant: "id-bar",
			Provider:       provider,
		},
	}

	t.Run("should return tenants", func(t *testing.T) {
		//when
		actualTenants, err := externaltenant.MapTenants(validSrcPath, provider)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedTenants, actualTenants)
	})

	t.Run("should fail while reading tenants file", func(t *testing.T) {
		//when
		_, err := externaltenant.MapTenants(invalidPath, provider)

		//then
		require.Error(t, err)
	})

	t.Run("should fail while unmarshalling tenants", func(t *testing.T) {
		//when
		_, err := externaltenant.MapTenants(pathToInvalidJSON, provider)

		//then
		require.Error(t, err)
	})
}
