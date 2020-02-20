package postsql

import (
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProvisioningToStorageConversion checks if the conversion between model and DTO works properly for provisioning operation
func TestProvisioningOperationToStorageConversion(t *testing.T) {
	// given
	provisioningOperation := fixProvisioningOperation()

	// when
	dto, err := provisioningOperationToDTO(&provisioningOperation)
	require.NoError(t, err)
	newObj, err := toProvisioningOperation(&dto)
	require.NoError(t, err)

	// then
	assert.Equal(t, provisioningOperation, *newObj)
}

func fixProvisioningOperation() internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:                     "id-001",
			InstanceID:             "instance-id-001",
			Version:                0,
			CreatedAt:              time.Now(),
			UpdatedAt:              time.Now().Add(time.Second),
			Description:            "some description",
			State:                  domain.InProgress,
			ProvisionerOperationID: "target-op-id",
		},
		LmsTenantID: "l-t-id",
	}
}
