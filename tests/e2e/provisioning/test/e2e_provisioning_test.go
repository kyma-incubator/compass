package test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/tests/e2e/provisioning/internal/hyperscaler/azure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_E2E_Provisioning(t *testing.T) {
	ts := newTestSuite(t)
	if ts.IsDummyTest {
		return
	}
	if ts.IsCleanupPhase {
		ts.Cleanup()
		return
	}
	configMap := ts.testConfigMap()

	operationID, err := ts.brokerClient.ProvisionRuntime()
	require.NoError(t, err)

	ts.log.Infof("Creating config map %s with test data", ts.ConfigName)
	err = ts.configMapClient.Create(configMap)
	require.NoError(t, err)

	err = ts.brokerClient.AwaitOperationSucceeded(operationID, ts.ProvisionTimeout)
	require.NoError(t, err)

	dashboardURL, err := ts.brokerClient.FetchDashboardURL()
	require.NoError(t, err)

	ts.log.Infof("Updating config map %s with dashboardUrl", ts.ConfigName)
	configMap.Data[dashboardUrlKey] = dashboardURL
	err = ts.configMapClient.Update(configMap)
	require.NoError(t, err)

	ts.log.Info("Fetching runtime's kubeconfig")
	config, err := ts.runtimeClient.FetchRuntimeConfig()
	require.NoError(t, err)

	ts.log.Infof("Creating a secret %s with test data", ts.ConfigName)
	err = ts.secretClient.Create(ts.testSecret(config))
	require.NoError(t, err)

	err = ts.dashboardChecker.AssertRedirectedToUAA(dashboardURL)
	assert.NoError(t, err)

	if ts.IsTestAzureEventHubsEnabled {
		checkAzureEventHubProperties(ts, t, operationID)
	}
}

func checkAzureEventHubProperties(ts *Suite, t *testing.T, operationID string) {
	resourceGroup, err := (*ts.azureClient).GetResourceGroup(context.TODO(), ts.brokerClient.GetClusterName())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resourceGroup.StatusCode, "HTTP GET fails for ResourceGroup")
	assert.NotNil(t, resourceGroup, "ResourceGroup name is nil")
	assert.Equal(t, ts.brokerClient.GetClusterName(), *resourceGroup.Name, "ResourceGroup Name is incorrect")

	// Check tags for ResourceGroup
	assert.Equal(t, operationID, *resourceGroup.Tags[azure.TagOperationID], "Tag OperationID for ResourceGroup is incorrect")
	assert.Equal(t, ts.brokerClient.InstanceID(), *resourceGroup.Tags[azure.TagInstanceID], "Tag InstanceID for ResourceGroup is incorrect")
	assert.Equal(t, ts.brokerClient.SubAccountID(), *resourceGroup.Tags[azure.TagSubAccountID], "Tag SubAccountID for ResourceGroup is incorrect")

	namespace, err := (*ts.azureClient).GetEHNamespace(context.TODO(), ts.brokerClient.GetClusterName(), ts.brokerClient.GetClusterName())
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resourceGroup.StatusCode, "HTTP GET fails for EH Namespace")
	assert.NotNil(t, namespace, "EH Namespace name is nil")
	assert.Equal(t, ts.brokerClient.GetClusterName(), *namespace.Name, "EH Namespace name is incorrect")

	// Check tags for EH Namespace
	assert.Equal(t, operationID, *namespace.Tags[azure.TagOperationID], "Tag OperationID for EH Namespace is incorrect")
	assert.Equal(t, ts.brokerClient.InstanceID(), *namespace.Tags[azure.TagInstanceID], "Tag InstanceID for EH Namespace is incorrect")
	assert.Equal(t, ts.brokerClient.SubAccountID(), *resourceGroup.Tags[azure.TagSubAccountID], "Tag SubAccountID for EH Namespace is incorrect")

}
