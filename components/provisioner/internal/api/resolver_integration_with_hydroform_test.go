package api

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/provisioner/internal/api/middlewares"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime"
	mocks2 "github.com/kyma-incubator/compass/components/provisioner/internal/runtime/clientbuilder/mocks"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	hydroformmocks "github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	"github.com/kyma-incubator/hydroform/types"

	"github.com/kyma-incubator/compass/components/provisioner/internal/util"

	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"

	installationMocks "github.com/kyma-incubator/compass/components/provisioner/internal/installation/mocks"

	"github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"
	"github.com/kyma-incubator/compass/components/provisioner/internal/uuid"

	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/testutils"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	kymaVersion                   = "1.8"
	kymaSystemNamespace           = "kyma-system"
	kymaIntegrationNamespace      = "kyma-integration"
	clusterEssentialsComponent    = "cluster-essentials"
	coreComponent                 = "core"
	applicationConnectorComponent = "application-connector"

	gardenerProject        = "gardener-project"
	runtimeAgentComponent  = "compass-runtime-agent"
	compassSystemNamespace = "compass-system"
	tenant                 = "tenant"
)

func waitForOperationCompleted(provisioningService provisioning.Service, operationID string, seconds uint) error {

	retryCount := seconds
	for ; retryCount > 0; retryCount-- {

		status, err := provisioningService.RuntimeOperationStatus(operationID)

		if err != nil {
			return err
		}

		if status.State == gqlschema.OperationStateSucceeded {
			return nil
		}

		time.Sleep(1 * time.Second)
	}
	return errors.New("timeout checking for operation state")
}

type provisionerTestConfig struct {
	runtimeID    string
	config       *gqlschema.ClusterConfigInput
	description  string
	runtimeInput *gqlschema.RuntimeInput
}

func getTestClusterConfigurations() []provisionerTestConfig {

	clusterConfigForGardenerWithGCP := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      1024,
			MachineType:       "n1-standard-1",
			Region:            "region",
			Provider:          "GCP",
			Seed:              util.StringPtr("gcp-eu1"),
			TargetSecret:      "secret",
			DiskType:          "ssd",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zone: "zone",
				},
			},
		},
	}

	clusterConfigForGardenerWithAzure := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      1024,
			MachineType:       "n1-standard-1",
			Region:            "region",
			Provider:          "Azure",
			Seed:              util.StringPtr("gcp-eu1"),
			TargetSecret:      "secret",
			DiskType:          "ssd",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AzureConfig: &gqlschema.AzureProviderConfigInput{
					VnetCidr: "cidr",
				},
			},
		},
	}

	clusterConfigForGardenerWithAWS := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      1024,
			MachineType:       "n1-standard-1",
			Region:            "region",
			Provider:          "AWS",
			Seed:              nil,
			TargetSecret:      "secret",
			DiskType:          "ssd",
			WorkerCidr:        "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
			ProviderSpecificConfig: &gqlschema.ProviderSpecificInput{
				AwsConfig: &gqlschema.AWSProviderConfigInput{
					Zone:         "zone",
					InternalCidr: "cidr",
					VpcCidr:      "cidr",
					PublicCidr:   "cidr",
				},
			},
		},
	}

	testConfig := []provisionerTestConfig{
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bbb", config: clusterConfigForGardenerWithGCP, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with GCP configuration 1",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime1",
				Description: new(string),
			}},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb4", config: clusterConfigForGardenerWithAzure, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with Azure configuration",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime2",
				Description: new(string),
			}},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb5", config: clusterConfigForGardenerWithAWS, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with AWS configuration",
			runtimeInput: &gqlschema.RuntimeInput{
				Name:        "test runtime3",
				Description: new(string),
			}},
	}
	return testConfig
}

var providerCredentials = &gqlschema.CredentialsInput{SecretName: "secret_1"}

func TestResolver_ProvisionRuntimeWithDatabaseAndHydroform(t *testing.T) {

	mockedKubeConfigValue := "test config value"
	mockedTerraformState := []byte(`{"test_key": "test_value"}`)

	hydroformServiceMock := &hydroformmocks.Service{}
	hydroformServiceMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: mockedKubeConfigValue, State: mockedTerraformState}, nil).
		Run(func(args mock.Arguments) {
			time.Sleep(1 * time.Second)
		})
	hydroformServiceMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			time.Sleep(1 * time.Second)
		})

	installationServiceMock := &installationMocks.Service{}
	installationServiceMock.On(
		"InstallKyma",
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string"),
		mock.AnythingOfType("model.Release"),
		mock.AnythingOfType("model.Configuration"),
		mock.AnythingOfType("[]model.KymaComponentConfig")).
		Return(nil)

	uuidGenerator := uuid.NewUUIDGenerator()

	ctx := context.WithValue(context.Background(), middlewares.Tenant, tenant)

	cleanupNetwork, err := testutils.EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "postgres_database")
	require.NoError(t, err)

	defer containerCleanupFunc()

	connection, err := database.InitializeDatabaseConnection(connString, 5)
	require.NoError(t, err)
	require.NotNil(t, connection)

	defer testutils.CloseDatabase(t, connection)

	err = database.SetupSchema(connection, testutils.SchemaFilePath)
	require.NoError(t, err)

	kymaConfig := fixKymaGraphQLConfigInput()
	clusterConfigurations := getTestClusterConfigurations()

	for _, cfg := range clusterConfigurations {
		t.Run(cfg.description, func(t *testing.T) {

			directorServiceMock := &directormock.DirectorClient{}
			directorServiceMock.On("CreateRuntime", mock.Anything, mock.Anything).Return(cfg.runtimeID, nil)
			directorServiceMock.On("DeleteRuntime", mock.Anything, mock.Anything).Return(nil)
			directorServiceMock.On("GetConnectionToken", mock.Anything, mock.Anything).Return(graphql.OneTimeTokenForRuntimeExt{}, nil)

			cmClientBuilder := &mocks2.ConfigMapClientBuilder{}
			configMapClient := fake.NewSimpleClientset().CoreV1().ConfigMaps(compassSystemNamespace)
			cmClientBuilder.On("CreateK8SConfigMapClient", mockedKubeConfigValue, compassSystemNamespace).Return(configMapClient, nil)
			runtimeConfigurator := runtime.NewRuntimeConfigurator(cmClientBuilder, directorServiceMock)

			fullConfig := gqlschema.ProvisionRuntimeInput{RuntimeInput: cfg.runtimeInput, ClusterConfig: cfg.config, Credentials: providerCredentials, KymaConfig: kymaConfig}

			dbSessionFactory := dbsession.NewFactory(connection)
			releaseRepository := release.NewReleaseRepository(connection, uuidGenerator)
			inputConverter := provisioning.NewInputConverter(uuidGenerator, releaseRepository, gardenerProject)
			graphQLConverter := provisioning.NewGraphQLConverter()
			hydroformProvisioner := hydroform.NewHydroformProvisioner(hydroformServiceMock, installationServiceMock, dbSessionFactory, directorServiceMock, runtimeConfigurator)
			provisioningService := provisioning.NewProvisioningService(inputConverter, graphQLConverter, directorServiceMock, dbSessionFactory, hydroformProvisioner, uuidGenerator)
			validator := NewValidator(dbSessionFactory.NewReadSession())
			provisioner := NewResolver(provisioningService, validator)

			err := insertDummyReleaseIfNotExist(releaseRepository, uuidGenerator.New(), kymaVersion)
			require.NoError(t, err)

			status, err := provisioner.ProvisionRuntime(ctx, fullConfig)
			require.NoError(t, err)

			require.NotNil(t, status)
			require.NotNil(t, status.RuntimeID)
			assert.Equal(t, cfg.runtimeID, *status.RuntimeID)

			messageProvisioningStarted := "Provisioning started"

			statusForProvisioningStarted := &gqlschema.OperationStatus{
				ID:        status.ID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateInProgress,
				RuntimeID: status.RuntimeID,
				Message:   &messageProvisioningStarted,
			}

			runtimeStatusProvisioningStarted, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, runtimeStatusProvisioningStarted)
			assert.Equal(t, statusForProvisioningStarted, runtimeStatusProvisioningStarted.LastOperationStatus)
			assert.Equal(t, fixKymaGraphQLConfig(), runtimeStatusProvisioningStarted.RuntimeConfiguration.KymaConfig)

			err = waitForOperationCompleted(provisioningService, *status.ID, 3)
			require.NoError(t, err)

			messageProvisioningSucceeded := "Operation succeeded."

			statusForProvisioningSucceeded := &gqlschema.OperationStatus{
				ID:        status.ID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateSucceeded,
				RuntimeID: status.RuntimeID,
				Message:   &messageProvisioningSucceeded,
			}

			runtimeStatusProvisioned, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, runtimeStatusProvisioned)
			assert.Equal(t, statusForProvisioningSucceeded, runtimeStatusProvisioned.LastOperationStatus)
			assert.Equal(t, fixKymaGraphQLConfig(), runtimeStatusProvisioningStarted.RuntimeConfiguration.KymaConfig)

			session := dbSessionFactory.NewReadSession()
			clusterData, err := session.GetCluster(cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, clusterData)
			require.NotNil(t, clusterData.Kubeconfig)
			assert.Equal(t, mockedTerraformState, clusterData.TerraformState)
			assert.Equal(t, mockedKubeConfigValue, *clusterData.Kubeconfig)

			deprovisionID, err := provisioner.DeprovisionRuntime(ctx, cfg.runtimeID)
			require.NoError(t, err)

			messageDeprovisioningStarted := "Deprovisioning started."

			statusForDeprovisioningStarted := &gqlschema.OperationStatus{
				ID:        &deprovisionID,
				Operation: gqlschema.OperationTypeDeprovision,
				State:     gqlschema.OperationStateInProgress,
				RuntimeID: &cfg.runtimeID,
				Message:   &messageDeprovisioningStarted,
			}

			runtimeStatusDeprovStarted, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, runtimeStatusDeprovStarted)
			assert.Equal(t, statusForDeprovisioningStarted, runtimeStatusDeprovStarted.LastOperationStatus)

			err = waitForOperationCompleted(provisioningService, deprovisionID, 3)
			require.NoError(t, err)

			messageDeprovSucceess := "Operation succeeded."

			runtimeStatusDeprovSuccess := &gqlschema.OperationStatus{
				ID:        &deprovisionID,
				Operation: gqlschema.OperationTypeDeprovision,
				State:     gqlschema.OperationStateSucceeded,
				RuntimeID: &cfg.runtimeID,
				Message:   &messageDeprovSucceess,
			}

			runtimeStatusDeprovisioned, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, runtimeStatusDeprovisioned)
			assert.Equal(t, runtimeStatusDeprovSuccess, runtimeStatusDeprovisioned.LastOperationStatus)
		})
	}
}

func insertDummyReleaseIfNotExist(releaseRepo release.Repository, id, version string) error {
	_, err := releaseRepo.GetReleaseByVersion(version)
	if err == nil {
		return nil
	}

	if err.Code() != dberrors.CodeNotFound {
		return err
	}
	_, err = releaseRepo.SaveRelease(model.Release{
		Id:            id,
		Version:       version,
		TillerYAML:    "tiller YAML",
		InstallerYAML: "installer YAML",
	})

	return err
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
			{
				Component: runtimeAgentComponent,
				Namespace: compassSystemNamespace,
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

func fixKymaGraphQLConfig() *gqlschema.KymaConfig {

	return &gqlschema.KymaConfig{
		Version: util.StringPtr(kymaVersion),
		Components: []*gqlschema.ComponentConfiguration{
			{
				Component:     clusterEssentialsComponent,
				Namespace:     kymaSystemNamespace,
				Configuration: make([]*gqlschema.ConfigEntry, 0, 0),
			},
			{
				Component: coreComponent,
				Namespace: kymaSystemNamespace,
				Configuration: []*gqlschema.ConfigEntry{
					fixGQLConfigEntry("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntry("test.config.key2", "value2", util.BoolPtr(false)),
				},
			},
			{
				Component: applicationConnectorComponent,
				Namespace: kymaIntegrationNamespace,
				Configuration: []*gqlschema.ConfigEntry{
					fixGQLConfigEntry("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntry("test.secret.key", "secretValue", util.BoolPtr(true)),
				},
			},
			{
				Component: runtimeAgentComponent,
				Namespace: compassSystemNamespace,
				Configuration: []*gqlschema.ConfigEntry{
					fixGQLConfigEntry("test.config.key", "value", util.BoolPtr(false)),
					fixGQLConfigEntry("test.secret.key", "secretValue", util.BoolPtr(true)),
				},
			},
		},
		Configuration: []*gqlschema.ConfigEntry{
			fixGQLConfigEntry("global.config.key", "globalValue", util.BoolPtr(false)),
			fixGQLConfigEntry("global.config.key2", "globalValue2", util.BoolPtr(false)),
			fixGQLConfigEntry("global.secret.key", "globalSecretValue", util.BoolPtr(true)),
		},
	}
}

func fixGQLConfigEntry(key, val string, secret *bool) *gqlschema.ConfigEntry {
	return &gqlschema.ConfigEntry{
		Key:    key,
		Value:  val,
		Secret: secret,
	}
}
