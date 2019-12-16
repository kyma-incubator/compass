package api

import (
	"context"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	"github.com/kyma-incubator/hydroform/types"

	configMock "github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/configuration/mocks"
	hydroformmocks "github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	directormock "github.com/kyma-incubator/compass/components/provisioner/internal/director/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/testutils"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	runtimeID   string
	config      *gqlschema.ClusterConfigInput
	description string
}

func getTestClusterConfigurations() []provisionerTestConfig {

	clusterConfigForGCP := &gqlschema.ClusterConfigInput{
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

	clusterConfigForGardenerWithGCP := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			Name:              "Something",
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
				GcpConfig: &gqlschema.GCPProviderConfigInput{
					Zone: "zone",
				},
			},
		},
	}

	clusterConfigForGardenerWithAzure := &gqlschema.ClusterConfigInput{
		GardenerConfig: &gqlschema.GardenerConfigInput{
			Name:              "Something",
			ProjectName:       "Project",
			KubernetesVersion: "version",
			NodeCount:         3,
			VolumeSizeGb:      1024,
			MachineType:       "n1-standard-1",
			Region:            "region",
			Provider:          "Azure",
			Seed:              "gcp-eu1",
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
			Name:              "Something",
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
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bba", config: clusterConfigForGCP, description: "Should provision and deprovision a runtime with happy flow using correct GCP configuration"},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bbb", config: clusterConfigForGardenerWithGCP, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with GCP configuration 1"},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb4", config: clusterConfigForGardenerWithAzure, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with Azure configuration"},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bb5", config: clusterConfigForGardenerWithAWS, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with AWS configuration"},
	}
	return testConfig
}

func TestResolver_ProvisionRuntimeWithDatabase(t *testing.T) {

	////////all mocks go here

	mockedKubeConfigValue := "test config value"
	mockedTerraformState := `{"test_key": "test_value"}`

	hydroformServiceMock := &hydroformmocks.Service{}
	factory := &configMock.BuilderFactory{}
	builder := &configMock.Builder{}
	hydroformServiceMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: mockedKubeConfigValue, State: mockedTerraformState}, nil).
		Run(func(args mock.Arguments) {
			time.Sleep(1 * time.Second)
		})

	hydroformServiceMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil).
		Run(func(args mock.Arguments) {
			time.Sleep(1 * time.Second)
		})

	factory.On("NewProvisioningBuilder", mock.Anything).Return(builder)
	factory.On("NewDeprovisioningBuilder", mock.Anything).Return(builder)

	uuidGenerator := persistence.NewUUIDGenerator()

	ctx := context.Background()

	cleanupNetwork, err := testutils.EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "postgres_database")
	require.NoError(t, err)

	defer containerCleanupFunc()

	connection, err := database.InitializeDatabase(connString, testutils.SchemaFilePath, 4)

	require.NoError(t, err)
	require.NotNil(t, connection)

	defer testutils.CloseDatabase(t, connection)

	kymaConfig := &gqlschema.KymaConfigInput{
		Version: "1.5",
		Modules: gqlschema.AllKymaModule,
	}

	providerCredentials := &gqlschema.CredentialsInput{SecretName: "secret_1"}

	runtimeInput := &gqlschema.RuntimeInput {
		Name : "test runtime",
		Description: new(string),
		Labels : &gqlschema.Labels{},
	}

	clusterConfigurations := getTestClusterConfigurations()

	for _, cfg := range clusterConfigurations {
		t.Run(cfg.description, func(t *testing.T) {

			directorServiceMock := &directormock.DirectorClient{}
			directorServiceMock.On("CreateRuntime", mock.Anything).Return(cfg.runtimeID, nil)
			directorServiceMock.On("DeleteRuntime", mock.Anything).Return(nil)

			fullConfig := gqlschema.ProvisionRuntimeInput{RuntimeInput : runtimeInput, ClusterConfig: cfg.config, Credentials: providerCredentials, KymaConfig: kymaConfig}

			dbSessionFactory := dbsession.NewFactory(connection)
			persistenceService := persistence.NewService(dbSessionFactory, uuidGenerator)
			provisioningService := provisioning.NewProvisioningService(persistenceService, uuidGenerator, hydroformServiceMock, directorServiceMock, factory)
			provisioner := NewResolver(provisioningService)

			operationID, err := provisioner.ProvisionRuntime(ctx, fullConfig)
			require.NoError(t, err)

			messageProvisioningStarted := "Provisioning started"

			statusForProvisioningStarted := &gqlschema.OperationStatus{
				ID:        &operationID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateInProgress,
				RuntimeID: &cfg.runtimeID,
				Message:   &messageProvisioningStarted,
			}

			runtimeStatusProvisioningStarted, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)

			require.NoError(t, err)
			require.NotNil(t, runtimeStatusProvisioningStarted)
			assert.Equal(t, statusForProvisioningStarted, runtimeStatusProvisioningStarted.LastOperationStatus)

			err = waitForOperationCompleted(provisioningService, operationID, 3)
			require.NoError(t, err)

			messageProvisioningSucceeded := "Operation succeeded."

			statusForProvisioningSucceeded := &gqlschema.OperationStatus{
				ID:        &operationID,
				Operation: gqlschema.OperationTypeProvision,
				State:     gqlschema.OperationStateSucceeded,
				RuntimeID: &cfg.runtimeID,
				Message:   &messageProvisioningSucceeded,
			}

			runtimeStatusProvisioned, err := provisioner.RuntimeStatus(ctx, cfg.runtimeID)
			require.NoError(t, err)
			require.NotNil(t, runtimeStatusProvisioned)
			assert.Equal(t, statusForProvisioningSucceeded, runtimeStatusProvisioned.LastOperationStatus)

			clusterData, err := persistenceService.GetClusterData(cfg.runtimeID)

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
