package event_hub

import (
	"context"
	"fmt"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

const (
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort = 9093

	k8sSecretName                     = "secret-name"
	k8sSecretNamespace                = "knative-eventing"
	componentNameKnativeEventing      = "knative-eventing"
	componentNameKnativeEventingKafka = "knative-eventing-kafka"
	kafkaProvider                     = "azure"
)

/*TODO(anishj0shi)
1) create an interface type "event-hub-provider" which exposes some functions like
GetEventHubNamespace, MarkNamespaceAsUsed, GetNamespaceAccessCredentials, inject this interface as an input
to NewProvisioningAzureEventHubStep
2) Refactor Azure Client Wrapper to conform to the above protocol
3) Implement retry logic for Namespace retrieval and NamespaceTagging operation.
*/

type ProvisionAzureEventHubStep struct {
	operationManager *process.OperationManager
	namespaceClient  azure.NamespaceClientInterface
	context          context.Context
}

func NewProvisionAzureEventHubStep(os storage.Operations, namespaceClient azure.NamespaceClientInterface, ctx context.Context) *ProvisionAzureEventHubStep {
	return &ProvisionAzureEventHubStep{
		operationManager: process.NewOperationManager(os),
		namespaceClient:  namespaceClient,
		context:          ctx,
	}
}

func (p *ProvisionAzureEventHubStep) Name() string {
	return "Provision Azure Event Hubs"
}

func (p *ProvisionAzureEventHubStep) Run(operation internal.ProvisioningOperation,
	log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	unusedEventHubNamespace, err := azure.GetFirstUnusedNamespaces(p.context, p.namespaceClient)
	if err != nil {
		return p.operationManager.OperationFailed(operation, "no azure event-hub namespace found in the given subscription")
	}

	log.Printf("Get Access Keys for Azure EventHubs Namespace [%s]\n", unusedEventHubNamespace)
	resourceGroup := azure.GetResourceGroup(unusedEventHubNamespace)

	log.Printf("Found unused EventHubs Namespace, name: %v, resourceGroup: %v", unusedEventHubNamespace.Name, resourceGroup)

	accessKeys, err := azure.GetEventHubsNamespaceAccessKeys(p.context, p.namespaceClient, resourceGroup, *unusedEventHubNamespace.Name, authorizationRuleName)
	if err != nil {
		log.Fatalf("Failed to get Access Keys for Azure EventHubs Namespace [%s] with error: %v\n", unusedEventHubNamespace, err)
	}

	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaEndpoint = appendPort(kafkaEndpoint, kafkaPort)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	if _, err := azure.MarkNamespaceAsUsed(p.context, p.namespaceClient, resourceGroup, unusedEventHubNamespace); err != nil {
		return p.operationManager.OperationFailed(operation, "no azure event-hub namespace found in the given subscription")
	}

	operation.InputCreator.SetOverrides(componentNameKnativeEventing, getKnativeEventingOverrides())
	operation.InputCreator.SetOverrides(componentNameKnativeEventingKafka, getKafkaChannelOverrides(kafkaEndpoint, k8sSecretNamespace, "$ConnectionString", kafkaPassword, kafkaProvider))

	return operation, 0, nil
}

func extractEndpoint(accessKeys *eventhub.AccessKeys) string {
	endpoint := strings.Split(*accessKeys.PrimaryConnectionString, ";")[0]
	endpoint = strings.TrimPrefix(endpoint, "Endpoint=sb://")
	endpoint = strings.TrimSuffix(endpoint, "/")
	return endpoint
}

func appendPort(endpoint string, port int64) string {
	return fmt.Sprintf("%s:%d", endpoint, port)
}

func getKnativeEventingOverrides() []*gqlschema.ConfigEntryInput {
	var knativeOverrides []*gqlschema.ConfigEntryInput
	knativeOverrides = []*gqlschema.ConfigEntryInput{
		{
			Key:   "knative-eventing.channel.default.apiVersion",
			Value: "knativekafka.kyma-project.io/v1alpha1",
		},
		{
			Key:   "knative-eventing.channel.default.kind",
			Value: "KafkaChannel",
		},
	}
	return knativeOverrides
}

func getKafkaChannelOverrides(broker, namespace, username, password, kafkaProvider string) []*gqlschema.ConfigEntryInput {
	kafkaOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "kafka.brokers",
			Value:  broker,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.namespace",
			Value:  namespace,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.password",
			Value:  password,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.username",
			Value:  username,
			Secret: ptr.Bool(true),
		},
		{
			Key:    "kafka.secretName",
			Value:  "knative-kafka",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "environment.kafkaProvider",
			Value:  kafkaProvider,
			Secret: ptr.Bool(true),
		},
	}
	return kafkaOverrides
}
