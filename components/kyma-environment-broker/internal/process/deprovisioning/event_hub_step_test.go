package deprovisioning

import (
	"context"
	"fmt"
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
		giveSteps           func(t *testing.T, memoryStorageOp storage.Operations, instanceStorage storage.Instances, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep
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
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, instanceStorage storage.Instances, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {
				namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
				namespaceClientResourceGroupInDeletionMode := azuretesting.NewFakeNamespaceClientResourceGroupInDeletionMode()
				namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

				stepResourceGroupExists := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
				stepResourceGroupInDeletionMode := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupInDeletionMode), accountProvider)
				stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)

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
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, instanceStorage storage.Instances, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {

				namespaceClientResourceGroupExists := azuretesting.NewFakeNamespaceClientResourceGroupExists()
				namespaceClientResourceGroupDoesNotExist := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()

				stepResourceGroupExists := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupExists), accountProvider)
				stepResourceGroupDoesNotExist := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClientResourceGroupDoesNotExist), accountProvider)
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
			giveSteps: func(t *testing.T, memoryStorageOp storage.Operations, instanceStorage storage.Instances, accountProvider *hyperscalerautomock.AccountProvider) []DeprovisionAzureEventHubStep {
				namespaceClient := azuretesting.NewFakeNamespaceClientResourceGroupDoesNotExist()
				step := fixEventHubStep(memoryStorageOp, instanceStorage, azuretesting.NewFakeHyperscalerProvider(namespaceClient), accountProvider)

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
			err = memoryStorage.Instances().Insert(fixInstance())
			require.NoError(t, err)
			steps := tt.giveSteps(t, memoryStorage.Operations(), memoryStorage.Instances(), accountProvider)
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

func Test_StepsUnhappyPath(t *testing.T) {
	tests := []struct {
		name                string
		giveOperation       func() internal.DeprovisioningOperation
		giveInstance		func() internal.Instance
		giveStep            func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep
		wantRepeatOperation bool
	}{
		{
			name:          "Instance provision parameter errors",
			giveOperation: fixDeprovisioningOperationWithParameters,
			giveInstance: fixInvalidInstance,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return fixEventHubStep(storage.Operations(), storage.Instances(), azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientHappyPath()), accountProvider)
			},
			wantRepeatOperation: false,
		},
		{
			name:          "AccountProvider cannot get gardener credentials",
			giveOperation: fixDeprovisioningOperationWithParameters,
			giveInstance: fixInstance,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep {
				accountProvider := fixAccountProviderGardenerCredentialsError()
				return fixEventHubStep(storage.Operations(), storage.Instances(), azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientHappyPath()), accountProvider)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "Error while getting EventHubs Namespace credentials",
			giveOperation: fixDeprovisioningOperationWithParameters,
			giveInstance: fixInstance,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep {
				accountProvider := fixAccountProviderGardenerCredentialsError()
				return NewDeprovisionAzureEventHubStep(storage.Operations(), storage.Instances(),
					// ups ... namespace cannot get listed
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientListError()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "Error while getting config from Credentials",
			giveOperation: fixDeprovisioningOperation,
			giveInstance: fixInstance,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep {
				accountProvider := fixAccountProviderGardenerCredentialsHAPError()
				return NewDeprovisionAzureEventHubStep(storage.Operations(), storage.Instances(),
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceAccessKeysNil()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: false,
		},
		{
			name:          "Error while getting client from HAP",
			giveOperation: fixDeprovisioningOperation,
			giveInstance: fixInstance,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) DeprovisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return NewDeprovisionAzureEventHubStep(storage.Operations(), storage.Instances(),
					// ups ... client cannot be created
					azuretesting.NewFakeHyperscalerProviderError(),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			memoryStorage := storage.NewMemoryStorage()
			op := tt.giveOperation()
			step := tt.giveStep(t, memoryStorage)
			// this is required to avoid storage retries (without this statement there will be an error => retry)
			// TODO(nachtmaar): can this be deleted ??
			//TODO(montaro): Had to get it back, otherwise the failed operation would be retried
			err := memoryStorage.Operations().InsertDeprovisioningOperation(op)
			require.NoError(t, err)
			err = memoryStorage.Instances().Insert(tt.giveInstance())
			require.NoError(t, err)

			// when
			op.UpdatedAt = time.Now()
			op, when, err := step.Run(op, fixLogger())
			require.NotNil(t, op)

			// then
			if tt.wantRepeatOperation {
				ensureOperationIsRepeated(t, op, when, err)
			} else {
				ensureOperationIsNotRepeated(t, err)
			}
		})
	}
}

func fixInstance() internal.Instance {
	// TODO(nachtmaar): share with provisiong test ?
	return internal.Instance{
		InstanceID: fixInstanceID,
		ProvisioningParameters: `{
			"plan_id": "4deee563-e5ec-4731-b9b1-53b42d855f0c",
			"ers_context": {
				"subaccount_id": "` + fixSubAccountID + `"
			},
			"parameters": {
				"name": "nachtmaar-15",
				"components": [],
				"region": "westeurope"
			}
		}`}
}

func fixInvalidInstance() internal.Instance {
	// TODO(nachtmaar): share with provisiong test ?
	return internal.Instance{
		InstanceID: fixInstanceID,
		ProvisioningParameters: `}{INVALID JSON}{`}
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

func fixEventHubStep(memoryStorageOp storage.Operations, instanceStorage storage.Instances, hyperscalerProvider azure.HyperscalerProvider,
	accountProvider *hyperscalerautomock.AccountProvider) DeprovisionAzureEventHubStep {
	return NewDeprovisionAzureEventHubStep(memoryStorageOp, instanceStorage, hyperscalerProvider, accountProvider, context.Background())
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

func fixAccountProviderGardenerCredentialsError() *hyperscalerautomock.AccountProvider {
	accountProvider := hyperscalerautomock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		HyperscalerType: hyperscaler.Azure,
		CredentialData:  map[string][]byte{},
	}, fmt.Errorf("ups ... gardener credentials could not be retrieved"))
	return &accountProvider
}

func fixAccountProviderGardenerCredentialsHAPError() *hyperscalerautomock.AccountProvider {
	accountProvider := hyperscalerautomock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		HyperscalerType: hyperscaler.AWS,
		CredentialData: map[string][]byte{
			"subscriptionID": []byte("subscriptionID"),
			"clientID":       []byte("clientID"),
			"clientSecret":   []byte("clientSecret"),
			"tenantID":       []byte("tenantID"),
		},
	}, nil)
	return &accountProvider
}
