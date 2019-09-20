package apitests

import (
	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	maxWaitTime = 60
	interval    = 1
)

var provisionRuntimeInput = gqlschema.ProvisionRuntimeInput{
	ClusterConfig: &gqlschema.ClusterConfigInput{
		Name:                   "Test",
		ComputeZone:            "Zone",
		Credentials:            "Credentials",
		InfrastructureProvider: gqlschema.InfrastructureProviderAks,
	},
	KymaConfig: &gqlschema.KymaConfigInput{
		Version: "1.6",
		Modules: []gqlschema.KymaModule{},
	},
}

var upgradeRuntimeInput = gqlschema.UpgradeRuntimeInput{
	ClusterConfig: &gqlschema.UpgradeClusterInput{
		Version: "2.0",
	},
	KymaConfig: &gqlschema.KymaConfigInput{
		Version: "1.7",
		Modules: []gqlschema.KymaModule{},
	},
}

func TestFullProvisionerFlow(t *testing.T) {
	config, e := testkit.ReadConfig()
	require.NoError(t, e)

	client := testkit.NewProvisionerClient(config.InternalProvisionerUrl)

	runtimeID := uuid.New().String()

	t.Logf("Provisioning runtime %s", runtimeID)
	provisionOperationID, e := client.ProvisionRuntime(runtimeID, provisionRuntimeInput)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is provisioned", runtimeID)
	waitUntilOperationIsFinished(e, client, provisionOperationID, t, runtimeID)
	t.Logf("Runtime %s provisioned succesfully", runtimeID)

	t.Logf("Reconnecting runtime agent with runtime %s", runtimeID)
	reconnectOperationID, e := client.ReconnectRuntimeAgent(runtimeID)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is provisioned", runtimeID)
	waitUntilOperationIsFinished(e, client, reconnectOperationID, t, runtimeID)
	t.Logf("Runtime agent for runtime %s reconnected succesfully", runtimeID)

	t.Logf("Upgrading runtime %s", runtimeID)

	upgradeOperationID, e := client.UpgradeRuntime(runtimeID, upgradeRuntimeInput)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is upgraded", runtimeID)
	waitUntilOperationIsFinished(e, client, upgradeOperationID, t, runtimeID)
	t.Logf("Runtime %s upgraded succesfully", runtimeID)

	t.Logf("Checking current status of runtime %s", runtimeID)
	status, e := client.RuntimeStatus(runtimeID)

	require.NoError(t, e)

	assert.Equal(t, gqlschema.OperationTypeUpgrade, status.LastOperationStatus.Operation)
	assert.Equal(t, gqlschema.OperationStateSucceeded, status.LastOperationStatus.State)

	t.Logf("Deprovisioning runtime %s", runtimeID)
	deprovisionOperationID, e := client.DeprovisionRuntime(runtimeID)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is deprovisioned", runtimeID)
	waitUntilOperationIsFinished(e, client, deprovisionOperationID, t, runtimeID)
	t.Logf("Runtime %s deprovisioned succesfully", runtimeID)
}

func waitUntilOperationIsFinished(e error, client testkit.ProvisionerClient, operationID string, t *testing.T, runtimeID string) {
	var operationStatus gqlschema.OperationStatus
	for i := 0; i <= maxWaitTime; i += interval {
		operationStatus, e = client.RuntimeOperationStatus(operationID)
		require.NoError(t, e)

		if operationStatus.State == gqlschema.OperationStateSucceeded {
			break
		}

		if operationStatus.State == gqlschema.OperationStateFailed {
			t.Logf("Failed to provision runtime %s", runtimeID)
			t.FailNow()
		}
		time.Sleep(interval * time.Second)
	}
	require.Equal(t, gqlschema.OperationStateSucceeded, operationStatus.State)
}
