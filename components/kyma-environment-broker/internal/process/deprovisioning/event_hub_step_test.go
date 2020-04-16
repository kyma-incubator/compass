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

type wantStateFunction = func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient)

func Test_StepsProvisionSucceeded(t *testing.T) {
	tests := []struct {
		name                string
		giveOperation       func() internal.DeprovisioningOperation
		giveSteps           func(t *testing.T, memoryStorageOp storage.Operations, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep
		wantRepeatOperation bool
		wantStates          func(t *testing.T) []wantStateFunction
	}{
		{
			// 1. a ResourceGroup exists before we call the deproviosioning step
			// 2. resourceGroup is in deletion state during retry wait time before we call the deproviosioning step again
			// 3. expectation is that no new deprovisioning is triggered
			// 4. after calling step again - expectation is that the deprovisioning succeeded now
			name:          "ResourceGroupInDeletionMode",
			giveOperation: fixDeprovisioningOperationWithParameters,
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {
				namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
				namespaceClientResourceGroupInDeletionMode := azuretesting.NewFakeNamespaceClientResourceGroupInDeletionMode()
				namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

				stepResourceGroupExists := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
				stepResourceGroupInDeletionMode := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupInDeletionMode), accountProvider)
				stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)

				return []DeprovisionAzureEventHubStep{
					stepResourceGroupExists,
					stepResourceGroupInDeletionMode,
					stepResourceGroupDoesNotExist,
				}
			},
			wantStates: func(t *testing.T) []wantStateFunction {
				return []wantStateFunction{
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						ensureOperationIsRepeated(t, operation, when, err)
					},
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						assert.False(t, azureClient.DeleteResourceGroupCalled)
						ensureOperationIsRepeated(t, operation, when, err)
					},
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						ensureOperationSuccessful(t, operation, when, err)
					},
				}
			},
		},
		{
			// Idea:
			// 1. a ResourceGroup exists before we call the deproviosioning step
			// 2. resourceGroup got deleted during retry wait time before we call the deproviosioning step again
			// 3. expectation is that the deprovisioning succeeded now
			name:          "ResourceGroupExists",
			giveOperation: fixDeprovisioningOperationWithParameters,
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {

				namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
				namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

				stepResourceGroupExists := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
				stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)
				return []DeprovisionAzureEventHubStep{
					stepResourceGroupExists,
					stepResourceGroupDoesNotExist,
				}
			},
			wantStates: func(t *testing.T) []wantStateFunction {
				return []wantStateFunction{
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						ensureOperationIsRepeated(t, operation, when, err)
					},
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						ensureOperationSuccessful(t, operation, when, err)
					},
				}
			},
		},
		{

			// Idea:
			// 1. a ResourceGroup does not exist before we call the deproviosioning step
			// 2. expectation is that the deprovisioning succeeded
			name: "ResourceGroupDoesNotExist",
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {
				namespaceClient := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()
				step := fixEventHubStep(memoryStorageOp, azuretesting.NewFakeHyperscalerProvider(namespaceClient), accountProvider)

				return []DeprovisionAzureEventHubStep{
					step,
				}
			},
			wantStates: func(t *testing.T) []wantStateFunction {
				return []wantStateFunction{
					func(t *testing.T, operation internal.DeprovisioningOperation, when time.Duration, err error, azureClient azuretesting.FakeNamespaceClient) {
						ensureOperationSuccessful(t, operation, when, err)
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			memoryStorage := storage.NewMemoryStorage()
			accountProvider := fixAccountProvider()

			op := fixDeprovisioningOperationWithParameters()
			// this is required to avoid storage retries (without this statement there will be an error => retry)
			err := memoryStorage.Operations().InsertDeprovisioningOperation(op)
			require.NoError(t, err)
			steps := tt.giveSteps(t, memoryStorage.Operations(), accountProvider)
			wantStates := tt.wantStates(t)
			for idx, step := range steps {
				// when
				op.UpdatedAt = time.Now()
				op, when, err := step.Run(op, fixLogger())
				require.NoError(t, err)

				fakeHyperscalerProvider, ok := step.HyperscalerProvider.(*azuretesting.FakeHyperscalerProvider)
				if !ok {
					require.True(t, ok)
				}
				fakeAzureClient, ok := fakeHyperscalerProvider.Client.(*azuretesting.FakeNamespaceClient)
				if !ok {
					require.True(t, ok)
				}

				// then
				wantStates[idx](t, op, when, err, *fakeAzureClient)
			}

		})
	}

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

func ensureOperationIsRepeated(t *testing.T, op internal.DeprovisioningOperation, when time.Duration, err error) {
	t.Helper()
	assert.Nil(t, err)
	assert.True(t, when != 0)
	assert.NotEqual(t, op.Operation.State, domain.Succeeded)
}

func ensureOperationIsNotRepeated(t *testing.T, err error) {
	t.Helper()
	assert.NotNil(t, err)
}

func ensureOperationSuccessful(t *testing.T, op internal.DeprovisioningOperation, when time.Duration, err error) {
	t.Helper()
	assert.Equal(t, when, time.Duration(0))
	assert.Equal(t, op.Operation.State, domain.Succeeded)
}
