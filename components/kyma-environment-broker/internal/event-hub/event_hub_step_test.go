package event_hub

import (
	"context"
	"testing"

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
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/hyperscaler/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	inputAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
)

const kymaVersion = "1.10"

func fixLogger() logrus.FieldLogger {
	return logrus.StandardLogger()
}

/// A fake client for Azure EventHubs Namespace handling
type FakeNamespaceClient struct {
	eventHubs []eventhub.EHNamespace
}

func (nc *FakeNamespaceClient) ListKeys(ctx context.Context, resourceGroupName string, namespaceName string, authorizationRuleName string) (result eventhub.AccessKeys, err error) {
	return eventhub.AccessKeys{
		PrimaryConnectionString: ptr.String("Endpoint=sb://name/;"),
	}, nil
}

func (nc *FakeNamespaceClient) Update(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	return parameters, nil
}

func (nc *FakeNamespaceClient) ListComplete(ctx context.Context) (result eventhub.EHNamespaceListResultIterator, err error) {
	return eventhub.EHNamespaceListResultIterator{}, nil
}

func (nc *FakeNamespaceClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, namespaceName string, parameters eventhub.EHNamespace) (result eventhub.EHNamespace, err error) {
	return eventhub.EHNamespace{}, nil
}

func (nc *FakeNamespaceClient) PersistResourceGroup(ctx context.Context, config *azure.Config, name string) (resources.Group, error) {
	return resources.Group{
		Name: ptr.String("my-resourcegroup"),
	}, nil
}

func (nc *FakeNamespaceClient) PersistEventHubsNamespace(ctx context.Context, azureCfg *azure.Config, namespaceClient azure.NamespaceClientInterface, groupName, namespace string) (*eventhub.EHNamespace, error) {
	return &eventhub.EHNamespace{
		Name: ptr.String("my-ehnamespace"),
	}, nil
}

type fakeAzureClient struct{}

func (ac *fakeAzureClient) GetNamespacesClientOrDie(config *azure.Config) azure.NamespaceClientInterface {
	return &FakeNamespaceClient{}
}

// ensure the fake client is implementing the interface
var _ azure.NamespaceClientInterface = (*FakeNamespaceClient)(nil)

func Test_Overrides(t *testing.T) {
	// given
	memoryStorageOp := storage.NewMemoryStorage().Operations()
	accountProvider := automock.AccountProvider{}
	accountProvider.On("GardenerCredentials", hyperscaler.Azure, mock.Anything).Return(hyperscaler.Credentials{
		CredentialName:  "",
		HyperscalerType: "",
		TenantName:      "",
		CredentialData: map[string][]byte{
			"subscriptionID": []byte("subscriptionID"),
			"clientID":       []byte("clientID"),
			"clientSecret":   []byte("clientSecret"),
			"tenantID":       []byte("tenantID"),
		},
	}, nil)

	step := NewProvisionAzureEventHubStep(memoryStorageOp,
		&fakeAzureClient{},
		&accountProvider,
		context.Background(),
	)
	op := internal.ProvisioningOperation{
		Operation: internal.Operation{},
		ProvisioningParameters: `{
			"parameters": {
        		"name": "nachtmaar-15",
        		"components": [],
				"region": "europe-west3"
			}
		}`,
		InputCreator: fixInputCreator(t),
	}

	// when
	op, _, err := step.Run(op, fixLogger())
	require.NoError(t, err)
	// require.Zero(t, when)
	provisionRuntimeInput, err := op.InputCreator.Create()
	require.NoError(t, err)

	// then
	allOverridesFound := ensureOverrides(t, provisionRuntimeInput)
	require.True(t, allOverridesFound[componentNameKnativeEventing], "overrides for %s were not found", componentNameKnativeEventing)
	require.True(t, allOverridesFound[componentNameKnativeEventingKafka], "overrides for %s were not found", componentNameKnativeEventingKafka)
}

// ensureOverrides ensures that the overrides for
// - the kafka channel controller
// - and the default knative channel
// is set
func ensureOverrides(t *testing.T, provisionRuntimeInput gqlschema.ProvisionRuntimeInput) map[string]bool {
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
				Secret: ptr.Bool(false),
			})
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "knative-eventing.channel.default.kind",
				Value:  "KafkaChannel",
				Secret: ptr.Bool(false),
			})
			allOverridesFound[componentNameKnativeEventing] = true
			break
		case componentNameKnativeEventingKafka:
			assert.Contains(t, component.Configuration, &gqlschema.ConfigEntryInput{
				Key:    "kafka.brokers",
				Value:  "name:9093",
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
			break
		}
	}

	return allOverridesFound
}

func fixInputCreator(t *testing.T) internal.ProvisionInputCreator {
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
	ibf := input.NewInputBuilderFactory(optComponentsSvc, kymaComponentList, input.Config{}, kymaVersion)

	creator, found := ibf.ForPlan(broker.AzurePlanID)
	if !found {
		t.Errorf("input creator for %q plan does not exist", broker.AzurePlanID)
	}

	return creator
}
