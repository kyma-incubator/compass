package test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstance(t *testing.T) {
	ctx := context.Background()

	cleanupNetwork, err := storage.EnsureTestNetworkForDB(t, ctx)
	require.NoError(t, err)
	defer cleanupNetwork()

	t.Run("Should create and update instance", func(t *testing.T) {
		// given
		containerCleanupFunc, connString, err := storage.InitTestDBContainer(t, ctx, "test_DB_2")
		require.NoError(t, err)
		defer containerCleanupFunc()

		// when
		brokerStorage, err := storage.New(connString)

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
