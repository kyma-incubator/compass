package apitests

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/kyma-incubator/compass/tests/provisioner-tests/test/testkit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	timeout  = 60 * time.Second
	interval = 1 * time.Second
)

var provisionRuntimeInput = gqlschema.ProvisionRuntimeInput{
	ClusterConfig: &gqlschema.ClusterConfigInput{
		Name:        "Test",
		ComputeZone: "Zone",
		Credentials: &gqlschema.CredentialsInput{SecretName: "secret"},
		ProviderConfig: &gqlschema.ProviderConfigInput{
			GardenerProviderConfig: &gqlschema.GardenerProviderConfigInput{
				TargetProvider: "gcp",
				TargetSecret:   "gardener-secret",
			},
		},
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
	waitUntilOperationIsFinished(t, client, provisionOperationID)
	t.Logf("Runtime %s provisioned succesfully", runtimeID)

	t.Logf("Reconnecting runtime agent with runtime %s", runtimeID)
	reconnectOperationID, e := client.ReconnectRuntimeAgent(runtimeID)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is provisioned", runtimeID)
	waitUntilOperationIsFinished(t, client, reconnectOperationID)
	t.Logf("Runtime agent for runtime %s reconnected succesfully", runtimeID)

	t.Logf("Upgrading runtime %s", runtimeID)

	upgradeOperationID, e := client.UpgradeRuntime(runtimeID, upgradeRuntimeInput)

	require.NoError(t, e)

	t.Logf("Waiting until runtime %s is upgraded", runtimeID)
	waitUntilOperationIsFinished(t, client, upgradeOperationID)
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
	waitUntilOperationIsFinished(t, client, deprovisionOperationID)
	t.Logf("Runtime %s deprovisioned succesfully", runtimeID)
}

func waitUntilOperationIsFinished(t *testing.T, client testkit.ProvisionerClient, operationID string) {
	err := waitForFunction(interval, timeout, func() bool {
		operationStatus, e := client.RuntimeOperationStatus(operationID)
		if e != nil {
			return false
		}

		if operationStatus.State == gqlschema.OperationStateSucceeded {
			return true
		}

		if operationStatus.State == gqlschema.OperationStateFailed {
			t.FailNow()
		}
		return false
	})
	require.NoError(t, err)
}

func waitForFunction(interval, timeout time.Duration, isDone func() bool) error {
	done := time.After(timeout)

	for {
		if isDone() {
			return nil
		}

		select {
		case <-done:
			return errors.New("timeout waiting for condition")
		default:
			time.Sleep(interval)
		}
	}
}
