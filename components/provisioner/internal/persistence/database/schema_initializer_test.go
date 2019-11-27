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
		connection, err := InitializeDatabase(connString, testutils.SchemaFilePath, 4)

		require.NoError(t, err)
		require.NotNil(t, connection)

		defer testutils.CloseDatabase(t, connection)

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
		connection, err := InitializeDatabase(connString, testutils.SchemaFilePath, 4)

		require.NoError(t, err)
		require.NotNil(t, connection)

		defer testutils.CloseDatabase(t, connection)

		badSchemaFilePath := "../../../assets/database/notfound.sql"

		connection2, secondAttemptInitError := InitializeDatabase(connString, badSchemaFilePath, 4)

		// then
		require.NoError(t, secondAttemptInitError)
		require.NotNil(t, connection2)

		defer testutils.CloseDatabase(t, connection2)

		err = testutils.CheckIfAllDatabaseTablesArePresent(connection2)

		assert.NoError(t, err)
	})

	t.Run("Should return error when failed to connect to the database", func(t *testing.T) {

		containerCleanupFunc, _, err := testutils.InitTestDBContainer(t, ctx, "test_DB_3")
		require.NoError(t, err)

		defer containerCleanupFunc()

		// given
		connString := "bad connection string"

		// when
		connection, err := InitializeDatabase(connString, testutils.SchemaFilePath, 4)

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
		connection, err := InitializeDatabase(connString, schemaFilePath, 4)

		// then
		assert.Error(t, err)
		assert.Nil(t, connection)
	})
}
