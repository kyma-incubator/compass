package provisioning

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2019-05-01/resources"
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

// ensure the fake client is implementing the interface
var _ azure.AzureInterface = (*FakeNamespaceClient)(nil)

/// A fake client for Azure EventHubs Namespace handling
type FakeNamespaceClient struct {
	persistEventhubsNamespaceError error
	resourceGroupError             error
	accessKeysError                error
	accessKeys                     *eventhub.AccessKeys
	tags                           azure.Tags
}

func (nc *FakeNamespaceClient) GetEventhubAccessKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	if nc.accessKeys != nil {
		return *nc.accessKeys, nil
	}
	return eventhub.AccessKeys{
		PrimaryConnectionString: ptr.String("Endpoint=sb://name/;"),
	}, nc.accessKeysError
}

func (nc *FakeNamespaceClient) CreateResourceGroup(ctx context.Context, config *azure.Config, name string, tags azure.Tags) (resources.Group, error) {
	nc.tags = tags
	return resources.Group{
		Name: ptr.String("my-resourcegroup"),
	}, nc.resourceGroupError
}

func (nc *FakeNamespaceClient) CreateNamespace(ctx context.Context, azureCfg *azure.Config, groupName, namespace string, tags azure.Tags) (*eventhub.EHNamespace, error) {
	nc.tags = tags
	return &eventhub.EHNamespace{
		Name: ptr.String(namespace),
	}, nc.persistEventhubsNamespaceError
}

func (nc *FakeNamespaceClient) DeleteResourceGroup(ctx context.Context) error {
	//TODO(montaro) double check me
	return nil
}

func NewFakeNamespaceClientCreationError() azure.AzureInterface {
	return &FakeNamespaceClient{persistEventhubsNamespaceError: fmt.Errorf("error while creating namespace")}
}

func NewFakeNamespaceClientListError() azure.AzureInterface {
	return &FakeNamespaceClient{accessKeysError: fmt.Errorf("cannot list namespaces")}
}

func NewFakeNamespaceResourceGroupError() azure.AzureInterface {
	return &FakeNamespaceClient{resourceGroupError: fmt.Errorf("cannot create resource group")}
}

func NewFakeNamespaceAccessKeysNil() azure.AzureInterface {
	return &FakeNamespaceClient{
		// no error here
		accessKeysError: nil,
		accessKeys: &eventhub.AccessKeys{
			// ups .. we got an AccessKeys with nil connection string even though there was no error
			PrimaryConnectionString: nil,
		},
	}
}

func NewFakeNamespaceClientHappyPath() *FakeNamespaceClient {
	return &FakeNamespaceClient{}
}

// ensure the fake client is implementing the interface
var _ azure.HyperscalerProvider = (*fakeHyperscalerProvider)(nil)

type fakeHyperscalerProvider struct {
	client azure.AzureInterface
	err    error
}

func (ac *fakeHyperscalerProvider) GetClient(config *azure.Config) (azure.AzureInterface, error) {
	return ac.client, ac.err
}

func NewFakeHyperscalerProvider(client azure.AzureInterface) azure.HyperscalerProvider {
	return &fakeHyperscalerProvider{
		client: client,
		err:    nil,
	}
}

func NewFakeHyperscalerProviderError() azure.HyperscalerProvider {
	return &fakeHyperscalerProvider{
		client: nil,
		err:    fmt.Errorf("ups ... hyperscaler provider could not provide a hyperscaler client"),
	}
}

func Test_HappyPath(t *testing.T) {
	// given
	tags := fixTags()
	memoryStorage := storage.NewMemoryStorage()
	accountProvider := fixAccountProvider()
	namespaceClient := NewFakeNamespaceClientHappyPath()
	step := fixEventHubStep(memoryStorage.Operations(), NewFakeHyperscalerProvider(namespaceClient), accountProvider)
	op := fixProvisioningOperation(t)
	// this is required to avoid storage retries (without this statement there will be an error => retry)
	err := memoryStorage.Operations().InsertProvisioningOperation(op)
	require.NoError(t, err)

	// when
	op.UpdatedAt = time.Now()
	op, _, err = step.Run(op, fixLogger())
	require.NoError(t, err)
	provisionRuntimeInput, err := op.InputCreator.Create()
	require.NoError(t, err)

	// then
	allOverridesFound := ensureOverrides(t, provisionRuntimeInput)
	assert.True(t, allOverridesFound[componentNameKnativeEventing], "overrides for %s were not found", componentNameKnativeEventing)
	assert.True(t, allOverridesFound[componentNameKnativeEventingKafka], "overrides for %s were not found", componentNameKnativeEventingKafka)
	assert.Equal(t, namespaceClient.tags, tags)
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
				return *fixEventHubStep(storage.Operations(), NewFakeHyperscalerProvider(NewFakeNamespaceClientHappyPath()), accountProvider)
			},
			wantRepeatOperation: false,
		},
		{
			name:          "AccountProvider cannot get gardener credentials",
			giveOperation: fixProvisioningOperation,
			giveStep: func(t *testing.T, storage storage.BrokerStorage) ProvisionAzureEventHubStep {
				accountProvider := fixAccountProviderGardenerCredentialsError()
				return *fixEventHubStep(storage.Operations(), NewFakeHyperscalerProvider(NewFakeNamespaceClientHappyPath()), accountProvider)
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
					NewFakeHyperscalerProvider(NewFakeNamespaceClientCreationError()),
					&accountProvider,
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
					NewFakeHyperscalerProvider(NewFakeNamespaceClientListError()),
					&accountProvider,
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
					NewFakeHyperscalerProvider(NewFakeNamespaceAccessKeysNil()),
					&accountProvider,
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
					NewFakeHyperscalerProviderError(),
					&accountProvider,
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
					NewFakeHyperscalerProvider(NewFakeNamespaceResourceGroupError()),
					&accountProvider,
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

	creator, err := ibf.ForPlan(broker.GcpPlanID, "")
	if err != nil {
		t.Errorf("cannot create input creator for %q plan", broker.GcpPlanID)
	}

	return creator
}

func fixAccountProvider() hyperscalerautomock.AccountProvider {
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
	return accountProvider
}

func fixAccountProviderGardenerCredentialsError() hyperscalerautomock.AccountProvider {
	accountProvider := hyperscalerautomock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		HyperscalerType: hyperscaler.Azure,
		CredentialData:  map[string][]byte{},
	}, fmt.Errorf("ups ... gardener credentials could not be retrieved"))
	return accountProvider
}

func fixEventHubStep(memoryStorageOp storage.Operations, hyperscalerProvider azure.HyperscalerProvider,
	accountProvider hyperscalerautomock.AccountProvider) *ProvisionAzureEventHubStep {
	return NewProvisionAzureEventHubStep(memoryStorageOp, hyperscalerProvider, &accountProvider, context.Background())
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
