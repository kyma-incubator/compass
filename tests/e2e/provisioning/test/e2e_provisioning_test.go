package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKymaProvisioningE2E(t *testing.T) {
	ts := newTestSuite(t)
	instanceID, operationID, err := ts.brokerClient.ProvisionRuntime()
	require.NoError(t, err)

	err = ts.brokerClient.AwaitProvisioningSucceeded(instanceID, operationID)
	require.NoError(t, err)

	dashboardURL, err := ts.brokerClient.GetInstanceDetails(instanceID)
	require.NoError(t, err)
	require.NotEmpty(t, dashboardURL)

	err = ts.kymaClient.CallDashboard(dashboardURL)
	require.NoError(t, err)

	ts.log.Info("Provisioning test end with success")
	ts.log.Info("Cleaning up...")

	// Fetch gardener kubeconfig which is inside gardener secret
	// From gardener we can fetch the runtime's kubeconfig and trigger the cleaning logic. Delete instances -> brokers..

	// deprovision runtime via broker?
	// service-manager service instance leftovers cleaning?
}
