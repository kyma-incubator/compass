package storage

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gocraft/dbr"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	DbUser            = "admin"
	DbPass            = "nimda"
	DbName            = "broker"
	DbPort            = "5432"
	SchemaName        = "public"
	DockerUserNetwork = "test_network"
	EnvPipelineBuild  = "PIPELINE_BUILD"
)

// partially copied from
// https://github.com/kyma-incubator/compass/blob/master/components/provisioner/internal/persistence/database/schema_initializer_test.go
func TestSchemaInitializer(t *testing.T) {
	ctx := context.Background()

	cleanupNetwork, err := EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	t.Run("Should initialize database when schema not applied", func(t *testing.T) {
		// given
		containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
		require.NoError(t, err)
		defer containerCleanupFunc()

		// when
		connection, err := InitializeDatabase(connString)

		require.NoError(t, err)
		require.NotNil(t, connection)

		defer CloseDatabase(t, connection)

		// then
		err = CheckIfAllDatabaseTablesArePresent(connection)

		assert.NoError(t, err)
	})

	t.Run("Should skip database initialization when schema already applied", func(t *testing.T) {
		//given
		containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_2")
		require.NoError(t, err)
		defer containerCleanupFunc()

		// when
		connection, err := InitializeDatabase(connString)

		require.NoError(t, err)
		require.NotNil(t, connection)
		defer CloseDatabase(t, connection)

		connection2, secondAttemptInitError := InitializeDatabase(connString)

		// then
		require.NoError(t, secondAttemptInitError)
		require.NotNil(t, connection2)
		defer CloseDatabase(t, connection2)

		err = CheckIfAllDatabaseTablesArePresent(connection2)
		assert.NoError(t, err)
	})

	t.Run("Should return error when failed to connect to the database", func(t *testing.T) {
		containerCleanupFunc, _, err := InitTestDBContainer(t, ctx, "test_DB_3")
		require.NoError(t, err)
		defer containerCleanupFunc()

		// given
		connString := "bad connection string"

		// when
		connection, err := InitializeDatabase(connString)

		// then
		assert.Error(t, err)
		assert.Nil(t, connection)
	})
}

func makeConnectionString(hostname string, port string) string {
	host := "localhost"
	if os.Getenv(EnvPipelineBuild) != "" {
		host = hostname
		port = DbPort
	}

	return fmt.Sprintf(connectionURLFormat, host, port, DbUser,
		DbPass, DbName, "disable")
}

func CloseDatabase(t *testing.T, connection *dbr.Connection) {
	if connection != nil {
		err := connection.Close()
		assert.Nil(t, err, "Failed to close db connection")
	}
}

func InitTestDBContainer(t *testing.T, ctx context.Context, hostname string) (func(), string, error) {
	_, err := isDockerTestNetworkPresent(ctx)
	if err != nil {
		return nil, "", err
	}

	req := testcontainers.ContainerRequest{
		Image:        "postgres:11",
		SkipReaper:   true,
		ExposedPorts: []string{fmt.Sprintf("%s", DbPort)},
		Networks:     []string{DockerUserNetwork},
		NetworkAliases: map[string][]string{
			DockerUserNetwork: {hostname},
		},
		Env: map[string]string{
			"POSTGRES_USER":     DbUser,
			"POSTGRES_PASSWORD": DbPass,
			"POSTGRES_DB":       DbName,
		},
		WaitingFor: wait.ForListeningPort(DbPort),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	if err != nil {
		t.Logf("Failed to create contianer: %s", err.Error())
		return nil, "", err
	}

	port, err := postgresContainer.MappedPort(ctx, DbPort)
	if err != nil {
		t.Logf("Failed to get mapped port for container %s : %s", postgresContainer.GetContainerID(), err.Error())
		errTerminate := postgresContainer.Terminate(ctx)
		if errTerminate != nil {
			t.Logf("Failed to terminate container %s after failing of getting mapped port: %s", postgresContainer.GetContainerID(), err.Error())
		}
		return nil, "", err
	}

	cleanupFunc := func() {
		err := postgresContainer.Terminate(ctx)
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)
	}

	connString := makeConnectionString(hostname, port.Port())

	return cleanupFunc, connString, nil
}

func CheckIfAllDatabaseTablesArePresent(db *dbr.Connection) error {
	tables := []string{TableInstances}

	for _, table := range tables {
		checkError := checkIfDBTableIsPresent(table, db)
		if checkError != nil {
			return checkError
		}
	}
	return nil
}

func checkIfDBTableIsPresent(tableNameToCheck string, db *dbr.Connection) error {
	checkQuery := fmt.Sprintf(`SELECT '%s.%s'::regclass;`, SchemaName, tableNameToCheck)
	row := db.QueryRow(checkQuery)

	var tableName string
	err := row.Scan(&tableName)

	if err != nil {
		psqlErr, converted := err.(*pq.Error)

		if converted && psqlErr.Code == TableNotExistsError {
			return errors.Wrap(err, "Table is missing in the database")
		}
		return errors.Wrap(err, "Failed to check if table is present in the database")
	}

	if tableName != InstancesTableName {
		return errors.Wrap(err, "Failed verify table name in the database")
	}

	return nil
}

func isDockerTestNetworkPresent(ctx context.Context) (bool, error) {
	netReq := testcontainers.NetworkRequest{
		Name:   DockerUserNetwork,
		Driver: "bridge",
	}
	provider, err := testcontainers.NewDockerProvider()

	if err != nil || provider == nil {
		return false, errors.Wrap(err, "Failed to use Docker provider to access network information")
	}

	_, err = provider.GetNetwork(ctx, netReq)

	if err == nil {
		return true, nil
	}

	return false, nil
}

func createTestNetworkForDB(ctx context.Context) (testcontainers.Network, error) {
	netReq := testcontainers.NetworkRequest{
		Name:   DockerUserNetwork,
		Driver: "bridge",
	}
	provider, err := testcontainers.NewDockerProvider()

	if err != nil || provider == nil {
		return nil, errors.Wrap(err, "Failed to use Docker provider to access network information")
	}

	createdNetwork, err := provider.CreateNetwork(ctx, netReq)

	if err != nil {
		return nil, errors.Wrap(err, "Failed to create docker user network")
	}

	return createdNetwork, nil
}

func EnsureTestNetworkForDB(t *testing.T, ctx context.Context) (func(), error) {
	networkPresent, err := isDockerTestNetworkPresent(ctx)
	if networkPresent && err == nil {
		return func() {}, nil
	}

	if os.Getenv(EnvPipelineBuild) != "" {
		return func() {}, errors.Errorf("Docker network %s does not exist", DockerUserNetwork)
	}

	createdNetwork, err := createTestNetworkForDB(ctx)

	if err != nil {
		return func() {}, err
	}

	cleanupFunc := func() {
		err = createdNetwork.Remove(ctx)
		assert.NoError(t, err)
		time.Sleep(2 * time.Second)
	}

	return cleanupFunc, nil
}
