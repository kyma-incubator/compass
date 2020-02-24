package storage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"

	"github.com/gocraft/dbr"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	DbUser            = "admin"
	DbPass            = "nimda"
	DbName            = "broker"
	DbPort            = "5432"
	DockerUserNetwork = "test_network"
	EnvPipelineBuild  = "PIPELINE_BUILD"
)

func makeConnectionString(hostname string, port string) string {
	host := "localhost"
	if os.Getenv(EnvPipelineBuild) != "" {
		host = hostname
		port = DbPort
	}

	cfg := Config{
		Host:     host,
		User:     DbUser,
		Password: DbPass,
		Port:     port,
		Name:     DbName,
		SSLMode:  "disable",
	}
	return cfg.ConnectionURL()
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

func InitTestDBTables(t *testing.T, connectionURL string) error {
	connection, err := postsql.WaitForDatabaseAccess(connectionURL, 10)
	if err != nil {
		t.Logf("Cannot connect to database with URL %s", connectionURL)
		return err
	}

	for name, v := range fixTables() {
		if _, err := connection.Exec(v); err != nil {
			t.Logf("Cannot create table %s", name)
			return err
		}
		t.Logf("Table %s added to database", name)
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
	exec.Command("systemctl start docker.service")

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

func fixTables() map[string]string {
	return map[string]string{
		postsql.InstancesTableName: fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
			instance_id varchar(255) PRIMARY KEY,
			runtime_id varchar(255) NOT NULL,
			global_account_id varchar(255) NOT NULL,
			service_id varchar(255) NOT NULL,
			service_plan_id varchar(255) NOT NULL,
			dashboard_url varchar(255) NOT NULL,
			provisioning_parameters text NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			delated_at TIMESTAMPTZ NOT NULL DEFAULT '0001-01-01 00:00:00+00'
		)`, postsql.InstancesTableName),
		postsql.OperationTableName: fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS %s (
			id varchar(255) PRIMARY KEY,
			instance_id varchar(255) NOT NULL,
			target_operation_id varchar(255) NOT NULL,
			version integer NOT NULL,
			state varchar(32) NOT NULL,
			description text NOT NULL,
			type varchar(32) NOT NULL,
			data json NOT NULL,
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
			)`, postsql.OperationTableName)}

}
