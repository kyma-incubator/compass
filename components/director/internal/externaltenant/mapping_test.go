package externaltenant_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/externaltenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapTenants(t *testing.T) {
	//given
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
		},
		{
			Name:           "foo",
			ExternalTenant: "id-foo",
			Provider:       firstProvider,
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
		//when
		actualTenants, err := externaltenant.MapTenants(validTenantSrcPath)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedTenants, actualTenants)
	})

	t.Run("should fail while reading tenants directory", func(t *testing.T) {
		//when
		_, err := externaltenant.MapTenants(invalidPath)

		//then
		require.Error(t, err)
		assert.Equal(t, err.Error(), fmt.Sprintf("while reading directory with tenant files [%s]: open %s: no such file or directory", invalidPath, invalidPath))
	})

	t.Run("should fail while reading file with tenants - unsupported file extension", func(t *testing.T) {
		//when
		_, err := externaltenant.MapTenants(directoryWithInvalidFiles)

		//then
		require.Error(t, err)
		assert.Equal(t, err.Error(), apperrors.NewInternalError("unsupported file format [.txt]").Error())
	})

	t.Run("should fail while unmarshalling tenants", func(t *testing.T) {
		//when
		_, err := externaltenant.MapTenants(directoryWithInvalidTenantJSON)

		//then
		require.Error(t, err)
		assert.Equal(t, err.Error(), fmt.Sprintf("while unmarshalling tenants from file [%s%s]: invalid character '\\n' in string literal", directoryWithInvalidTenantJSON, "invalid.json"))
	})
}
