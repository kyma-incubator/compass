package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_E2E_Upgrade(t *testing.T) {

	ts := newTestSuite(t)
	if ts.IsDummyTest {
		return
	}
	if !ts.IsUpgradeTest {
		t.SkipNow()
	}
	if ts.IsCleanupPhase {
		ts.Cleanup()
		return
	}
	configMap := ts.testConfigMap()

	operationID, err := ts.brokerClient.ProvisionRuntime(ts.upgradeSuite.PreUpgradeKymaVersion)
	require.NoError(t, err)

	ts.log.Infof("Creating config map %s with test data", ts.ConfigName)
	err = ts.configMapClient.Create(configMap)
	require.NoError(t, err)

	err = ts.brokerClient.AwaitOperationSucceeded(operationID, ts.ProvisionTimeout)
	require.NoError(t, err)

	ts.log.Info("Fetching runtime's kubeconfig")
	config, err := ts.runtimeClient.FetchRuntimeConfig()
	require.NoError(t, err)

	ts.log.Infof("Creating a secret %s with test data", ts.ConfigName)
	err = ts.secretClient.Create(ts.testSecret(config))
	require.NoError(t, err)

	ts.log.Info("Starting upgrade")
	upgradeOperationID, err := ts.upgradeSuite.upgradeClient.UpgradeRuntimeToVersion(ts.upgradeSuite.UpgradeKymaVersion)
	require.NoError(t, err, "failed to upgrade Runtime")

	ts.log.Info("Waiting for upgrade to finish...")
	err = ts.upgradeSuite.upgradeClient.AwaitOperationFinished(upgradeOperationID, ts.upgradeSuite.UpgradeTimeout)
	require.NoError(t, err, "error waiting for upgrade to finish")

}
