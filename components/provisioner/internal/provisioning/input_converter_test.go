package provisioning

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler"
	"testing"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	realeaseMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/release/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	hyperscalerMocks "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/hyperscaler/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid/mocks"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/stretchr/testify/assert"
)

const (
	kymaVersion                   = "1.5"
	clusterEssentialsComponent    = "cluster-essentials"
	coreComponent                 = "core"
	applicationConnectorComponent = "application-connector"
)

type InputConverterTester interface {
	createGardenerClusterName(provider string) string
}

func NewInputConverterTester(uuidGenerator uuid.UUIDGenerator, releaseRepo release.ReadRepository) InputConverterTester {
	return &converter{
		uuidGenerator: uuidGenerator,
		releaseRepo:   releaseRepo,
	}
}

func Test_ProvisioningInputToCluster(t *testing.T) {

	readSession := &realeaseMocks.Repository{}
	readSession.On("GetReleaseByVersion", kymaVersion).Return(fixKymaRelease(), nil)

	createGQLRuntimeInputGCP := func(zone *string) gqlschema.ProvisionRuntimeInput {
		return gqlschema.ProvisionRuntimeInput{
			RuntimeInput: &gqlschema.RuntimeInput{
				Name:        "runtimeName",
				Description: nil,
				Labels:      &gqlschema.Labels{},
			},
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GcpConfig: &gqlschema.GCPConfigInput{
					Name:              "Something",
					ProjectName:       "Project",
					NumberOfNodes:     3,
					BootDiskSizeGb:    256,
					MachineType:       "n1-standard-1",
					Region:            "region",
					Zone:              zone,
					KubernetesVersion: "version",
				},
			},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: "secretName",
			},
			KymaConfig: fixKymaGraphQLConfigInput(),
		}
	}

	createExpectedRuntimeInputGCP := func(zone string) model.Cluster {
		return model.Cluster{
			ID:          "runtimeID",
			RuntimeName: "runtimeName",
			ClusterConfig: model.GCPConfig{
				ID:                "id",
				Name:              "Something",
				ProjectName:       "Project",
				NumberOfNodes:     3,
				BootDiskSizeGB:    256,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Zone:              zone,
				KubernetesVersion: "version",
				ClusterID:         "runtimeID",
			},
			Kubeconfig:            nil,
			KymaConfig:            fixKymaConfig(),
			CredentialsSecretName: "secretName",
		}
	}

	gcpGardenerProvider := &gqlschema.GCPProviderConfigInput{Zone: "zone"}

	gardenerGCPGQLInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name:        "runtimeName",
			Description: nil,
			Labels:      &gqlschema.Labels{},
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "GCP",
				Seed:              "gcp-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					GcpConfig: gcpGardenerProvider,
				},
			},
		},
		Credentials: &gqlschema.CredentialsInput{
			SecretName: "secretName",
		},
		KymaConfig: fixKymaGraphQLConfigInput(),
	}

	expectedGCPProviderCfg, err := model.NewGCPGardenerConfig(gcpGardenerProvider)
	require.NoError(t, err)

	expectedGardenerGCPRuntimeConfig := model.Cluster{
		ID:          "runtimeID",
		RuntimeName: "runtimeName",
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "gcp-verylon",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "GCP",
			Seed:                   "gcp-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			GardenerProviderConfig: expectedGCPProviderCfg,
		},
		Kubeconfig:            nil,
		KymaConfig:            fixKymaConfig(),
		CredentialsSecretName: "secretName",
	}

	azureGardenerProvider := &gqlschema.AzureProviderConfigInput{VnetCidr: "cidr"}

	gardenerAzureGQLInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name:        "runtimeName",
			Description: nil,
			Labels:      &gqlschema.Labels{},
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "Azure",
				Seed:              "az-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					AzureConfig: azureGardenerProvider,
				},
			},
		},
		Credentials: &gqlschema.CredentialsInput{
			SecretName: "secretName",
		},
		KymaConfig: fixKymaGraphQLConfigInput(),
	}

	expectedAzureProviderCfg, err := model.NewAzureGardenerConfig(azureGardenerProvider)
	require.NoError(t, err)

	expectedGardenerAzureRuntimeConfig := model.Cluster{
		ID:          "runtimeID",
		RuntimeName: "runtimeName",
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "azu-verylon",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "Azure",
			Seed:                   "az-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			GardenerProviderConfig: expectedAzureProviderCfg,
		},
		Kubeconfig:            nil,
		KymaConfig:            fixKymaConfig(),
		CredentialsSecretName: "secretName",
	}

	awsGardenerProvider := &gqlschema.AWSProviderConfigInput{
		Zone:         "zone",
		InternalCidr: "cidr",
		VpcCidr:      "cidr",
		PublicCidr:   "cidr",
	}

	gardenerAWSGQLInput := gqlschema.ProvisionRuntimeInput{
		RuntimeInput: &gqlschema.RuntimeInput{
			Name:        "runtimeName",
			Description: nil,
			Labels:      &gqlschema.Labels{},
		},
		ClusterConfig: &gqlschema.ClusterConfigInput{
			GardenerConfig: &gqlschema.GardenerConfigInput{
				ProjectName:       "Project",
				KubernetesVersion: "version",
				NodeCount:         3,
				VolumeSizeGb:      1024,
				MachineType:       "n1-standard-1",
				Region:            "region",
				Provider:          "AWS",
				Seed:              "aws-eu1",
				TargetSecret:      "secret",
				DiskType:          "ssd",
				WorkerCidr:        "cidr",
				AutoScalerMin:     1,
				AutoScalerMax:     5,
				MaxSurge:          1,
				MaxUnavailable:    2,
				ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
					AwsConfig: awsGardenerProvider,
				},
			},
		},
		Credentials: &gqlschema.CredentialsInput{
			SecretName: "secretName",
		},
		KymaConfig: fixKymaGraphQLConfigInput(),
	}

	expectedAWSProviderCfg, err := model.NewAWSGardenerConfig(awsGardenerProvider)
	require.NoError(t, err)

	expectedGardenerAWSRuntimeConfig := model.Cluster{
		ID:          "runtimeID",
		RuntimeName: "runtimeName",
		ClusterConfig: model.GardenerConfig{
			ID:                     "id",
			Name:                   "aws-verylon",
			ProjectName:            "Project",
			MachineType:            "n1-standard-1",
			Region:                 "region",
			KubernetesVersion:      "version",
			NodeCount:              3,
			VolumeSizeGB:           1024,
			DiskType:               "ssd",
			Provider:               "AWS",
			Seed:                   "aws-eu1",
			TargetSecret:           "secret",
			WorkerCidr:             "cidr",
			AutoScalerMin:          1,
			AutoScalerMax:          5,
			MaxSurge:               1,
			MaxUnavailable:         2,
			ClusterID:              "runtimeID",
			GardenerProviderConfig: expectedAWSProviderCfg,
		},
		Kubeconfig:            nil,
		KymaConfig:            fixKymaConfig(),
		CredentialsSecretName: "secretName",
	}

	zone := "zone"

	configurations := []struct {
		input       gqlschema.ProvisionRuntimeInput
		expected    model.Cluster
		description string
	}{
		{
			input:       createGQLRuntimeInputGCP(&zone),
			expected:    createExpectedRuntimeInputGCP(zone),
			description: "Should create proper runtime config struct with GCP input",
		},
		{
			input:       createGQLRuntimeInputGCP(nil),
			expected:    createExpectedRuntimeInputGCP(""),
			description: "Should create proper runtime config struct with GCP input (empty zone)",
		},
		{
			input:       gardenerGCPGQLInput,
			expected:    expectedGardenerGCPRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for GCP provider",
		},
		{
			input:       gardenerAzureGQLInput,
			expected:    expectedGardenerAzureRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for Azure provider",
		},
		{
			input:       gardenerAWSGQLInput,
			expected:    expectedGardenerAWSRuntimeConfig,
			description: "Should create proper runtime config struct with Gardener input for AWS provider",
		},
	}

	accountProvider := hyperscaler.NewAccountProvider(nil, nil, "default-tenant")

	for _, testCase := range configurations {
		t.Run(testCase.description, func(t *testing.T) {
			//given
			uuidGeneratorMock := &mocks.UUIDGenerator{}
			uuidGeneratorMock.On("New").Return("id").Times(5)
			uuidGeneratorMock.On("New").Return("very-Long-ID-That-Has-More-Than-Fourteen-Characters-And-Even-Some-Hypens")

			inputConverter := NewInputConverter(uuidGeneratorMock, readSession, accountProvider)

			//when
			runtimeConfig, err := inputConverter.ProvisioningInputToCluster("runtimeID", testCase.input)

			//then
			require.NoError(t, err)
			assert.Equal(t, testCase.expected, runtimeConfig)
			uuidGeneratorMock.AssertExpectations(t)
		})
	}
}

func TestConverter_ProvisioningInputToCluster_Error(t *testing.T) {

	t.Run("should return error when failed to get kyma release", func(t *testing.T) {
		// given
		uuidGeneratorMock := &mocks.UUIDGenerator{}
		readSession := &realeaseMocks.Repository{}
		readSession.On("GetReleaseByVersion", kymaVersion).Return(model.Release{}, dberrors.NotFound("error"))

		input := gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GcpConfig: &gqlschema.GCPConfigInput{},
			},
			Credentials: &gqlschema.CredentialsInput{
				SecretName: "secretName",
			},
			KymaConfig: &gqlschema.KymaConfigInput{
				Version: kymaVersion,
			},
		}
		accountProvider := &hyperscalerMocks.AccountProvider{}

		inputConverter := NewInputConverter(uuidGeneratorMock, readSession, accountProvider)

		//when
		_, err := inputConverter.ProvisioningInputToCluster("runtimeID", input)

		//then
		require.Error(t, err)
		uuidGeneratorMock.AssertExpectations(t)
	})

	t.Run("should return error when no provider input provided", func(t *testing.T) {
		// given
		input := gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{},
		}

		inputConverter := NewInputConverter(nil, nil, nil)

		//when
		_, err := inputConverter.ProvisioningInputToCluster("runtimeID", input)

		//then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not match any provider")
	})

	t.Run("should return error when no Gardener provider specified", func(t *testing.T) {
		// given
		uuidGeneratorMock := &mocks.UUIDGenerator{}
		uuidGeneratorMock.On("New").Return("id").Times(4)

		input := gqlschema.ProvisionRuntimeInput{
			ClusterConfig: &gqlschema.ClusterConfigInput{
				GardenerConfig: &gqlschema.GardenerConfigInput{},
			},
		}

		inputConverter := NewInputConverter(uuidGeneratorMock, nil, &hyperscalerMocks.AccountProvider{})

		//when
		_, err := inputConverter.ProvisioningInputToCluster("runtimeID", input)

		//then
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ProviderSpecificInput not specified (was nil)")
	})

}

func TestConverter_CreateGardenerClusterName(t *testing.T) {

	providerInputs := []struct {
		provider     string
		expectedName string
		description  string
	}{
		{
			provider:     "gcp",
			expectedName: "gcp-id",
			description:  "regular GCP provider name",
		},
		{
			provider:     "aws",
			expectedName: "aws-id",
			description:  "regular AWS provider name",
		},
		{
			provider:     "azure",
			expectedName: "azu-id",
			description:  "regular Azure provider name",
		},
		{
			provider:     "GCP",
			expectedName: "gcp-id",
			description:  "capitalized GCP provider name",
		},
		{
			provider:     "AWS",
			expectedName: "aws-id",
			description:  "capitalized AWS provider name",
		},
		{
			provider:     "AZURE",
			expectedName: "azu-id",
			description:  "capitalized Azure provider name",
		},
		{
			provider:     "-",
			expectedName: "c--id",
			description:  "wrong provider name that contains only hyphen: \"-\"",
		},
		{
			provider:     "!#$@^%&*gcp",
			expectedName: "gcp-id",
			description:  "wrong provider name with non-alphanumeric characters",
		},
		{
			provider:     "912740131aws---",
			expectedName: "aws-id",
			description:  "wrong provider name with numbers",
		},
	}

	for _, testCase := range providerInputs {
		t.Run(testCase.description, func(t *testing.T) {
			uuidGeneratorMock := &mocks.UUIDGenerator{}
			uuidGeneratorMock.On("New").Return("id")

			inputConverter := NewInputConverterTester(uuidGeneratorMock, nil)
			generatedName := inputConverter.createGardenerClusterName(testCase.provider)

			assert.Equal(t, testCase.expectedName, generatedName)
		})
	}
}

func fixKymaGraphQLConfigInput() *gqlschema.KymaConfigInput {
	return &gqlschema.KymaConfigInput{
		Version: kymaVersion,
		Components: []*gqlschema.ComponentConfigurationInput{
			{
				Component: clusterEssentialsComponent,
				Namespace: kymaSystemNamespace,
			},
			{
				Component: coreComponent,
				Namespace: kymaSystemNamespace,
				Configuration: []*gqlschema.ConfigEntryInput{
					fixGQLConfigEntryInput("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntryInput("test.config.key2", "value2", util.BoolPtr(false)),
				},
			},
			{
				Component: applicationConnectorComponent,
				Namespace: kymaIntegrationNamespace,
				Configuration: []*gqlschema.ConfigEntryInput{
					fixGQLConfigEntryInput("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntryInput("test.secret.key", "secretValue", util.BoolPtr(true)),
				},
			},
		},
		Configuration: []*gqlschema.ConfigEntryInput{
			fixGQLConfigEntryInput("global.config.key", "globalValue", util.BoolPtr(false)),
			fixGQLConfigEntryInput("global.config.key2", "globalValue2", util.BoolPtr(false)),
			fixGQLConfigEntryInput("global.secret.key", "globalSecretValue", util.BoolPtr(true)),
		},
	}
}

func fixGQLConfigEntryInput(key, val string, secret *bool) *gqlschema.ConfigEntryInput {
	return &gqlschema.ConfigEntryInput{
		Key:    key,
		Value:  val,
		Secret: secret,
	}
}
