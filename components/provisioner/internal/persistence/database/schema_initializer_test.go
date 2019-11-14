package database

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/gocraft/dbr"
	"github.com/lib/pq"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
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

const (
	tableCluster          = "cluster"
	tableGardenerConfig   = "gardener_config"
	tableGCPConfig        = "gcp_config"
	tableOperation        = "kyma_config"
	tableKymaConfigModule = "kyma_config_module"
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

func checkIfDBTableIsPresent(tableNameToCheck string, db *dbr.Connection) error {

	checkQuery := fmt.Sprintf(`SELECT '%s.%s'::regclass;`, schemaName, tableNameToCheck)

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

	if tableName != clusterTableName {
		return errors.Wrap(err, "Failed verify table name in the database")
	}
	return nil
}

func checkIfAllDatabaseTablesArePresent(db *dbr.Connection) error {

	tables := []string{tableCluster, tableGardenerConfig, tableGCPConfig, tableOperation, tableKymaConfigModule}

	for _, table := range tables {
		checkError := checkIfDBTableIsPresent(table, db)

		if checkError != nil {
			return checkError
		}
	}
	return nil
}

func ensureBridgeNetworkForDB(ctx context.Context) (testcontainers.Network, error) {

	netReq := testcontainers.NetworkRequest{
		Name:   "test_network",
		Driver: "bridge",
	}

	provider, err := testcontainers.NewDockerProvider()

	if err != nil || provider == nil {
		return nil, errors.Wrap(err, "Failed to use Docker provider to access network information")
	}

	_, err = provider.GetNetwork(ctx, netReq)

	if err != nil && os.Getenv("PIPELINE_BUILD") == "" {

		// make own network in local mode
		createdNetwork, err := provider.CreateNetwork(ctx, netReq)

		if err == nil {
			return createdNetwork, nil
		} else {
			return nil, err
		}
	} else {
		return nil, nil
	}
}

func tearDownBridgeNetworkForDB(t *testing.T, ctx context.Context, network testcontainers.Network) {

	if os.Getenv("PIPELINE_BUILD") == "" && network != nil {
		// tear down testing network in local mode
		err := network.Remove(ctx)

		if err != nil {
			assert.Nil(t, err, "Failed to delete Bridge Network for Database")
		}
	}
}

func initDBContainer(t *testing.T, ctx context.Context) (testcontainers.Container, string, string, error) {

	req := testcontainers.ContainerRequest{
		Image:        "postgres:11",
		SkipReaper:   true,
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

	// for local testing return localhost for CI return postgres_database
	return postgresContainer, port.Port(), "postgres_database", nil
}

func TestSchemaInitializer(t *testing.T) {

	ctx := context.Background()

	createdNetwork, err := ensureBridgeNetworkForDB(ctx)
	require.NoError(t, err)

	defer tearDownBridgeNetworkForDB(t, ctx, createdNetwork)

	t.Run("Should initialize database when schema not applied", func(t *testing.T) {
		// given
		ctx := context.Background()
		postgresContainer, mappedPort, hostName, err := initDBContainer(t, ctx)
		require.NoError(t, err)
		require.NotNil(t, postgresContainer)

		defer postgresContainer.Terminate(ctx)

		connString := makeConnectionString(hostName, mappedPort)
		schemaFilePath := "../../../assets/database/provisioner.sql"

		// when
		connection, err := InitializeDatabase(connString, schemaFilePath, 4)

		require.NoError(t, err)
		require.NotNil(t, connection)

		defer closeDatabase(t, connection)

		// then
		err = checkIfAllDatabaseTablesArePresent(connection)

		assert.NoError(t, err)
	})

	t.Run("Should skip database initialization when schema already applied", func(t *testing.T) {
		//given
		ctx := context.Background()
		postgresContainer, mappedPort, hostName, err := initDBContainer(t, ctx)
		require.NoError(t, err)
		require.NotNil(t, postgresContainer)

		defer postgresContainer.Terminate(ctx)

		connString := makeConnectionString(hostName, mappedPort)
		schemaFilePath := "../../../assets/database/provisioner.sql"

		// when
		connection, err := InitializeDatabase(connString, schemaFilePath, 4)

		require.NoError(t, err)
		require.NotNil(t, connection)

		defer closeDatabase(t, connection)

		badSchemaFilePath := "../../../assets/database/notfound.sql"

		connection2, secondAttemptInitError := InitializeDatabase(connString, badSchemaFilePath, 4)

		// then
		require.NoError(t, secondAttemptInitError)
		require.NotNil(t, connection2)

		defer closeDatabase(t, connection2)

		err = checkIfAllDatabaseTablesArePresent(connection2)

		assert.NoError(t, err)
	})

	t.Run("Should return error when failed to connect to the database", func(t *testing.T) {
		// given
		connString := "bad connection string"
		schemaFilePath := "../../../assets/database/provisioner.sql"

		// when
		connection, err := InitializeDatabase(connString, schemaFilePath, 4)

		// then
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("Should return error when failed to read database schema", func(t *testing.T) {
		//given
		ctx := context.Background()
		postgresContainer, mappedPort, hostName, err := initDBContainer(t, ctx)
		require.NoError(t, err)
		require.NotNil(t, postgresContainer)

		defer postgresContainer.Terminate(ctx)

		connString := makeConnectionString(hostName, mappedPort)
		schemaFilePath := "../../../assets/database/notfound.sql"

		//when
		connection, err := InitializeDatabase(connString, schemaFilePath, 4)

		// then
		assert.Error(t, err)
		assert.Nil(t, connection)
	})
}
