package api

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/gocraft/dbr"
	"github.com/kyma-incubator/compass/components/provisioner/internal/hydroform"
	hydroformmocks "github.com/kyma-incubator/compass/components/provisioner/internal/hydroform/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/database"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dbsession"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/hydroform/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	dbUser = "admin"
	dbPass = "nimda"
	dbName = "provisioner"
	dbPort = "5432"
)

func makeConnectionString(host string, port string) string {

	if os.Getenv("PIPELINE_BUILD") == "" {
		host = "localhost"
	} else {
		port = dbPort
	}

	const connStringFormat string = "host=%s port=%s user=%s password=%s dbname=%s sslmode=%s"

	return fmt.Sprintf(connStringFormat, host, port, dbUser,
		dbPass, dbName, "disable")
}

func closeDatabase(t *testing.T, connection *dbr.Connection) {
	if connection != nil {
		err := connection.Close()
		assert.Nil(t, err, "Failed to close db connection")
	}
}

func initDBContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string, string, error) {

	netReq := testcontainers.NetworkRequest{
		Name:   "test_network",
		Driver: "bridge",
	}

	provider, err := testcontainers.NewDockerProvider()

	if err != nil {
		return nil, "", "", err
	}

	_, err = provider.GetNetwork(ctx, netReq)

	if err != nil {
		return nil, "", "", err
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:11",
		ExposedPorts: []string{fmt.Sprintf("%s", dbPort)},
		Networks:     []string{"test_network"},
		NetworkAliases: map[string][]string{
			"test_network": {"postgres_database"},
		},
		Env: map[string]string{
			"POSTGRES_USER":     "admin",
			"POSTGRES_PASSWORD": "nimda",
			"POSTGRES_DB":       "provisioner",
		},
		WaitingFor: wait.ForListeningPort(nat.Port(dbPort)),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Logf("Failed to create contianer: %s", err.Error())
		return nil, "", "", err
	}

	port, err := postgresContainer.MappedPort(ctx, dbPort)
	if err != nil {
		t.Logf("Failed to get mapped port for container %s : %s", postgresContainer.GetContainerID(), err.Error())
		errTerminate := postgresContainer.Terminate(ctx)
		if errTerminate != nil {
			t.Logf("Failed to terminate container %s after failing of getting mapped port: %s", postgresContainer.GetContainerID(), err.Error())
		}
		return nil, "", "", err
	}

	return postgresContainer, port.Port(), "postgres_database", nil
}

func waitForOperationCompleted(provisioningService provisioning.Service, operationID string) error {

	retryCount := 3
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
			BootDiskSize:      "256",
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
			VolumeSize:        "1TB",
			MachineType:       "n1-standard-1",
			Region:            "region",
			TargetProvider:    "GCP",
			TargetSecret:      "secret",
			DiskType:          "ssd",
			Zone:              "zone",
			Cidr:              "cidr",
			AutoScalerMin:     1,
			AutoScalerMax:     5,
			MaxSurge:          1,
			MaxUnavailable:    2,
		},
	}

	testConfig := []provisionerTestConfig{
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bbd", config: clusterConfigForGCP, description: "Should provision and deprovision a runtime with happy flow using correct GCP configuration"},
		{runtimeID: "1100bb59-9c40-4ebb-b846-7477c4dc5bbe", config: clusterConfigForGardenerWithGCP, description: "Should provision and deprovision a runtime with happy flow using correct Gardener with GCP configuration"},
	}
	return testConfig
}

func TestResolver_ProvisionRuntimeWithDatabase(t *testing.T) {

	hydroformServiceMock := &hydroformmocks.Service{}
	hydroformServiceMock.On("ProvisionCluster", mock.Anything, mock.Anything).Return(hydroform.ClusterInfo{ClusterStatus: types.Provisioned, KubeConfig: "", State: "{}"}, nil)
	hydroformServiceMock.On("DeprovisionCluster", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	uuidGenerator := persistence.NewUUIDGenerator()

	ctx := context.Background()
	postgresContainer, mappedPort, hostName, err := initDBContainer(t, ctx)
	require.NoError(t, err)
	require.NotNil(t, postgresContainer)

	defer postgresContainer.Terminate(ctx)

	connString := makeConnectionString(hostName, mappedPort)
	schemaFilePath := "../../assets/database/provisioner.sql"

	connection, err := database.InitializeDatabase(connString, schemaFilePath, 4)

	require.NoError(t, err)
	require.NotNil(t, connection)

	defer closeDatabase(t, connection)

	kymaConfig := &gqlschema.KymaConfigInput{
		Version: "1.5",
		Modules: gqlschema.AllKymaModule,
	}

	kymaCredentials := &gqlschema.CredentialsInput{SecretName: "secret_1"}

	clusterConfigurations := getTestClusterConfigurations()

	for _, cfg := range clusterConfigurations {
		t.Run(cfg.description, func(t *testing.T) {

			fullConfig := gqlschema.ProvisionRuntimeInput{ClusterConfig: cfg.config, Credentials: kymaCredentials, KymaConfig: kymaConfig}

			dbSessionFactory := dbsession.NewFactory(connection)
			persistenceService := persistence.NewService(dbSessionFactory, uuidGenerator)
			provisioningService := provisioning.NewProvisioningService(persistenceService, uuidGenerator, hydroformServiceMock)
			provisioner := NewResolver(provisioningService)

			operationID, err := provisioner.ProvisionRuntime(ctx, cfg.runtimeID, fullConfig)
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

			err = waitForOperationCompleted(provisioningService, operationID)
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

			err = waitForOperationCompleted(provisioningService, operationID)
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
