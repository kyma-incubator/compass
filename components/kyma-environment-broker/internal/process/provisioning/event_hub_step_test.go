package provisioning

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pivotal-cf/brokerapi/v7/domain"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	hyperscalerautomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure"
	azuretesting "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/azure/testing"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	inputAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

const (
	fixSubAccountID = "test-sub-account-id"
	fixInstanceID   = "test-instance-id"
	fixOperationID  = "test-operation-id"
)

func fixLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}

func Test_HappyPath(t *testing.T) {
	// given
	tags := fixTags()
	memoryStorage := storage.NewMemoryStorage()
	accountProvider := fixAccountProvider()
	namespaceClient := azuretesting.NewFakeNamespaceClientHappyPath()
	step := fixEventHubStep(memoryStorage.Operations(), azuretesting.NewFakeHyperscalerProvider(namespaceClient), accountProvider)
	op := fixProvisioningOperation(t)
	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := memoryStorage.Operations().InsertProvisioningOperation(op)
	require.NoError(t, err)

	// when
	op.UpdatedAt = time.Now()
	op, when, err := step.Run(op, fixLogger())
	require.NoError(t, err)
	provisionRuntimeInput, err := op.InputCreator.Create()
	require.NoError(t, err)

	// then
	ensureOperationSuccessful(t, op, when, err)
	allOverridesFound := ensureOverrides(t, provisionRuntimeInput)
	assert.True(t, allOverridesFound[componentNameKnativeEventing], "overrides for %s were not found", componentNameKnativeEventing)
	assert.True(t, allOverridesFound[componentNameKnativeEventingKafka], "overrides for %s were not found", componentNameKnativeEventingKafka)
	assert.Equal(t, namespaceClient.Tags, tags)
}

func Test_StepsUnhappyPath(t *testing.T) {
	tests := []struct {
		name                string
		giveOperation       func(t *testing.T) internal.ProvisioningOperation
		giveStep            func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep
		wantRepeatOperation bool
	}{
		{
			name:          "Provision parameter errors",
			giveOperation: fixInvalidProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *fixEventHubStep(storage.Operations(), azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientHappyPath()), accountProvider)
			},
			wantRepeatOperation: false,
		},
		{
			name:          "AccountProvider cannot get gardener credentials",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProviderGardenerCredentialsError()
				return *fixEventHubStep(storage.Operations(), azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientHappyPath()), accountProvider)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "EventHubs Namespace creation error",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *NewProvisionAzureEventHubStep(storage.Operations(),
					// ups ... namespace cannot get created
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientCreationError()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "Error while getting EventHubs Namespace credentials",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *NewProvisionAzureEventHubStep(storage.Operations(),
					// ups ... namespace cannot get listed
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceClientListError()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "No error while getting EventHubs Namespace credentials, but PrimaryConnectionString in AccessKey is nil",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *NewProvisionAzureEventHubStep(storage.Operations(),
					// ups ... PrimaryConnectionString is nil
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceAccessKeysNil()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: true,
		},
		{
			name:          "Error while getting config from HAP",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *NewProvisionAzureEventHubStep(storage.Operations(),
					// ups ... client cannot be created
					azuretesting.NewFakeHyperscalerProviderError(),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: false,
		},
		{
			name:          "Error while creating Azure ResourceGroup",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProvider()
				return *NewProvisionAzureEventHubStep(storage.Operations(),
					// ups ... resource group cannot be created
					azuretesting.NewFakeHyperscalerProvider(azuretesting.NewFakeNamespaceResourceGroupError()),
					accountProvider,
					context.Background(),
				)
			},
			wantRepeatOperation: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// given
			memoryStorage := storage.NewMemoryStorage()
			op := tt.giveOperation(t)
			step := tt.giveStep(t, memoryStorage)
			// this is required to avoid storage retries (without this statement there will be an error => retry)
			err := memoryStorage.Operations().InsertProvisioningOperation(op)
			require.NoError(t, err)

			// when
			op.UpdatedAt = time.Now()
			op, when, err := step.Run(op, fixLogger())
			require.NotNil(t, op)

			// then
			if tt.wantRepeatOperation {
				ensureOperationIsRepeated(t, err, when)
			} else {
				ensureOperationIsNotRepeated(t, err)
			}
		})
	}
}

// operationManager.OperationFailed(...)
// manager.go: if processedOperation.State != domain.InProgress { return 0, nil } => repeat
// queue.go: if err == nil && when != 0 => repeat

func ensureOperationIsRepeated(t *testing.T, err error, when time.Duration) {
	t.Helper()
	assert.Nil(t, err)
	assert.True(t, when != 0)
}

func ensureOperationIsNotRepeated(t *testing.T, err error) {
	t.Helper()
	assert.NotNil(t, err)
}

// ensureOverrides ensures that the overrides for
// - the kafka channel controller
// - and the default knative channel
// are set
func ensureOverrides(t *testing.T, provisionRuntimeInput gqlschema.ProvisionRuntimeInput) map[string]bool {
	t.Helper()

	allOverridesFound := map[string]bool{
		componentNameKnativeEventing:      false,
		componentNameKnativeEventingKafka: false,
	}

	kymaConfig := provisionRuntimeInput.KymaConfig
	for _, component := range kymaConfig.Components {
		switch component.Component {
		case componentNameKnativeEventing:
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "knative-eventing.channel.default.apiVersion",
				Value:  "knativekafka.kyma-project.io/v1alpha1",
				Secret: nil,
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "knative-eventing.channel.default.kind",
				Value:  "KafkaChannel",
				Secret: nil,
			})
			allOverridesFound[componentNameKnativeEventing] = true
		case componentNameKnativeEventingKafka:
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.brokers.hostname",
				Value:  "name",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.brokers.port",
				Value:  "9093",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.namespace",
				Value:  "knative-eventing",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.password",
				Value:  "Endpoint=sb://name/;",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.username",
				Value:  "$ConnectionString",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.secretName",
				Value:  "knative-kafka",
				Secret: ptr.Bool(true),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "environment.kafkaProvider",
				Value:  kafkaProvider,
				Secret: ptr.Bool(true),
			})
			allOverridesFound[componentNameKnativeEventingKafka] = true
		}
	}

	return allOverridesFound
}

func fixKnativeKafkaInputCreator(t *testing.T) internal.ProvisionInputCreator {
	optComponentsSvc := &inputAutomock.OptionalComponentService{}
	componentConfigurationInputList := internal.ComponentConfigurationInputList{
		{
			Component:     "keb",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
		{
			Component: componentNameKnativeEventing,
			Namespace: "knative-eventing",
		},
		{
			Component: componentNameKnativeEventingKafka,
			Namespace: "knative-eventing",
		},
	}

	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return([]string{})
	optComponentsSvc.On("ExecuteDisablers", mock.Anything).Return(componentConfigurationInputList, nil)

	kymaComponentList := []v1alpha1.KymaComponent{
		{
			Name:      "keb",
			Namespace: "kyma-system",
		},
		{
			Name:      componentNameKnativeEventing,
			Namespace: "knative-eventing",
		},
		{
			Name:      componentNameKnativeEventingKafka,
			Namespace: "knative-eventing",
		},
	}
	componentsProvider := &inputAutomock.ComponentListProvider{}
	componentsProvider.On("AllComponents", kymaVersion).Return(kymaComponentList, nil)
	defer componentsProvider.AssertExpectations(t)

	ibf, err := input.NewInputBuilderFactory(optComponentsSvc, componentsProvider, input.Config{}, kymaVersion)
	assert.NoError(t, err)

	creator, err := ibf.ForPlan(broker.GCPPlanID, "")
	if err != nil {
		t.Errorf("cannot create input creator for %q plan", broker.GCPPlanID)
	}

	return creator
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

func fixAccountProviderGardenerCredentialsError() *hyperscalerautomock.AccountProvider {
	accountProvider := hyperscalerautomock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		HyperscalerType: hyperscaler.Azure,
		CredentialData:  map[string][]byte{},
	}, fmt.Errorf("ups ... gardener credentials could not be retrieved"))
	return &accountProvider
}

func fixEventHubStep(memoryStorageOp storage.Operations, hyperscalerProvider azure.HyperscalerProvider,
	accountProvider *hyperscalerautomock.AccountProvider) *ProvisionAzureEventHubStep {
	return NewProvisionAzureEventHubStep(memoryStorageOp, hyperscalerProvider, accountProvider, context.Background())
}

func fixProvisioningOperation(t *testing.T) internal.ProvisioningOperation {
	op := internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:         fixOperationID,
			InstanceID: fixInstanceID,
		},
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
		}`,
		InputCreator: fixKnativeKafkaInputCreator(t),
	}
	return op
}

func fixInvalidProvisioningOperation(t *testing.T) internal.ProvisioningOperation {
	op := internal.ProvisioningOperation{
		Operation: internal.Operation{},
		// ups .. invalid json
		ProvisioningParameters: `{
			"parameters": a{}a
		}`,
		InputCreator: fixKnativeKafkaInputCreator(t),
	}
	return op
}

func fixTags() azure.Tags {
	return azure.Tags{
		azure.TagSubAccountID: ptr.String(fixSubAccountID),
		azure.TagOperationID:  ptr.String(fixOperationID),
		azure.TagInstanceID:   ptr.String(fixInstanceID),
	}
}

func ensureOperationSuccessful(t *testing.T, op internal.ProvisioningOperation, when time.Duration, err error) {
	t.Helper()
	assert.Equal(t, when, time.Duration(0))
	assert.Equal(t, op.Operation.State, domain.LastOperationState(""))
	assert.Nil(t, err)
}
