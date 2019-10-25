package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolver_ProvisionRuntime(t *testing.T) {
	ctx := context.Background()
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	clusterConfig := &gqlschema.ClusterConfigInput{
		GcpConfig: &gqlschema.GCPConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			NumberOfNodes:     3,
			BootDiskSize:      "256",
			MachineType:       "machine",
			Region:            "region",
			Zone:              new(string),
			KubernetesVersion: "version",
		},
	}

	t.Run("Should start provisioning and return operation ID", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Modules: gqlschema.AllKymaModule,
		}

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, KymaConfig: kymaConfig}

		provisioningService.On("ProvisionRuntime", runtimeID, config).Return(expectedID, nil, nil)

		//when
		operationID, err := provisioner.ProvisionRuntime(ctx, runtimeID, config)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedID, operationID)
	})

	t.Run("Should return error when Kyma config validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
		}

		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, KymaConfig: kymaConfig}

		//when
		operationID, err := provisioner.ProvisionRuntime(ctx, runtimeID, config)

		//then
		require.Error(t, err)
		assert.Empty(t, operationID)
	})

	t.Run("Should return error when provisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Modules: gqlschema.AllKymaModule,
		}

		config := gqlschema.ProvisionRuntimeInput{ClusterConfig: clusterConfig, KymaConfig: kymaConfig}

		provisioningService.On("ProvisionRuntime", runtimeID, config).Return("", nil, errors.New("Provisioning failed"))

		//when
		operationID, err := provisioner.ProvisionRuntime(ctx, runtimeID, config)

		//then
		require.Error(t, err)
		assert.Empty(t, operationID)
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	ctx := context.Background()
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	t.Run("Should start deprovisioning and return operation ID", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		credentials := gqlschema.CredentialsInput{SecretName: "secretName"}

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		provisioningService.On("DeprovisionRuntime", runtimeID, credentials).Return(expectedID, nil, nil)

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID, credentials)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedID, operationID)
	})

	t.Run("Should return error when deprovisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		credentials := gqlschema.CredentialsInput{SecretName: "secretName"}

		provisioningService.On("DeprovisionRuntime", runtimeID, credentials).Return("", nil, errors.New("Deprovisioning fails because reasons"))

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID, credentials)

		//then
		require.Error(t, err)
		assert.Empty(t, operationID)
	})
}

func TestResolver_RuntimeStatus(t *testing.T) {
	ctx := context.Background()
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		operationID := "acc5040c-3bb6-47b8-8651-07f6950bd0a7"
		message := "some message"

		status := &gqlschema.RuntimeStatus{
			LastOperationStatus: &gqlschema.OperationStatus{
				ID:        &operationID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateInProgress,
				RuntimeID: &runtimeID,
				Message:   &message,
			},
			RuntimeConfiguration:    &gqlschema.RuntimeConfig{},
			RuntimeConnectionStatus: &gqlschema.RuntimeConnectionStatus{},
		}

		provisioningService.On("RuntimeStatus", runtimeID).Return(status, nil)

		//when
		runtimeStatus, err := provisioner.RuntimeStatus(ctx, runtimeID)

		//then
		require.NoError(t, err)
		assert.Equal(t, status, runtimeStatus)
	})

	t.Run("Should return error when runtime status fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		provisioningService.On("RuntimeStatus", runtimeID).Return(nil, errors.New("Runtime status fails"))

		//when
		status, err := provisioner.RuntimeStatus(ctx, runtimeID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})
}

func TestResolver_RuntimeOperationStatus(t *testing.T) {
	ctx := context.Background()
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		operationID := "acc5040c-3bb6-47b8-8651-07f6950bd0a7"
		message := "some message"

		operationStatus := &gqlschema.OperationStatus{
			ID:        &operationID,
			Operation: gqlschema.OperationTypeProvision,
			State:     gqlschema.OperationStateInProgress,
			RuntimeID: &runtimeID,
			Message:   &message,
		}

		provisioningService.On("RuntimeOperationStatus", operationID).Return(operationStatus, nil)

		//when
		status, err := provisioner.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, err)
		assert.Equal(t, operationStatus, status)
	})

	t.Run("Should return error when Runtime Operation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		operationID := "acc5040c-3bb6-47b8-8651-07f6950bd0a7"

		provisioningService.On("RuntimeOperationStatus", operationID).Return(nil, errors.New("Some error"))

		//when
		status, err := provisioner.RuntimeOperationStatus(ctx, operationID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})
}
