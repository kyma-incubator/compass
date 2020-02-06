package database

import (
	"context"
	"testing"

	testutils "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaInitializer(t *testing.T) {

	ctx := context.Background()

	cleanupNetwork, err := testutils.EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	t.Run("Should initialize database when schema not applied", func(t *testing.T) {
		// given
		containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "test_DB_1")
		require.NoError(t, err)

		defer containerCleanupFunc()

		// when
		connection, err := InitializeDatabaseConnection(connString, 4)
		require.NoError(t, err)
		require.NotNil(t, connection)

		defer testutils.CloseDatabase(t, connection)

		err = SetupSchema(connection, testutils.SchemaFilePath)
		require.NoError(t, err)

		// then
		err = testutils.CheckIfAllDatabaseTablesArePresent(connection)

		assert.NoError(t, err)
	})

	t.Run("Should skip database initialization when schema already applied", func(t *testing.T) {
		//given
		containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "test_DB_2")
		require.NoError(t, err)

		defer containerCleanupFunc()

		// when
		connection, err := InitializeDatabaseConnection(connString, 4)
		require.NoError(t, err)
		require.NotNil(t, connection)

		defer testutils.CloseDatabase(t, connection)

		err = SetupSchema(connection, testutils.SchemaFilePath)
		require.NoError(t, err)

		err = SetupSchema(connection, testutils.SchemaFilePath)
		require.NoError(t, err)

		// then
		err = testutils.CheckIfAllDatabaseTablesArePresent(connection)

		assert.NoError(t, err)
	})

	t.Run("Should return error when failed to connect to the database", func(t *testing.T) {

		containerCleanupFunc, _, err := testutils.InitTestDBContainer(t, ctx, "test_DB_3")
		require.NoError(t, err)

		defer containerCleanupFunc()

		// given
		connString := "bad connection string"

		// when
		connection, err := InitializeDatabaseConnection(connString, 4)

		// then
		assert.Error(t, err)
		assert.Nil(t, connection)
	})

	t.Run("Should return error when failed to read database schema", func(t *testing.T) {
		//given
		containerCleanupFunc, connString, err := testutils.InitTestDBContainer(t, ctx, "test_DB_4")
		require.NoError(t, err)

		defer containerCleanupFunc()

		schemaFilePath := "../../../assets/database/notfound.sql"

		//when
		connection, err := InitializeDatabaseConnection(connString, 4)
		require.NoError(t, err)

		err = SetupSchema(connection, schemaFilePath)

		// then
		assert.Error(t, err)
	})
}
