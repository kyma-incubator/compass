package api

import (
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	dbMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/require"
)

func TestValidator_ValidateInput(t *testing.T) {

	clusterConfig := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion:      "1.15.4",
			NodeCount:              3,
			VolumeSizeGb:           30,
			MachineType:            "n1-standard-4",
			Region:                 "europe",
			Provider:               "gcp",
			Seed:                   util.StringPtr("2"),
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

	t.Run("Should return nil when config is correct", func(t *testing.T) {
		//given
		validator := NewValidator(nil)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
				{
					Component:     "compass-runtime-agent",
					Configuration: nil,
				},
			},
		}

		providerCredentials := &gqlschema.CredentialsInput{SecretName: "secret_1"}

		config := gqlschema.ProvisionRuntimeInput{
			RuntimeInput:  runtimeInput,
			ClusterConfig: clusterConfig,
			Credentials:   providerCredentials,
			KymaConfig:    kymaConfig,
		}

		//when
		err := validator.ValidateInput(config)

		//then
		require.NoError(t, err)
	})

	t.Run("Should return error when config is incorrect", func(t *testing.T) {
		//given
		validator := NewValidator(nil)

		config := gqlschema.ProvisionRuntimeInput{}

		//when
		err := validator.ValidateInput(config)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when Runtime Agent component is not passed in installation config", func(t *testing.T) {
		//given
		validator := NewValidator(nil)

		kymaConfig := &gqlschema.KymaConfigInput{
			Version: "1.5",
			Components: []*gqlschema.ComponentConfigurationInput{
				{
					Component:     "core",
					Configuration: nil,
				},
			},
		}

		providerCredentials := &gqlschema.CredentialsInput{SecretName: "secret_1"}

		config := gqlschema.ProvisionRuntimeInput{
			RuntimeInput:  runtimeInput,
			ClusterConfig: clusterConfig,
			Credentials:   providerCredentials,
			KymaConfig:    kymaConfig,
		}

		//when
		err := validator.ValidateInput(config)

		//then
		require.Error(t, err)
	})
}

func TestValidator_ValidateTenant(t *testing.T) {
	tenant := "tenant"
	runtimeID := "123-123-123"
	t.Run("Should return nil when tenant matches tenant provided for Runtime", func(t *testing.T) {
		//given
		readSession := &dbMocks.ReadSession{}
		validator := NewValidator(readSession)

		expectedTenant := "tenant"

		readSession.On("GetTenant", runtimeID).Return(expectedTenant, nil)

		//when
		err := validator.ValidateTenant(runtimeID, tenant)

		//then
		require.NoError(t, err)
	})

	t.Run("Should return error when tenant does not match tenant provided for Runtime", func(t *testing.T) {
		//given
		readSession := &dbMocks.ReadSession{}
		validator := NewValidator(readSession)

		expectedTenant := "otherTenant"

		readSession.On("GetTenant", runtimeID).Return(expectedTenant, nil)

		//when
		err := validator.ValidateTenant(runtimeID, tenant)

		//then
		require.Error(t, err)
	})

	t.Run("Should return error when persistence service returns error", func(t *testing.T) {
		//given
		readSession := &dbMocks.ReadSession{}
		validator := NewValidator(readSession)

		readSession.On("GetTenant", runtimeID).Return("", dberrors.Internal("Some db error"))

		//when
		err := validator.ValidateTenant(runtimeID, tenant)

		//then
		require.Error(t, err)
	})
}
