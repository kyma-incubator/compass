package event_hub

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/eventhub/mgmt/2017-04-01/eventhub"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/azure"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/event-hub/k8s"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

const (
	authorizationRuleName = "RootManageSharedAccessKey"

	kafkaPort = 9093

	k8sSecretName      = "secret-name"
	k8sSecretNamespace = "secret-namespace"
	componentName      = "knative-eventing"
)

type ProvisionAzureEventHubStep struct {
	operationManager *process.OperationManager
	config           azure.Config
	context          context.Context
}

func NewProvisionAzureEventHubStep(os storage.Operations, cfg azure.Config, ctx context.Context) *ProvisionAzureEventHubStep {
	return &ProvisionAzureEventHubStep{
		operationManager: process.NewOperationManager(os),
		config:           cfg,
		context:          ctx,
	}
}

func (p *ProvisionAzureEventHubStep) Name() string {
	return "Provision Azure Event Hubs"
}

func (p *ProvisionAzureEventHubStep) Run(operation internal.ProvisioningOperation,
	log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {

	unusedEventHubNamespace, err := azure.GetFirstUnusedNamespaces(p.context, &p.config)
	if err != nil {
		return p.operationManager.OperationFailed(operation, "no azure event-hub namespace found in the given subscription")
	}

	log.Printf("Get Access Keys for Azure EventHubs Namespace [%s]\n", unusedEventHubNamespace)
	resourceGroup := azure.GetResourceGroup(unusedEventHubNamespace)

	log.Printf("Found unused EventHubs Namespace, name: %v, resourceGroup: %v", unusedEventHubNamespace.Name, resourceGroup)

	accessKeys, err := azure.GetEventHubsNamespaceAccessKeys(p.context, &p.config, resourceGroup, *unusedEventHubNamespace.Name, authorizationRuleName)
	if err != nil {
		log.Fatalf("Failed to get Access Keys for Azure EventHubs Namespace [%s] with error: %v\n", unusedEventHubNamespace, err)
	}

	kafkaEndpoint := extractEndpoint(accessKeys)
	kafkaEndpoint = appendPort(kafkaEndpoint, kafkaPort)
	kafkaPassword := *accessKeys.PrimaryConnectionString

	secret := k8s.SecretFrom(k8sSecretName, k8sSecretNamespace, kafkaEndpoint, kafkaPassword, *unusedEventHubNamespace.Name)
	secretBytes, _ := json.MarshalIndent(secret, "", "    ")
	log.Printf("Kubernetes secret:\n%s", secretBytes)

	if _, err := azure.MarkNamespaceAsUsed(p.context, &p.config, resourceGroup, unusedEventHubNamespace); err != nil {
		return operation, 5 * time.Second, nil
	}

	operation.InputCreator.SetOverrides(componentName, getKnativeEventingOverrides())

	return p.operationManager.OperationSucceeded(operation, "azure event hub provisioned successfully.")
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
			Key:   "channel.default.apiVersion",
			Value: "knativekafka.kyma-project.io/v1alpha1",
		},
		{
			Key:   "channel.default.kind",
			Value: "KafkaChannel",
		},
	}
	return knativeOverrides
}
