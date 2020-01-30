package test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/require"
)

func Test_E2E_Provisioning(t *testing.T) {
	ts := newTestSuite(t)
	operationID, err := ts.brokerClient.ProvisionRuntime()
	require.NoError(t, err)
	defer ts.TearDown()

	err = ts.brokerClient.AwaitProvisioningSucceeded(operationID)
	assert.NoError(t, err)

	dashboardURL, err := ts.brokerClient.FetchDashboardURL()
	assert.NoError(t, err)

	err = ts.kymaClient.CallDashboardURL(dashboardURL)
	assert.NoError(t, err)
}
