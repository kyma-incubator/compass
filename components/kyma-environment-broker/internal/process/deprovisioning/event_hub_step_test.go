package deprovisioning

import (
	"context"
	"testing"
	"time"

	"github.com/pivotal-cf/brokerapi/v7/domain"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	hyperscalerautomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	azuretesting "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure/testing"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

const (
	fixSubAccountID = "test-sub-account-id"
)

// TODO: table test ?
// Idea:
// 1. a ResourceGroup exists before we call the deproviosioning step
// 2. resourceGroup is in deletion state during retry wait time before we call the deproviosioning step again
// 3. expectation is that no new deprovisioning is triggered
// 4. after calling step again - expectation is that the deprovisioning succeeded now
func Test_DeprovisionSucceeded_ResourceGroupInDeletionMode(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	accountProvider := fixAccountProvider()
	namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
	namespaceClientResourceGroupInDeletionMode := azuretesting.NewFakeNamespaceClientResourceGroupInDeletionMode()
	namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

	stepResourceGroupExists := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
	stepResourceGroupInDeletionMode := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupInDeletionMode), accountProvider)
	stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)
	op := fixDeprovisioningOperationWithParameters()
	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := memoryStorage.Operations().InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// when
	op.UpdatedAt = time.Now()
	op, _, err = stepResourceGroupExists.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	// retry is triggered to wait for deletion
	ensureOperationIsRepeated(t, err, time.Minute, op)

	// when
	op, when, err := stepResourceGroupInDeletionMode.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	// do not try to delete again
	assert.False(t, namespaceClientResourceGroupInDeletionMode.DeleteResourceGroupCalled)
	ensureOperationIsRepeated(t, err, when, op)

	// when
	op, when, err = stepResourceGroupDoesNotExist.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	ensureOperationSuccessful(t, when, op)
}

// TODO: table test ?
// Idea:
// 1. a ResourceGroup exists before we call the deproviosioning step
// 2. resourceGroup got deleted during retry wait time before we call the deproviosioning step again
// 3. expectation is that the deprovisioning succeeded now
func Test_DeprovisionSucceeded_ResourceGroupExists(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	accountProvider := fixAccountProvider()
	namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
	namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

	stepResourceGroupExists := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
	stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)
	op := fixDeprovisioningOperationWithParameters()
	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := memoryStorage.Operations().InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// when
	op.UpdatedAt = time.Now()
	op, _, err = stepResourceGroupExists.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	// retry is triggered to wait for deletion
	ensureOperationIsRepeated(t, err, time.Minute, op)

	// when
	op, when, err := stepResourceGroupDoesNotExist.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	ensureOperationSuccessful(t, when, op)
}

func Test_DeprovisionSucceeded_ResourceGroupDoesNotExist(t *testing.T) {
	// given
	memoryStorage := storage.NewMemoryStorage()
	accountProvider := fixAccountProvider()
	namespaceClient := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()
	step := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClient), accountProvider)
	op := fixDeprovisioningOperationWithParameters()
	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := memoryStorage.Operations().InsertDeprovisioningOperation(op)
	require.NoError(t, err)

	// when
	op.UpdatedAt = time.Now()
	op, when, err := step.Run(op, fixLogger())
	require.NoError(t, err)

	// then
	ensureOperationSuccessful(t, when, op)
}

func fixTags() azure.Tags {
	return azure.Tags{
		azure.TagSubAccountID: ptr.String(fixSubAccountID),
		azure.TagOperationID:  ptr.String(fixOperationID),
		azure.TagInstanceID:   ptr.String(fixInstanceID),
	}
}

func fixAccountProvider() *hyperscalerautomock.AccountProvider {
	accountProvider := hyperscalerautomock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		HyperscalerType: hyperscaler.Azure,
		CredentialData: map[string][]byte{
			"subscriptionID": []byte("subscriptionID"),
			"clientID":       []byte("clientID"),
			"clientSecret":   []byte("clientSecret"),
			"tenantID":       []byte("tenantID"),
		},
	}, nil)
	return &accountProvider
}

func fixEventHubStep(memoryStorageOp storage.Operations, hyperscalerProvider azure.HyperscalerProvider,
	accountProvider *hyperscalerautomock.AccountProvider) DeprovisionAzureEventHubStep {
	return NewDeprovisionAzureEventHubStep(memoryStorageOp, hyperscalerProvider, accountProvider, context.Background())
}

func fixLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}

func fixDeprovisioningOperationWithParameters() internal.DeprovisioningOperation {
	return internal.DeprovisioningOperation{
		Operation: internal.Operation{
			ID:                     fixOperationID,
			InstanceID:             fixInstanceID,
			ProvisionerOperationID: fixProvisionerOperationID,
			Description:            "",
			UpdatedAt:              time.Now(),
		},
		DeprovisioningParameters: `{
			"plan_id": "4deee563-e5ec-4731-b9b1-53b42d855f0c",
			"ers_context": {
				"subaccount_id": "` + fixSubAccountID + `"
			},
			"parameters": {
				"name": "nachtmaar-15",
				"components": [],
				"region": "westeurope"
			}
		}`,
	}
}

// operationManager.OperationFailed(...)
// manager.go: if processedOperation.State != domain.InProgress { return 0, nil } => repeat
// queue.go: if err == nil && when != 0 => repeat

func ensureOperationIsRepeated(t *testing.T, err error, when time.Duration, op internal.DeprovisioningOperation) {
	t.Helper()
	assert.Nil(t, err)
	assert.True(t, when != 0)
	assert.NotEqual(t, op.Operation.State, domain.Succeeded)
}

func ensureOperationIsNotRepeated(t *testing.T, err error) {
	t.Helper()
	assert.NotNil(t, err)
}

func ensureOperationSuccessful(t *testing.T, when time.Duration, op internal.DeprovisioningOperation) {
	t.Helper()
	assert.Equal(t, when, time.Duration(0))
	assert.Equal(t, op.Operation.State, domain.Succeeded)
}
