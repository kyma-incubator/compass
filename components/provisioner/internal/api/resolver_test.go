package api

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	operationID = "ec781980-0533-4098-aab7-96b535569732"
	runtimeID   = "1100bb59-9c40-4ebb-b846-7477c4dc5bbb"
)

func TestResolver_ProvisionRuntime(t *testing.T) {
	ctx := context.Background()

	clusterConfig := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion:      "1.15.4",
			NodeCount:              3,
			VolumeSizeGb:           30,
			MachineType:            "n1-standard-4",
			Region:                 "europe",
			Provider:               "gcp",
			Seed:                   util.StringPtr(""),
			TargetSecret:           "test-secret",
			DiskType:               "ssd",
			WorkerCidr:             "10.10.10.10/255",
			AutoScalerMin:          1,
			AutoScalerMax:          3,
			MaxSurge:               40,
			MaxUnavailable:         1,
			ProviderSpecificConfig: nil,
		},
	}

	runtimeInput := &gqlschema.RuntimeInput{
		Name:        "test runtime",
		Description: new(string),
	}

	providerCredentials := &gqlschema.CredentialsInput{SecretName: "secret_1"}

	t.Run("Should start provisioning and return operation ID", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
			},
		}

		operation := &gqlschema.OperationStatus{
			ID:        util.StringPtr(operationID),
			Operation: operationID,
			State:     gqlschema.OperationStateInProgress,
			Message:   util.StringPtr("Message"),
			RuntimeID: util.StringPtr(runtimeID),
		}

		config := gqlschema.ProvisionRuntimeInput{
			RuntimeInput:  runtimeInput,
			ClusterConfig: clusterConfig,
			Credentials:   providerCredentials,
			KymaConfig:    kymaConfig,
		}

		provisioningService.On("ProvisionRuntime", config).Return(operation, nil)

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.NoError(t, err)
		require.NotNil(t, status)
		require.NotNil(t, status.ID)
		require.NotNil(t, status.RuntimeID)
		assert.Equal(t, operationID, *status.ID)
		assert.Equal(t, runtimeID, *status.RuntimeID)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.State)
		assert.Equal(t, util.StringPtr("Message"), status.Message)
	})

	t.Run("Should return error when requested provisioning on GCP", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		clusterConfig := &gqlschema.ClusterConfigInput{
			GcpConfig: &gqlschema.GCPConfigInput{
				Name:              "Something",
				ProjectName:       "Project",
				NumberOfNodes:     3,
				BootDiskSizeGb:    256,
				MachineType:       "machine",
				Region:            "region",
				Zone:              new(string),
				KubernetesVersion: "version",
			},
		}

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
			},
		}

		config := gqlschema.ProvisionRuntimeInput{RuntimeInput: runtimeInput, ClusterConfig: clusterConfig, KymaConfig: kymaConfig}

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("Should return error when Kyma config validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
		}

		config := gqlschema.ProvisionRuntimeInput{RuntimeInput: runtimeInput, ClusterConfig: clusterConfig, Credentials: providerCredentials, KymaConfig: kymaConfig}

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("Should return error when provisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
			},
		}

		config := gqlschema.ProvisionRuntimeInput{RuntimeInput: runtimeInput, ClusterConfig: clusterConfig, Credentials: providerCredentials, KymaConfig: kymaConfig}

		provisioningService.On("ProvisionRuntime", config).Return(nil, errors.New("Provisioning failed"))

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	ctx := context.Background()

	t.Run("Should start deprovisioning and return operation ID", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		provisioningService.On("DeprovisionRuntime", runtimeID).Return(expectedID, nil)

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedID, operationID)
	})

	t.Run("Should return error when deprovisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		provisioner := NewResolver(provisioningService)

		provisioningService.On("DeprovisionRuntime", runtimeID).Return("", errors.New("Deprovisioning fails because reasons"))

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

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
