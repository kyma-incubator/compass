package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/api/middlewares"
	validatorMocks "github.com/kyma-incubator/compass/components/provisioner/internal/api/mocks"

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
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	clusterConfig := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion:      "1.15.4",
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
		validator := &validatorMocks.Validator{}
		resolver := NewResolver(provisioningService, validator)

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
			Operation: gqlschema.OperationTypeProvision,
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

		provisioningService.On("ProvisionRuntime", config, tenant, "").Return(operation, nil)
		validator.On("ValidateProvisioningInput", config).Return(nil)

		//when
		status, err := resolver.ProvisionRuntime(ctx, config)

		//then
		require.NoError(t, err)
		require.NotNil(t, status)
		require.NotNil(t, status.ID)
		require.NotNil(t, status.RuntimeID)
		assert.Equal(t, operationID, *status.ID)
		assert.Equal(t, runtimeID, *status.RuntimeID)
		assert.Equal(t, gqlschema.OperationStateInProgress, status.State)
		assert.Equal(t, gqlschema.OperationTypeProvision, status.Operation)
		assert.Equal(t, util.StringPtr("Message"), status.Message)
	})

	t.Run("Should return error when requested provisioning on GCP", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

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

		validator.On("ValidateProvisioningInput", config).Return(nil)

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("Should return error when Kyma config validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
		}

		config := gqlschema.ProvisionRuntimeInput{RuntimeInput: runtimeInput, ClusterConfig: clusterConfig, Credentials: providerCredentials, KymaConfig: kymaConfig}

		validator.On("ValidateProvisioningInput", config).Return(errors.New("Some error"))

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("Should return error when provisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

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

		provisioningService.On("ProvisionRuntime", config, tenant, "").Return(nil, errors.New("Provisioning failed"))
		validator.On("ValidateProvisioningInput", config).Return(nil)

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})

	t.Run("Should fail when tenant header is not passed to context", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
			},
		}

		config := gqlschema.ProvisionRuntimeInput{
			RuntimeInput:  runtimeInput,
			ClusterConfig: clusterConfig,
			Credentials:   providerCredentials,
			KymaConfig:    kymaConfig,
		}

		validator.On("ValidateProvisioningInput", config).Return(nil)

		ctx := context.Background()

		//when
		status, err := provisioner.ProvisionRuntime(ctx, config)

		//then
		require.Error(t, err)
		assert.Nil(t, status)
	})
}

func TestResolver_DeprovisionRuntime(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	t.Run("Should start deprovisioning and return operation ID", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		provisioningService.On("DeprovisionRuntime", runtimeID, tenant).Return(expectedID, nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.NoError(t, err)
		assert.Equal(t, expectedID, operationID)
	})

	t.Run("Should return error when deprovisioning fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		provisioningService.On("DeprovisionRuntime", runtimeID, tenant).Return("", errors.New("Deprovisioning fails because reasons"))
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.Error(t, err)
		assert.Empty(t, operationID)
	})

	t.Run("Should fail when tenant header is not passed to context", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		provisioningService.On("DeprovisionRuntime", runtimeID, tenant).Return(expectedID, nil, nil)

		ctx := context.Background()

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.Error(t, err)
		require.Empty(t, operationID)
	})

	t.Run("Should fail deprovisioning when tenant validation fail", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		expectedID := "ec781980-0533-4098-aab7-96b535569732"

		provisioningService.On("DeprovisionRuntime", runtimeID, tenant).Return(expectedID, nil, nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(errors.New("Very bad error"))

		//when
		operationID, err := provisioner.DeprovisionRuntime(ctx, runtimeID)

		//then
		require.Error(t, err)
		assert.Empty(t, operationID)
	})
}

func TestResolver_UpgradeRuntime(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	upgradeInput := gqlschema.UpgradeRuntimeInput{
		KymaConfig: fixKymaGraphQLConfigInput(),
	}

	t.Run("Should start upgrade and return operation id", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		operation := &gqlschema.OperationStatus{
			ID:        util.StringPtr(operationID),
			Operation: gqlschema.OperationTypeUpgrade,
			State:     gqlschema.OperationStateInProgress,
			Message:   util.StringPtr("Message"),
			RuntimeID: util.StringPtr(runtimeID),
		}

		provisioningService.On("UpgradeRuntime", runtimeID, upgradeInput).Return(operation, nil)
		validator.On("ValidateUpgradeInput", upgradeInput).Return(nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		resolver := NewResolver(provisioningService, validator)

		//when
		status, err := resolver.UpgradeRuntime(ctx, runtimeID, upgradeInput)

		//then
		require.NoError(t, err)
		require.NotNil(t, status)
		require.NotNil(t, status.ID)
		require.NotNil(t, status.RuntimeID)
		assert.Equal(t, operation, status)
	})

	t.Run("Should return error when upgrade fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		provisioningService.On("UpgradeRuntime", runtimeID, upgradeInput).Return(nil, errors.Errorf("error"))
		validator.On("ValidateUpgradeInput", upgradeInput).Return(nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		resolver := NewResolver(provisioningService, validator)

		//when
		_, err := resolver.UpgradeRuntime(ctx, runtimeID, upgradeInput)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when tenant validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		validator.On("ValidateUpgradeInput", upgradeInput).Return(nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(errors.Errorf("error"))

		resolver := NewResolver(provisioningService, validator)

		//when
		_, err := resolver.UpgradeRuntime(ctx, runtimeID, upgradeInput)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		validator.On("ValidateUpgradeInput", upgradeInput).Return(errors.Errorf("error"))
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		resolver := NewResolver(provisioningService, validator)

		//when
		_, err := resolver.UpgradeRuntime(ctx, runtimeID, upgradeInput)

		//then
		require.Error(t, err)
	})
}

func TestResolver_RollBackUpgradeOperation(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	runtimeStatus := gqlschema.RuntimeStatus{}

	t.Run("Should start upgrade and return operation id", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		provisioningService.On("RollBackLastUpgrade", runtimeID).Return(&runtimeStatus, nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		resolver := NewResolver(provisioningService, validator)

		//when
		status, err := resolver.RollBackUpgradeOperation(ctx, runtimeID)

		//then
		require.NoError(t, err)
		require.NotNil(t, status)
	})

	t.Run("Should return error when failed to roll back upgrade", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}

		provisioningService.On("RollBackLastUpgrade", runtimeID).Return(nil, fmt.Errorf("error"))
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		resolver := NewResolver(provisioningService, validator)

		//when
		_, err := resolver.RollBackUpgradeOperation(ctx, runtimeID)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when failed to validate tenant", func(t *testing.T) {
		//given
		validator := &validatorMocks.Validator{}
		validator.On("ValidateTenant", runtimeID, tenant).Return(fmt.Errorf("error"))

		resolver := NewResolver(nil, validator)

		//when
		_, err := resolver.RollBackUpgradeOperation(ctx, runtimeID)

		//then
		require.Error(t, err)
	})
}

func TestResolver_RuntimeStatus(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

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
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		//when
		runtimeStatus, err := provisioner.RuntimeStatus(ctx, runtimeID)

		//then
		require.NoError(t, err)
		assert.Equal(t, status, runtimeStatus)
	})

	t.Run("Should return error when runtime status fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		provisioningService.On("RuntimeStatus", runtimeID).Return(nil, errors.New("Runtime status fails"))
		validator.On("ValidateTenant", runtimeID, tenant).Return(nil)

		//when
		status, err := provisioner.RuntimeStatus(ctx, runtimeID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})

	t.Run("Should return error when tenant header does not match tenant provided during provisioning", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

		provisioningService.On("RuntimeStatus", runtimeID).Return(nil, nil)
		validator.On("ValidateTenant", runtimeID, tenant).Return(errors.New("Bad error"))

		//when
		status, err := provisioner.RuntimeStatus(ctx, runtimeID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})
}

func TestResolver_RuntimeOperationStatus(t *testing.T) {
	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)
	runtimeID := "1100bb59-9c40-4ebb-b846-7477c4dc5bbd"

	t.Run("Should return operation status", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

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
		validator.On("ValidateTenantForOperation", operationID, tenant).Return(nil)

		//when
		status, err := provisioner.RuntimeOperationStatus(ctx, operationID)

		//then
		require.NoError(t, err)
		assert.Equal(t, operationStatus, status)
	})

	t.Run("Should return error when tenant validation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		provisioner := NewResolver(provisioningService, validator)

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
		validator.On("ValidateTenantForOperation", operationID, tenant).Return(errors.New("oh no"))
		//when
		status, err := provisioner.RuntimeOperationStatus(ctx, operationID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})

	t.Run("Should return error when Runtime Operation fails", func(t *testing.T) {
		//given
		provisioningService := &mocks.Service{}
		validator := &validatorMocks.Validator{}
		validator.On("ValidateTenantForOperation", operationID, tenant).Return(nil)
		provisioner := NewResolver(provisioningService, validator)

		provisioningService.On("RuntimeOperationStatus", operationID).Return(nil, errors.New("Some error"))

		//when
		status, err := provisioner.RuntimeOperationStatus(ctx, operationID)

		//then
		require.Error(t, err)
		require.Empty(t, status)
	})
}
