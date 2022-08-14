package externaltenant_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTenants(t *testing.T) {
	// GIVEN
	validTenantSrcPath := "./testdata/tenants/"
	directoryWithInvalidTenantJSON := "./testdata/invalidtenants/"
	directoryWithInvalidFiles := "./testdata/invalidfiletypes/"
	invalidPath := "foo"

	firstProvider := "valid_tenants.json"
	secondProvider := "valid_tenants2.json"

	expectedTenants := []model.BusinessTenantMappingInput{
		{
			Name:           "default",
			ExternalTenant: "id-default",
			Provider:       firstProvider,
			Region:         "eu-1",
			Type:           string(tenant.Subaccount),
		},
		{
			Name:           "foo",
			ExternalTenant: "id-foo",
			Provider:       firstProvider,
			Region:         "eu-2",
			Type:           string(tenant.Subaccount),
		},
		{
			Name:           "bar",
			ExternalTenant: "id-bar",
			Provider:       secondProvider,
		},
		{
			Name:           "baz",
			ExternalTenant: "id-baz",
			Provider:       secondProvider,
		},
	}

	t.Run("should return tenants", func(t *testing.T) {
		// WHEN
		actualTenants, err := externaltenant.MapTenants(validTenantSrcPath, "eu-1")

		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedTenants, actualTenants)
	})

	t.Run("should fail while reading tenants directory", func(t *testing.T) {
		// WHEN
		_, err := externaltenant.MapTenants(invalidPath, "")

		// THEN
		require.Error(t, err)
		assert.Equal(t, err.Error(), fmt.Sprintf("while reading directory with tenant files [%s]: open %s: no such file or directory", invalidPath, invalidPath))
	})

	t.Run("should fail while reading file with tenants - unsupported file extension", func(t *testing.T) {
		// WHEN
		_, err := externaltenant.MapTenants(directoryWithInvalidFiles, "")

		// THEN
		require.Error(t, err)
		assert.Equal(t, err.Error(), apperrors.NewInternalError("unsupported file format [.txt]").Error())
	})

	t.Run("should fail while unmarshalling tenants", func(t *testing.T) {
		// WHEN
		_, err := externaltenant.MapTenants(directoryWithInvalidTenantJSON, "")

		// THEN
		require.Error(t, err)
		assert.Equal(t, err.Error(), fmt.Sprintf("while unmarshalling tenants from file [%s%s]: invalid character '\\n' in string literal", directoryWithInvalidTenantJSON, "invalid.json"))
	})
}
