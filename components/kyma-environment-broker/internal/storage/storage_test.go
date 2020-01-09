// +build integration

package storage

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage/postsql"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSchemaInitializer(t *testing.T) {
	ctx := context.Background()

	cleanupNetwork, err := EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	t.Run("Init tests", func(t *testing.T) {
		t.Run("Should initialize database when schema not applied", func(t *testing.T) {
			// given
			containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			// when
			connection, err := postsql.InitializeDatabase(connString)

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
			connection, err := postsql.InitializeDatabase(connString)

			require.NoError(t, err)
			require.NotNil(t, connection)
			defer CloseDatabase(t, connection)

			connection2, secondAttemptInitError := postsql.InitializeDatabase(connString)

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
			connection, err := postsql.InitializeDatabase(connString)

			// then
			assert.Error(t, err)
			assert.Nil(t, connection)
		})
	})

	// Instances
	t.Run("Instances", func(t *testing.T) {
		t.Run("Should create and update instance", func(t *testing.T) {
			// given
			containerCleanupFunc, connString, err := InitTestDBContainer(t, ctx, "test_DB_1")
			require.NoError(t, err)
			defer containerCleanupFunc()

			// when
			brokerStorage, err := New(connString)

			require.NoError(t, err)
			require.NotNil(t, brokerStorage)

			testData := "test"
			fixInstance := fixInstance(testData)
			err = brokerStorage.Instances().Insert(*fixInstance)
			require.NoError(t, err)

			fixInstance.DashboardURL = "diff"
			err = brokerStorage.Instances().Update(*fixInstance)
			require.NoError(t, err)

			// then
			inst, err := brokerStorage.Instances().GetByID(testData)
			assert.NoError(t, err)
			require.NotNil(t, inst)

			assert.Equal(t, fixInstance, inst)
		})
	})

}

func fixInstance(testData string) *internal.Instance {
	return &internal.Instance{
		InstanceID:             testData,
		RuntimeID:              testData,
		GlobalAccountID:        testData,
		ServiceID:              testData,
		ServicePlanID:          testData,
		DashboardURL:           testData,
		ProvisioningParameters: testData,
	}
}
