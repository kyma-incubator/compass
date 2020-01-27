package test

import (
	"github.com/stretchr/testify/require"
	"testing"
)


func TestKymaProvisioningE2E(t *testing.T) {
	ts := newTestSuite(t)
	instanceID, err := ts.brokerClient.ProvisionRuntime()
	require.NoError(t, err)

	err = ts.brokerClient.AwaitProvisioningSucceeded(instanceID)
	require.NoError(t, err)

	dashboardURL, err := ts.brokerClient.GetInstanceDetails(instanceID)
	require.NoError(t, err)
	require.NotEmpty(t, dashboardURL)



}