package test

import (
	"context"
	"fmt"
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
	filter := fmt.Sprintf("tagName eq 'InstanceID' and tagValue eq '%s'", ts.InstanceID)
	groupListResultPage, err := (*ts.azureClient).ListResourceGroup(context.TODO(), filter, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, groupListResultPage.Response().StatusCode, "HTTP GET fails for ListResourceGroup")
	assert.Equal(t, 1, len(groupListResultPage.Values()), "groupListResultPage should return 1 ResourceGroup")

	rg := groupListResultPage.Values()[0]
	tagsRG := groupListResultPage.Values()[0].Tags

	// Check tags for ResourceGroup
	assert.Equal(t, operationID, *tagsRG[azure.TagOperationID], "Tag OperationID for ResourceGroup is incorrect")
	assert.Equal(t, ts.brokerClient.InstanceID(), *tagsRG[azure.TagInstanceID], "Tag InstanceID for ResourceGroup is incorrect")
	assert.Equal(t, ts.brokerClient.SubAccountID(), *tagsRG[azure.TagSubAccountID], "Tag SubAccountID for ResourceGroup is incorrect")

	ehNamespaceListResultPage, err := (*ts.azureClient).ListEHNamespaceByResourceGroup(context.TODO(), *rg.Name)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, ehNamespaceListResultPage.Response().StatusCode, "HTTP GET fails for EH Namespace")
	assert.Equal(t, 1, len(ehNamespaceListResultPage.Values()), "ehNamespaceListResultPage should return 1 EHNamespace")

	tagsEHNamespace := ehNamespaceListResultPage.Values()[0].Tags

	// Check tags for EH Namespace
	assert.Equal(t, operationID, *tagsEHNamespace[azure.TagOperationID], "Tag OperationID for EH Namespace is incorrect")
	assert.Equal(t, ts.brokerClient.InstanceID(), *tagsEHNamespace[azure.TagInstanceID], "Tag InstanceID for EH Namespace is incorrect")
	assert.Equal(t, ts.brokerClient.SubAccountID(), *tagsEHNamespace[azure.TagSubAccountID], "Tag SubAccountID for EH Namespace is incorrect")

}
